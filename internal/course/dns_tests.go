package course

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"magister/internal/magister"
)

const (
	dnsLocalName     = "magister.test"
	dnsLocalIPv4     = "203.0.113.10"
	dnsForwardName   = "forward.magister.test"
	dnsForwardIPv4   = "198.51.100.42"
	dnsRecursiveName = "recursive.magister.test"
	dnsRecursiveIPv4 = "198.51.100.99"
	dnsCacheName     = "cache.magister.test"
	dnsCacheIPv4     = "192.0.2.55"
	dnsIPv6Name      = "ipv6.magister.test"
	dnsIPv6Address   = "2001:db8::53"
	dnsLargeName     = "large.magister.test"
	dnsQuestionName  = "google.com"
)

func testDNSUDPServer(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	query := buildDNSQuery(id, dnsLocalName, dnsTypeA, true)
	response, err := dnsUDPExchange(srv.addr(), query, spec.RequestTimeout)
	if err != nil {
		return wrapServerError("send UDP DNS query", err, srv)
	}
	if len(response) == 0 {
		return wrapServerError("validate UDP response", fmt.Errorf("expected at least one response byte"), srv)
	}
	return nil
}

func testDNSHeader(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp, err := queryLearnerDNS(srv, spec, buildDNSQuery(id, dnsLocalName, dnsTypeA, true))
	if err != nil {
		return err
	}
	if resp.ID != id {
		return wrapServerError("validate DNS transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	if resp.Flags&dnsFlagQR == 0 {
		return wrapServerError("validate DNS QR flag", fmt.Errorf("response flag QR is not set"), srv)
	}
	if got := resp.Flags & 0x000f; got != dnsRCodeNoError {
		return wrapServerError("validate DNS RCODE", fmt.Errorf("got %d want %d", got, dnsRCodeNoError), srv)
	}
	return nil
}

func testDNSQuestion(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp, err := queryLearnerDNS(srv, spec, buildDNSQuery(id, dnsQuestionName, dnsTypeA, true))
	if err != nil {
		return err
	}
	if len(resp.Questions) != 1 {
		return wrapServerError("validate DNS question count", fmt.Errorf("got %d want 1", len(resp.Questions)), srv)
	}
	if err := expectDNSQuestion(resp.Questions[0], dnsQuestionName, dnsTypeA); err != nil {
		return wrapServerError("validate echoed DNS question", err, srv)
	}
	if resp.ID != id {
		return wrapServerError("validate DNS transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	return nil
}

func testDNSAnswer(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp, err := queryLearnerDNS(srv, spec, buildDNSQuery(id, dnsLocalName, dnsTypeA, true))
	if err != nil {
		return err
	}
	if err := expectAAnswer(resp, dnsLocalName, dnsLocalIPv4); err != nil {
		return wrapServerError("validate A answer", err, srv)
	}
	if resp.ID != id {
		return wrapServerError("validate DNS transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	return nil
}

func testDNSParseHeader(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	query := buildDNSQueryWithQuestions(id, 1<<11, []dnsQuestion{{
		Name:  dnsLocalName,
		Type:  dnsTypeA,
		Class: dnsClassIN,
	}})
	resp, err := queryLearnerDNS(srv, spec, query)
	if err != nil {
		return err
	}
	if resp.ID != id {
		return wrapServerError("validate unsupported opcode transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	if resp.Flags&dnsFlagQR == 0 {
		return wrapServerError("validate unsupported opcode QR flag", fmt.Errorf("response flag QR is not set"), srv)
	}
	if got := resp.Flags & 0x000f; got != dnsRCodeNotImpl {
		return wrapServerError("validate unsupported opcode RCODE", fmt.Errorf("got %d want %d", got, dnsRCodeNotImpl), srv)
	}
	return nil
}

func testDNSParseQuestion(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	idOK, err := dnsRandomID()
	if err != nil {
		return err
	}
	okResp, err := queryLearnerDNS(srv, spec, buildDNSQuery(idOK, dnsLocalName, dnsTypeA, true))
	if err != nil {
		return err
	}
	if err := expectAAnswer(okResp, dnsLocalName, dnsLocalIPv4); err != nil {
		return wrapServerError("validate known question", err, srv)
	}
	if okResp.ID != idOK {
		return wrapServerError("validate known question transaction ID", fmt.Errorf("got %#x want %#x", okResp.ID, idOK), srv)
	}

	idMissing, err := dnsRandomID()
	if err != nil {
		return err
	}
	missingResp, err := queryLearnerDNS(srv, spec, buildDNSQuery(idMissing, "missing.magister.test", dnsTypeA, true))
	if err != nil {
		return err
	}
	if got := missingResp.Flags & 0x000f; got != dnsRCodeNXDomain {
		return wrapServerError("validate missing-name RCODE", fmt.Errorf("got %d want %d", got, dnsRCodeNXDomain), srv)
	}
	if missingResp.ID != idMissing {
		return wrapServerError("validate missing-name transaction ID", fmt.Errorf("got %#x want %#x", missingResp.ID, idMissing), srv)
	}
	return nil
}

func testDNSCompressedPacket(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp, err := queryLearnerDNS(srv, spec, buildCompressedDuplicateQuestionQuery(id, dnsLocalName))
	if err != nil {
		return err
	}
	if len(resp.Questions) != 2 {
		return wrapServerError("validate compressed DNS question count", fmt.Errorf("got %d want 2", len(resp.Questions)), srv)
	}
	for _, question := range resp.Questions {
		if err := expectDNSQuestion(question, dnsLocalName, dnsTypeA); err != nil {
			return wrapServerError("validate compressed DNS question", err, srv)
		}
	}
	if got := countAAnswers(resp, dnsLocalName, dnsLocalIPv4); got < 2 {
		return wrapServerError("validate answers for compressed questions", fmt.Errorf("got %d matching A answers want at least 2", got), srv)
	}
	if resp.ID != id {
		return wrapServerError("validate compressed DNS transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	return nil
}

func testDNSForwarding(ctx context.Context, spec magister.RunnerConfig) error {
	upstream, err := startAnsweringDNSServer(dnsForwardName, dnsARecord(dnsForwardName, dnsForwardIPv4, 60))
	if err != nil {
		return err
	}
	defer upstream.stop()

	h := newHarness(spec)
	srv, err := h.startDNS(ctx, dnsUpstreamEnv(upstream.addr))
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp, err := queryLearnerDNS(srv, spec, buildDNSQuery(id, dnsForwardName, dnsTypeA, true))
	if err != nil {
		return err
	}
	if err := expectAAnswer(resp, dnsForwardName, dnsForwardIPv4); err != nil {
		return wrapServerError("validate forwarded DNS answer", err, srv)
	}
	if resp.ID != id {
		return wrapServerError("validate forwarded DNS transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	if count := upstream.queryCountFor(dnsForwardName); count == 0 {
		return wrapServerError("validate upstream forwarding", fmt.Errorf("upstream did not receive query for %s", dnsForwardName), srv)
	}
	return nil
}

func testDNSRecursiveResolution(ctx context.Context, spec magister.RunnerConfig) error {
	auth, err := startAnsweringDNSServer(dnsRecursiveName, dnsARecord(dnsRecursiveName, dnsRecursiveIPv4, 120))
	if err != nil {
		return err
	}
	defer auth.stop()

	tld, err := startFakeDNSServer(func(req dnsMessage) []byte {
		flags := uint16(dnsFlagQR | dnsFlagAA | dnsFlagRA)
		return buildDNSResponse(req, flags, nil,
			[]dnsRecord{dnsNSRecord("magister.test", "ns.magister.test", 120)},
			[]dnsRecord{dnsARecord("ns.magister.test", "127.0.0.1", 120)},
		)
	})
	if err != nil {
		return err
	}
	defer tld.stop()

	root, err := startFakeDNSServer(func(req dnsMessage) []byte {
		flags := uint16(dnsFlagQR | dnsFlagAA | dnsFlagRA)
		return buildDNSResponse(req, flags, nil,
			[]dnsRecord{dnsNSRecord("test", "ns.test.magister", 120)},
			[]dnsRecord{dnsARecord("ns.test.magister", "127.0.0.1", 120)},
		)
	})
	if err != nil {
		return err
	}
	defer root.stop()

	h := newHarness(spec)
	env := []string{
		"DNS_ROOT_HINTS=" + root.addr,
		"DNS_ROOT_ADDR=" + root.addr,
		"DNS_AUTHORITY_ADDRS=ns.test.magister=" + tld.addr + ",ns.magister.test=" + auth.addr,
		"DNS_AUTHORITY_MAP=ns.test.magister=" + tld.addr + ",ns.magister.test=" + auth.addr,
	}
	srv, err := h.startDNS(ctx, env)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp, err := queryLearnerDNS(srv, spec, buildDNSQuery(id, dnsRecursiveName, dnsTypeA, true))
	if err != nil {
		return err
	}
	if err := expectAAnswer(resp, dnsRecursiveName, dnsRecursiveIPv4); err != nil {
		return wrapServerError("validate recursive DNS answer", err, srv)
	}
	if resp.ID != id {
		return wrapServerError("validate recursive DNS transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	for label, server := range map[string]*fakeDNSServer{
		"root":          root,
		"TLD":           tld,
		"authoritative": auth,
	} {
		if count := server.queryCountFor(dnsRecursiveName); count == 0 {
			return wrapServerError("validate "+label+" query", fmt.Errorf("%s server was not queried for %s", label, dnsRecursiveName), srv)
		}
	}
	return nil
}

func testDNSCachingTTL(ctx context.Context, spec magister.RunnerConfig) error {
	upstream, err := startAnsweringDNSServer(dnsCacheName, dnsARecord(dnsCacheName, dnsCacheIPv4, 1))
	if err != nil {
		return err
	}
	defer upstream.stop()

	h := newHarness(spec)
	srv, err := h.startDNS(ctx, dnsUpstreamEnv(upstream.addr))
	if err != nil {
		return err
	}
	defer srv.stop()

	id1, err := dnsRandomID()
	if err != nil {
		return err
	}
	query := buildDNSQuery(id1, dnsCacheName, dnsTypeA, true)
	resp1, err := queryLearnerDNS(srv, spec, query)
	if err != nil {
		return err
	}
	if err := expectAAnswer(resp1, dnsCacheName, dnsCacheIPv4); err != nil {
		return wrapServerError("validate first cached-name answer", err, srv)
	}
	if resp1.ID != id1 {
		return wrapServerError("validate first cache response transaction ID", fmt.Errorf("got %#x want %#x", resp1.ID, id1), srv)
	}
	id2, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp2, err := queryLearnerDNS(srv, spec, buildDNSQuery(id2, dnsCacheName, dnsTypeA, true))
	if err != nil {
		return err
	}
	if err := expectAAnswer(resp2, dnsCacheName, dnsCacheIPv4); err != nil {
		return wrapServerError("validate second cached-name answer", err, srv)
	}
	if resp2.ID != id2 {
		return wrapServerError("validate second cache response transaction ID", fmt.Errorf("got %#x want %#x", resp2.ID, id2), srv)
	}
	if count := upstream.queryCountFor(dnsCacheName); count != 1 {
		return wrapServerError("validate cache hit", fmt.Errorf("upstream query count got %d want 1 before TTL expiry", count), srv)
	}

	time.Sleep(1200 * time.Millisecond)
	id3, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp3, err := queryLearnerDNS(srv, spec, buildDNSQuery(id3, dnsCacheName, dnsTypeA, true))
	if err != nil {
		return err
	}
	if err := expectAAnswer(resp3, dnsCacheName, dnsCacheIPv4); err != nil {
		return wrapServerError("validate post-expiry cached-name answer", err, srv)
	}
	if resp3.ID != id3 {
		return wrapServerError("validate post-expiry cache response transaction ID", fmt.Errorf("got %#x want %#x", resp3.ID, id3), srv)
	}
	if count := upstream.queryCountFor(dnsCacheName); count < 2 {
		return wrapServerError("validate cache expiry", fmt.Errorf("upstream query count got %d want at least 2 after TTL expiry", count), srv)
	}
	return nil
}

func testDNSIPv6AAAA(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	resp, err := queryLearnerDNS(srv, spec, buildDNSQuery(id, dnsIPv6Name, dnsTypeAAAA, true))
	if err != nil {
		return err
	}
	if err := expectAAAAAnswer(resp, dnsIPv6Name, dnsIPv6Address); err != nil {
		return wrapServerError("validate AAAA answer", err, srv)
	}
	if resp.ID != id {
		return wrapServerError("validate AAAA transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	return nil
}

func testDNSTCPFallback(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.startDNS(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	id, err := dnsRandomID()
	if err != nil {
		return err
	}
	query := buildDNSQuery(id, dnsLargeName, dnsTypeA, true)
	udpBytes, err := dnsUDPExchange(srv.addr(), query, spec.RequestTimeout)
	if err != nil {
		return wrapServerError("send large UDP DNS query", err, srv)
	}
	if len(udpBytes) < 4 {
		return wrapServerError("validate truncated UDP response", fmt.Errorf("response too short: %d bytes", len(udpBytes)), srv)
	}
	gotUDPID := binary.BigEndian.Uint16(udpBytes[0:2])
	if gotUDPID != id {
		return wrapServerError("validate truncated UDP transaction ID", fmt.Errorf("got %#x want %#x", gotUDPID, id), srv)
	}
	flags := binary.BigEndian.Uint16(udpBytes[2:4])
	if flags&dnsFlagTC == 0 {
		return wrapServerError("validate UDP truncation flag", fmt.Errorf("TC flag is not set"), srv)
	}

	tcpBytes, err := dnsTCPExchange(srv.addr(), query, spec.RequestTimeout)
	if err != nil {
		return wrapServerError("send TCP DNS query", err, srv)
	}
	resp, err := parseDNSMessage(tcpBytes)
	if err != nil {
		return wrapServerError("parse TCP DNS response", err, srv)
	}
	if resp.ID != id {
		return wrapServerError("validate TCP DNS transaction ID", fmt.Errorf("got %#x want %#x", resp.ID, id), srv)
	}
	if resp.Flags&dnsFlagTC != 0 {
		return wrapServerError("validate TCP truncation flag", fmt.Errorf("TC flag should be clear in TCP response"), srv)
	}
	if got := countAAnswers(resp, dnsLargeName, ""); got < 20 {
		return wrapServerError("validate full TCP answer set", fmt.Errorf("got %d A answers want at least 20", got), srv)
	}
	if len(resp.Raw) <= 512 {
		return wrapServerError("validate large TCP response size", fmt.Errorf("got %d bytes want more than 512", len(resp.Raw)), srv)
	}
	return nil
}

func queryLearnerDNS(srv *managedServer, spec magister.RunnerConfig, query []byte) (dnsMessage, error) {
	response, err := dnsUDPExchange(srv.addr(), query, spec.RequestTimeout)
	if err != nil {
		return dnsMessage{}, wrapServerError("exchange DNS over UDP", err, srv)
	}
	msg, err := parseDNSMessage(response)
	if err != nil {
		return dnsMessage{}, wrapServerError("parse DNS response", err, srv)
	}
	return msg, nil
}

func expectDNSQuestion(question dnsQuestion, name string, qtype uint16) error {
	if normalizeDNSName(question.Name) != normalizeDNSName(name) {
		return fmt.Errorf("question name mismatch: got %q want %q", question.Name, normalizeDNSName(name))
	}
	if question.Type != qtype {
		return fmt.Errorf("question type mismatch: got %d want %d", question.Type, qtype)
	}
	if question.Class != dnsClassIN {
		return fmt.Errorf("question class mismatch: got %d want %d", question.Class, dnsClassIN)
	}
	return nil
}

func expectAAnswer(msg dnsMessage, name, ip string) error {
	if count := countAAnswers(msg, name, ip); count == 0 {
		return fmt.Errorf("missing A answer for %s -> %s", name, ip)
	}
	return nil
}

func expectAAAAAnswer(msg dnsMessage, name, ip string) error {
	want := net.ParseIP(ip).To16()
	if want == nil {
		return fmt.Errorf("invalid IPv6 address %q", ip)
	}
	for _, answer := range msg.Answers {
		if normalizeDNSName(answer.Name) == normalizeDNSName(name) && answer.Type == dnsTypeAAAA && answer.Class == dnsClassIN && string(answer.Data) == string(want) {
			return nil
		}
	}
	return fmt.Errorf("missing AAAA answer for %s -> %s", name, ip)
}

func countAAnswers(msg dnsMessage, name, ip string) int {
	var want []byte
	if strings.TrimSpace(ip) != "" {
		parsed := net.ParseIP(ip).To4()
		if parsed == nil {
			return 0
		}
		want = []byte(parsed)
	}
	count := 0
	for _, answer := range msg.Answers {
		if normalizeDNSName(answer.Name) != normalizeDNSName(name) || answer.Type != dnsTypeA || answer.Class != dnsClassIN {
			continue
		}
		if want != nil && string(answer.Data) != string(want) {
			continue
		}
		count++
	}
	return count
}

func startAnsweringDNSServer(name string, answer dnsRecord) (*fakeDNSServer, error) {
	return startFakeDNSServer(func(req dnsMessage) []byte {
		flags := uint16(dnsFlagQR | dnsFlagRA)
		answers := []dnsRecord(nil)
		if len(req.Questions) > 0 && normalizeDNSName(req.Questions[0].Name) == normalizeDNSName(name) {
			answers = []dnsRecord{answer}
		} else {
			flags |= dnsRCodeNXDomain
		}
		return buildDNSResponse(req, flags, answers, nil, nil)
	})
}

func dnsUpstreamEnv(addr string) []string {
	return []string{
		"DNS_UPSTREAM_ADDR=" + addr,
		"UPSTREAM_DNS_ADDR=" + addr,
		"COURSE_DNS_UPSTREAM_ADDR=" + addr,
	}
}
