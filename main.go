package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const usage = `Usage: ppping <host> <port> [proto] [count]

  host   IP address or FQDN (required)
  port   Port number (required)
  proto  Protocol: tcp or udp (default: tcp)
  count  Number of attempts (default: 4)

Examples:
  ppping 172.26.104.10 3389
  ppping 172.26.104.10 3389 tcp
  ppping 172.26.104.10 4433 udp 10
  ppping myapp.internal.com 443`

func main() {
	if len(os.Args) < 3 || len(os.Args) > 5 {
		fmt.Println(usage)
		os.Exit(1)
	}

	host := os.Args[1]
	port := os.Args[2]
	proto := "tcp"
	count := 4

	if _, err := strconv.Atoi(port); err != nil {
		fmt.Printf("Error: invalid port %q\n", port)
		os.Exit(1)
	}

	if len(os.Args) >= 4 {
		proto = os.Args[3]
		if proto != "tcp" && proto != "udp" {
			fmt.Printf("Error: unsupported protocol %q (use tcp or udp)\n", proto)
			os.Exit(1)
		}
	}

	if len(os.Args) == 5 {
		var err error
		count, err = strconv.Atoi(os.Args[4])
		if err != nil || count < 1 {
			fmt.Printf("Error: invalid count %q\n", os.Args[4])
			os.Exit(1)
		}
	}

	ips := resolveHost(host)

	for idx, ip := range ips {
		if idx > 0 {
			fmt.Println()
		}
		probe(ip, port, proto, count)
	}
}

func resolveHost(host string) []string {
	// If it's already an IP address, return it directly.
	if net.ParseIP(host) != nil {
		return []string{host}
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		fmt.Printf("Error: could not resolve %s: %v\n", host, err)
		os.Exit(1)
	}
	if len(addrs) == 0 {
		fmt.Printf("Error: no addresses found for %s\n", host)
		os.Exit(1)
	}

	fmt.Printf("Resolved %s -> %d address", host, len(addrs))
	if len(addrs) != 1 {
		fmt.Print("es")
	}
	fmt.Println()
	for i, addr := range addrs {
		fmt.Printf("  [%d] %s\n", i+1, addr)
	}
	fmt.Println()

	return addrs
}

func probe(ip, port, proto string, count int) {
	target := net.JoinHostPort(ip, port)
	fmt.Printf("Probing %s (%s) x%d\n", target, proto, count)

	var successes int
	var totalLatency time.Duration

	for i := 1; i <= count; i++ {
		if i > 1 {
			time.Sleep(1 * time.Second)
		}

		latency, err := doProbe(proto, target)
		if err != nil {
			fmt.Printf("  Attempt %d: Failed   %s\n", i, formatError(err))
		} else {
			successes++
			totalLatency += latency
			fmt.Printf("  Attempt %d: Success  %s\n", i, formatLatency(latency))
		}
	}

	fmt.Printf("  Summary: %d/%d succeeded", successes, count)
	if successes > 0 {
		avg := totalLatency / time.Duration(successes)
		fmt.Printf(", avg %s", formatLatency(avg))
	}
	fmt.Println()
}

func doProbe(proto, target string) (time.Duration, error) {
	timeout := 5 * time.Second

	switch proto {
	case "tcp":
		start := time.Now()
		conn, err := net.DialTimeout("tcp", target, timeout)
		latency := time.Since(start)
		if err != nil {
			return 0, err
		}
		conn.Close()
		return latency, nil

	case "udp":
		conn, err := net.DialTimeout("udp", target, timeout)
		if err != nil {
			return 0, err
		}
		defer conn.Close()

		conn.SetDeadline(time.Now().Add(timeout))
		start := time.Now()
		_, err = conn.Write([]byte("\x00"))
		if err != nil {
			return 0, err
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		latency := time.Since(start)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// No response — open|filtered (typical for UDP).
				return latency, nil
			}
			return 0, err
		}
		return latency, nil

	default:
		return 0, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func formatLatency(d time.Duration) string {
	ms := float64(d.Microseconds()) / 1000.0
	return fmt.Sprintf("%.1fms", ms)
}

func formatError(err error) string {
	if opErr, ok := err.(*net.OpError); ok {
		switch {
		case opErr.Timeout():
			return "timeout"
		default:
			if opErr.Err != nil {
				return opErr.Err.Error()
			}
		}
	}
	return err.Error()
}
