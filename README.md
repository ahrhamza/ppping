# ppping

A lightweight, portable Windows CLI tool for probing TCP and UDP port connectivity with latency measurement. Think of it as `ping`, but for ports.

No installation required — single standalone `.exe` with zero runtime dependencies.

## Quick Start

```
ppping.exe <host> <port> [proto] [count]
```

| Argument | Required | Default | Description                    |
|----------|----------|---------|--------------------------------|
| `host`   | Yes      | —       | IP address or FQDN to probe   |
| `port`   | Yes      | —       | Port number (1–65535)          |
| `proto`  | No       | `tcp`   | Protocol: `tcp` or `udp`      |
| `count`  | No       | `4`     | Number of probe attempts       |

## Examples

**Probe RDP on a known IP (TCP, 4 attempts):**

```
> ppping.exe 172.26.104.10 3389

Probing 172.26.104.10:3389 (tcp) x4
  Attempt 1: Success  12.4ms
  Attempt 2: Success  11.2ms
  Attempt 3: Success  13.1ms
  Attempt 4: Success  11.8ms
  Summary: 4/4 succeeded, avg 12.1ms
```

**Probe HTTPS on an FQDN (resolves and probes each IP):**

```
> ppping.exe myapp.internal.com 443

Resolved myapp.internal.com -> 3 addresses
  [1] 172.26.104.10
  [2] 172.26.104.11
  [3] 172.26.104.12

Probing 172.26.104.10:443 (tcp) x4
  Attempt 1: Success  12.4ms
  Attempt 2: Success  11.2ms
  Attempt 3: Success  13.1ms
  Attempt 4: Success  11.8ms
  Summary: 4/4 succeeded, avg 12.1ms

Probing 172.26.104.11:443 (tcp) x4
  Attempt 1: Success  14.0ms
  Attempt 2: Success  13.5ms
  Attempt 3: Failed   timeout
  Attempt 4: Success  14.2ms
  Summary: 3/4 succeeded, avg 13.9ms

Probing 172.26.104.12:443 (tcp) x4
  Attempt 1: Failed   connection refused
  Attempt 2: Failed   connection refused
  Attempt 3: Failed   connection refused
  Attempt 4: Failed   connection refused
  Summary: 0/4 succeeded
```

**UDP probe with custom count:**

```
> ppping.exe 172.26.104.10 4433 udp 10

Probing 172.26.104.10:4433 (udp) x10
  Attempt 1: Success  5001.2ms
  Attempt 2: Success  5000.8ms
  ...
  Summary: 10/10 succeeded, avg 5001.0ms
```

## How It Works

### TCP Mode

Each attempt performs a full TCP three-way handshake (`SYN → SYN-ACK → ACK`) and immediately closes the connection. The reported latency is the time from initiating the connection to completing the handshake.

### UDP Mode

Each attempt sends a single-byte probe packet and waits for a response. Since UDP is connectionless:

- **Response received** → port is open, latency is measured
- **Timeout (no response)** → reported as success, but the port may be open (silent), filtered, or the service simply doesn't reply to unknown packets
- **ICMP port unreachable** → reported as a failure, indicating the port is closed

### DNS Resolution

When the host is an FQDN (not an IP address), ppping resolves it via the system DNS resolver and probes **every** resolved IP independently. Each IP gets its own full set of attempts and its own summary line. This is useful for verifying connectivity to all backends behind a DNS record.

### Timing

- **1 second** between consecutive attempts
- **5 second** connection timeout per attempt

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

| Error                              | Meaning                                                  |
|------------------------------------|----------------------------------------------------------|
| `connection refused`               | Host is reachable but nothing is listening on that port   |
| `timeout`                          | No response within 5 seconds — host or port is filtered  |
| `host unreachable` / `no route`    | Network-level connectivity issue                         |
| `could not resolve <host>`         | DNS lookup failed                                        |

## License

This project is provided as-is with no warranty. Use at your own discretion.
