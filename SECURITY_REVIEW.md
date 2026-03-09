# Security Review: ppping

Date: 2026-03-09
Scope: `main.go`, runtime behavior, and user-facing output.

## Executive Summary

`ppping` is a small network probing CLI with a limited attack surface (no privileged operations, no file writes, no shell execution). The most relevant security concerns are **operator safety** and **resource exhaustion** from unbounded user input.

## Findings

### 1) Unbounded probe count could cause local resource exhaustion (Medium)

**Issue:** `count` previously accepted any non-negative integer, which could result in extremely long runtimes and repeated network traffic if a very large value was supplied.

**Impact:** Accidental or malicious invocation could consume CPU/network resources for a prolonged period.

**Fix applied:** Added a hard limit of `1,000,000` attempts (`maxCount`), while keeping `0` as explicit nonstop mode.

### 2) Terminal output injection risk via unsanitized host display (Low)

**Issue:** Hostnames were previously printed with `%s`. If terminal control characters are present in input, output rendering could be manipulated in some terminals.

**Impact:** Mostly operational/log readability risk.

**Fix applied:** Hostnames are now rendered with `%q` in DNS resolution messages and errors, escaping control characters.

### 3) Protocol parsing accepted only lowercase (Hardening)

**Issue:** Protocol parsing was case-sensitive (`tcp`/`udp` only).

**Impact:** Not a direct vulnerability, but inconsistent UX and avoidable input edge cases.

**Fix applied:** Protocol values are normalized with `strings.ToLower` before validation.

## Additional Notes

- The tool does not execute shell commands or deserialize untrusted structured data.
- Network behavior is intentionally scanner-like; usage should remain within authorized environments.
- UDP "success" semantics (timeout treated as open|filtered) are documented and should be preserved to avoid false assumptions.

## Recommended Future Hardening

1. Add rate limiting/jitter option for safer use in sensitive environments.
2. Add `--json` output mode to reduce log parsing ambiguity.
3. Consider optional global deadline for multi-IP runs.
