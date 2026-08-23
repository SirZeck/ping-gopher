# 🛡️ Validator Module (`internal/validator`)

`internal/validator` provides security validation utilities for the PingGopher runtime engine, with primary emphasis on Server-Side Request Forgery (SSRF) prevention.

---

## 🎯 Security Features

1. **URL Protocol Scheme Enforcer**:
   Only permits `http://` and `https://` target URL schemes. Protocols such as `file://`, `gopher://`, `dict://`, and `ftp://` are rejected.

2. **CIDR Blocklist Validation**:
   Validates hostname resolutions against private RFC 1918 networks (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), loopback (`127.0.0.0/8`, `::1/128`), link-local (`169.254.0.0/16`, `fe80::/10`), and cloud provider Instance Metadata Service (IMDS) endpoints (`169.254.169.254`).

3. **IPv4-Mapped IPv6 Normalization**:
   Normalizes IPv4-mapped IPv6 addresses (`::ffff:169.254.169.254`) via `ip.To4()` prior to CIDR evaluation.

4. **Connect-Time Socket Dialer (`SafeDialContext`)**:
   Provides `validator.SafeDialContext(timeout)` for HTTP transports and TLS dialers to re-verify resolved target IP addresses at TCP socket creation time, protecting against DNS TTL=0 TOCTOU rebinding attacks.
