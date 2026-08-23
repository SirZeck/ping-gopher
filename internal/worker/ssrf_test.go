package worker

import (
	"errors"
	"testing"
)

func TestValidateSafeURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errType error
	}{
		{
			name:    "Blocked: Loopback IPv4",
			url:     "http://127.0.0.1/admin",
			wantErr: true,
			errType: ErrSSRFForbiddenIP,
		},
		{
			name:    "Blocked: AWS Cloud Metadata Endpoint",
			url:     "http://169.254.169.254/latest/meta-data/",
			wantErr: true,
			errType: ErrSSRFForbiddenIP,
		},
		{
			name:    "Blocked: Private RFC 1918 (10.0.0.1)",
			url:     "http://10.0.0.1:8080/metrics",
			wantErr: true,
			errType: ErrSSRFForbiddenIP,
		},
		{
			name:    "Blocked: Private RFC 1918 (192.168.1.1)",
			url:     "http://192.168.1.1/router",
			wantErr: true,
			errType: ErrSSRFForbiddenIP,
		},
		{
			name:    "Blocked: Invalid Scheme (ftp)",
			url:     "ftp://example.com/file.txt",
			wantErr: true,
			errType: ErrInvalidScheme,
		},
		{
			name:    "Blocked: Localhost Hostname",
			url:     "http://localhost:8080",
			wantErr: true,
			errType: ErrSSRFForbiddenIP,
		},
		{
			name:    "Blocked: IPv4-Mapped IPv6 Metadata",
			url:     "http://[::ffff:169.254.169.254]/latest/meta-data",
			wantErr: true,
			errType: ErrSSRFForbiddenIP,
		},
		{
			name:    "Allowed: Valid Public HTTPS Target (httpbin.org)",
			url:     "https://httpbin.org/status/200",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSafeURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSafeURL(%s) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("ValidateSafeURL(%s) error = %v, wantErrType %v", tt.url, err, tt.errType)
			}
		})
	}
}
