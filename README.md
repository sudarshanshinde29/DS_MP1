## MP1: Distributed Grep (Coordinator + Workers)

### Overview
- Three gRPC workers scan local log files and stream results to a coordinator.
- The coordinator aggregates results in either lines mode or count mode and prints per-worker output plus a final total when counting.

### Prerequisites
~~- Go installed (per repo `go.mod`).
- Logs present:
  - `logs/VM1.logs/machine.*.log`
  - `logs/VM2.logs/machine.*.log`
  - `logs/VM3.logs/machine.*.log`

### Project layout (key paths)
- Coordinator: `coordinator/main.go`~~
- Worker: `worker/main.go`
- Properties: `cluster.properties`
- Logs: `logs/VM{1,2,3}.logs/`
- Protobuf builds: `protoBuilds/`

### Configure cluster
The coordinator reads workers from `cluster.properties`:
```properties
no.of.machines=3
peer.machine.ip0=127.0.0.1
peer.machine.port0=6001
peer.machine.name0=vm1
peer.machine.ip1=127.0.0.1
peer.machine.port1=6002
peer.machine.name1=vm2
peer.machine.ip2=127.0.0.1
peer.machine.port2=6003
peer.machine.name2=vm3
```
- Indices start at 0 and go up to `no.of.machines - 1`.
- Names are labels for printing.

### Start the workers (3 terminals)
Run each in its own terminal so you can see logs. Use a glob that matches your files (e.g., `machine.*.log`).

Terminal 1:
```bash
cd "/Users/sunvegnalwar/Fall 25 UIUC/CS 425/MP1/DS_MP1"
go run ./worker/main.go -addr :6001 -logdir ./logs/VM1.logs -glob "machine.*.log" -label vm1 2>&1 | cat
```

Terminal 2:
```bash
cd "/Users/sunvegnalwar/Fall 25 UIUC/CS 425/MP1/DS_MP1"
go run ./worker/main.go -addr :6002 -logdir ./logs/VM2.logs -glob "machine.*.log" -label vm2 2>&1 | cat
```

Terminal 3:
```bash
cd "/Users/sunvegnalwar/Fall 25 UIUC/CS 425/MP1/DS_MP1"
go run ./worker/main.go -addr :6003 -logdir ./logs/VM3.logs -glob "machine.*.log" -label vm3 2>&1 | cat
```

Notes:
- If you see “address already in use”, free the port (see Troubleshooting).
- Workers print matched files and the exact grep command they execute.

### Run the coordinator
Count mode (case-insensitive for “error”):
```bash
cd "/Users/sunvegnalwar/Fall 25 UIUC/CS 425/MP1/DS_MP1"
go run ./coordinator/main.go -props cluster.properties -mode count -- -i -e "error"
```

Lines mode (stream matching lines):
```bash
go run ./coordinator/main.go -props cluster.properties -mode lines -- -i -e "error"
```

What you’ll see:
- Coordinator prints loaded properties and computed targets/labels.
- In count mode, each worker prints its count and the coordinator prints a TOTAL.
- In lines mode, the coordinator prints matching lines with source filename and worker label.

### Grep options (passed through)
Add standard grep flags after `--`. Examples:
- Case-insensitive single pattern:
  ```bash
  -- -i -e "error"
  ```
- Multiple patterns:
  ```bash
  -- -i -E "(error|warn)"
  ```
- Fixed string match:
  ```bash
  -- -F "ERROR"
  ```

### Troubleshooting
- Ports in use:
  ```bash
  lsof -nP -iTCP:6001-6003 -sTCP:LISTEN
  kill <PID>              # or kill -9 <PID>
  # Or bulk kill by command line:
  pkill -f "/Users/sunvegnalwar/Fall 25 UIUC/CS 425/MP1/DS_MP1/worker/main.go"
  ```
- No matches but logs contain hits:
  - Verify glob matches expected files (e.g., `machine.*.log`, not `machine..log`).
  - Run grep locally to confirm:
    ```bash
    /usr/bin/grep -H -c -i -e error logs/VM1.logs/machine.1.log
    ```
  - Watch worker stderr; it prints matched files and the exact grep command.
- Properties loaded but no connections:
  - Ensure workers are running and listening on the ports in `cluster.properties`.

### Common globs
- All `.log` files: `-glob "*.log"`
- Specific app/date prefix: `-glob "app-2025-09-*.log"`

### Clean shutdown
- Press Ctrl+C in each worker terminal to stop.
- Coordinator exits when done; Ctrl+C to stop early.

### Example end-to-end
- Start 3 workers (6001, 6002, 6003) with `-glob "machine.*.log"`.
- Run the coordinator in count mode:
  ```bash
  go run ./coordinator/main.go -props cluster.properties -mode count -- -i -e "error"
  ```
- Expect per-worker counts and a nonzero `TOTAL` if logs contain “error”.


