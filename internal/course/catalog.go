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

func validateHeaderValue(name, got, want string) error {
	if got != want {
		return fmt.Errorf("%s mismatch: got %q want %q", name, got, want)
	}
	return nil
}
