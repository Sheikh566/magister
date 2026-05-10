package course

// DNS wire-format helpers, UDP/TCP exchange, and fake servers for dns_tests.

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	dnsClassIN = 1

	dnsTypeA    = 1
	dnsTypeNS   = 2
	dnsTypeAAAA = 28

	dnsFlagQR = 0x8000
	dnsFlagAA = 0x0400
	dnsFlagTC = 0x0200
	dnsFlagRD = 0x0100
	dnsFlagRA = 0x0080

	dnsRCodeNoError  = 0
	dnsRCodeNXDomain = 3
	dnsRCodeNotImpl  = 4
)

func dnsRandomID() (uint16, error) {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b[:]), nil
}

type dnsQuestion struct {
	Name  string
	Type  uint16
	Class uint16
}

type dnsRecord struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  []byte
}

type dnsMessage struct {
	ID          uint16
	Flags       uint16
	Questions   []dnsQuestion
	Answers     []dnsRecord
	Authorities []dnsRecord
	Additionals []dnsRecord
	Raw         []byte
}

func (h *Harness) startDNS(ctx context.Context, extraEnv []string) (*managedServer, error) {
	if err := waitForUDPPortFree(ctx, h.spec.Host, h.spec.Port); err != nil {
		return nil, err
	}

	return h.startWithReady(ctx, extraEnv, func(srv *managedServer) error {
		return srv.waitForDNSReady(ctx)
	})
}

func waitForUDPPortFree(ctx context.Context, host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	deadline := time.After(3 * time.Second)
	for {
		conn, err := net.ListenPacket("udp", addr)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("UDP port %d still in use: %w", port, ctx.Err())
		case <-deadline:
			return fmt.Errorf("UDP port %d is already in use — a previous server may still be running", port)
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (s *managedServer) waitForDNSReady(ctx context.Context) error {
	deadline := time.NewTimer(s.spec.StartupTimeout)
	defer deadline.Stop()

	for {
		id, err := dnsRandomID()
		if err != nil {
			return fmt.Errorf("generate DNS probe id: %w", err)
		}
		query := buildDNSQuery(id, "ready.magister.test", dnsTypeA, true)

		select {
		case <-ctx.Done():
			return fmt.Errorf("DNS server startup canceled: %w\n\n%s", ctx.Err(), s.output.String())
		case err := <-s.done:
			if err == nil {
				return fmt.Errorf("DNS server exited before becoming ready\n\n%s", s.output.String())
			}
			return fmt.Errorf("DNS server exited before becoming ready: %w\n\n%s", err, s.output.String())
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for %s to answer DNS over UDP on %s\n\n%s", s.spec.Command, s.addr(), s.output.String())
		default:
		}

		if _, err := dnsUDPExchange(s.addr(), query, 150*time.Millisecond); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func buildDNSQuery(id uint16, name string, qtype uint16, recursionDesired bool) []byte {
	flags := uint16(0)
	if recursionDesired {
		flags |= dnsFlagRD
	}
	return buildDNSQueryWithQuestions(id, flags, []dnsQuestion{{
		Name:  name,
		Type:  qtype,
		Class: dnsClassIN,
	}})
}

func buildDNSQueryWithQuestions(id, flags uint16, questions []dnsQuestion) []byte {
	msg := make([]byte, 12)
	binary.BigEndian.PutUint16(msg[0:2], id)
	binary.BigEndian.PutUint16(msg[2:4], flags)
	binary.BigEndian.PutUint16(msg[4:6], uint16(len(questions)))

	for _, question := range questions {
		msg = append(msg, encodeDNSName(question.Name)...)
		msg = appendUint16(msg, question.Type)
		msg = appendUint16(msg, question.Class)
	}
	return msg
}

func buildCompressedDuplicateQuestionQuery(id uint16, name string) []byte {
	msg := make([]byte, 12)
	binary.BigEndian.PutUint16(msg[0:2], id)
	binary.BigEndian.PutUint16(msg[2:4], dnsFlagRD)
	binary.BigEndian.PutUint16(msg[4:6], 2)
	msg = append(msg, encodeDNSName(name)...)
	msg = appendUint16(msg, dnsTypeA)
	msg = appendUint16(msg, dnsClassIN)
	msg = append(msg, 0xc0, 0x0c)
	msg = appendUint16(msg, dnsTypeA)
	msg = appendUint16(msg, dnsClassIN)
	return msg
}

func buildDNSResponse(req dnsMessage, flags uint16, answers, authorities, additionals []dnsRecord) []byte {
	msg := make([]byte, 12)
	binary.BigEndian.PutUint16(msg[0:2], req.ID)
	binary.BigEndian.PutUint16(msg[2:4], flags)
	binary.BigEndian.PutUint16(msg[4:6], uint16(len(req.Questions)))
	binary.BigEndian.PutUint16(msg[6:8], uint16(len(answers)))
	binary.BigEndian.PutUint16(msg[8:10], uint16(len(authorities)))
	binary.BigEndian.PutUint16(msg[10:12], uint16(len(additionals)))

	for _, question := range req.Questions {
		msg = append(msg, encodeDNSName(question.Name)...)
		msg = appendUint16(msg, question.Type)
		msg = appendUint16(msg, question.Class)
	}
	msg = appendDNSRecords(msg, answers)
	msg = appendDNSRecords(msg, authorities)
	msg = appendDNSRecords(msg, additionals)
	return msg
}

func appendDNSRecords(msg []byte, records []dnsRecord) []byte {
	for _, record := range records {
		msg = append(msg, encodeDNSName(record.Name)...)
		msg = appendUint16(msg, record.Type)
		msg = appendUint16(msg, record.Class)
		msg = appendUint32(msg, record.TTL)
		msg = appendUint16(msg, uint16(len(record.Data)))
		msg = append(msg, record.Data...)
	}
	return msg
}

func parseDNSMessage(packet []byte) (dnsMessage, error) {
	if len(packet) < 12 {
		return dnsMessage{}, fmt.Errorf("DNS packet too short: got %d bytes", len(packet))
	}

	msg := dnsMessage{
		ID:    binary.BigEndian.Uint16(packet[0:2]),
		Flags: binary.BigEndian.Uint16(packet[2:4]),
		Raw:   append([]byte(nil), packet...),
	}
	qd := int(binary.BigEndian.Uint16(packet[4:6]))
	an := int(binary.BigEndian.Uint16(packet[6:8]))
	ns := int(binary.BigEndian.Uint16(packet[8:10]))
	ar := int(binary.BigEndian.Uint16(packet[10:12]))

	offset := 12
	for i := 0; i < qd; i++ {
		name, next, err := decodeDNSName(packet, offset)
		if err != nil {
			return dnsMessage{}, err
		}
		if next+4 > len(packet) {
			return dnsMessage{}, errors.New("truncated DNS question")
		}
		msg.Questions = append(msg.Questions, dnsQuestion{
			Name:  name,
			Type:  binary.BigEndian.Uint16(packet[next : next+2]),
			Class: binary.BigEndian.Uint16(packet[next+2 : next+4]),
		})
		offset = next + 4
	}

	var err error
	msg.Answers, offset, err = parseDNSRecords(packet, offset, an)
	if err != nil {
		return dnsMessage{}, err
	}
	msg.Authorities, offset, err = parseDNSRecords(packet, offset, ns)
	if err != nil {
		return dnsMessage{}, err
	}
	msg.Additionals, _, err = parseDNSRecords(packet, offset, ar)
	if err != nil {
		return dnsMessage{}, err
	}

	return msg, nil
}

func parseDNSRecords(packet []byte, offset, count int) ([]dnsRecord, int, error) {
	records := make([]dnsRecord, 0, count)
	for i := 0; i < count; i++ {
		name, next, err := decodeDNSName(packet, offset)
		if err != nil {
			return nil, 0, err
		}
		if next+10 > len(packet) {
			return nil, 0, errors.New("truncated DNS record header")
		}
		rdLen := int(binary.BigEndian.Uint16(packet[next+8 : next+10]))
		dataStart := next + 10
		dataEnd := dataStart + rdLen
		if dataEnd > len(packet) {
			return nil, 0, errors.New("truncated DNS record data")
		}
		records = append(records, dnsRecord{
			Name:  name,
			Type:  binary.BigEndian.Uint16(packet[next : next+2]),
			Class: binary.BigEndian.Uint16(packet[next+2 : next+4]),
			TTL:   binary.BigEndian.Uint32(packet[next+4 : next+8]),
			Data:  append([]byte(nil), packet[dataStart:dataEnd]...),
		})
		offset = dataEnd
	}
	return records, offset, nil
}

func decodeDNSName(packet []byte, offset int) (string, int, error) {
	labels := make([]string, 0, 4)
	next := offset
	jumped := false
	seen := map[int]bool{}

	for {
		if offset >= len(packet) {
			return "", 0, errors.New("truncated DNS name")
		}
		if seen[offset] {
			return "", 0, errors.New("DNS compression pointer loop")
		}
		seen[offset] = true

		length := int(packet[offset])
		switch {
		case length&0xc0 == 0xc0:
			if offset+1 >= len(packet) {
				return "", 0, errors.New("truncated DNS compression pointer")
			}
			ptr := int(binary.BigEndian.Uint16(packet[offset:offset+2]) & 0x3fff)
			if !jumped {
				next = offset + 2
				jumped = true
			}
			offset = ptr
		case length&0xc0 != 0:
			return "", 0, errors.New("unsupported DNS label encoding")
		case length == 0:
			if !jumped {
				next = offset + 1
			}
			return strings.Join(labels, "."), next, nil
		default:
			start := offset + 1
			end := start + length
			if end > len(packet) {
				return "", 0, errors.New("truncated DNS label")
			}
			labels = append(labels, strings.ToLower(string(packet[start:end])))
			offset = end
		}
	}
}

func encodeDNSName(name string) []byte {
	name = normalizeDNSName(name)
	if name == "" {
		return []byte{0}
	}
	var out []byte
	for _, label := range strings.Split(name, ".") {
		out = append(out, byte(len(label)))
		out = append(out, label...)
	}
	return append(out, 0)
}

func normalizeDNSName(name string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(name), "."))
}

func appendUint16(msg []byte, value uint16) []byte {
	return append(msg, byte(value>>8), byte(value))
}

func appendUint32(msg []byte, value uint32) []byte {
	return append(msg, byte(value>>24), byte(value>>16), byte(value>>8), byte(value))
}

func dnsARecord(name, ip string, ttl uint32) dnsRecord {
	parsed := net.ParseIP(ip).To4()
	if parsed == nil {
		panic(fmt.Sprintf("invalid IPv4 address %q", ip))
	}
	return dnsRecord{Name: name, Type: dnsTypeA, Class: dnsClassIN, TTL: ttl, Data: []byte(parsed)}
}

func dnsAAAARecord(name, ip string, ttl uint32) dnsRecord {
	parsed := net.ParseIP(ip).To16()
	if parsed == nil {
		panic(fmt.Sprintf("invalid IPv6 address %q", ip))
	}
	return dnsRecord{Name: name, Type: dnsTypeAAAA, Class: dnsClassIN, TTL: ttl, Data: []byte(parsed)}
}

func dnsNSRecord(name, target string, ttl uint32) dnsRecord {
	return dnsRecord{Name: name, Type: dnsTypeNS, Class: dnsClassIN, TTL: ttl, Data: encodeDNSName(target)}
}

func dnsUDPExchange(addr string, query []byte, timeout time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	if _, err := conn.Write(query); err != nil {
		return nil, err
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), buf[:n]...), nil
}

func dnsTCPExchange(addr string, query []byte, timeout time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	prefixed := appendUint16(nil, uint16(len(query)))
	prefixed = append(prefixed, query...)
	if _, err := conn.Write(prefixed); err != nil {
		return nil, err
	}
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	size := int(binary.BigEndian.Uint16(header))
	body := make([]byte, size)
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, err
	}
	return body, nil
}

type fakeDNSServer struct {
	conn    *net.UDPConn
	addr    string
	handler func(dnsMessage) []byte

	mu      sync.Mutex
	queries []dnsMessage
}

func startFakeDNSServer(handler func(dnsMessage) []byte) (*fakeDNSServer, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	server := &fakeDNSServer{
		conn:    conn,
		addr:    conn.LocalAddr().String(),
		handler: handler,
	}
	go server.serve()
	return server, nil
}

func (s *fakeDNSServer) serve() {
	buf := make([]byte, 4096)
	for {
		n, remote, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		msg, err := parseDNSMessage(buf[:n])
		if err != nil {
			continue
		}
		s.mu.Lock()
		s.queries = append(s.queries, msg)
		s.mu.Unlock()
		if response := s.handler(msg); response != nil {
			_, _ = s.conn.WriteToUDP(response, remote)
		}
	}
}

func (s *fakeDNSServer) stop() {
	_ = s.conn.Close()
}

func (s *fakeDNSServer) queryCountFor(name string) int {
	name = normalizeDNSName(name)
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, query := range s.queries {
		for _, question := range query.Questions {
			if normalizeDNSName(question.Name) == name {
				count++
			}
		}
	}
	return count
}
