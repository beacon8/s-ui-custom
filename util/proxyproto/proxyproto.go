// Package proxyproto parses the HAProxy PROXY protocol header that a trusted
// TCP relay writes in front of each proxied connection.
//
// The parser is deliberately strict: only PROXY v1 TCP4/TCP6/UNKNOWN and PROXY
// v2 LOCAL/TCP4/TCP6 are accepted. UDP and UNIX address families are rejected,
// because a UDP association or a UNIX socket cannot carry the per-connection
// client address this feature depends on.
package proxyproto

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"strconv"
	"strings"
)

// MaxV1Line is the longest legal PROXY v1 line, CRLF included.
const MaxV1Line = 107

// maxV2Payload bounds the v2 address block plus TLV section. The spec allows
// up to 65535 bytes, but a real header is well under 256 and an unbounded
// allocation driven by the peer is not acceptable.
const maxV2Payload = 4096

var (
	// ErrNoHeader reports that the connection does not start with a PROXY
	// header. Callers must not fall back to the peer address in that case.
	ErrNoHeader = errors.New("proxyproto: no PROXY header")
)

var v2Signature = [12]byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A}

// Version is the PROXY protocol major version.
type Version uint8

const (
	Version1 Version = 1
	Version2 Version = 2
)

// Command is the v2 command nibble. v1 only carries the equivalent of Proxy.
type Command uint8

const (
	// CommandLocal marks a v2 LOCAL header: the sender's own health check.
	CommandLocal Command = 0
	// CommandProxy marks a forwarded connection with a known client address.
	CommandProxy Command = 1
)

func (c Command) String() string {
	switch c {
	case CommandLocal:
		return "LOCAL"
	case CommandProxy:
		return "PROXY"
	default:
		return "UNKNOWN"
	}
}

// Family is the address family of a parsed header.
type Family uint8

const (
	FamilyUnspec Family = 0
	FamilyInet   Family = 1
	FamilyInet6  Family = 2
)

func (f Family) String() string {
	switch f {
	case FamilyInet:
		return "INET"
	case FamilyInet6:
		return "INET6"
	default:
		return "UNSPEC"
	}
}

// Header is a parsed PROXY header.
type Header struct {
	Version Version
	Command Command
	Family  Family
	Source  netip.Addr
	// SourcePort is zero when the family carries no port (UNSPEC).
	SourcePort      uint16
	Destination     netip.Addr
	DestinationPort uint16
}

// ReadHeader consumes exactly one PROXY header from reader. Bytes read past
// the header stay buffered in reader; callers must hand those bytes to the
// wrapped connection instead of dropping them.
func ReadHeader(reader *bufio.Reader) (Header, error) {
	first, err := reader.Peek(1)
	if err != nil {
		return Header{}, err
	}
	switch first[0] {
	case 'P':
		return readV1(reader)
	case v2Signature[0]:
		return readV2(reader)
	default:
		return Header{}, ErrNoHeader
	}
}

func readV1(reader *bufio.Reader) (Header, error) {
	line := make([]byte, 0, MaxV1Line)
	for len(line) <= MaxV1Line {
		b, err := reader.ReadByte()
		if err != nil {
			return Header{}, err
		}
		line = append(line, b)
		if b == '\n' {
			break
		}
	}
	if len(line) > MaxV1Line {
		return Header{}, errors.New("proxyproto: v1 line exceeds 107 bytes")
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return Header{}, errors.New("proxyproto: v1 line is not CRLF terminated")
	}
	fields := strings.Split(string(line[:len(line)-2]), " ")
	if len(fields) < 2 || fields[0] != "PROXY" {
		return Header{}, errors.New("proxyproto: invalid v1 prefix")
	}
	header := Header{Version: Version1, Command: CommandProxy}
	switch fields[1] {
	case "UNKNOWN":
		// Everything after UNKNOWN is implementation-specific and ignored.
		header.Family = FamilyUnspec
		return header, nil
	case "TCP4":
		header.Family = FamilyInet
	case "TCP6":
		header.Family = FamilyInet6
	default:
		return Header{}, fmt.Errorf("proxyproto: unsupported v1 protocol %q", fields[1])
	}
	if len(fields) != 6 {
		return Header{}, fmt.Errorf("proxyproto: v1 expected 6 fields, got %d", len(fields))
	}
	source, err := netip.ParseAddr(fields[2])
	if err != nil || !source.IsValid() {
		return Header{}, fmt.Errorf("proxyproto: invalid v1 source address %q", fields[2])
	}
	destination, err := netip.ParseAddr(fields[3])
	if err != nil || !destination.IsValid() {
		return Header{}, fmt.Errorf("proxyproto: invalid v1 destination address %q", fields[3])
	}
	if header.Family == FamilyInet && (!source.Is4() || !destination.Is4()) {
		return Header{}, errors.New("proxyproto: v1 TCP4 with a non-IPv4 address")
	}
	if header.Family == FamilyInet6 && (!source.Is6() || source.Is4In6() || !destination.Is6() || destination.Is4In6()) {
		return Header{}, errors.New("proxyproto: v1 TCP6 with a non-IPv6 address")
	}
	header.Source = source
	header.Destination = destination
	if header.SourcePort, err = parsePort(fields[4]); err != nil {
		return Header{}, err
	}
	if header.DestinationPort, err = parsePort(fields[5]); err != nil {
		return Header{}, err
	}
	return header, nil
}

func parsePort(value string) (uint16, error) {
	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("proxyproto: invalid port %q", value)
	}
	return uint16(port), nil
}

func readV2(reader *bufio.Reader) (Header, error) {
	var signature [12]byte
	if _, err := io.ReadFull(reader, signature[:]); err != nil {
		return Header{}, err
	}
	if signature != v2Signature {
		return Header{}, ErrNoHeader
	}
	var fixed [4]byte
	if _, err := io.ReadFull(reader, fixed[:]); err != nil {
		return Header{}, err
	}
	versionCommand := fixed[0]
	familyProtocol := fixed[1]
	payloadLength := int(binary.BigEndian.Uint16(fixed[2:4]))
	if versionCommand>>4 != 2 {
		return Header{}, fmt.Errorf("proxyproto: unsupported v2 version %d", versionCommand>>4)
	}
	command := Command(versionCommand & 0x0F)
	if command != CommandLocal && command != CommandProxy {
		return Header{}, fmt.Errorf("proxyproto: unsupported v2 command %d", command)
	}
	if payloadLength > maxV2Payload {
		return Header{}, fmt.Errorf("proxyproto: v2 payload %d exceeds %d bytes", payloadLength, maxV2Payload)
	}
	payload := make([]byte, payloadLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return Header{}, err
	}
	if command == CommandLocal {
		// A LOCAL header must not be trusted for a client address, even when
		// its address block is present.
		return Header{Version: Version2, Command: CommandLocal, Family: FamilyUnspec}, nil
	}
	family := Family(familyProtocol >> 4)
	protocol := familyProtocol & 0x0F
	header := Header{Version: Version2, Command: CommandProxy, Family: family}
	switch family {
	case FamilyInet:
		if protocol != 1 {
			return Header{}, fmt.Errorf("proxyproto: unsupported v2 INET protocol %d (only TCP stream)", protocol)
		}
		if len(payload) < 12 {
			return Header{}, errors.New("proxyproto: v2 INET payload shorter than 12 bytes")
		}
		header.Source = netip.AddrFrom4([4]byte(payload[0:4]))
		header.Destination = netip.AddrFrom4([4]byte(payload[4:8]))
		header.SourcePort = binary.BigEndian.Uint16(payload[8:10])
		header.DestinationPort = binary.BigEndian.Uint16(payload[10:12])
	case FamilyInet6:
		if protocol != 1 {
			return Header{}, fmt.Errorf("proxyproto: unsupported v2 INET6 protocol %d (only TCP stream)", protocol)
		}
		if len(payload) < 36 {
			return Header{}, errors.New("proxyproto: v2 INET6 payload shorter than 36 bytes")
		}
		header.Source = netip.AddrFrom16([16]byte(payload[0:16]))
		header.Destination = netip.AddrFrom16([16]byte(payload[16:32]))
		header.SourcePort = binary.BigEndian.Uint16(payload[32:34])
		header.DestinationPort = binary.BigEndian.Uint16(payload[34:36])
	default:
		return Header{}, fmt.Errorf("proxyproto: unsupported v2 address family %d", family)
	}
	return header, nil
}
