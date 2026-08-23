package worker

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SirZeck/ping-gopher/internal/validator"
)

// HTTPProbeResult contains the metrics collected from an HTTP/HTTPS request probe.
type HTTPProbeResult struct {
	StatusCode     int
	ResponseTimeMS int
	ErrorMessage   string
	IsUp           bool
}

// SSLProbeResult contains SSL/TLS certificate metadata collected from a target.
type SSLProbeResult struct {
	ExpirationDate *time.Time
	DaysRemaining  int
	Issuer         string
	ErrorMessage   string
	IsValid        bool
}

var sharedHTTPTransport = &http.Transport{
	DialContext:         validator.SafeDialContext(10 * time.Second),
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 20,
	IdleConnTimeout:     90 * time.Second,
}

var sharedHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: sharedHTTPTransport,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// ExecuteHTTPProbe performs a synthetic HTTP/HTTPS probe against targetURL with timeout and SSRF protection.
func ExecuteHTTPProbe(targetURL string, timeout time.Duration) *HTTPProbeResult {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if err := ValidateSafeURL(targetURL); err != nil {
		return &HTTPProbeResult{
			StatusCode:     0,
			ResponseTimeMS: 0,
			ErrorMessage:   fmt.Sprintf("SSRF protection blocked probe: %v", err),
			IsUp:           false,
		}
	}

	client := sharedHTTPClient
	if timeout != 10*time.Second {
		client = &http.Client{
			Timeout:       timeout,
			Transport:     sharedHTTPTransport,
			CheckRedirect: sharedHTTPClient.CheckRedirect,
		}
	}

	startTime := time.Now()
	resp, err := client.Get(targetURL)
	duration := time.Since(startTime)

	result := &HTTPProbeResult{
		ResponseTimeMS: int(duration.Milliseconds()),
	}

	if err != nil {
		result.StatusCode = 0
		result.ErrorMessage = err.Error()
		result.IsUp = false
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	// 2xx and 3xx response codes are considered UP by default
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.IsUp = true
	} else {
		result.IsUp = false
		result.ErrorMessage = fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return result
}

// ExecuteSSLProbe connects via TLS to targetURL to inspect server certificate validity and expiry.
func ExecuteSSLProbe(targetURL string, timeout time.Duration) *SSLProbeResult {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if err := ValidateSafeURL(targetURL); err != nil {
		return &SSLProbeResult{
			ErrorMessage: fmt.Sprintf("SSRF protection blocked SSL probe: %v", err),
			IsValid:      false,
		}
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return &SSLProbeResult{
			ErrorMessage: fmt.Sprintf("Invalid URL format: %v", err),
			IsValid:      false,
		}
	}

	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "443"
	}

	targetAddr := fmt.Sprintf("%s:%s", host, port)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	safeDialer := validator.SafeDialContext(timeout)
	rawConn, err := safeDialer(ctx, "tcp", targetAddr)
	if err != nil {
		return &SSLProbeResult{
			ErrorMessage: fmt.Sprintf("TLS Connection failed: %v", err),
			IsValid:      false,
		}
	}

	tlsConn := tls.Client(rawConn, &tls.Config{
		InsecureSkipVerify: false, // Enforce strict TLS validation
		ServerName:         host,
	})

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		rawConn.Close()
		return &SSLProbeResult{
			ErrorMessage: fmt.Sprintf("TLS Handshake failed: %v", err),
			IsValid:      false,
		}
	}
	defer tlsConn.Close()

	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return &SSLProbeResult{
			ErrorMessage: "No TLS peer certificates found",
			IsValid:      false,
		}
	}

	leafCert := certs[0]
	expiry := leafCert.NotAfter
	daysRemaining := int(time.Until(expiry).Hours() / 24)

	issuer := leafCert.Issuer.CommonName
	if issuer == "" && len(leafCert.Issuer.Organization) > 0 {
		issuer = leafCert.Issuer.Organization[0]
	}

	return &SSLProbeResult{
		ExpirationDate: &expiry,
		DaysRemaining:  daysRemaining,
		Issuer:         issuer,
		IsValid:        daysRemaining > 0,
	}
}

// ExecuteHTTPAssertionProbe performs an HTTP check validating exact expected status codes and body keyword assertions.
func ExecuteHTTPAssertionProbe(targetURL string, expectedStatus int, expectedKeyword string, timeout time.Duration) *HTTPProbeResult {
	result := ExecuteHTTPProbe(targetURL, timeout)
	if !result.IsUp {
		return result
	}

	if expectedStatus > 0 && result.StatusCode != expectedStatus {
		result.IsUp = false
		result.ErrorMessage = fmt.Sprintf("HTTP status assertion failed: expected %d, got %d", expectedStatus, result.StatusCode)
		return result
	}

	if expectedKeyword != "" {
		client := sharedHTTPClient
		if timeout > 0 && timeout != 10*time.Second {
			client = &http.Client{
				Timeout:       timeout,
				Transport:     sharedHTTPTransport,
				CheckRedirect: sharedHTTPClient.CheckRedirect,
			}
		}
		resp, err := client.Get(targetURL)
		if err == nil {
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
			if err == nil && !strings.Contains(string(bodyBytes), expectedKeyword) {
				result.IsUp = false
				result.ErrorMessage = fmt.Sprintf("HTTP body assertion failed: keyword '%s' not found in response", expectedKeyword)
			}
		}
	}

	return result
}

// ExecuteTCPProbe dials a TCP socket connection with connect-time SSRF validation and measures latency.
func ExecuteTCPProbe(targetAddr string, timeout time.Duration) *HTTPProbeResult {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	// Add default port 80 if missing
	host, port, err := net.SplitHostPort(targetAddr)
	if err != nil {
		host = targetAddr
		port = "80"
		targetAddr = fmt.Sprintf("%s:%s", host, port)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	safeDialer := validator.SafeDialContext(timeout)
	startTime := time.Now()
	conn, err := safeDialer(ctx, "tcp", targetAddr)
	duration := time.Since(startTime)

	if err != nil {
		return &HTTPProbeResult{
			StatusCode:     0,
			ResponseTimeMS: int(duration.Milliseconds()),
			ErrorMessage:   fmt.Sprintf("TCP Connection failed to %s: %v", targetAddr, err),
			IsUp:           false,
		}
	}
	defer conn.Close()

	return &HTTPProbeResult{
		StatusCode:     200,
		ResponseTimeMS: int(duration.Milliseconds()),
		IsUp:           true,
	}
}

// ExecuteDNSProbe resolves domain records (A, AAAA, MX, TXT) and measures lookup latency.
func ExecuteDNSProbe(domain string, recordType string, timeout time.Duration) *HTTPProbeResult {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}
	if host, _, err := net.SplitHostPort(domain); err == nil {
		domain = host
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	startTime := time.Now()
	var err error
	var recordsCount int

	switch strings.ToUpper(recordType) {
	case "AAAA":
		var ips []net.IPAddr
		ips, err = net.DefaultResolver.LookupIPAddr(ctx, domain)
		for _, ip := range ips {
			if ip.IP.To4() == nil {
				recordsCount++
			}
		}
	case "MX":
		var mxs []*net.MX
		mxs, err = net.DefaultResolver.LookupMX(ctx, domain)
		recordsCount = len(mxs)
	case "TXT":
		var txts []string
		txts, err = net.DefaultResolver.LookupTXT(ctx, domain)
		recordsCount = len(txts)
	default: // Default "A" record
		var ips []net.IPAddr
		ips, err = net.DefaultResolver.LookupIPAddr(ctx, domain)
		for _, ip := range ips {
			if ip.IP.To4() != nil {
				recordsCount++
			}
		}
	}
	duration := time.Since(startTime)

	if err != nil || recordsCount == 0 {
		errMsg := "No DNS records found"
		if err != nil {
			errMsg = err.Error()
		}
		return &HTTPProbeResult{
			StatusCode:     0,
			ResponseTimeMS: int(duration.Milliseconds()),
			ErrorMessage:   fmt.Sprintf("DNS lookup failed for %s (%s): %s", domain, recordType, errMsg),
			IsUp:           false,
		}
	}

	return &HTTPProbeResult{
		StatusCode:     200,
		ResponseTimeMS: int(duration.Milliseconds()),
		IsUp:           true,
	}
}
