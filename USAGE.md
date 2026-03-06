# ppping — Usage Guide

## Syntax

```
ppping <host> <port> [proto] [count]
```

All arguments are positional (no flags or dashes).

### Arguments

**`host`** (required)
The target to probe. Accepts:
- IPv4 address: `192.168.1.1`
- IPv6 address: `::1`
- FQDN: `myapp.internal.com`

**`port`** (required)
The port number to probe (1–65535).

**`proto`** (optional, default: `tcp`)
The protocol to use. Must be `tcp` or `udp`.

**`count`** (optional, default: `4`)
The number of probe attempts to make. Must be a non-negative integer. Use `0` for nonstop mode (probes continuously until Ctrl+C).

## Common Use Cases

### Verify a service is listening

```
> ppping.exe 10.0.0.5 443
```

A successful TCP probe confirms something is accepting connections on that port.

### Diagnose intermittent connectivity

```
> ppping.exe 10.0.0.5 3389 tcp 20
```

Run 20 attempts to catch intermittent drops. The summary shows the success rate and average latency.

### Test connectivity to all IPs behind a DNS name

```
> ppping.exe loadbalancer.corp.com 443
```

If the name resolves to multiple IPs, each one is probed separately, making it easy to identify a single unhealthy backend.

### Check if a UDP port is reachable

```
> ppping.exe 10.0.0.5 53 udp
```

UDP probing is inherently less reliable — a "success" with high latency typically means no response was received (open|filtered), not that the service responded.

### Continuous monitoring

```
> ppping.exe 10.0.0.5 443 tcp 0
```

Probes indefinitely until you press Ctrl+C. When interrupted, a final summary is printed with total attempts, success rate, and average latency. Useful for monitoring a service over time or waiting for it to come up.

### Quick single-shot check

```
> ppping.exe 10.0.0.5 22 tcp 1
```

A single attempt for a fast pass/fail check.

## Understanding the Output

### Attempt Lines

```
  Attempt 1: Success  12.4ms     # Connection established in 12.4ms
  Attempt 2: Failed   timeout    # No response within 5 seconds
  Attempt 3: Failed   connection refused  # Port is closed
```

### Summary Line

```
  Summary: 3/4 succeeded, avg 12.1ms
```

- `3/4 succeeded` — 3 out of 4 attempts connected
- `avg 12.1ms` — average latency across successful attempts only (omitted if all failed)

### DNS Resolution Block

Only shown when the host is an FQDN:

```
Resolved myapp.internal.com -> 2 addresses
  [1] 10.0.0.5
  [2] 10.0.0.6
```

## Exit Codes

| Code | Meaning                                      |
|------|----------------------------------------------|
| `0`  | Ran successfully (regardless of probe results)|
| `1`  | Invalid arguments or DNS resolution failure   |

## Tips

- **Firewall rules**: If all attempts show `timeout`, the port is likely filtered by a firewall. Try from a different network segment to isolate the issue.
- **UDP caveats**: Most UDP services don't respond to arbitrary probe packets. A "success" with ~5000ms latency means the probe timed out without receiving an ICMP error — the port is likely open or filtered, not necessarily responsive.
- **IPv6**: Works with IPv6 addresses. Use the bare address (e.g., `::1`), not bracket notation.
- **Run as Administrator**: On Windows, some network errors provide more detail when run with elevated privileges.
