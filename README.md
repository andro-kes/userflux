# userflux

A load testing tool for generating HTTP traffic and analyzing system performance.

## Overview

userflux is a local-only load testing framework that generates HTTP load through a Gateway. It supports configurable test scenarios via YAML scripts, random delay strategies, and automatic result collection.

## Quick Start

```bash
# Run a load test with a script
./loadtest <script_name>

# Example
./loadtest register.yml
```

## Project Structure

```
userflux/
├── cmd/main.go              # Entry point
├── internal/
│   ├── agent/               # Load generator (vUsers, HTTP requests)
│   ├── logging/             # Production-grade logger
│   ├── orchestrator/        # Script parsing and session management
│   └── session/             # Data models and session state
├── scripts/                 # YAML test scenarios
├── sessions/                # Session metadata (auto-created)
├── results/                 # Test results (auto-created)
└── logs/                    # Log files (auto-created)
```

## Writing Test Scripts

Test scripts are YAML files placed in the `scripts/` directory:

```yaml
config:
  users: 1                   # Number of virtual users (planned)
  time: "100s"               # Test duration

script:
  name: "register"
  flow:
    - name: "health"
      url: "http://localhost:8081"
      request:
        method: "GET"
        path: "/health"
        headers:
          accept: "application/json"
    - name: "register"
      url: "http://localhost:8081"
      body: [username, password]  # Auto-generated random values
      request:
        method: "POST"
        path: "/auth/register"
        headers:
          accept: "application/json"
          content-type: "application/json"
```

## Features

- **Random delay strategies**: Supports uniform, exponential, and normal distributions for realistic load patterns
- **Automatic body generation**: Generates random values for specified body fields
- **Concurrent execution**: Spawns virtual users as goroutines with configurable timing
- **Result tracking**: Counts successful and failed requests per session

## Logging

userflux includes production-grade logging that writes to both a file and stderr:

- **Log File**: `./logs/userflux.log` (created automatically if missing)
- **Format**: Timestamped messages with level prefixes (INFO/ERROR) and file/line information
- **Behavior**: Log file is **truncated** on each run (fresh log per execution)
- **Console Output**: Logs are also visible on stderr for immediate feedback

The log file captures:
- Application startup and shutdown
- Script selection and parsing
- Session and result file creation
- Agent execution and user goroutine lifecycle
- HTTP request execution details
- Errors and warnings throughout the execution

To monitor logs in real-time:
```bash
tail -f ./logs/userflux.log
```

## Output Files

After each run, userflux creates:

- `sessions/session_<N>` — Copy of the executed script configuration
- `results/result_<N>` — JSON summary with success/failure counts

Example result:
```json
{
  "Script": "register",
  "Total": 150,
  "Success": 148,
  "Failure": 2
}
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed architecture diagrams and design principles.

## Requirements

- Go 1.24.2+
- Target HTTP service running (e.g., Gateway on `localhost:8081`)