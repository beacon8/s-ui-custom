package service

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/alireza0/s-ui/util/common"
)

// Outbound address statuses:
//   - ok: the outbound has a configured server address
//   - direct: the outbound is direct, so there is no proxy server address
//   - multi: selector/urltest groups have no single configured address
//   - none: the outbound type has no server field
//   - split: current route rules can send this user through more than one
//     outbound
//   - unbound: no outbound (or only reject) applies to this user
const (
	OutboundAddressStatusOK      = "ok"
	OutboundAddressStatusDirect  = "direct"
	OutboundAddressStatusMulti   = "multi"
	OutboundAddressStatusNone    = "none"
	OutboundAddressStatusSplit   = "split"
	OutboundAddressStatusUnbound = "unbound"
)

// OutboundAddress is the configured server address of the outbound a client is
// bound to. It is read from the current effective config; it is not a measured
// public egress address and never causes a network request.
type OutboundAddress struct {
	Status   string `json:"status"`
	Outbound string `json:"outbound,omitempty"`
	Address  string `json:"address,omitempty"`
	Port     uint16 `json:"port,omitempty"`
	Type     string `json:"type,omitempty"`
}

type OutboundAddressResponse struct {
	Users map[string]OutboundAddress `json:"users"`
}

// OutboundAddressService maps the visible clients to the server address of the
// outbound their route selects. It only reads the current config.
type OutboundAddressService struct{}

func (s *OutboundAddressService) GetOutboundAddresses(users []string) (*OutboundAddressResponse, error) {
	response := &OutboundAddressResponse{Users: make(map[string]OutboundAddress)}
	users = uniqueUsers(users)
	if len(users) == 0 {
		return response, nil
	}

	raw, err := (&ConfigService{}).GetConfig("")
	if err != nil {
		return nil, common.NewError("cannot read the effective config: ", err)
	}
	var parsed egressRouteConfig
	if err := json.Unmarshal(*raw, &parsed); err != nil {
		return nil, common.NewError("cannot parse the effective config: ", err)
	}
	response.Users = resolveOutboundAddresses(parsed, users)
	return response, nil
}

// resolveOutboundAddresses is the pure part of the lookup: it maps route rules
// to outbound tags and reads each outbound's configured server. It never dials.
func resolveOutboundAddresses(parsed egressRouteConfig, users []string) map[string]OutboundAddress {
	addresses := make(map[string]OutboundAddress, len(users))
	targets := resolveTargetsFromParsed(parsed, users)

	outbounds := make(map[string]egressOutbound, len(parsed.Outbounds))
	for _, outbound := range parsed.Outbounds {
		if outbound.Tag != "" {
			outbounds[outbound.Tag] = outbound
		}
	}

	for _, user := range users {
		target := targets[user]
		switch target.kind {
		case egressTargetSplit:
			addresses[user] = OutboundAddress{Status: OutboundAddressStatusSplit}
		case egressTargetUnbound:
			addresses[user] = OutboundAddress{Status: OutboundAddressStatusUnbound}
		case egressTargetOutbound:
			outbound, ok := outbounds[target.tag]
			if !ok {
				addresses[user] = OutboundAddress{Status: OutboundAddressStatusUnbound}
				continue
			}
			entry := OutboundAddress{Outbound: outbound.Tag, Type: outbound.Type}
			switch {
			case outbound.Type == "direct":
				entry.Status = OutboundAddressStatusDirect
			case outbound.Type == "selector" || outbound.Type == "urltest":
				entry.Status = OutboundAddressStatusMulti
			case outbound.Type == "block":
				entry = OutboundAddress{Status: OutboundAddressStatusUnbound}
			case strings.TrimSpace(outbound.Server) != "":
				entry.Status = OutboundAddressStatusOK
				entry.Address = strings.TrimSpace(outbound.Server)
				entry.Port = outbound.ServerPort
			default:
				entry.Status = OutboundAddressStatusNone
			}
			addresses[user] = entry
		default:
			addresses[user] = OutboundAddress{Status: OutboundAddressStatusUnbound}
		}
	}
	return addresses
}

func uniqueUsers(users []string) []string {
	seen := make(map[string]bool, len(users))
	result := make([]string, 0, len(users))
	for _, user := range users {
		if user == "" || seen[user] {
			continue
		}
		seen[user] = true
		result = append(result, user)
	}
	sort.Strings(result)
	return result
}

// ---- route mapping -------------------------------------------------------

type egressTargetKind int

const (
	egressTargetUnbound egressTargetKind = iota
	egressTargetOutbound
	egressTargetSplit
)

type egressTarget struct {
	kind egressTargetKind
	tag  string
}

type egressOutbound struct {
	Tag        string `json:"tag"`
	Type       string `json:"type"`
	Server     string `json:"server"`
	ServerPort uint16 `json:"server_port"`
}

type egressRouteConfig struct {
	Route struct {
		Rules []json.RawMessage `json:"rules"`
		Final string            `json:"final"`
	} `json:"route"`
	Outbounds []egressOutbound `json:"outbounds"`
}

func resolveEgressTargets(config json.RawMessage, users []string) (map[string]egressTarget, error) {
	var parsed egressRouteConfig
	if err := json.Unmarshal(config, &parsed); err != nil {
		return nil, common.NewError("cannot parse the effective config: ", err)
	}
	return resolveTargetsFromParsed(parsed, users), nil
}

func resolveTargetsFromParsed(parsed egressRouteConfig, users []string) map[string]egressTarget {
	outboundTags := make(map[string]bool, len(parsed.Outbounds))
	firstOutbound := ""
	for _, outbound := range parsed.Outbounds {
		if outbound.Tag == "" || outboundTags[outbound.Tag] {
			continue
		}
		outboundTags[outbound.Tag] = true
		if firstOutbound == "" {
			firstOutbound = outbound.Tag
		}
	}
	def := parsed.Route.Final
	if def == "" {
		// sing-box falls back to the first declared outbound when route.final is
		// absent. Repeating that rule here keeps the table honest for old configs.
		def = firstOutbound
	}
	if def != "" && !outboundTags[def] {
		def = ""
	}

	targets := make(map[string]egressTarget, len(users))
	for _, user := range users {
		targets[user] = resolveUserEgress(parsed.Route.Rules, def, outboundTags, user)
	}
	return targets
}

type egressDestinationKind int

const (
	egressDestinationNone egressDestinationKind = iota
	egressDestinationRoute
	egressDestinationReject
)

type egressDestination struct {
	kind egressDestinationKind
	tag  string
}

// resolveUserEgress walks the rules in order. It is deliberately conservative:
// a rule that might apply but has any condition other than auth_user makes the
// answer "split", because destination, protocol, inbound or an earlier rule can
// send the same user through another outbound.
func resolveUserEgress(rules []json.RawMessage, final string, outboundTags map[string]bool, user string) egressTarget {
	var outcomes []egressTarget
	seen := map[egressTarget]bool{}
	add := func(target egressTarget) {
		if seen[target] {
			return
		}
		seen[target] = true
		outcomes = append(outcomes, target)
	}

	decided := false
	for _, raw := range rules {
		verdict, destination := evaluateEgressRule(raw, outboundTags, user)
		if verdict == verdictNo {
			continue
		}
		switch destination.kind {
		case egressDestinationRoute:
			add(egressTarget{kind: egressTargetOutbound, tag: destination.tag})
		case egressDestinationReject:
			add(egressTarget{kind: egressTargetUnbound})
		case egressDestinationNone:
			// sniff/hijack-dns/resolve and route rules without a target do not
			// pick an egress; keep walking.
			continue
		}
		if verdict == verdictYes {
			// A rule that matches every packet of this user is final: later
			// rules never get a chance.
			decided = true
			break
		}
	}
	if !decided {
		// Traffic that does not match any of the conditional rules above falls
		// through to the default outbound, so it is another possible exit.
		if final != "" {
			add(egressTarget{kind: egressTargetOutbound, tag: final})
		} else {
			add(egressTarget{kind: egressTargetUnbound})
		}
	}
	if len(outcomes) > 1 {
		return egressTarget{kind: egressTargetSplit}
	}
	if len(outcomes) == 0 {
		return egressTarget{kind: egressTargetUnbound}
	}
	return outcomes[0]
}

type ruleVerdict int

const (
	verdictNo ruleVerdict = iota
	verdictMaybe
	verdictYes
)

func andVerdict(a, b ruleVerdict) ruleVerdict {
	if a == verdictNo || b == verdictNo {
		return verdictNo
	}
	if a == verdictYes && b == verdictYes {
		return verdictYes
	}
	return verdictMaybe
}

func orVerdict(a, b ruleVerdict) ruleVerdict {
	if a == verdictYes || b == verdictYes {
		return verdictYes
	}
	if a == verdictNo && b == verdictNo {
		return verdictNo
	}
	return verdictMaybe
}

func notVerdict(v ruleVerdict) ruleVerdict {
	switch v {
	case verdictYes:
		return verdictNo
	case verdictNo:
		return verdictYes
	default:
		return verdictMaybe
	}
}

// Keys that describe the action rather than a match condition. Anything else
// in a rule is treated as a condition that may not hold for this user.
var egressRuleActionKeys = map[string]bool{
	"action":                       true,
	"outbound":                     true,
	"override_address":             true,
	"override_port":                true,
	"network_strategy":             true,
	"fallback_delay":               true,
	"udp_disable_domain_unmapping": true,
	"udp_connect":                  true,
	"udp_timeout":                  true,
	"tls_fragment":                 true,
	"tls_fragment_fallback_delay":  true,
	"tls_record_fragment":          true,
	"tls_spoof":                    true,
	"tls_spoof_method":             true,
	"method":                       true,
	"sniffer":                      true,
	"timeout":                      true,
	"server":                       true,
	"strategy":                     true,
	"disable_cache":                true,
	"disable_optimistic_cache":     true,
	"rewrite_ttl":                  true,
	"client_subnet":                true,
}

func evaluateEgressRule(raw json.RawMessage, outboundTags map[string]bool, user string) (ruleVerdict, egressDestination) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return verdictMaybe, egressDestination{}
	}
	verdict := egressRuleVerdict(object, user)
	return verdict, egressRuleDestination(object, outboundTags)
}

func egressRuleVerdict(object map[string]json.RawMessage, user string) ruleVerdict {
	var ruleType string
	_ = json.Unmarshal(object["type"], &ruleType)

	verdict := verdictYes
	if ruleType == "logical" {
		verdict = egressLogicalVerdict(object, user)
	} else {
		for key, raw := range object {
			if key == "type" || key == "invert" || egressRuleActionKeys[key] {
				continue
			}
			if key == "auth_user" {
				verdict = andVerdict(verdict, egressAuthUserVerdict(raw, user))
				continue
			}
			// Any other match condition (domain, inbound, protocol, port,
			// rule_set, process, ...) can make this rule apply to some traffic
			// and not the rest.
			verdict = andVerdict(verdict, verdictMaybe)
		}
	}

	var invert bool
	if err := json.Unmarshal(object["invert"], &invert); err == nil && invert {
		verdict = notVerdict(verdict)
	}
	return verdict
}

func egressLogicalVerdict(object map[string]json.RawMessage, user string) ruleVerdict {
	var mode string
	_ = json.Unmarshal(object["mode"], &mode)
	var nested []json.RawMessage
	if err := json.Unmarshal(object["rules"], &nested); err != nil {
		return verdictMaybe
	}
	if len(nested) == 0 {
		return verdictMaybe
	}

	if mode == "or" {
		verdict := verdictNo
		for _, raw := range nested {
			var child map[string]json.RawMessage
			if err := json.Unmarshal(raw, &child); err != nil {
				return verdictMaybe
			}
			verdict = orVerdict(verdict, egressRuleVerdict(child, user))
		}
		return verdict
	}

	verdict := verdictYes
	for _, raw := range nested {
		var child map[string]json.RawMessage
		if err := json.Unmarshal(raw, &child); err != nil {
			return verdictMaybe
		}
		verdict = andVerdict(verdict, egressRuleVerdict(child, user))
	}
	return verdict
}

func egressAuthUserVerdict(raw json.RawMessage, user string) ruleVerdict {
	users, ok := jsonStringList(raw)
	if !ok {
		return verdictMaybe
	}
	if len(users) == 0 {
		return verdictYes
	}
	for _, name := range users {
		if name == user {
			return verdictYes
		}
	}
	return verdictNo
}

func egressRuleDestination(object map[string]json.RawMessage, outboundTags map[string]bool) egressDestination {
	var action string
	if raw, ok := object["action"]; ok {
		_ = json.Unmarshal(raw, &action)
	}
	switch strings.TrimSpace(action) {
	case "", "route":
		var outbound string
		if err := json.Unmarshal(object["outbound"], &outbound); err != nil || strings.TrimSpace(outbound) == "" {
			// A route action without an explicit outbound uses the default
			// outbound. The caller's final/default path handles it only when no
			// later rule exists, so treat it as no destination here.
			return egressDestination{}
		}
		outbound = strings.TrimSpace(outbound)
		if !outboundTags[outbound] {
			// The rule points at a tag that no longer exists. Reporting an
			// outbound would be a lie; treat it as "no usable binding".
			return egressDestination{kind: egressDestinationReject}
		}
		return egressDestination{kind: egressDestinationRoute, tag: outbound}
	case "reject":
		return egressDestination{kind: egressDestinationReject}
	default:
		return egressDestination{}
	}
}

func jsonStringList(raw json.RawMessage) ([]string, bool) {
	if len(raw) == 0 {
		return nil, false
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, true
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{single}, true
	}
	return nil, false
}
