package course

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	mathrand "math/rand/v2"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"magister/internal/magister"
)

func testTCPAccept(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	conn, err := h.dial()
	if err != nil {
		return wrapServerError("connect to server", err, srv)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
		return err
	}

	buf := make([]byte, 1)
	n, err := conn.Read(buf)
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return nil
	}
	if err == io.EOF {
		return wrapServerError("connection closed immediately after accept", err, srv)
	}
	if err != nil {
		return wrapServerError("read after connect", err, srv)
	}
	if n >= 0 {
		return nil
	}
	return nil
}

func testMinimal200(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	resp, body, err := h.exchange("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("request to /", err, srv)
	}
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return wrapServerError("validate status", err, srv)
	}
	if err := expectContentLength(resp, 0); err != nil {
		return wrapServerError("validate Content-Length", err, srv)
	}
	if len(body) != 0 {
		return wrapServerError("validate body", fmt.Errorf("expected empty body, got %q", string(body)), srv)
	}

	respWrong, _, err := h.exchange("POST / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("request to / with wrong method", err, srv)
	}
	if err := expectStatus(respWrong, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate wrong method status", err, srv)
	}

	return nil
}

func testRouting(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	respRoot, _, err := h.exchange("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("request to /", err, srv)
	}
	if err := expectStatus(respRoot, http.StatusOK); err != nil {
		return wrapServerError("validate root route", err, srv)
	}

	respMissing, _, err := h.exchange("GET /missing HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("request to /missing", err, srv)
	}
	if err := expectStatus(respMissing, http.StatusNotFound); err != nil {
		return wrapServerError("validate missing route", err, srv)
	}

	respWrong, _, err := h.exchange("POST / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("request to / with wrong method", err, srv)
	}
	if err := expectStatus(respWrong, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate wrong method status", err, srv)
	}

	return nil
}

func testEchoPath(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	want := hex.EncodeToString(bytes)
	path := fmt.Sprintf("/echo/%s", want)
	
	resp, body, err := h.exchange(fmt.Sprintf("GET %s HTTP/1.1\r\nHost: localhost\r\n\r\n", path))
	if err != nil {
		return wrapServerError(fmt.Sprintf("request to %s", path), err, srv)
	}
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return wrapServerError("validate status", err, srv)
	}
	if err := expectBody(body, want); err != nil {
		return wrapServerError("validate body", err, srv)
	}
	if err := expectContentLength(resp, len(want)); err != nil {
		return wrapServerError("validate Content-Length", err, srv)
	}
	if err := validateHeaderValue("Content-Type", resp.Header.Get("Content-Type"), "text/plain"); err != nil {
		return wrapServerError("validate Content-Type", err, srv)
	}

	respWrong, _, err := h.exchange(fmt.Sprintf("POST %s HTTP/1.1\r\nHost: localhost\r\n\r\n", path))
	if err != nil {
		return wrapServerError("request with wrong method", err, srv)
	}
	if err := expectStatus(respWrong, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate wrong method status", err, srv)
	}

	return nil
}

func testMirrorBody(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	body := "ping-pong"
	req := "POST /mirror HTTP/1.1\r\nHost: localhost\r\nContent-Length: 9\r\n\r\nping-pong"
	resp, gotBody, err := h.exchange(req)
	if err != nil {
		return wrapServerError("request to /mirror", err, srv)
	}
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return wrapServerError("validate status", err, srv)
	}
	if err := expectBody(gotBody, body); err != nil {
		return wrapServerError("validate echoed body", err, srv)
	}
	if err := expectContentLength(resp, len(body)); err != nil {
		return wrapServerError("validate Content-Length", err, srv)
	}

	return nil
}

func testReadHeader(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	want := "magister-client/1.0"
	req := "GET /headers/user-agent HTTP/1.1\r\nHost: localhost\r\nUser-Agent: magister-client/1.0\r\n\r\n"
	resp, body, err := h.exchange(req)
	if err != nil {
		return wrapServerError("request to /headers/user-agent", err, srv)
	}
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return wrapServerError("validate status", err, srv)
	}
	if err := expectBody(body, want); err != nil {
		return wrapServerError("validate returned header value", err, srv)
	}

	respWrong, _, err := h.exchange("POST /headers/user-agent HTTP/1.1\r\nHost: localhost\r\nUser-Agent: magister-client/1.0\r\n\r\n")
	if err != nil {
		return wrapServerError("request to /headers/user-agent with wrong method", err, srv)
	}
	if err := expectStatus(respWrong, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate wrong method status", err, srv)
	}

	return nil
}

func testKeepAlive(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	conn, err := h.dial()
	if err != nil {
		return wrapServerError("connect to server", err, srv)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(spec.RequestTimeout)); err != nil {
		return wrapServerError("set connection deadline", err, srv)
	}

	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	if _, err := bw.WriteString("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"); err != nil {
		return wrapServerError("write first request", err, srv)
	}
	if err := bw.Flush(); err != nil {
		return wrapServerError("flush first request", err, srv)
	}

	resp1, _, err := readHTTPResponse(br, requestFromRaw("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	if err != nil {
		return wrapServerError("read first response", err, srv)
	}
	if err := expectStatus(resp1, http.StatusOK); err != nil {
		return wrapServerError("validate first response", err, srv)
	}

	if _, err := bw.WriteString("GET /echo/again HTTP/1.1\r\nHost: localhost\r\n\r\n"); err != nil {
		return wrapServerError("write second request", err, srv)
	}
	if err := bw.Flush(); err != nil {
		return wrapServerError("flush second request", err, srv)
	}

	resp2, body2, err := readHTTPResponse(br, requestFromRaw("GET /echo/again HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	if err != nil {
		return wrapServerError("read second response", err, srv)
	}
	if err := expectStatus(resp2, http.StatusOK); err != nil {
		return wrapServerError("validate second response", err, srv)
	}
	if err := expectBody(body2, "again"); err != nil {
		return wrapServerError("validate second response body", err, srv)
	}

	if _, err := bw.WriteString("POST / HTTP/1.1\r\nHost: localhost\r\n\r\n"); err != nil {
		return wrapServerError("write wrong method request", err, srv)
	}
	if err := bw.Flush(); err != nil {
		return wrapServerError("flush wrong method request", err, srv)
	}

	resp3, _, err := readHTTPResponse(br, requestFromRaw("POST / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	if err != nil {
		return wrapServerError("read wrong method response", err, srv)
	}
	if err := expectStatus(resp3, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate wrong method response", err, srv)
	}

	return nil
}

func testResponsivenessUnderSlowClients(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	var slowConns []net.Conn
	for i := 0; i < 3; i++ {
		conn, err := h.dial()
		if err != nil {
			return wrapServerError("open slow client connection", err, srv)
		}
		slowConns = append(slowConns, conn)
		req := "POST /mirror HTTP/1.1\r\nHost: localhost\r\nContent-Length: 4\r\n\r\np"
		if _, err := io.WriteString(conn, req); err != nil {
			return wrapServerError("write partial slow request", err, srv)
		}
	}
	defer func() {
		for _, conn := range slowConns {
			_ = conn.Close()
		}
	}()

	time.Sleep(150 * time.Millisecond)

	resp, body, err := h.exchange("GET /echo/fast HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("serve fast request while slow clients are open", err, srv)
	}
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return wrapServerError("validate fast request status", err, srv)
	}
	if err := expectBody(body, "fast"); err != nil {
		return wrapServerError("validate fast request body", err, srv)
	}

	respWrong, _, err := h.exchange("POST /echo/fast HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("fast request with wrong method", err, srv)
	}
	if err := expectStatus(respWrong, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate wrong method status", err, srv)
	}

	return nil
}

func testHEAD(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	respGet, bodyGet, err := h.exchange("GET /echo/headline HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("GET /echo/headline", err, srv)
	}
	if err := expectStatus(respGet, http.StatusOK); err != nil {
		return wrapServerError("validate GET response", err, srv)
	}
	if err := expectBody(bodyGet, "headline"); err != nil {
		return wrapServerError("validate GET body", err, srv)
	}

	respHead, bodyHead, err := h.exchange("HEAD /echo/headline HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("HEAD /echo/headline", err, srv)
	}
	if err := expectStatus(respHead, http.StatusOK); err != nil {
		return wrapServerError("validate HEAD response", err, srv)
	}
	if err := expectContentLength(respHead, len("headline")); err != nil {
		return wrapServerError("validate HEAD Content-Length", err, srv)
	}
	if len(bodyHead) != 0 {
		return wrapServerError("validate HEAD body", fmt.Errorf("expected empty body, got %q", string(bodyHead)), srv)
	}

	respWrong, _, err := h.exchange("POST /echo/headline HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("request to /echo/headline with wrong method", err, srv)
	}
	if err := expectStatus(respWrong, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate wrong method status", err, srv)
	}

	return nil
}

func testBadRequest(ctx context.Context, spec magister.RunnerConfig) error {
	h := newHarness(spec)
	srv, err := h.start(ctx, nil)
	if err != nil {
		return err
	}
	defer srv.stop()

	malformedReq := "BROKEN\r\n\r\n"
	validReq := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"

	// Randomize order so learners cannot hardcode "first request = 400"
	if mathrand.IntN(2) == 0 {
		// Malformed first, then valid
		resp, _, err := h.exchange(malformedReq)
		if err != nil {
			return wrapServerError("send malformed request", err, srv)
		}
		if err := expectStatus(resp, http.StatusBadRequest); err != nil {
			return wrapServerError("validate malformed request handling", err, srv)
		}
		respOK, _, err := h.exchange(validReq)
		if err != nil {
			return wrapServerError("follow-up request after malformed input", err, srv)
		}
		if err := expectStatus(respOK, http.StatusOK); err != nil {
			return wrapServerError("validate follow-up request", err, srv)
		}
	} else {
		// Valid first, then malformed
		respOK, _, err := h.exchange(validReq)
		if err != nil {
			return wrapServerError("send valid request", err, srv)
		}
		if err := expectStatus(respOK, http.StatusOK); err != nil {
			return wrapServerError("validate valid request", err, srv)
		}
		resp, _, err := h.exchange(malformedReq)
		if err != nil {
			return wrapServerError("send malformed request", err, srv)
		}
		if err := expectStatus(resp, http.StatusBadRequest); err != nil {
			return wrapServerError("validate malformed request handling", err, srv)
		}
	}

	respWrong, _, err := h.exchange("UNKNOWNMETHOD / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("request with unknown method", err, srv)
	}
	if err := expectStatus(respWrong, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate unknown method status", err, srv)
	}

	return nil
}

func testStaticFiles(ctx context.Context, spec magister.RunnerConfig) error {
	dir, err := os.MkdirTemp("", "http-course-files-*")
	if err != nil {
		return err
	}
	dir = filepath.Clean(dir)
	defer os.RemoveAll(dir)

	filename := "hello.txt"
	contents := "served-from-disk"
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(contents), 0o644); err != nil {
		return err
	}

	h := newHarness(spec)
	srv, err := h.start(ctx, []string{
		fmt.Sprintf("FILES_ROOT=%s", dir),
		fmt.Sprintf("COURSE_FILES_ROOT=%s", dir),
	})
	if err != nil {
		return err
	}
	defer srv.stop()

	resp, body, err := h.exchange(fmt.Sprintf("GET /files/%s HTTP/1.1\r\nHost: localhost\r\n\r\n", filename))
	if err != nil {
		return wrapServerError("request existing file", err, srv)
	}
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return wrapServerError("validate existing file status", err, srv)
	}
	if err := expectBody(body, contents); err != nil {
		return wrapServerError("validate existing file body", err, srv)
	}
	if err := expectContentLength(resp, len(contents)); err != nil {
		return wrapServerError("validate existing file Content-Length", err, srv)
	}

	respMissing, _, err := h.exchange("GET /files/missing.txt HTTP/1.1\r\nHost: localhost\r\n\r\n")
	if err != nil {
		return wrapServerError("request missing file", err, srv)
	}
	if err := expectStatus(respMissing, http.StatusNotFound); err != nil {
		return wrapServerError("validate missing file status", err, srv)
	}

	respWrong, _, err := h.exchange(fmt.Sprintf("POST /files/%s HTTP/1.1\r\nHost: localhost\r\n\r\n", filename))
	if err != nil {
		return wrapServerError("request existing file with wrong method", err, srv)
	}
	if err := expectStatus(respWrong, http.StatusMethodNotAllowed); err != nil {
		return wrapServerError("validate wrong method status", err, srv)
	}

	return nil
}
