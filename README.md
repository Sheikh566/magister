# Magister

Magister is a local, implementation-driven learning environment for building systems. 

It is designed for developers who want to learn how complex systems work under the hood by building them from scratch. Magister provides the curriculum and an automated black-box testing engine, while you provide the code in whatever language or architecture you prefer.

> ⚠️ **Disclaimer:** This project was completely *vibe-coded* (AI-generated) and is currently in its initial phase. You may encounter bugs, edge cases in tests, or rough edges in the UI. Please open an issue if you run into any problems!

## Included Courses

* **Build Your Own HTTP Server** (`http-server`)
  * Start with raw TCP sockets and progressively implement routing, header parsing, connection keep-alive, and concurrency. 
  * *Stuck? Check out a complete solution written in Go: [magister-http-server.go](https://gist.github.com/Sheikh566/96aa020ac19a4258779ea118213ca1bd).*

**Looking for Contributors:** The Magister engine is designed to be completely generic. I am actively looking for contributors to help add more courses (e.g., *Build Your Own Redis*, *Build Your Own SQLite*, *Build Your Own Git*). If you're interested in writing test harnesses and markdown lessons, please reach out!

---

## How it works

1. **Pick a course:** Browse the available courses and read the lesson requirements.
2. **Write your code:** Build the target application in your preferred language (Python, Go, Rust, Node, etc.).
3. **Run the tests:** Tell Magister how to start your server. Magister will boot your process, fire real network requests against it, and validate the responses to ensure they meet the protocol specification.

## Installation

### Download Pre-compiled Binaries (Recommended)

Magister is distributed as a single, standalone binary. You don't need Go or any other dependencies installed to run it.

1. Go to the [Releases page](https://github.com/sheikh566/magister/releases).
2. Download the archive for your operating system and architecture.
3. Extract the archive.
4. Move the `magister` binary to a folder in your `$PATH` (e.g., `/usr/local/bin` on Mac/Linux).

### Install via Go
If you already have Go 1.25+ installed on your system, you can build and install the latest version directly:
```bash
go install github.com/sheikh566/magister/cmd/magister@latest
```

---

## Usage

Magister provides both a modern Web UI and a fast CLI. Your progress is saved locally in a `.magister/state.json` file in whatever directory you run the command from.

### Web UI

The easiest way to use Magister is through the web interface. Run this command inside the directory where you are writing your code:

```bash
magister serve
```

Then open `http://127.0.0.1:8787` in your browser. From here, you can read the lessons, configure your runner (e.g., tell it to run `python3 server.py`), and execute tests with a click of a button.

### CLI

If you prefer staying in the terminal, the CLI is fully featured:

**List courses:**
```bash
magister courses
```

**List lessons for a course:**
```bash
magister lessons http-server
```

**Show a specific lesson's requirements:**
```bash
magister show http-server http-01
```

**Run the test for a specific lesson:**
```bash
magister test http-server tcp-01 --cmd "python3 my_server.py"
```

**Run all tests in a course at once:**
```bash
magister test http-server all --cmd "python3 my_server.py"
```

*Note: Once you pass `--cmd` to the test runner once, Magister saves it to your local state. You don't need to pass it on subsequent runs unless your command changes.*

## Local Development
To run the project locally from source:
```bash
git clone https://github.com/sheikh566/magister.git
cd magister
go run ./cmd/magister serve
```

To run the internal verification tests:
```bash
go test ./...
```