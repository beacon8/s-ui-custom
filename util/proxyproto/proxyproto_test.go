package proxyproto

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net/netip"
	"strings"
	"testing"
)

func v2Header(command byte, familyProtocol byte, payload []byte) []byte {
	header := make([]byte, 0, 16+len(payload))
	header = append(header, v2Signature[:]...)
	header = append(header, 0x20|command, familyProtocol, 0, 0)
	binary.BigEndian.PutUint16(header[14:16], uint16(len(payload)))
	header = append(header, payload...)
	return header
}

func v2TCP4(source, destination netip.Addr, sourcePort, destinationPort uint16, tlv []byte) []byte {
	payload := make([]byte, 12, 12+len(tlv))
	copy(payload[0:4], source.AsSlice())
	copy(payload[4:8], destination.AsSlice())
	binary.BigEndian.PutUint16(payload[8:10], sourcePort)
	binary.BigEndian.PutUint16(payload[10:12], destinationPort)
	payload = append(payload, tlv...)
	return v2Header(1, 0x11, payload)
}

func v2TCP6(source, destination netip.Addr, sourcePort, destinationPort uint16) []byte {
	payload := make([]byte, 36)
	copy(payload[0:16], source.AsSlice())
	copy(payload[16:32], destination.AsSlice())
	binary.BigEndian.PutUint16(payload[32:34], sourcePort)
	binary.BigEndian.PutUint16(payload[34:36], destinationPort)
	return v2Header(1, 0x21, payload)
}

func readAll(t *testing.T, data []byte) (Header, *bufio.Reader) {
	t.Helper()
	reader := bufio.NewReader(bytes.NewReader(data))
	header, err := ReadHeader(reader)
	if err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	return header, reader
}

func TestReadV1TCP4(t *testing.T) {
	header, reader := readAll(t, []byte("PROXY TCP4 192.0.2.10 198.51.100.20 54321 443\r\n"))
	if header.Version != Version1 || header.Command != CommandProxy || header.Family != FamilyInet {
		t.Fatalf("unexpected header metadata: %+v", header)
	}
	if header.Source != netip.MustParseAddr("192.0.2.10") || header.SourcePort != 54321 {
		t.Fatalf("unexpected source: %v:%d", header.Source, header.SourcePort)
	}
	if header.Destination != netip.MustParseAddr("198.51.100.20") || header.DestinationPort != 443 {
		t.Fatalf("unexpected destination: %v:%d", header.Destination, header.DestinationPort)
	}
	if reader.Buffered() != 0 {
		t.Fatalf("unexpected buffered bytes: %d", reader.Buffered())
	}
}

func TestReadV1TCP6(t *testing.T) {
	header, _ := readAll(t, []byte("PROXY TCP6 2001:db8::1 2001:db8::2 1234 443\r\n"))
	if header.Family != FamilyInet6 {
		t.Fatalf("family = %v", header.Family)
	}
	if header.Source != netip.MustParseAddr("2001:db8::1") || header.SourcePort != 1234 {
		t.Fatalf("unexpected source: %v:%d", header.Source, header.SourcePort)
	}
}

func TestReadV1Unknown(t *testing.T) {
	header, _ := readAll(t, []byte("PROXY UNKNOWN\r\n"))
	if header.Command != CommandProxy || header.Family != FamilyUnspec || header.Source.IsValid() {
		t.Fatalf("unexpected UNKNOWN header: %+v", header)
	}
}

func TestReadV1SplitByteByByte(t *testing.T) {
	payload := []byte("PROXY TCP4 203.0.113.7 198.51.100.20 1000 443\r\nGET / HTTP/1.1\r\n\r\n")
	reader := bufio.NewReaderSize(&oneByteReader{data: payload}, 1)
	header, err := ReadHeader(reader)
	if err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	if header.Source != netip.MustParseAddr("203.0.113.7") {
		t.Fatalf("source = %v", header.Source)
	}
	rest, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read rest: %v", err)
	}
	if string(rest) != "GET / HTTP/1.1\r\n\r\n" {
		t.Fatalf("rest = %q", rest)
	}
}

func TestReadV1TooLong(t *testing.T) {
	data := "PROXY TCP4 " + strings.Repeat("1", 200) + "\r\n"
	reader := bufio.NewReader(strings.NewReader(data))
	if _, err := ReadHeader(reader); err == nil {
		t.Fatal("expected an error for an overlong v1 line")
	}
}

func TestReadV1BadTermination(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("PROXY TCP4 192.0.2.1 198.51.100.1 1 2\n"))
	if _, err := ReadHeader(reader); err == nil {
		t.Fatal("expected an error for a v1 line without CR")
	}
}

func TestReadV1RejectsUDP(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("PROXY UDP4 192.0.2.1 198.51.100.1 1 2\r\n"))
	if _, err := ReadHeader(reader); err == nil {
		t.Fatal("expected an error for UDP4")
	}
}

func TestReadV2TCP4WithTLV(t *testing.T) {
	header, reader := readAll(t, v2TCP4(
		netip.MustParseAddr("192.0.2.10"),
		netip.MustParseAddr("198.51.100.20"),
		54321, 443,
		[]byte{0x03, 0x00, 0x04, 0xde, 0xad, 0xbe, 0xef},
	))
	if header.Version != Version2 || header.Family != FamilyInet {
		t.Fatalf("unexpected header: %+v", header)
	}
	if header.Source != netip.MustParseAddr("192.0.2.10") || header.SourcePort != 54321 {
		t.Fatalf("unexpected source: %v:%d", header.Source, header.SourcePort)
	}
	if reader.Buffered() != 0 {
		t.Fatalf("unexpected buffered bytes: %d", reader.Buffered())
	}
}

func TestReadV2TCP6(t *testing.T) {
	header, _ := readAll(t, v2TCP6(
		netip.MustParseAddr("2001:db8::1"),
		netip.MustParseAddr("2001:db8::2"),
		1234, 443,
	))
	if header.Source != netip.MustParseAddr("2001:db8::1") || header.SourcePort != 1234 {
		t.Fatalf("unexpected source: %v:%d", header.Source, header.SourcePort)
	}
}

func TestReadV2Local(t *testing.T) {
	header, _ := readAll(t, v2Header(0, 0x00, nil))
	if header.Command != CommandLocal || header.Source.IsValid() {
		t.Fatalf("unexpected LOCAL header: %+v", header)
	}
}

func TestReadV2RejectsUDP(t *testing.T) {
	reader := bufio.NewReader(bytes.NewReader(v2Header(1, 0x12, make([]byte, 12))))
	if _, err := ReadHeader(reader); err == nil {
		t.Fatal("expected an error for a UDP v2 header")
	}
}

func TestReadV2RejectsUnix(t *testing.T) {
	reader := bufio.NewReader(bytes.NewReader(v2Header(1, 0x31, make([]byte, 216))))
	if _, err := ReadHeader(reader); err == nil {
		t.Fatal("expected an error for a UNIX v2 header")
	}
}

func TestReadV2BadSignature(t *testing.T) {
	data := v2TCP4(netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("198.51.100.1"), 1, 2, nil)
	data[8] = 0
	reader := bufio.NewReader(bytes.NewReader(data))
	if _, err := ReadHeader(reader); err == nil {
		t.Fatal("expected an error for a bad v2 signature")
	}
}

func TestReadV2ShortPayload(t *testing.T) {
	reader := bufio.NewReader(bytes.NewReader(v2Header(1, 0x11, make([]byte, 6))))
	if _, err := ReadHeader(reader); err == nil {
		t.Fatal("expected an error for a truncated INET payload")
	}
}

func TestReadV2OversizePayloadRejectedBeforeAllocation(t *testing.T) {
	header := v2Header(1, 0x11, nil)
	binary.BigEndian.PutUint16(header[14:16], 5000)
	reader := bufio.NewReader(bytes.NewReader(header))
	if _, err := ReadHeader(reader); err == nil {
		t.Fatal("expected an error for an oversized payload")
	}
}

func TestReadKeepsPayloadBuffered(t *testing.T) {
	data := append(v2TCP4(netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("198.51.100.1"), 1, 2, nil), []byte("hello")...)
	reader := bufio.NewReader(bytes.NewReader(data))
	if _, err := ReadHeader(reader); err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	if reader.Buffered() != 5 {
		t.Fatalf("buffered = %d, want 5", reader.Buffered())
	}
	rest, _ := io.ReadAll(reader)
	if string(rest) != "hello" {
		t.Fatalf("rest = %q", rest)
	}
}

func TestReadNoHeader(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\n\r\n"))
	if _, err := ReadHeader(reader); err != ErrNoHeader {
		t.Fatalf("err = %v, want ErrNoHeader", err)
	}
}

type oneByteReader struct {
	data   []byte
	offset int
}

func (r *oneByteReader) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	p[0] = r.data[r.offset]
	r.offset++
	return 1, nil
}
