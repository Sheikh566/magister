package course

import (
	"fmt"
	"strings"

	"magister/internal/magister"
)

func HTTPServerCourse() magister.Course {
	return magister.Course{
		ID:          "http-server",
		Title:       "Build Your Own HTTP Server",
		Summary:     "Master the web's fundamental protocol by building a fully functional HTTP/1.1 server from scratch.",
		Description: "Welcome to the ultimate challenge. Starting with nothing but raw TCP sockets, you will implement routing, header parsing, connection keep-alive, and concurrency. Build the server using any programming language or architecture you prefer, and the Magister testing engine will validate its behavior step-by-step.",
		Lessons: []magister.Lesson{
			{
				ID:      "tcp-01",
				Chapter: "1. TCP Foundations",
				Title:   "Accept One TCP Connection",
				Summary: "Open a TCP listener on the configured port and accept an inbound connection without immediately tearing it down.",
				Objectives: []string{
					"Listen on the port provided by the environment.",
					"Allow a client from localhost to complete a TCP handshake.",
					"Keep the connection alive long enough for the client to observe that it was accepted.",
				},
				TestFocus: []string{
					"The tester opens a raw TCP connection to your process.",
					"It fails if the port is closed or if the connection is dropped immediately.",
				},
				Run: testTCPAccept,
			},
			{
				ID:      "http-01",
				Chapter: "1. TCP Foundations",
				Title:   "Return a Minimal HTTP Response",
				Summary: "When a well-formed request arrives, answer with a valid `200 OK` HTTP/1.1 response and an empty body.",
				Objectives: []string{
					"Recognize the end of a basic HTTP request.",
					"Write a parseable HTTP/1.1 response.",
					"Return status `200 OK` with `Content-Length: 0`.",
				},
				TestFocus: []string{
					"The tester sends `GET / HTTP/1.1` and parses your response as HTTP.",
					"It checks the status code, the body length, and whether the response framing is valid.",
				},
				WireText: strings.TrimSpace(`
Raw HTTP/1.1 request bytes on the TCP connection:

GET / HTTP/1.1
Host: localhost


One valid minimal HTTP/1.1 response shape:

HTTP/1.1 200 OK
Content-Length: 0


On the wire, each visible line ends with CRLF, and the blank line between headers and body is required.
For this lesson the body is empty, so the response ends immediately after that blank line.
`),
				Run: testMinimal200,
			},
			{
				ID:      "http-02",
				Chapter: "2. HTTP Message Handling",
				Title:   "Route Root and Unknown Paths",
				Summary: "Different paths should now produce different outcomes: root succeeds, unknown paths return not found.",
				Objectives: []string{
					"`GET /` returns `200 OK`.",
					"`GET` for an unknown path returns `404 Not Found`.",
					"Both responses remain valid HTTP/1.1 messages.",
				},
				TestFocus: []string{
					"The tester makes separate requests to `/` and to an unknown path.",
					"It checks that the two routes produce distinct status codes.",
				},
				Run: testRouting,
			},
			{
				ID:      "http-03",
				Chapter: "2. HTTP Message Handling",
				Title:   "Echo One Path Segment",
				Summary: "A route with a variable final segment should return that segment as the response body.",
				Objectives: []string{
					"`GET /echo/<text>` returns `200 OK`.",
					"The body is exactly `<text>`.",
					"`Content-Length` matches the echoed bytes.",
					"`Content-Type` is `text/plain`.",
				},
				TestFocus: []string{
					"The tester generates a random string and calls `/echo/<random-string>`.",
					"It checks body bytes and response headers only, not your internal routing design.",
				},
				Run: testEchoPath,
			},
			{
				ID:      "http-04",
				Chapter: "2. HTTP Message Handling",
				Title:   "Echo the Request Body",
				Summary: "Read a request body and return those exact bytes in the response.",
				Objectives: []string{
					"`POST /mirror` returns `200 OK`.",
					"The response body exactly matches the request body.",
					"`Content-Length` matches the echoed body size.",
				},
				TestFocus: []string{
					"The tester sends a short body with a valid `Content-Length`.",
					"It checks that the returned bytes are unchanged.",
				},
				Run: testMirrorBody,
			},
			{
				ID:      "http-05",
				Chapter: "2. HTTP Message Handling",
				Title:   "Read a Request Header",
				Summary: "Look up a request header named in the path and return its value in the response body.",
				Objectives: []string{
					"`GET /headers/<name>` returns the value of request header `<name>`.",
					"Header lookup is case-insensitive.",
					"The response body contains only the header value.",
				},
				TestFocus: []string{
					"The tester sends `User-Agent: blackbox-client/1.0`.",
					"It requests `/headers/user-agent` and expects the exact header value back.",
				},
				Run: testReadHeader,
			},
			{
				ID:      "http-06",
				Chapter: "3. Real Server Behavior",
				Title:   "Reuse One TCP Connection",
				Summary: "Support two sequential HTTP/1.1 requests over the same already-open TCP connection.",
				Objectives: []string{
					"Do not require a brand-new TCP connection for every request.",
					"After a successful first response, allow a second request on the same socket.",
					"Return valid responses for both requests.",
				},
				TestFocus: []string{
					"The tester writes one request, reads the response, then writes a second request on the same connection.",
					"It fails if the server closes the socket after the first response.",
				},
				Run: testKeepAlive,
			},
			{
				ID:      "http-07",
				Chapter: "3. Real Server Behavior",
				Title:   "Stay Responsive While Slow Clients Are Connected",
				Summary: "A few slow or incomplete clients should not prevent the server from handling another ready request.",
				Objectives: []string{
					"Accept multiple clients at the same time.",
					"Do not let one partially-sent request block all other work.",
					"Continue serving a normal request while other connections are still incomplete.",
				},
				TestFocus: []string{
					"The tester opens several slow `POST /mirror` connections and sends only part of each body.",
					"It then sends a normal `GET /echo/fast` and expects a timely response.",
				},
				Run: testResponsivenessUnderSlowClients,
			},
			{
				ID:      "http-08",
				Chapter: "3. Real Server Behavior",
				Title:   "Support HEAD",
				Summary: "A `HEAD` request should describe the same resource as `GET` without including the body bytes.",
				Objectives: []string{
					"`HEAD /echo/<text>` returns the same status code as `GET` for that route.",
					"`Content-Length` reflects the size of the matching `GET` body.",
					"The `HEAD` response body is empty.",
				},
				TestFocus: []string{
					"The tester compares the `GET` and `HEAD` responses for the same resource.",
					"It checks headers and the absence of a response body.",
				},
				Run: testHEAD,
			},
			{
				ID:      "http-09",
				Chapter: "3. Real Server Behavior",
				Title:   "Reject Malformed Requests",
				Summary: "Bad input should return `400 Bad Request` instead of crashing the process or hanging forever.",
				Objectives: []string{
					"A malformed request line returns status `400 Bad Request`.",
					"The server stays alive after handling malformed input.",
					"A later well-formed request still succeeds.",
				},
				TestFocus: []string{
					"The tester sends malformed and valid requests in random order.",
					"It verifies that malformed input returns 400 and that the process stays alive for subsequent requests.",
				},
				Run: testBadRequest,
			},
			{
				ID:      "http-10",
				Chapter: "3. Real Server Behavior",
				Title:   "Serve Files From a Directory",
				Summary: "Expose file bytes over HTTP by reading from a configurable directory on disk.",
				Objectives: []string{
					"`GET /files/<name>` returns the file contents from the configured root directory.",
					"Missing files return `404 Not Found`.",
					"`Content-Length` matches the number of bytes served.",
				},
				TestFocus: []string{
					"The tester creates a temporary directory and passes it to your process as `FILES_ROOT` and `COURSE_FILES_ROOT`.",
					"It requests one real file and one missing file.",
				},
				Run: testStaticFiles,
			},
		},
	}
}

func DNSServerCourse() magister.Course {
	return magister.Course{
		ID:          "dns-server",
		Title:       "Build Your Own DNS Server",
		Summary:     "Learn DNS packet structure, UDP servers, forwarding, recursion, caching, IPv6, and TCP fallback by building a resolver from scratch.",
		Description: "Build a DNS server one protocol feature at a time. Starting with UDP datagrams and DNS headers, you will encode and parse questions, answer records, follow upstream and authoritative nameservers, cache records by TTL, support IPv6, and handle TCP fallback for large responses.",
		Lessons: []magister.Lesson{
			{
				ID:      "dns-01",
				Chapter: "1. DNS Packet Foundations",
				Title:   "Setup UDP Server",
				Summary: "Open a UDP listener on the configured port and send a response datagram when a DNS query arrives.",
				Objectives: []string{
					"Listen on the port provided by the environment.",
					"Receive DNS queries over UDP from localhost.",
					"Send at least one response datagram without crashing or hanging.",
				},
				TestFocus: []string{
					"The tester sends a well-formed DNS query over UDP.",
					"It fails if no response datagram is received before the timeout.",
				},
				Run: testDNSUDPServer,
			},
			{
				ID:      "dns-02",
				Chapter: "1. DNS Packet Foundations",
				Title:   "Write Header Section",
				Summary: "Return a valid DNS response header with the same transaction ID and the response bit set.",
				Objectives: []string{
					"Read the 12-byte DNS header from the request.",
					"Copy the request transaction ID into the response.",
					"Set `QR=1` and return `RCODE=0` for a basic query.",
				},
				TestFocus: []string{
					"The tester sends an `A` query with a known transaction ID.",
					"It checks the response ID, `QR` flag, and response code.",
				},
				Run: testDNSHeader,
			},
			{
				ID:      "dns-03",
				Chapter: "1. DNS Packet Foundations",
				Title:   "Write Question Section",
				Summary: "Echo the DNS question section so the response describes the query it is answering.",
				Objectives: []string{
					"Encode domain names as length-prefixed labels.",
					"Copy `QTYPE` and `QCLASS` into the response question.",
					"Keep `QDCOUNT` consistent with the encoded question section.",
				},
				TestFocus: []string{
					"The tester asks for the `A` record for `google.com`.",
					"It parses the response question and checks name, type, and class.",
				},
				Run: testDNSQuestion,
			},
			{
				ID:      "dns-04",
				Chapter: "1. DNS Packet Foundations",
				Title:   "Write Answer Section",
				Summary: "Answer a known `A` query with one IPv4 resource record and a valid TTL.",
				Objectives: []string{
					"Encode a DNS answer resource record.",
					"Return `magister.test` as `203.0.113.10` for `A` queries.",
					"Set `ANCOUNT` and `RDLENGTH` correctly.",
				},
				TestFocus: []string{
					"The tester queries `magister.test` with type `A` and class `IN`.",
					"It expects one matching IPv4 address in the answer section.",
				},
				Run: testDNSAnswer,
			},
			{
				ID:      "dns-05",
				Chapter: "2. DNS Packet Parsing",
				Title:   "Parse Header Section",
				Summary: "Interpret incoming DNS header flags and return a DNS error for unsupported opcodes.",
				Objectives: []string{
					"Parse transaction ID, flags, opcode, and section counts.",
					"Recognize standard queries with `OPCODE=0`.",
					"Return `RCODE=4` for unsupported opcodes.",
				},
				TestFocus: []string{
					"The tester sends a query with an unsupported opcode.",
					"It checks that the response preserves the ID and returns `Not Implemented`.",
				},
				Run: testDNSParseHeader,
			},
			{
				ID:      "dns-06",
				Chapter: "2. DNS Packet Parsing",
				Title:   "Parse Question Section",
				Summary: "Decode incoming names, types, and classes so behavior can depend on the requested record.",
				Objectives: []string{
					"Parse `QNAME`, `QTYPE`, and `QCLASS` from the request.",
					"Answer the known `magister.test` `A` record.",
					"Return `NXDOMAIN` for a missing name.",
				},
				TestFocus: []string{
					"The tester sends one known-name query and one missing-name query.",
					"It expects the known answer and `RCODE=3` for the missing name.",
				},
				Run: testDNSParseQuestion,
			},
			{
				ID:      "dns-07",
				Chapter: "2. DNS Packet Parsing",
				Title:   "Parse Compressed Packet",
				Summary: "Support DNS name compression pointers in incoming packets.",
				Objectives: []string{
					"Decode names that use compression pointers.",
					"Guard against pointer loops and out-of-bounds pointers.",
					"Answer each parsed `A` question for `magister.test`.",
				},
				TestFocus: []string{
					"The tester sends two questions where the second name is a pointer to the first.",
					"It expects both questions to be parsed and answered.",
				},
				Run: testDNSCompressedPacket,
			},
			{
				ID:      "dns-08",
				Chapter: "3. Resolver Behavior",
				Title:   "Forward Queries Upstream",
				Summary: "Forward unknown questions to a configured upstream DNS server and relay the response.",
				Objectives: []string{
					"Read the upstream resolver address from the environment.",
					"Forward queries your local zone cannot answer.",
					"Relay upstream answers while preserving DNS transaction semantics.",
				},
				TestFocus: []string{
					"The tester starts a fake upstream DNS server and passes its address to your process.",
					"It verifies that the upstream receives the query and that your server returns its answer.",
				},
				Run: testDNSForwarding,
			},
			{
				ID:      "dns-09",
				Chapter: "3. Resolver Behavior",
				Title:   "Recursive Resolution",
				Summary: "Resolve a name by following referrals from root hints through authoritative nameservers.",
				Objectives: []string{
					"Read root hints and local authority address hints from the environment.",
					"Follow `NS` referrals toward the authoritative server.",
					"Return the final answer from the authoritative nameserver.",
				},
				TestFocus: []string{
					"The tester starts fake root, TLD, and authoritative DNS servers.",
					"It checks that each server is queried and that the final `A` answer is returned.",
				},
				Run: testDNSRecursiveResolution,
			},
			{
				ID:      "dns-10",
				Chapter: "3. Resolver Behavior",
				Title:   "Cache Answers With TTL",
				Summary: "Cache successful upstream answers and expire them according to their DNS TTL values.",
				Objectives: []string{
					"Store upstream answers by question name, type, and class.",
					"Serve cache hits without contacting the upstream resolver.",
					"Expire cached answers after their TTL has elapsed.",
				},
				TestFocus: []string{
					"The tester returns an upstream answer with a one-second TTL.",
					"It checks for a cache hit before expiry and a new upstream query after expiry.",
				},
				Run: testDNSCachingTTL,
			},
			{
				ID:      "dns-11",
				Chapter: "3. Resolver Behavior",
				Title:   "Support IPv6 and AAAA Records",
				Summary: "Handle IPv6 resource data by answering `AAAA` questions with 16-byte addresses.",
				Objectives: []string{
					"Recognize `QTYPE=AAAA` queries.",
					"Encode IPv6 addresses as 16-byte `RDATA` values.",
					"Return `ipv6.magister.test` as `2001:db8::53`.",
				},
				TestFocus: []string{
					"The tester asks for the `AAAA` record for `ipv6.magister.test`.",
					"It expects an IPv6 address in the answer section.",
				},
				Run: testDNSIPv6AAAA,
			},
			{
				ID:      "dns-12",
				Chapter: "4. Production DNS Features",
				Title:   "Handle TCP Fallback and Truncation",
				Summary: "Set the UDP truncation flag for oversized responses and serve the full response over DNS-over-TCP.",
				Objectives: []string{
					"Detect when a UDP response would exceed the classic 512-byte DNS limit.",
					"Set the `TC` flag in the UDP response for `large.magister.test`.",
					"Accept DNS-over-TCP queries with the two-byte length prefix and return the full answer set.",
				},
				TestFocus: []string{
					"The tester queries `large.magister.test` over UDP and expects `TC=1`.",
					"It retries over TCP and expects the full non-truncated answer set.",
				},
				Run: testDNSTCPFallback,
			},
		},
	}
}

func validateHeaderValue(name, got, want string) error {
	if got != want {
		return fmt.Errorf("%s mismatch: got %q want %q", name, got, want)
	}
	return nil
}
