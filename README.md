# userflux

A load testing tool for generating HTTP traffic and analyzing system performance.

## Logging

userflux includes production-grade logging that writes to both a file and stderr:

- **Log File**: `./logs/userflux.log` (created automatically if missing)
- **Format**: Timestamped messages with level prefixes (INFO/ERROR) and file/line information
- **Behavior**: Logs are appended across runs, so the file grows over time
- **Console Output**: During development, logs are also visible on stderr for immediate feedback

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
