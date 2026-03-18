# ppping

A lightweight, portable CLI tool for probing TCP and UDP port connectivity with latency measurement. Think of it as `ping`, but for ports.

No installation required — single standalone binary with zero runtime dependencies.

## Quick Start

```
ppping <host> <port> [proto] [count]
```

| Argument | Required | Default | Description                                        |
|----------|----------|---------|----------------------------------------------------|
| `host`   | Yes      | —       | IP address or FQDN to probe                        |
| `port`   | Yes      | —       | Port number (1–65535)                              |
| `proto`  | No       | `tcp`   | Protocol: `tcp` or `udp` (case-insensitive)        |
| `count`  | No       | `4`     | Number of attempts (1–10,000), or `0` for nonstop |

## Examples

**Probe RDP on a known IP (TCP, 4 attempts):**

```
> ppping 172.26.104.10 3389

Probing 172.26.104.10:3389 (tcp) x4
  Attempt 1: Success       12.4ms
  Attempt 2: Success       11.2ms
  Attempt 3: Success       13.1ms
  Attempt 4: Success       11.8ms
  Summary: 4/4 succeeded, avg 12.1ms
```

**Probe HTTPS on an FQDN (finite count — probes each IP in sequence):**

```
> ppping myapp.internal.com 443

Resolved "myapp.internal.com" -> 3 addresses
  [1] 172.26.104.10
  [2] 172.26.104.11
  [3] 172.26.104.12

Probing 172.26.104.10:443 (tcp) x4
  Attempt 1: Success       12.4ms
  ...
  Summary: 4/4 succeeded, avg 12.1ms

Probing 172.26.104.11:443 (tcp) x4
  Attempt 1: Success       14.0ms
  Attempt 2: Success       13.5ms
  Attempt 3: Failed        timeout
  Attempt 4: Success       14.2ms
  Summary: 3/4 succeeded, avg 13.9ms

Probing 172.26.104.12:443 (tcp) x4
  Attempt 1: Failed        connection refused
  ...
  Summary: 0/4 succeeded
```

**Nonstop probing on an FQDN — round-robin across all IPs (runs until Ctrl+C):**

```
> ppping myapp.internal.com 443 0

Resolved "myapp.internal.com" -> 3 addresses
  [1] 172.26.104.10
  [2] 172.26.104.11
  [3] 172.26.104.12

Round-robin nonstop mode: cycling through 3 addresses on port 443 (tcp) — Ctrl+C to stop

  Attempt    1  [1/3]  172.26.104.10:443     Success       12.4ms
  Attempt    2  [2/3]  172.26.104.11:443     Success       13.5ms
  Attempt    3  [3/3]  172.26.104.12:443     Failed        connection refused
  Attempt    4  [1/3]  172.26.104.10:443     Success       12.1ms
  ...
  ^C

Summary:
  [1] 172.26.104.10:443: 2/2 succeeded, avg 12.3ms
  [2] 172.26.104.11:443: 1/1 succeeded, avg 13.5ms
  [3] 172.26.104.12:443: 0/1 succeeded
```

**Nonstop probing on a single IP:**

```
> ppping 172.26.104.10 3389 0

Probing 172.26.104.10:3389 (tcp) nonstop — Ctrl+C to stop
  Attempt 1: Success       12.4ms
  Attempt 2: Success       11.2ms
  ^C
  Summary: 2/2 succeeded, avg 11.8ms
```

**UDP probe:**

```
> ppping 172.26.104.10 4433 udp 4

Probing 172.26.104.10:4433 (udp) x4
  Attempt 1: open|filtered 5001.2ms
  Attempt 2: open|filtered 5000.8ms
  Attempt 3: open|filtered 5001.0ms
  Attempt 4: open|filtered 5000.9ms
  Summary: 4/4 succeeded, avg 5001.0ms
```

## How It Works

### TCP Mode

Each attempt performs a full TCP three-way handshake (`SYN → SYN-ACK → ACK`) and immediately closes the connection. The reported latency is the time from initiating the connection to completing the handshake.

### UDP Mode

Each attempt sends a single-byte probe packet and waits for a response. Since UDP is connectionless:

- **Response received** → reported as `Success`, latency is measured
- **Timeout (no response)** → reported as `open|filtered` — the port may be open (silent), filtered, or the service simply doesn't reply to unknown packets
- **ICMP port unreachable** → reported as `Failed`, indicating the port is closed

Note: `open|filtered` is counted as a successful probe in the summary (the probe itself didn't fail), but it does **not** confirm the remote application replied.

### DNS Resolution

When the host is an FQDN, ppping resolves it via the system DNS resolver and probes every resolved IP:

- **Finite count mode** — each IP is probed sequentially with its own full set of attempts and summary line
- **Nonstop mode (`count=0`)** — all IPs are cycled in round-robin order, one probe per IP per second, with a per-IP summary printed on Ctrl+C

If a hostname resolves to **10 or more addresses**, a warning is printed before probing begins.

If the requested count multiplied by the number of resolved IPs would exceed the **10,000 total ping limit**, the per-address count is automatically capped and a warning is shown.

### Timing

- **1 second** between consecutive attempts
- **5 second** connection timeout per attempt

## Limits

| Limit | Value |
|-------|-------|
| Maximum `count` (single IP) | 10,000 |
| Maximum total pings (multi-IP) | 10,000 across all addresses |
| Nonstop mode max attempts | 10,000 (then stops automatically) |

## Building from Source

Requires [Go](https://go.dev/dl/) 1.24 or later.

**Windows (amd64):**

```
GOOS=windows GOARCH=amd64 go build -o ppping.exe .
```

**Other targets:**

```
# Windows ARM64
GOOS=windows GOARCH=arm64 go build -o ppping.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o ppping .

# macOS
GOOS=darwin GOARCH=arm64 go build -o ppping .
```

The output is always a single standalone binary with no runtime dependencies.

## Error Messages

| Output                             | Meaning                                                        |
|------------------------------------|----------------------------------------------------------------|
| `Success`                          | Connection completed (TCP handshake, or UDP reply received)    |
| `open\|filtered`                   | UDP timeout — no reply, but no ICMP unreachable either         |
| `Failed: connection refused`       | Host is reachable but nothing is listening on that port        |
| `Failed: timeout`                  | No response within 5 seconds — host or port is filtered        |
| `Failed: host unreachable / no route` | Network-level connectivity issue                            |
| `Error: could not resolve <host>`  | DNS lookup failed                                              |

## License

This project is provided as-is with no warranty. Use at your own discretion.
