package priceDB

import (
	"fmt"
	"io/ioutil"

	"crypto/x509"
)

// Build a x509.CertPool from the cert in certs
func grabCert(loc string) (*x509.CertPool, error) {

	cert, err := ioutil.ReadFile(loc)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire cert", err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(cert))
	if !ok {
		return nil, fmt.Errorf("failed to add cert to trust chain", err)
	}

	return roots, nil

}
