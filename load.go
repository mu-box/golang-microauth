package microauth

import (
	"crypto/tls"
	"encoding/pem"
	"io/ioutil"

	"go.step.sm/crypto/pemutil"
)

// Load is a helper function to load a certificate and key from password protected files.
func Load(certFile, keyFile, password string) (*tls.Certificate, error) {
	certPEMBlock, err := ioutil.ReadFile(certFile)
	if err != nil {
		return &tls.Certificate{}, err
	}

	rawKeyPEMBlock, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return &tls.Certificate{}, err
	}
	temp, _ := pem.Decode(rawKeyPEMBlock)
	keyPEMBlock, err := pemutil.DecryptPEMBlock(temp, []byte(password))
	if err != nil {
		if err.Error() == "unsupported encrypted PEM" {
			keyPEMBlock = rawKeyPEMBlock
		} else {
			return &tls.Certificate{}, err
		}
	}

	crt, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	return &crt, err
}
