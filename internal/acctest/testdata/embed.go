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

func init() { //nolint:gochecknoinits // here the init makes sense
	SslKey = strings.TrimSpace(SslKey)
	SslCert1 = strings.TrimSpace(SslCert1)
	SslCert2 = strings.TrimSpace(SslCert2)
}
