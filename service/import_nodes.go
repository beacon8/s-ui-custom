package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alireza0/s-ui/core"
	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/util"
	"github.com/alireza0/s-ui/util/common"
	"github.com/alireza0/s-ui/util/rateconv"

	"gorm.io/gorm"
)

// ImportService turns one importClientNodes request into a batch of clients,
// each with its own inbound identity, outbound and route.
//
// Everything that can be rejected is collected before the first write, the
// writes share one transaction, and the core is only touched after that
// transaction commits. A failed preflight therefore cannot leave a half batch
// in the database or in the running core.
type ImportService struct{}

const (
	maxImportNodes = 1000
	minIDWidth     = 3
	maxIDWidth     = 12
	maxStartID     = 999999999
)

var importPrefixPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// ImportRequest is the JSON body of importClientNodes. startId, idWidth and
// clientDefaults are optional; nodes is required.
type ImportRequest struct {
	Prefix          string               `json:"prefix"`
	StartID         *int                 `json:"startId,omitempty"`
	IDWidth         *int                 `json:"idWidth,omitempty"`
	InboundIDs      []uint               `json:"inboundIds"`
	ClientDefaults  *ImportClientDefault `json:"clientDefaults,omitempty"`
	DefaultProtocol string               `json:"defaultProtocol,omitempty"`
	Nodes           []ImportNodeInput    `json:"nodes"`
	DryRun          bool                 `json:"dryRun,omitempty"`
}

// ImportNodeInput is one node. Exactly one representation is used: structured
// proxy fields, raw text, or a share link. Line is the original paste line
// number and is only echoed in errors and previews.
type ImportNodeInput struct {
	Protocol string          `json:"protocol,omitempty"`
	Host     string          `json:"host,omitempty"`
	Port     *int            `json:"port,omitempty"`
	Username string          `json:"username,omitempty"`
	Password string          `json:"password,omitempty"`
	TLS      json.RawMessage `json:"tls,omitempty"`
	Raw      *string         `json:"raw,omitempty"`
	URI      *string         `json:"uri,omitempty"`
	Line     *int            `json:"line,omitempty"`
}

// ImportClientDefault mirrors the shared fields of one client form. Volume is
// bytes; volumeGiB is the form's more convenient unit. Sending both is rejected
// because the same value must not have two sources.
type ImportClientDefault struct {
	Enable     *bool                     `json:"enable,omitempty"`
	Desc       string                    `json:"desc,omitempty"`
	Remark     string                    `json:"remark,omitempty"`
	Volume     *int64                    `json:"volume,omitempty"`
	VolumeGiB  *float64                  `json:"volumeGiB,omitempty"`
	Expiry     *int64                    `json:"expiry,omitempty"`
	DelayStart *bool                     `json:"delayStart,omitempty"`
	AutoReset  *bool                     `json:"autoReset,omitempty"`
	ResetDays  *int                      `json:"resetDays,omitempty"`
	MaxConns   *int                      `json:"maxConns,omitempty"`
	IPLimit    *int                      `json:"ipLimit,omitempty"`
	UpMbps     *int                      `json:"upMbps,omitempty"`
	DownMbps   *int                      `json:"downMbps,omitempty"`
	Config     map[string]map[string]any `json:"config,omitempty"`
	Links      []ImportLink              `json:"links,omitempty"`
}

// ImportLink is one non-local link every member of the batch receives.
type ImportLink struct {
	Type   string `json:"type"`
	Remark string `json:"remark,omitempty"`
	URI    string `json:"uri"`
}

// ImportIssue is one predictable problem. Index is zero based; Line is the
// original paste line. Conflict carries the clashing client name or tag.
type ImportIssue struct {
	Code     string `json:"code"`
	Index    *int   `json:"index,omitempty"`
	Line     *int   `json:"line,omitempty"`
	Field    string `json:"field,omitempty"`
	Conflict string `json:"conflict,omitempty"`
	Message  string `json:"message"`
}

// ImportItem is one created (or, in a dry run, predicted) client. Credentials
// are deliberately absent: previews, logs and audit rows may not carry them.
type ImportItem struct {
	Index               int    `json:"index"`
	Line                int    `json:"line,omitempty"`
	ClientID            uint   `json:"clientId"`
	ClientName          string `json:"clientName"`
	OutboundTag         string `json:"outboundTag"`
	Protocol            string `json:"protocol"`
	Address             string `json:"address"`
	SubscriptionID      string `json:"subscriptionId,omitempty"`
	SubscriptionLink    string `json:"subscriptionLink,omitempty"`
	SubscriptionRevoked bool   `json:"subscriptionRevoked,omitempty"`
}

// ImportGroupSubscription is the group identity created with the batch. The
// link is only filled on an authorised create/replay/reapply response.
type ImportGroupSubscription struct {
	SubscriptionID string `json:"subscriptionId,omitempty"`
	Name           string `json:"name,omitempty"`
	Link           string `json:"link,omitempty"`
	Revoked        bool   `json:"revoked,omitempty"`
}

// ImportCoreStatus says whether the committed batch reached the running core.
type ImportCoreStatus struct {
	Status string `json:"status"` // pending, applied, skipped, failed
	Reason string `json:"reason,omitempty"`
	Error  string `json:"error,omitempty"`
	At     int64  `json:"at,omitempty"`
}

// ImportResult is the obj of a successful import, dry run or replay.
type ImportResult struct {
	BatchID           string                  `json:"batchId"`
	BatchStatus       string                  `json:"batchStatus,omitempty"`
	Prefix            string                  `json:"prefix"`
	Group             string                  `json:"group"`
	Count             int                     `json:"count"`
	StartID           int                     `json:"startId"`
	EndID             int                     `json:"endId"`
	IDWidth           int                     `json:"idWidth"`
	InboundIDs        []uint                  `json:"inboundIds"`
	InboundTags       []string                `json:"inboundTags"`
	Items             []ImportItem            `json:"items"`
	GroupSubscription ImportGroupSubscription `json:"groupSubscription,omitempty"`
	Errors            []ImportIssue           `json:"errors,omitempty"`
	Core              ImportCoreStatus        `json:"core"`
	DryRun            bool                    `json:"dryRun"`
	Valid             bool                    `json:"valid"`
	Replayed          bool                    `json:"replayed,omitempty"`
	Retry             string                  `json:"retry,omitempty"`
}

// ImportError is the obj of a rejected import. Code is stable; the HTTP status
// is assigned by the API layer.
type ImportError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  []ImportIssue     `json:"errors,omitempty"`
	BatchID string            `json:"batchId,omitempty"`
	Core    *ImportCoreStatus `json:"core,omitempty"`
	Retry   string            `json:"retry,omitempty"`
}

func (e *ImportError) Error() string { return e.Message }

// ImportBatchView is what GET importBatch returns. Request is the original
// body without idempotency keys or credentials (the request never carries
// outbound credentials for the batch itself).
type ImportBatchView struct {
	BatchID   string           `json:"batchId"`
	Actor     string           `json:"actor"`
	Group     string           `json:"group"`
	Count     int              `json:"count"`
	Status    string           `json:"status"`
	Error     string           `json:"error,omitempty"`
	Core      ImportCoreStatus `json:"core"`
	CreatedAt int64            `json:"createdAt"`
	AppliedAt int64            `json:"appliedAt"`
	Request   json.RawMessage  `json:"request,omitempty"`
	Result    *ImportResult    `json:"result,omitempty"`
}

type importMeta struct {
	prefix   string
	startID  int
	idWidth  int
	group    string
	protocol string
	template clientTemplate
}

type clientTemplate struct {
	enable     bool
	desc       string
	remark     string
	volume     int64
	expiry     int64
	delayStart bool
	autoReset  bool
	resetDays  int
	maxConns   int
	ipLimit    int
	upMbps     int
	downMbps   int
	config     map[string]map[string]any
	links      []ImportLink
}

type parsedNode struct {
	index       int
	line        int
	protocol    string
	address     string
	outbound    map[string]any
	clientName  string
	outboundTag string
}

type preparedBatch struct {
	meta        importMeta
	nodes       []parsedNode
	clients     []*model.Client
	outbounds   []*model.Outbound
	items       []ImportItem
	issues      []ImportIssue
	baseConfig  json.RawMessage
	inboundIDs  []uint
	inboundTags []string
}

// applyImportCore is a seam for tests. Production leaves it nil and restarts
// the real core; tests can inject a one-shot failure and then let reapply pass.
var applyImportCore func() error

var errImportCoreSkipped = errors.New("core is not running")

// importValidationCore decodes configs without starting a box. It shares the
// same protocol registries as the running core and is built lazily: creating
// a Core registers protocols that log, and package init runs before the test
// logger is ready.
var (
	importValidationOnce sync.Once
	importValidationCore *core.Core
)

func validationCore() *core.Core {
	importValidationOnce.Do(func() {
		importValidationCore = core.NewCore()
	})
	return importValidationCore
}

func nowUnix() int64 { return time.Now().Unix() }

func intPtr(v int) *int { return &v }

func issue(code, field, message string, index, line int, conflict string) ImportIssue {
	is := ImportIssue{Code: code, Field: field, Message: message, Conflict: conflict}
	if index >= 0 {
		i := index
		is.Index = &i
	}
	if line > 0 {
		l := line
		is.Line = &l
	}
	return is
}

func issueCount(issues []ImportIssue) int { return len(issues) }

// Import runs a dry run immediately, or validates and commits a batch inside
// one transaction and then applies it to the core.
func (s *ImportService) Import(req *ImportRequest, actor, idemKey, hostname string) (*ImportResult, *ImportError) {
	if actor == "" {
		return nil, &ImportError{Code: "UNAUTHORIZED", Message: "a token or panel session is required"}
	}
	hash, err := importRequestHash(req)
	if err != nil {
		return nil, &ImportError{Code: "BAD_REQUEST", Message: err.Error()}
	}

	if req.DryRun {
		prepared := s.prepare(database.GetDB(), req)
		if len(prepared.issues) == 0 {
			config, err := prospectiveConfig(database.GetDB(), prepared.baseConfig, prepared.outbounds)
			if err != nil {
				prepared.issues = append(prepared.issues, issue("CONFIG_INVALID", "config", err.Error(), -1, 0, ""))
			} else if err := validationCore().ValidateConfig(config); err != nil {
				prepared.issues = append(prepared.issues, issue("CONFIG_INVALID", "config", err.Error(), -1, 0, ""))
			}
		}
		res := s.resultFor(prepared, true)
		return res, nil
	}

	db := database.GetDB()
	// Load the encryption key before taking the write lock. Reading the
	// settings row inside the migration transaction can wait on itself.
	subscriptionAEAD, aeadErr := subscriptionCipher()
	if aeadErr != nil {
		return nil, &ImportError{Code: "INTERNAL", Message: "unable to initialise subscription encryption: " + aeadErr.Error()}
	}
	tx := db.Begin()
	if tx.Error != nil {
		return nil, &ImportError{Code: "DB_BUSY", Message: tx.Error.Error()}
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if idemKey != "" {
		var existing model.ImportBatch
		err := tx.Where("actor = ? AND idem_key = ?", actor, idemKey).First(&existing).Error
		if err == nil {
			if existing.RequestHash != hash {
				return nil, &ImportError{Code: "IDEMPOTENCY_CONFLICT", Message: "this Idempotency-Key was already used with a different request body", BatchID: existing.BatchId}
			}
			var stored ImportResult
			if err := json.Unmarshal([]byte(existing.Result), &stored); err != nil {
				return nil, &ImportError{Code: "BATCH_CORRUPT", Message: "the stored result of this idempotency key cannot be read", BatchID: existing.BatchId}
			}
			stored.Replayed = true
			if existing.Status == "apply_failed" {
				core := stored.Core
				return nil, &ImportError{
					Code:    "CORE_APPLY_FAILED",
					Message: "the batch was committed but the core rejected it; retry with reapplyImportBatch instead of creating it again",
					BatchID: existing.BatchId,
					Core:    &core,
					Retry:   "POST apiv2/reapplyImportBatch {\"batchId\":\"" + existing.BatchId + "\"}",
				}
			}
			tx.Rollback()
			committed = true
			s.EnrichImportResult(db, &stored, hostname, true)
			return &stored, nil
		}
		if !database.IsNotFound(err) {
			return nil, &ImportError{Code: "DB_ERROR", Message: err.Error()}
		}
	}

	prepared := s.prepare(tx, req)
	if len(prepared.issues) > 0 {
		return nil, &ImportError{
			Code:    "VALIDATION_FAILED",
			Message: "the batch was not created; every listed issue must be fixed first",
			Errors:  prepared.issues,
		}
	}

	batchID, err := newBatchID()
	if err != nil {
		return nil, &ImportError{Code: "INTERNAL", Message: "unable to allocate a batch id"}
	}

	savedClients, err := s.writeBatch(tx, prepared, hostname)
	if err != nil {
		return nil, &ImportError{Code: "DB_ERROR", Message: err.Error()}
	}

	res := s.resultFor(prepared, false)
	res.BatchID = batchID
	res.BatchStatus = "committed"
	res.Core = ImportCoreStatus{Status: "pending"}
	for i := range res.Items {
		res.Items[i].ClientID = savedClients[i].Id
	}

	// The group and personal subscription identities are created in the same
	// transaction as the clients. Their IDs -- never their tokens -- go into
	// the stored result, so a replay or panel restart finds the same links.
	group, members, err := CreateImportSubscriptions(tx, subscriptionAEAD, batchID, res.Group, savedClients)
	if err != nil {
		return nil, &ImportError{Code: "DB_ERROR", Message: "unable to create subscription identities: " + err.Error()}
	}
	res.GroupSubscription.SubscriptionID = group.SubID
	for i := range res.Items {
		member := members[res.Items[i].ClientID]
		if member != nil {
			res.Items[i].SubscriptionID = member.SubID
		}
	}

	// Stored before commit, together with the data, so a retry after a lost
	// response can find the exact batch that was created.
	storedKey := idemKey
	if storedKey == "" {
		storedKey = "_" + batchID
	}
	resultJSON, err := json.Marshal(res)
	if err != nil {
		return nil, &ImportError{Code: "INTERNAL", Message: err.Error()}
	}
	requestJSON, err := redactedImportRequest(req, prepared)
	if err != nil {
		return nil, &ImportError{Code: "INTERNAL", Message: err.Error()}
	}
	batch := &model.ImportBatch{
		BatchId:     batchID,
		Actor:       actor,
		IdemKey:     storedKey,
		RequestHash: hash,
		Group:       res.Group,
		Count:       res.Count,
		Status:      "committed",
		Request:     string(requestJSON),
		Result:      string(resultJSON),
		CreatedAt:   nowUnix(),
	}
	if err := tx.Create(batch).Error; err != nil {
		return nil, &ImportError{Code: "IDEMPOTENCY_CONFLICT", Message: "another request committed this batch first"}
	}

	summary := map[string]any{
		"batchId":   batchID,
		"group":     res.Group,
		"count":     res.Count,
		"clients":   itemNames(res.Items),
		"outbounds": itemTags(res.Items),
	}
	summaryJSON, _ := json.Marshal(summary)
	if err := tx.Create(&model.Changes{
		DateTime: nowUnix(),
		Actor:    actor,
		Key:      "clients",
		Action:   "import",
		Obj:      summaryJSON,
	}).Error; err != nil {
		return nil, &ImportError{Code: "DB_ERROR", Message: err.Error()}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, &ImportError{Code: "DB_ERROR", Message: err.Error()}
	}
	committed = true
	LastUpdate = nowUnix()

	// The data is durable from here on. A core failure is reported as a
	// recoverable batch, never rolled back and never recreated. Store the
	// link-free result first, then enrich only the caller's response.
	applyErr := s.applyAndRecord(batch, res)
	if applyErr != nil {
		return nil, applyErr
	}
	s.EnrichImportResult(db, res, hostname, true)
	return res, nil
}

// Reapply retries the core application of a batch that already committed. It
// never recreates clients, outbounds or route rules.
func (s *ImportService) Reapply(batchID, actor, hostname string) (*ImportResult, *ImportError) {
	if actor == "" {
		return nil, &ImportError{Code: "UNAUTHORIZED", Message: "a token or panel session is required"}
	}
	db := database.GetDB()
	var batch model.ImportBatch
	if err := db.Where("batch_id = ?", batchID).First(&batch).Error; err != nil {
		if database.IsNotFound(err) {
			return nil, &ImportError{Code: "BATCH_NOT_FOUND", Message: "no such import batch", BatchID: batchID}
		}
		return nil, &ImportError{Code: "DB_ERROR", Message: err.Error()}
	}
	if batch.Actor != actor {
		return nil, &ImportError{Code: "BATCH_FORBIDDEN", Message: "this batch belongs to another administrator", BatchID: batchID}
	}
	var res ImportResult
	if err := json.Unmarshal([]byte(batch.Result), &res); err != nil {
		return nil, &ImportError{Code: "BATCH_CORRUPT", Message: "the stored result of this batch cannot be read", BatchID: batchID}
	}
	if err := s.verifyBatchData(&res); err != nil {
		return nil, &ImportError{Code: "BATCH_DATA_MISSING", Message: err.Error(), BatchID: batchID}
	}
	if batch.Status == "applied" {
		core := ImportCoreStatus{Status: "applied", At: batch.AppliedAt}
		if core.At == 0 {
			core.At = nowUnix()
		}
		res.Core = core
		res.Replayed = true
		s.EnrichImportResult(db, &res, hostname, true)
		return &res, nil
	}
	if err := s.applyAndRecord(&batch, &res); err != nil {
		return nil, err
	}
	res.Replayed = true
	s.EnrichImportResult(db, &res, hostname, true)
	return &res, nil
}

// GetBatch lets a caller locate a committed batch before deciding to reapply.
func (s *ImportService) GetBatch(batchID, actor, hostname string) (*ImportBatchView, *ImportError) {
	if actor == "" {
		return nil, &ImportError{Code: "UNAUTHORIZED", Message: "a token or panel session is required"}
	}
	db := database.GetDB()
	var batch model.ImportBatch
	if err := db.Where("batch_id = ?", batchID).First(&batch).Error; err != nil {
		if database.IsNotFound(err) {
			return nil, &ImportError{Code: "BATCH_NOT_FOUND", Message: "no such import batch", BatchID: batchID}
		}
		return nil, &ImportError{Code: "DB_ERROR", Message: err.Error()}
	}
	if batch.Actor != actor {
		return nil, &ImportError{Code: "BATCH_FORBIDDEN", Message: "this batch belongs to another administrator", BatchID: batchID}
	}
	view := &ImportBatchView{
		BatchID:   batch.BatchId,
		Actor:     batch.Actor,
		Group:     batch.Group,
		Count:     batch.Count,
		Status:    batch.Status,
		Error:     batch.Error,
		CreatedAt: batch.CreatedAt,
		AppliedAt: batch.AppliedAt,
		Request:   json.RawMessage(batch.Request),
	}
	var res ImportResult
	if err := json.Unmarshal([]byte(batch.Result), &res); err == nil {
		view.Core = res.Core
		// Historical batch views carry subscription ids, never tokens/links.
		s.EnrichImportResult(db, &res, hostname, false)
		view.Result = &res
	}
	return view, nil
}

// redactedImportRequest is what is kept on the batch row and later returned by
// GET importBatch. It records the shape of the request (lines, protocols,
// addresses, shared flags) without node credentials, upstream links or custom
// subscription URLs, any of which may carry a secret.
func redactedImportRequest(req *ImportRequest, prepared *preparedBatch) (json.RawMessage, error) {
	type nodeView struct {
		Index    int    `json:"index"`
		Line     int    `json:"line"`
		Protocol string `json:"protocol"`
		Address  string `json:"address"`
	}
	nodes := make([]nodeView, 0, len(prepared.nodes))
	for _, node := range prepared.nodes {
		nodes = append(nodes, nodeView{Index: node.index, Line: node.line, Protocol: node.protocol, Address: node.address})
	}
	summary := map[string]any{
		"prefix":          req.Prefix,
		"startId":         prepared.meta.startID,
		"idWidth":         prepared.meta.idWidth,
		"inboundIds":      prepared.inboundIDs,
		"defaultProtocol": prepared.meta.protocol,
		"dryRun":          req.DryRun,
		"count":           len(prepared.items),
		"group":           prepared.meta.group,
		"nodes":           nodes,
	}
	return json.MarshalIndent(summary, "", "  ")
}

func importRequestHash(req *ImportRequest) (string, error) {
	canonical, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

func newBatchID() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "imp_" + hex.EncodeToString(buf), nil
}

func itemNames(items []ImportItem) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.ClientName
	}
	return names
}

func itemTags(items []ImportItem) []string {
	tags := make([]string, len(items))
	for i, item := range items {
		tags[i] = item.OutboundTag
	}
	return tags
}

func (s *ImportService) resultFor(prepared *preparedBatch, dryRun bool) *ImportResult {
	res := &ImportResult{
		Prefix:      prepared.meta.prefix,
		Group:       prepared.meta.group,
		Count:       len(prepared.items),
		StartID:     prepared.meta.startID,
		EndID:       prepared.meta.startID + len(prepared.items) - 1,
		IDWidth:     prepared.meta.idWidth,
		InboundIDs:  prepared.inboundIDs,
		InboundTags: prepared.inboundTags,
		Items:       prepared.items,
		Errors:      prepared.issues,
		DryRun:      dryRun,
		Valid:       len(prepared.issues) == 0,
	}
	if res.Count == 0 {
		res.EndID = prepared.meta.startID - 1
	}
	if res.Items == nil {
		res.Items = []ImportItem{}
	}
	if res.Errors == nil {
		res.Errors = []ImportIssue{}
	}
	if dryRun {
		res.Core = ImportCoreStatus{Status: "not_applied"}
	}
	return res
}

// prepare does every check and builds the exact rows that would be written.
// It never writes; the caller supplies either the read database (dry run) or
// the transaction that will commit the batch.
func (s *ImportService) prepare(db *gorm.DB, req *ImportRequest) *preparedBatch {
	prepared := &preparedBatch{}
	issues := []ImportIssue{}

	if db == nil {
		db = database.GetDB()
	}

	prefix := strings.TrimSpace(req.Prefix)
	if prefix == "" {
		issues = append(issues, issue("PREFIX_REQUIRED", "prefix", "prefix must not be empty", -1, 0, ""))
	} else if len(prefix) > 32 {
		issues = append(issues, issue("PREFIX_INVALID", "prefix", "prefix must be at most 32 characters", -1, 0, ""))
	} else if !importPrefixPattern.MatchString(prefix) {
		issues = append(issues, issue("PREFIX_INVALID", "prefix", "prefix may only contain letters, digits, underscore, dot and hyphen", -1, 0, ""))
	}

	startID := 1
	if req.StartID != nil {
		startID = *req.StartID
	}
	if startID < 0 || startID > maxStartID {
		issues = append(issues, issue("START_ID_INVALID", "startId", fmt.Sprintf("startId must be between 0 and %d", maxStartID), -1, 0, ""))
	}
	idWidth := minIDWidth
	if req.IDWidth != nil {
		idWidth = *req.IDWidth
	}
	if idWidth < minIDWidth {
		idWidth = minIDWidth
	}
	if idWidth > maxIDWidth {
		issues = append(issues, issue("ID_WIDTH_INVALID", "idWidth", fmt.Sprintf("idWidth must be at most %d", maxIDWidth), -1, 0, ""))
		idWidth = maxIDWidth
	}

	protocol := strings.ToLower(strings.TrimSpace(req.DefaultProtocol))
	if protocol == "" {
		protocol = "socks5"
	}
	switch protocol {
	case "socks", "socks5":
		protocol = "socks5"
	case "http":
		protocol = "http"
	default:
		issues = append(issues, issue("DEFAULT_PROTOCOL_INVALID", "defaultProtocol", "defaultProtocol must be socks5 or http", -1, 0, ""))
		protocol = "socks5"
	}

	if len(req.Nodes) == 0 {
		issues = append(issues, issue("NODES_REQUIRED", "nodes", "nodes must contain at least one entry", -1, 0, ""))
	}
	if len(req.Nodes) > maxImportNodes {
		issues = append(issues, issue("NODES_TOO_MANY", "nodes", fmt.Sprintf("a batch may contain at most %d nodes", maxImportNodes), -1, 0, ""))
	}

	inboundIDs, inboundTags, inboundIssues := validateImportInbounds(db, req.InboundIDs)
	issues = append(issues, inboundIssues...)

	template, templateIssues := resolveClientTemplate(req.ClientDefaults)
	issues = append(issues, templateIssues...)

	// Parse first, then number only the entries that parsed. The contract is
	// endId = startId + valid nodes - 1, so a dry run can show the numbering a
	// corrected request would receive.
	parsed := make([]parsedNode, 0, len(req.Nodes))
	for i := range req.Nodes {
		node, nodeIssues := parseImportNode(&req.Nodes[i], i, protocol)
		if len(nodeIssues) > 0 {
			issues = append(issues, nodeIssues...)
			continue
		}
		// Schema-validate before assigning a tag. The tag is not part of the
		// outbound's option schema, so this catches bad TLS/transport fields.
		probe := cloneMap(node.outbound)
		probe["tag"] = "preview"
		encoded, err := json.Marshal(probe)
		if err != nil {
			issues = append(issues, issue("NODE_INVALID", "nodes", "node cannot be encoded", i, node.line, ""))
			continue
		}
		if err := validationCore().ValidateOutbound(encoded); err != nil {
			issues = append(issues, issue("NODE_INVALID", "nodes", "invalid "+node.protocol+" outbound: "+err.Error(), i, node.line, ""))
			continue
		}
		parsed = append(parsed, node)
	}

	names := make([]string, len(parsed))
	tags := make([]string, len(parsed))
	for i := range parsed {
		number := startID + i
		parsed[i].clientName = fmt.Sprintf("%s-%0*d", prefix, idWidth, number)
		parsed[i].outboundTag = "out-" + parsed[i].clientName
		parsed[i].outbound["tag"] = parsed[i].outboundTag
		names[i] = parsed[i].clientName
		tags[i] = parsed[i].outboundTag
		if len(parsed[i].clientName) > 64 {
			issues = append(issues, issue("CLIENT_NAME_TOO_LONG", "prefix", "generated client name exceeds 64 characters", parsed[i].index, parsed[i].line, parsed[i].clientName))
		}
	}

	if len(parsed) > 0 {
		seenName := map[string]bool{}
		for _, name := range names {
			if seenName[name] {
				issues = append(issues, issue("DUPLICATE_CLIENT_NAME", "nodes", "two nodes produce the same client name", -1, 0, name))
			}
			seenName[name] = true
		}
		existingNames, err := existingValues(db, "clients", "name", names)
		if err != nil {
			issues = append(issues, issue("DB_ERROR", "nodes", err.Error(), -1, 0, ""))
		}
		existingTags, err := existingValues(db, "outbounds", "tag", tags)
		if err != nil {
			issues = append(issues, issue("DB_ERROR", "nodes", err.Error(), -1, 0, ""))
		}
		for i := range parsed {
			if existingNames[names[i]] {
				issues = append(issues, issue("CLIENT_NAME_CONFLICT", "nodes", "client name is already in use", parsed[i].index, parsed[i].line, names[i]))
			}
			if existingTags[tags[i]] {
				issues = append(issues, issue("OUTBOUND_TAG_CONFLICT", "nodes", "outbound tag is already in use", parsed[i].index, parsed[i].line, tags[i]))
			}
		}
	}

	prepared.meta = importMeta{
		prefix:   prefix,
		startID:  startID,
		idWidth:  idWidth,
		protocol: protocol,
		template: template,
	}
	if len(parsed) > 0 {
		prepared.meta.group = fmt.Sprintf("%s-%d-%d", prefix, startID, startID+len(parsed)-1)
	} else {
		prepared.meta.group = ""
	}
	prepared.nodes = parsed
	prepared.issues = issues
	prepared.inboundIDs = inboundIDs
	prepared.inboundTags = inboundTags
	prepared.items = make([]ImportItem, 0, len(parsed))

	// Build rows only when there is something to build. Items and clients are
	// 1:1 in the same order.
	inboundsJSON, err := json.Marshal(inboundIDs)
	if err != nil {
		prepared.issues = append(prepared.issues, issue("INTERNAL", "inboundIds", err.Error(), -1, 0, ""))
		return prepared
	}
	for i := range parsed {
		client, err := buildImportClient(parsed[i], prepared, template, inboundsJSON)
		if err != nil {
			prepared.issues = append(prepared.issues, issue("CLIENT_INVALID", "clientDefaults", err.Error(), parsed[i].index, parsed[i].line, parsed[i].clientName))
			continue
		}
		prepared.clients = append(prepared.clients, client)
		prepared.items = append(prepared.items, ImportItem{
			Index:       parsed[i].index,
			Line:        parsed[i].line,
			ClientName:  parsed[i].clientName,
			OutboundTag: parsed[i].outboundTag,
			Protocol:    parsed[i].protocol,
			Address:     parsed[i].address,
		})
		outbound, err := importOutboundModel(parsed[i].outbound)
		if err != nil {
			prepared.issues = append(prepared.issues, issue("NODE_INVALID", "nodes", err.Error(), parsed[i].index, parsed[i].line, parsed[i].outboundTag))
			continue
		}
		prepared.outbounds = append(prepared.outbounds, outbound)
	}

	// Mismatches mean some rows failed after numbering; never commit a batch
	// whose items and clients disagree.
	if len(prepared.clients) != len(parsed) || len(prepared.outbounds) != len(parsed) {
		if len(prepared.issues) == 0 {
			prepared.issues = append(prepared.issues, issue("INTERNAL", "nodes", "unable to build every member of the batch", -1, 0, ""))
		}
	}

	if len(prepared.issues) == 0 {
		baseConfig, err := addImportRouteRules(db, tags, names)
		if err != nil {
			prepared.issues = append(prepared.issues, issue("CONFIG_INVALID", "config", err.Error(), -1, 0, ""))
		} else {
			prepared.baseConfig = baseConfig
		}
	}
	return prepared
}

func validateImportInbounds(db *gorm.DB, ids []uint) ([]uint, []string, []ImportIssue) {
	issues := []ImportIssue{}
	if len(ids) == 0 {
		issues = append(issues, issue("INBOUNDS_REQUIRED", "inboundIds", "select at least one inbound", -1, 0, ""))
		return nil, nil, issues
	}
	unique := make([]uint, 0, len(ids))
	seen := map[uint]bool{}
	for _, id := range ids {
		if id == 0 {
			issues = append(issues, issue("INBOUND_NOT_FOUND", "inboundIds", "inbound id 0 does not exist", -1, 0, "0"))
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	sort.Slice(unique, func(i, j int) bool { return unique[i] < unique[j] })

	var rows []model.Inbound
	if err := db.Model(model.Inbound{}).Where("id IN ?", unique).Find(&rows).Error; err != nil {
		issues = append(issues, issue("DB_ERROR", "inboundIds", err.Error(), -1, 0, ""))
		return nil, nil, issues
	}
	byID := make(map[uint]*model.Inbound, len(rows))
	for i := range rows {
		byID[rows[i].Id] = &rows[i]
	}
	tags := make([]string, 0, len(unique))
	for _, id := range unique {
		inbound, ok := byID[id]
		if !ok {
			issues = append(issues, issue("INBOUND_NOT_FOUND", "inboundIds", "the selected inbound does not exist", -1, 0, strconv.FormatUint(uint64(id), 10)))
			continue
		}
		if strings.TrimSpace(inbound.Tag) == "" {
			issues = append(issues, issue("INBOUND_NO_TAG", "inboundIds", "the inbound has no tag and cannot be linked", -1, 0, strconv.FormatUint(uint64(id), 10)))
			continue
		}
		if !inboundSupportsIdentity(inbound) {
			issues = append(issues, issue("INBOUND_NO_USER", "inboundIds", "the inbound cannot match a client identity, so a per-client route would never hit", -1, 0, inbound.Tag))
			continue
		}
		tags = append(tags, inbound.Tag)
	}
	return unique, tags, issues
}

// inboundSupportsIdentity mirrors the rule the inbound config generator uses
// when it exposes clients as users: only types with users, and not the two
// cases where the inbound itself owns the credentials.
func inboundSupportsIdentity(inbound *model.Inbound) bool {
	if !(&InboundService{}).hasUser(inbound.Type) {
		return false
	}
	switch inbound.Type {
	case "shadowtls":
		var options struct {
			Version int `json:"version"`
		}
		_ = json.Unmarshal(inbound.Options, &options)
		return options.Version >= 3
	case "shadowsocks":
		var options struct {
			Managed bool `json:"managed"`
		}
		_ = json.Unmarshal(inbound.Options, &options)
		return !options.Managed
	}
	return true
}

func resolveClientTemplate(def *ImportClientDefault) (clientTemplate, []ImportIssue) {
	template := clientTemplate{enable: true, resetDays: 0}
	issues := []ImportIssue{}
	if def == nil {
		return template, issues
	}
	if def.Enable != nil {
		template.enable = *def.Enable
	}
	template.desc = def.Desc
	template.remark = def.Remark
	if def.Volume != nil && def.VolumeGiB != nil {
		issues = append(issues, issue("DEFAULTS_VOLUME_AMBIGUOUS", "clientDefaults", "send either volume (bytes) or volumeGiB, not both", -1, 0, ""))
	}
	if def.Volume != nil {
		if *def.Volume < 0 {
			issues = append(issues, issue("DEFAULTS_VOLUME_INVALID", "clientDefaults", "volume must not be negative", -1, 0, ""))
		} else {
			template.volume = *def.Volume
		}
	}
	if def.VolumeGiB != nil {
		if *def.VolumeGiB < 0 {
			issues = append(issues, issue("DEFAULTS_VOLUME_INVALID", "clientDefaults", "volumeGiB must not be negative", -1, 0, ""))
		} else if *def.VolumeGiB > float64(1<<53) {
			issues = append(issues, issue("DEFAULTS_VOLUME_INVALID", "clientDefaults", "volumeGiB is too large", -1, 0, ""))
		} else {
			template.volume = int64(*def.VolumeGiB * (1 << 30))
		}
	}
	if def.Expiry != nil {
		if *def.Expiry < 0 {
			issues = append(issues, issue("DEFAULTS_EXPIRY_INVALID", "clientDefaults", "expiry must not be negative", -1, 0, ""))
		} else {
			template.expiry = *def.Expiry
		}
	}
	if def.DelayStart != nil {
		template.delayStart = *def.DelayStart
	}
	if def.AutoReset != nil {
		template.autoReset = *def.AutoReset
	}
	if def.ResetDays != nil {
		if *def.ResetDays < 0 {
			issues = append(issues, issue("DEFAULTS_RESET_DAYS_INVALID", "clientDefaults", "resetDays must not be negative", -1, 0, ""))
		} else {
			template.resetDays = *def.ResetDays
		}
	}
	if def.MaxConns != nil {
		if *def.MaxConns < 0 {
			issues = append(issues, issue("DEFAULTS_MAX_CONNS_INVALID", "clientDefaults", "maxConns must not be negative", -1, 0, ""))
		} else {
			template.maxConns = *def.MaxConns
		}
	}
	if def.IPLimit != nil {
		if *def.IPLimit < 0 {
			issues = append(issues, issue("DEFAULTS_IP_LIMIT_INVALID", "clientDefaults", "ipLimit must not be negative", -1, 0, ""))
		} else {
			template.ipLimit = *def.IPLimit
		}
	}
	if def.UpMbps != nil {
		if *def.UpMbps < 0 || int64(*def.UpMbps) > rateconv.MaxMbps {
			issues = append(issues, issue("DEFAULTS_RATE_INVALID", "clientDefaults", fmt.Sprintf("upMbps must be between 0 and %d", rateconv.MaxMbps), -1, 0, ""))
		} else {
			template.upMbps = *def.UpMbps
		}
	}
	if def.DownMbps != nil {
		if *def.DownMbps < 0 || int64(*def.DownMbps) > rateconv.MaxMbps {
			issues = append(issues, issue("DEFAULTS_RATE_INVALID", "clientDefaults", fmt.Sprintf("downMbps must be between 0 and %d", rateconv.MaxMbps), -1, 0, ""))
		} else {
			template.downMbps = *def.DownMbps
		}
	}
	if template.delayStart && !template.autoReset {
		// The single client form writes no expiry when a delayed client has no
		// periodic reset; the same rule applies to every member of the batch.
		template.expiry = 0
	}
	if template.delayStart || template.autoReset {
		if template.resetDays <= 0 {
			template.resetDays = 1
		}
	}
	for protocol, fields := range def.Config {
		if _, known := generatedProtocolNames()[protocol]; !known {
			issues = append(issues, issue("DEFAULTS_CONFIG_UNKNOWN", "clientDefaults.config", "unknown protocol key: "+protocol, -1, 0, protocol))
			continue
		}
		for _, identity := range []string{"name", "username", "password", "uuid", "auth_str", "userkey"} {
			if _, exists := fields[identity]; exists {
				issues = append(issues, issue("DEFAULTS_CONFIG_IDENTITY", "clientDefaults.config", "shared config must not set the per-client identity field "+identity, -1, 0, protocol+"."+identity))
			}
		}
	}
	if template.config == nil {
		template.config = map[string]map[string]any{}
	}
	for protocol, fields := range def.Config {
		copied := make(map[string]any, len(fields))
		for k, v := range fields {
			copied[k] = v
		}
		template.config[protocol] = copied
	}
	for _, link := range def.Links {
		if link.Type != "external" && link.Type != "sub" {
			issues = append(issues, issue("DEFAULTS_LINK_INVALID", "clientDefaults.links", "link type must be external or sub", -1, 0, link.Type))
			continue
		}
		if strings.TrimSpace(link.URI) == "" {
			issues = append(issues, issue("DEFAULTS_LINK_INVALID", "clientDefaults.links", "link uri must not be empty", -1, 0, link.Remark))
			continue
		}
		template.links = append(template.links, link)
	}
	return template, issues
}

func generatedProtocolNames() map[string]struct{} {
	return map[string]struct{}{
		"mixed": {}, "socks": {}, "http": {}, "shadowsocks": {}, "shadowsocks16": {},
		"shadowtls": {}, "vmess": {}, "vless": {}, "anytls": {}, "trojan": {}, "naive": {},
		"hysteria": {}, "snell": {}, "tuic": {}, "hysteria2": {},
	}
}

func parseImportNode(input *ImportNodeInput, index int, defaultProtocol string) (parsedNode, []ImportIssue) {
	line := index + 1
	if input.Line != nil && *input.Line > 0 {
		line = *input.Line
	}
	node := parsedNode{index: index, line: line}
	issues := []ImportIssue{}

	hasRaw := input.Raw != nil
	hasURI := input.URI != nil
	hasStructured := input.Protocol != "" || input.Host != "" || input.Port != nil ||
		input.Username != "" || input.Password != "" || len(input.TLS) > 0
	representations := 0
	for _, present := range []bool{hasRaw, hasURI, hasStructured} {
		if present {
			representations++
		}
	}
	if representations == 0 {
		return node, []ImportIssue{issue("NODE_EMPTY", "nodes", "node must use structured fields, raw or uri", index, line, "")}
	}
	if representations > 1 {
		return node, []ImportIssue{issue("NODE_REPRESENTATION", "nodes", "node must use exactly one representation: structured fields, raw or uri", index, line, "")}
	}

	var outbound map[string]any
	var protocol, address string
	var err error
	switch {
	case hasRaw:
		outbound, protocol, address, err = parseImportText(strings.TrimSpace(*input.Raw), defaultProtocol)
	case hasURI:
		outbound, protocol, address, err = parseImportText(strings.TrimSpace(*input.URI), defaultProtocol)
	default:
		outbound, protocol, address, err = parseStructuredNode(input)
	}
	if err != nil {
		code := "NODE_FORMAT"
		if strings.Contains(err.Error(), "unsupported") {
			code = "NODE_URI_UNSUPPORTED"
		}
		return node, []ImportIssue{issue(code, "nodes", err.Error(), index, line, "")}
	}
	node.outbound = outbound
	node.protocol = protocol
	node.address = address
	return node, issues
}

func parseImportText(text, defaultProtocol string) (map[string]any, string, string, error) {
	if text == "" {
		return nil, "", "", common.NewError("empty node line")
	}
	if strings.Contains(text, "://") {
		return parseImportURI(text)
	}
	return parseRawProxyLine(text, defaultProtocol)
}

func parseRawProxyLine(line, defaultProtocol string) (map[string]any, string, string, error) {
	var host, user, pass, portText string
	if strings.HasPrefix(line, "[") {
		end := strings.Index(line, "]:")
		if end < 0 {
			return nil, "", "", common.NewError("invalid IPv6 node: expected [address]:port:username:password")
		}
		host = line[1:end]
		rest := line[end+2:]
		parts := strings.SplitN(rest, ":", 3)
		if len(parts) != 3 {
			return nil, "", "", common.NewError("invalid node: expected [address]:port:username:password")
		}
		portText, user, pass = parts[0], parts[1], parts[2]
	} else {
		parts := strings.SplitN(line, ":", 4)
		if len(parts) != 4 {
			return nil, "", "", common.NewError("invalid node: expected address:port:username:password")
		}
		host, portText, user, pass = parts[0], parts[1], parts[2], parts[3]
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, "", "", common.NewError("node address must not be empty")
	}
	port, err := strconv.Atoi(strings.TrimSpace(portText))
	if err != nil || port < 1 || port > 65535 {
		return nil, "", "", common.NewErrorf("invalid node port %q", strings.TrimSpace(portText))
	}
	outbound, protocol, err := buildProxyOutbound(defaultProtocol, host, port, user, pass, nil)
	if err != nil {
		return nil, "", "", err
	}
	return outbound, protocol, net.JoinHostPort(host, strconv.Itoa(port)), nil
}

// importURISchemes is the set the API documents. Other upstream parsers exist,
// but accepting a scheme this contract does not mention would make the preview
// promise a protocol the caller did not ask for.
var importURISchemes = map[string]bool{
	"socks5": true, "socks": true, "http": true, "https": true, "vless": true, "vmess": true,
}

func parseImportURI(link string) (map[string]any, string, string, error) {
	parsed, err := url.Parse(link)
	if err != nil {
		return nil, "", "", common.NewError("invalid share link: ", err)
	}
	scheme := strings.ToLower(parsed.Scheme)
	if !importURISchemes[scheme] {
		return nil, "", "", common.NewErrorf("unsupported share link scheme %q; supported: socks5, http, https, vless, vmess", parsed.Scheme)
	}
	outbound, _, err := util.GetOutbound(link, 0)
	if err != nil || outbound == nil {
		if err == nil {
			err = common.NewError("parser returned no outbound")
		}
		return nil, "", "", common.NewError("cannot parse share link: ", err)
	}
	out := *outbound
	server, _ := out["server"].(string)
	port, ok := toInt(out["server_port"])
	if !ok || server == "" || port < 1 || port > 65535 {
		return nil, "", "", common.NewError("share link has no usable server and port")
	}
	out["server"] = strings.Trim(server, "[]")
	// The VMess decoder keeps the JSON port as whatever type the share link
	// carried (often a string); sing-box expects a number.
	out["server_port"] = port
	protocol := scheme
	kind, _ := out["type"].(string)
	switch kind {
	case "socks":
		protocol = "socks5"
	case "http":
		if tlsEnabled(out["tls"]) {
			protocol = "https"
		} else {
			protocol = "http"
		}
	case "vless", "vmess":
		protocol = kind
	}
	return out, protocol, net.JoinHostPort(strings.Trim(server, "[]"), strconv.Itoa(port)), nil
}

func parseStructuredNode(input *ImportNodeInput) (map[string]any, string, string, error) {
	protocol := strings.ToLower(strings.TrimSpace(input.Protocol))
	switch protocol {
	case "socks", "socks5", "http", "https":
	default:
		return nil, "", "", common.NewErrorf("unsupported structured protocol %q; use socks5, http, https or a share link for vless/vmess", input.Protocol)
	}
	host := strings.TrimSpace(strings.Trim(input.Host, "[]"))
	if host == "" {
		return nil, "", "", common.NewError("structured node needs host")
	}
	if input.Port == nil {
		return nil, "", "", common.NewError("structured node needs port")
	}
	port := *input.Port
	if port < 1 || port > 65535 {
		return nil, "", "", common.NewErrorf("invalid node port %d", port)
	}
	outbound, label, err := buildProxyOutbound(protocol, host, port, input.Username, input.Password, input.TLS)
	if err != nil {
		return nil, "", "", err
	}
	return outbound, label, net.JoinHostPort(host, strconv.Itoa(port)), nil
}

func buildProxyOutbound(protocol, host string, port int, username, password string, tlsRaw json.RawMessage) (map[string]any, string, error) {
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return nil, "", common.NewError("node address must not be empty")
	}
	if port < 1 || port > 65535 {
		return nil, "", common.NewErrorf("invalid node port %d", port)
	}
	outbound := map[string]any{
		"server":      host,
		"server_port": port,
	}
	switch protocol {
	case "socks", "socks5":
		outbound["type"] = "socks"
		outbound["version"] = "5"
	case "http", "https":
		outbound["type"] = "http"
	default:
		return nil, "", common.NewErrorf("unsupported protocol %q", protocol)
	}
	if username != "" || password != "" {
		outbound["username"] = username
		outbound["password"] = password
	}
	if len(tlsRaw) > 0 && string(tlsRaw) != "null" {
		var tlsOptions map[string]any
		if err := json.Unmarshal(tlsRaw, &tlsOptions); err != nil {
			return nil, "", common.NewError("tls must be a JSON object: ", err)
		}
		outbound["tls"] = tlsOptions
	}
	if protocol == "https" {
		tlsOptions, _ := outbound["tls"].(map[string]any)
		if tlsOptions == nil {
			tlsOptions = map[string]any{}
			outbound["tls"] = tlsOptions
		}
		tlsOptions["enabled"] = true
	}
	label := protocol
	if outbound["type"] == "socks" {
		label = "socks5"
	}
	if outbound["type"] == "http" && tlsEnabled(outbound["tls"]) {
		label = "https"
	}
	return outbound, label, nil
}

func tlsEnabled(raw any) bool {
	tlsOptions, ok := raw.(map[string]any)
	if !ok {
		return false
	}
	enabled, _ := tlsOptions["enabled"].(bool)
	return enabled
}

func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		if v != float64(int(v)) {
			return 0, false
		}
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	case json.Number:
		n, err := v.Int64()
		return int(n), err == nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		return n, err == nil
	}
	return 0, false
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func buildImportClient(node parsedNode, prepared *preparedBatch, template clientTemplate, inboundsJSON json.RawMessage) (*model.Client, error) {
	config := generateClientConfig()
	for protocol, overrides := range template.config {
		fields, ok := config[protocol]
		if !ok {
			fields = map[string]any{}
			config[protocol] = fields
		}
		for k, v := range overrides {
			fields[k] = v
		}
	}
	client := &model.Client{
		Enable:     template.enable,
		Name:       node.clientName,
		Inbounds:   inboundsJSON,
		Volume:     template.volume,
		Expiry:     template.expiry,
		Desc:       template.desc,
		Group:      prepared.meta.group,
		Remark:     template.remark,
		DelayStart: template.delayStart,
		AutoReset:  template.autoReset,
		ResetDays:  template.resetDays,
		MaxConns:   template.maxConns,
		IPLimit:    template.ipLimit,
		UpMbps:     template.upMbps,
		DownMbps:   template.downMbps,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	client.Config = configJSON
	if len(template.links) > 0 {
		links := make([]map[string]string, 0, len(template.links))
		for _, link := range template.links {
			links = append(links, map[string]string{"type": link.Type, "remark": link.Remark, "uri": link.URI})
		}
		linkJSON, err := json.Marshal(links)
		if err != nil {
			return nil, err
		}
		client.Links = linkJSON
	} else {
		client.Links = json.RawMessage("[]")
	}
	if err := setConfigIdentity(client); err != nil {
		return nil, err
	}
	return client, nil
}

func generateClientConfig() map[string]map[string]any {
	password := common.Random(10)
	ss16 := randomBase64(16)
	ss32 := randomBase64(32)
	uuid := randomUUID()
	return map[string]map[string]any{
		"mixed":         {"username": "", "password": password},
		"socks":         {"username": "", "password": password},
		"http":          {"username": "", "password": password},
		"shadowsocks":   {"name": "", "password": ss32},
		"shadowsocks16": {"name": "", "password": ss16},
		"shadowtls":     {"name": "", "password": ss32},
		"vmess":         {"name": "", "uuid": uuid, "alterId": 0},
		"vless":         {"name": "", "uuid": uuid, "flow": "xtls-rprx-vision"},
		"anytls":        {"name": "", "password": password},
		"trojan":        {"name": "", "password": password},
		"naive":         {"username": "", "password": password},
		"hysteria":      {"name": "", "auth_str": password},
		"snell":         {"name": "", "userkey": common.Random(32)},
		"tuic":          {"name": "", "uuid": uuid, "password": password},
		"hysteria2":     {"name": "", "password": password},
	}
}

func randomBase64(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return common.Random(n)
	}
	return base64.StdEncoding.EncodeToString(buf)
}

func randomUUID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return common.Random(32)
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}

func importOutboundModel(outbound map[string]any) (*model.Outbound, error) {
	encoded, err := json.Marshal(outbound)
	if err != nil {
		return nil, err
	}
	var row model.Outbound
	if err := row.UnmarshalJSON(encoded); err != nil {
		return nil, err
	}
	if row.Type == "" || row.Tag == "" {
		return nil, common.NewError("outbound has no type or tag")
	}
	return &row, nil
}

// existingValues returns the subset of values already present in a table.
// Chunked because SQLite rejects a statement with more bound variables than
// its compile-time limit; a 1000-node batch must not fail on that alone.
func existingValues(db *gorm.DB, table, column string, values []string) (map[string]bool, error) {
	found := make(map[string]bool)
	const chunkSize = 400
	for start := 0; start < len(values); start += chunkSize {
		end := start + chunkSize
		if end > len(values) {
			end = len(values)
		}
		var rows []string
		if err := db.Table(table).Where(column+" IN ?", values[start:end]).Pluck(column, &rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			found[row] = true
		}
	}
	return found, nil
}

func clientsByNames(db *gorm.DB, names []string) ([]*model.Client, error) {
	byName := make(map[string]*model.Client, len(names))
	const chunkSize = 400
	for start := 0; start < len(names); start += chunkSize {
		end := start + chunkSize
		if end > len(names) {
			end = len(names)
		}
		var rows []*model.Client
		if err := db.Model(model.Client{}).Select("id, name").Where("name IN ?", names[start:end]).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			byName[row.Name] = row
		}
	}
	ordered := make([]*model.Client, 0, len(names))
	for _, name := range names {
		row, ok := byName[name]
		if !ok {
			return nil, common.NewError("client was not saved: ", name)
		}
		ordered = append(ordered, row)
	}
	return ordered, nil
}

func readImportBaseConfig(db *gorm.DB) string {
	var row model.Setting
	if err := db.Where("key = ?", "config").Limit(1).Find(&row).Error; err == nil && strings.TrimSpace(row.Value) != "" {
		return row.Value
	}
	return defaultConfig
}

// addImportRouteRules inserts one auth_user route per client after any leading
// sniff/hijack-dns rules. Existing rules keep their order and purpose; the new
// rules simply get the first chance to match a client they own.
func addImportRouteRules(db *gorm.DB, tags, names []string) (json.RawMessage, error) {
	base := readImportBaseConfig(db)
	var singbox SingBoxConfig
	if err := json.Unmarshal([]byte(base), &singbox); err != nil {
		return nil, common.NewError("cannot read the base config: ", err)
	}
	route := map[string]json.RawMessage{}
	if len(singbox.Route) > 0 {
		if err := json.Unmarshal(singbox.Route, &route); err != nil {
			return nil, common.NewError("cannot read route: ", err)
		}
	}
	var rules []json.RawMessage
	if raw, ok := route["rules"]; ok && len(raw) > 0 {
		if err := json.Unmarshal(raw, &rules); err != nil {
			return nil, common.NewError("cannot read route rules: ", err)
		}
	}
	insertAt := 0
	for insertAt < len(rules) {
		var probe struct {
			Action string `json:"action"`
		}
		if err := json.Unmarshal(rules[insertAt], &probe); err != nil {
			break
		}
		if probe.Action == "sniff" || probe.Action == "hijack-dns" {
			insertAt++
			continue
		}
		break
	}
	merged := make([]json.RawMessage, 0, len(rules)+len(names))
	merged = append(merged, rules[:insertAt]...)
	for i := range names {
		rule, err := json.Marshal(map[string]any{
			"auth_user": []string{names[i]},
			"action":    "route",
			"outbound":  tags[i],
		})
		if err != nil {
			return nil, err
		}
		merged = append(merged, rule)
	}
	merged = append(merged, rules[insertAt:]...)
	encodedRules, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}
	route["rules"] = encodedRules
	encodedRoute, err := json.Marshal(route)
	if err != nil {
		return nil, err
	}
	singbox.Route = encodedRoute
	return json.Marshal(singbox)
}

func saveImportBaseConfig(tx *gorm.DB, config json.RawMessage) error {
	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	res := tx.Model(model.Setting{}).Where("key = ?", "config").Update("value", string(encoded))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return tx.Create(&model.Setting{Key: "config", Value: string(encoded)}).Error
	}
	return nil
}

// prospectiveConfig assembles exactly what the running core would read: the
// stored base config, the current rows, plus (for a dry run) outbounds that are
// not in the table yet. The result is decoded by sing-box itself.
func prospectiveConfig(db *gorm.DB, baseConfig json.RawMessage, extraOutbounds []*model.Outbound) ([]byte, error) {
	var singbox SingBoxConfig
	base := []byte(baseConfig)
	if len(base) == 0 {
		base = []byte(readImportBaseConfig(db))
	}
	if err := json.Unmarshal(base, &singbox); err != nil {
		return nil, common.NewError("cannot decode the base config: ", err)
	}
	var err error
	singbox.Inbounds, err = (&InboundService{}).GetAllConfig(db)
	if err != nil {
		return nil, err
	}
	singbox.Outbounds, err = (&OutboundService{}).GetAllConfig(db)
	if err != nil {
		return nil, err
	}
	for _, extra := range extraOutbounds {
		encoded, err := extra.MarshalJSON()
		if err != nil {
			return nil, err
		}
		singbox.Outbounds = append(singbox.Outbounds, encoded)
	}
	singbox.Services, err = (&ServicesService{}).GetAllConfig(db)
	if err != nil {
		return nil, err
	}
	singbox.Endpoints, err = (&EndpointService{}).GetAllConfig(db)
	if err != nil {
		return nil, err
	}
	if err := ensureDefaultHTTPClient(&singbox); err != nil {
		return nil, err
	}
	return json.MarshalIndent(singbox, "", "  ")
}

func (s *ImportService) writeBatch(tx *gorm.DB, prepared *preparedBatch, hostname string) ([]*model.Client, error) {
	names := make([]string, len(prepared.clients))
	for i, client := range prepared.clients {
		names[i] = client.Name
	}
	clientsJSON, err := json.Marshal(prepared.clients)
	if err != nil {
		return nil, err
	}
	// Reuse the single-client batch path: it validates names, preserves
	// non-local links, generates the local links and stamps createdAt.
	if _, err := (&ClientService{}).Save(tx, "addbulk", clientsJSON, hostname); err != nil {
		return nil, err
	}
	saved, err := clientsByNames(tx, names)
	if err != nil {
		return nil, err
	}
	for _, outbound := range prepared.outbounds {
		if err := tx.Create(outbound).Error; err != nil {
			return nil, common.NewError("unable to save outbound ", outbound.Tag, ": ", err)
		}
	}
	if err := saveImportBaseConfig(tx, prepared.baseConfig); err != nil {
		return nil, err
	}
	config, err := prospectiveConfig(tx, nil, nil)
	if err != nil {
		return nil, err
	}
	if err := validationCore().ValidateConfig(config); err != nil {
		return nil, common.NewError("the resulting configuration is invalid: ", err)
	}
	return saved, nil
}

func (s *ImportService) applyAndRecord(batch *model.ImportBatch, res *ImportResult) *ImportError {
	coreStatus := applyImportCoreStatus()
	res.Core = coreStatus
	status := "committed"
	errorText := ""
	appliedAt := int64(0)
	switch coreStatus.Status {
	case "applied":
		status = "applied"
		appliedAt = coreStatus.At
	case "failed":
		status = "apply_failed"
		errorText = coreStatus.Error
	}
	res.BatchStatus = status
	encoded, err := json.Marshal(res)
	if err != nil {
		return &ImportError{Code: "INTERNAL", Message: err.Error(), BatchID: batch.BatchId}
	}
	if err := database.GetDB().Model(&model.ImportBatch{}).Where("batch_id = ?", batch.BatchId).Updates(map[string]any{
		"status":     status,
		"error":      errorText,
		"result":     string(encoded),
		"applied_at": appliedAt,
	}).Error; err != nil {
		// The data is committed; only the bookkeeping failed. Say so rather
		// than tell the caller the creation failed.
		return &ImportError{Code: "BATCH_STATUS_UNSAVED", Message: "the batch was committed but its core status could not be stored; retry reapplyImportBatch", BatchID: batch.BatchId, Retry: "POST apiv2/reapplyImportBatch {\"batchId\":\"" + batch.BatchId + "\"}"}
	}
	batch.Status = status
	batch.Error = errorText
	batch.AppliedAt = appliedAt
	batch.Result = string(encoded)
	if status == "apply_failed" {
		core := coreStatus
		return &ImportError{
			Code:    "CORE_APPLY_FAILED",
			Message: "the batch is committed but the core rejected it; use reapplyImportBatch, do not create it again",
			BatchID: batch.BatchId,
			Core:    &core,
			Retry:   "POST apiv2/reapplyImportBatch {\"batchId\":\"" + batch.BatchId + "\"}",
		}
	}
	return nil
}

func applyImportCoreStatus() ImportCoreStatus {
	if applyImportCore != nil {
		if err := applyImportCore(); err != nil {
			if errors.Is(err, errImportCoreSkipped) {
				return ImportCoreStatus{Status: "skipped", Reason: err.Error(), At: nowUnix()}
			}
			return ImportCoreStatus{Status: "failed", Error: err.Error(), At: nowUnix()}
		}
		return ImportCoreStatus{Status: "applied", At: nowUnix()}
	}
	if corePtr == nil {
		return ImportCoreStatus{Status: "skipped", Reason: "core is not initialized", At: nowUnix()}
	}
	if maintenance, err := (&SettingService{}).GetMaintenance(); err == nil && maintenance {
		return ImportCoreStatus{Status: "skipped", Reason: "core is stopped for maintenance", At: nowUnix()}
	}
	if err := (&ConfigService{}).RestartCore(); err != nil {
		return ImportCoreStatus{Status: "failed", Error: err.Error(), At: nowUnix()}
	}
	return ImportCoreStatus{Status: "applied", At: nowUnix()}
}

// verifyBatchData checks that reapply would apply the batch that is actually
// on disk, not recreate it. Missing rows are an error; an empty batch is not.
func (s *ImportService) verifyBatchData(res *ImportResult) error {
	if len(res.Items) == 0 {
		return nil
	}
	db := database.GetDB()
	names := itemNames(res.Items)
	saved, err := clientsByNames(db, names)
	if err != nil {
		return common.NewError("the clients of this batch are missing: ", err)
	}
	if len(saved) != len(names) {
		return common.NewErrorf("the batch expected %d clients, found %d", len(names), len(saved))
	}
	tags := itemTags(res.Items)
	found, err := existingValues(db, "outbounds", "tag", tags)
	if err != nil {
		return err
	}
	for _, tag := range tags {
		if !found[tag] {
			return common.NewError("the outbound of this batch is missing: ", tag)
		}
	}
	var singbox SingBoxConfig
	if err := json.Unmarshal([]byte(readImportBaseConfig(db)), &singbox); err != nil {
		return err
	}
	route := map[string]json.RawMessage{}
	if err := json.Unmarshal(singbox.Route, &route); err != nil {
		return err
	}
	var rules []struct {
		AuthUser []string `json:"auth_user"`
		Outbound string   `json:"outbound"`
		Action   string   `json:"action"`
	}
	if raw, ok := route["rules"]; ok {
		if err := json.Unmarshal(raw, &rules); err != nil {
			return err
		}
	}
	matched := map[string]string{}
	for _, rule := range rules {
		if rule.Action != "route" || len(rule.AuthUser) != 1 {
			continue
		}
		matched[rule.AuthUser[0]] = rule.Outbound
	}
	for _, item := range res.Items {
		if matched[item.ClientName] != item.OutboundTag {
			return common.NewError("the route for this batch is missing or changed: ", item.ClientName)
		}
	}
	return nil
}
