package course

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"magister/internal/magister"
)

type Harness struct {
	spec magister.RunnerConfig
}

func newHarness(spec magister.RunnerConfig) *Harness {
	return &Harness{spec: spec}
}

type managedServer struct {
	spec   magister.RunnerConfig
	cmd    *exec.Cmd
	output bytes.Buffer
	done   chan error
}

func (h *Harness) start(ctx context.Context, extraEnv []string) (*managedServer, error) {
	shell, shellArgs := shellCommand(h.spec.Command)
	cmd := exec.CommandContext(ctx, shell, shellArgs...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", h.spec.Port),
		fmt.Sprintf("HTTP_SERVER_PORT=%d", h.spec.Port),
		fmt.Sprintf("COURSE_PORT=%d", h.spec.Port),
	)
	cmd.Env = append(cmd.Env, extraEnv...)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	srv := &managedServer{
		spec: h.spec,
		cmd:  cmd,
		done: make(chan error, 1),
	}
	cmd.Stdout = &srv.output
	cmd.Stderr = &srv.output

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}

	go func() {
		srv.done <- cmd.Wait()
	}()

	if err := srv.waitForReady(ctx); err != nil {
		_ = srv.stop()
		return nil, err
	}

	return srv, nil
}

func (s *managedServer) waitForReady(ctx context.Context) error {
	deadline := time.NewTimer(s.spec.StartupTimeout)
	defer deadline.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server startup canceled: %w\n\n%s", ctx.Err(), s.output.String())
		case err := <-s.done:
			if err == nil {
				return fmt.Errorf("server exited before becoming ready\n\n%s", s.output.String())
			}
			return fmt.Errorf("server exited before becoming ready: %w\n\n%s", err, s.output.String())
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for %s to accept connections on %s\n\n%s", s.spec.Command, s.addr(), s.output.String())
		default:
		}

		conn, err := net.DialTimeout("tcp", s.addr(), 150*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (s *managedServer) stop() error {
	if s.cmd.Process == nil {
		return nil
	}
	if s.cmd.ProcessState != nil && s.cmd.ProcessState.Exited() {
		return nil
	}

	if runtime.GOOS != "windows" {
		if s.cmd.Process != nil {
			_ = s.cmd.Process.Kill()
		}
	} else {
		if s.cmd.Process != nil {
			_ = s.cmd.Process.Kill()
		}
	}

	select {
	case <-time.After(500 * time.Millisecond):
		if s.cmd.Process != nil {
			_ = s.cmd.Process.Kill()
		}
		select {
		case <-time.After(500 * time.Millisecond):
			return fmt.Errorf("timed out waiting for server process to stop")
		case <-s.done:
		}
	case <-s.done:
	}

	return nil
}

func (s *managedServer) addr() string {
	return fmt.Sprintf("%s:%d", s.spec.Host, s.spec.Port)
}

func (s *managedServer) outputString() string {
	return strings.TrimSpace(s.output.String())
}

func (h *Harness) dial() (net.Conn, error) {
	return net.DialTimeout("tcp", fmt.Sprintf("%s:%d", h.spec.Host, h.spec.Port), h.spec.RequestTimeout)
}

func (h *Harness) exchange(raw string) (*http.Response, []byte, error) {
	conn, err := h.dial()
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(h.spec.RequestTimeout)); err != nil {
		return nil, nil, err
	}

	if _, err := io.WriteString(conn, raw); err != nil {
		return nil, nil, err
	}

	br := bufio.NewReader(conn)
	req := requestFromRaw(raw)
	return readHTTPResponse(br, req)
}

// requestFromRaw extracts the HTTP method from a raw request line to build
// an http.Request. This is needed so http.ReadResponse knows not to read
// a body for HEAD responses (which have Content-Length but 0 body bytes).
func requestFromRaw(raw string) *http.Request {
	firstLine := strings.SplitN(raw, "\r\n", 2)[0]
	parts := strings.SplitN(firstLine, " ", 2)
	method := "GET"
	if len(parts) >= 1 {
		method = strings.TrimSpace(parts[0])
	}
	if method == "" {
		method = "GET"
	}
	req, _ := http.NewRequest(method, "/", nil)
	return req
}

func readHTTPResponse(br *bufio.Reader, req *http.Request) (*http.Response, []byte, error) {
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

func expectStatus(resp *http.Response, want int) error {
	if resp.StatusCode != want {
		return fmt.Errorf("status mismatch: got %d want %d", resp.StatusCode, want)
	}
	return nil
}

func expectBody(body []byte, want string) error {
	if string(body) != want {
		return fmt.Errorf("body mismatch: got %q want %q", string(body), want)
	}
	return nil
}

func expectContentLength(resp *http.Response, want int) error {
	got := resp.Header.Get("Content-Length")
	if got != fmt.Sprintf("%d", want) {
		return fmt.Errorf("Content-Length mismatch: got %q want %d", got, want)
	}
	return nil
}

func wrapServerError(prefix string, err error, srv *managedServer) error {
	if err == nil {
		return nil
	}

	var builder strings.Builder
	builder.WriteString(prefix)
	builder.WriteString(": ")
	builder.WriteString(err.Error())

	if srv != nil {
		out := srv.outputString()
		if out != "" && !errors.Is(err, context.DeadlineExceeded) {
			builder.WriteString("\n\nserver output:\n")
			builder.WriteString(out)
		}
	}

	return errors.New(builder.String())
}

func shellCommand(command string) (string, []string) {
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		shell = "sh"
	}
	// Do not use "exec" prefix: commands like "cd /path && go run ." would fail
	// because exec tries to run "cd" (a shell builtin) as an executable.
	return shell, []string{"-c", command}
}
