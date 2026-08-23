package validator

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

var (
	ErrInvalidScheme   = errors.New("invalid URL scheme: only http and https are permitted")
	ErrEmptyHost       = errors.New("invalid URL: missing host")
	ErrSSRFForbiddenIP = errors.New("prohibited target IP address: private, loopback, or cloud metadata endpoints are not allowed")
)

// privateIPBlocks contains parsed IP CIDR networks that synthetic probes must not dial.
var privateIPBlocks []*net.IPNet

func init() {
	cidrs := []string{
		"127.0.0.0/8",    // IPv4 Loopback
		"10.0.0.0/8",     // RFC 1918 Private
		"172.16.0.0/12",  // RFC 1918 Private
		"192.168.0.0/16", // RFC 1918 Private
		"169.254.0.0/16", // Link-Local & AWS/GCP/Azure Metadata (169.254.169.254)
		"100.64.0.0/10",  // Carrier-Grade NAT
		"0.0.0.0/8",      // Current network
		"::1/128",        // IPv6 Loopback
		"fc00::/7",       // IPv6 Unique Local Address
		"fe80::/10",      // IPv6 Link-Local
	}

	for _, cidr := range cidrs {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			privateIPBlocks = append(privateIPBlocks, block)
		}
	}
}

// ValidateSafeURL checks if a target URL is safe for synthetic probing and protects against SSRF attacks.
func ValidateSafeURL(targetURL string) error {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL structure: %w", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrInvalidScheme
	}

	host := parsedURL.Hostname()
	if host == "" {
		return ErrEmptyHost
	}

	allowLoopback := os.Getenv("PINGGOPHER_ALLOW_LOOPBACK") == "true"

	// Parse host IP directly if given as IP address
	ip := net.ParseIP(host)
	if ip != nil {
		if allowLoopback && ip.IsLoopback() {
			return nil
		}
		if isPrivateIP(ip) {
			return ErrSSRFForbiddenIP
		}
		return nil
	}

	// Resolve hostname to IP addresses
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("failed to resolve host IP: %w", err)
	}

	for _, resolvedIP := range ips {
		if allowLoopback && resolvedIP.IsLoopback() {
			continue
		}
		if isPrivateIP(resolvedIP) {
			return fmt.Errorf("%w: host '%s' resolved to internal IP '%s'", ErrSSRFForbiddenIP, host, resolvedIP.String())
		}
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
