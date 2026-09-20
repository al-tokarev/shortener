package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	filehelpers "github.com/al-tokarev/shortener/internal/helpers/file_helpers"
)

func CreateX509Cert() (*Cert, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	exist1, err1 := filehelpers.FileExist(filepath.Join(homeDir, "cert.pem"))
	exist2, err2 := filehelpers.FileExist(filepath.Join(homeDir, "private.pem"))

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	if !exist1 && !exist2 {

	} else if exist1 && exist2 {

	} else {
		return nil, fmt.Errorf("inconsistent cert state")
	}

	switch {
	case !exist1 && !exist2:
		cert := &x509.Certificate{
			SerialNumber: big.NewInt(1462),
			Subject: pkix.Name{
				Organization: []string{"Shortener"},
				Country:      []string{"RU"},
			},
			IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
			NotBefore:    time.Now(),
			NotAfter:     time.Now().AddDate(10, 0, 0),
			SubjectKeyId: []byte{1, 2, 3, 4, 6},
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:     x509.KeyUsageDigitalSignature,
		}

		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, err
		}

		certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
		if err != nil {
			return nil, err
		}

		var certPEM bytes.Buffer
		err = pem.Encode(&certPEM, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		})
		if err != nil {
			return nil, err
		}

		var privateKeyPEM bytes.Buffer
		err = pem.Encode(&privateKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})
		if err != nil {
			return nil, err
		}

		if err = os.WriteFile(filepath.Join(homeDir, "cert.pem"), certPEM.Bytes(), 0644); err != nil {
			return nil, err
		}

		if err = os.WriteFile(filepath.Join(homeDir, "private.pem"), privateKeyPEM.Bytes(), 0600); err != nil {
			return nil, err
		}
	case exist1 && exist2:
	default:
		return nil, fmt.Errorf("inconsistent cert state")
	}

	return &Cert{
		CertFile: filepath.Join(homeDir, "cert.pem"),
		KeyFile:  filepath.Join(homeDir, "private.pem"),
	}, nil
}
