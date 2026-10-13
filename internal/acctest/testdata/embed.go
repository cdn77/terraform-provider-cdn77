package testdata

import (
	_ "embed"
	"strings"
)

//go:embed key.pem
var SslKey string

//go:embed cert1.pem
var SslCert1 string

//go:embed cert2.pem
var SslCert2 string

// SslCert3 covers the CNAMEs used by the CDN tests, which cert1 does not.
//
//go:embed cert3.pem
var SslCert3 string

func init() { //nolint:gochecknoinits // here the init makes sense
	SslKey = strings.TrimSpace(SslKey)
	SslCert1 = strings.TrimSpace(SslCert1)
	SslCert2 = strings.TrimSpace(SslCert2)
	SslCert3 = strings.TrimSpace(SslCert3)
}
