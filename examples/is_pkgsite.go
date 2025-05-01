package examples

import (
	"net"
)

// IsPkgsite checks if the code is running in the pkg.go.dev environment.
func IsPkgsite() bool {
	_, err := net.LookupIP("google.com")
	dnsErr, ok := err.(*net.DNSError)
	if !ok {
		return false
	}

	return dnsErr.Server == "169.254.169.254:53" && dnsErr.Err == "dial udp 169.254.169.254:53: connect: no route to host"
}
