# Magister

Magister is a local course runner for black-box, implementation-driven systems projects.

It is designed for courses where:

- lessons define observable behavior only
- you keep full control over architecture and code structure
- tests are run from the outside against your program
- progress is stored per lesson as `passed`, `failed`, or `not attempted`

The first course included here is:

- `http-server` — Build Your Own HTTP Server

The codebase is generic enough to add more courses later, such as Redis or SQLite tracks, without redesigning the app.

## What Magister provides

- a Go CLI for browsing courses and running lesson tests
- a local web UI for course browsing and progress tracking
- persistent runner settings and lesson status in `.magister/state.json`
- course definitions that describe behavior, not solutions

## CLI

List courses:

```bash
go run ./cmd/magister courses
```

List lessons for the HTTP course:

```bash
go run ./cmd/magister lessons http-server
```

Show a lesson:

```bash
go run ./cmd/magister show http-server http-01
```

Run one test:

```bash
go run ./cmd/magister test http-server tcp-01 --cmd "go run server.go"
```

Run all tests:

```bash
go run ./cmd/magister test http-server all --cmd "go run server.go"
```

You can override connection details when needed:

```bash
go run ./cmd/magister test http-server http-07 \
  --cmd "go run server.go" \
  --host 127.0.0.1 \
  --port 8080 \
  --startup-timeout 10s \
  --request-timeout 3s
```

The runner stores these settings for the course, so you do not need to pass them every time.

## Web UI

Start the local site:

```bash
go run ./cmd/magister serve
```

Then open:

```text
http://127.0.0.1:8787
```

The UI includes:

- a course home page
- a lesson browser
- per-lesson status badges
- saved runner settings
- lesson detail pages with contracts and raw wire examples
- latest test output for each lesson

## Current HTTP course

Course ID: `http-server`

Lessons:

1. `tcp-01` Accept one TCP connection
2. `http-01` Return a minimal HTTP response
3. `http-02` Route root and unknown paths
4. `http-03` Echo one path segment
5. `http-04` Echo the request body
6. `http-05` Read a request header
7. `http-06` Reuse one TCP connection
8. `http-07` Stay responsive while slow clients are connected
9. `http-08` Support `HEAD`
10. `http-09` Reject malformed requests
11. `http-10` Serve files from a directory

Each lesson stays at the black-box level. The course tells you what your server must do, not how to implement it.

## Verification

The core packages and web routes were verified with:

```bash
env GOCACHE=/tmp/go-build go test ./cmd/magister ./internal/magister ./internal/course
```