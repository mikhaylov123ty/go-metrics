package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

const (
	keysPairDir = "./"
	certName    = "cert.pem"
	keyName     = "key.pem"
)

func main() {
	if err := checkCert(); err != nil {
		log.Printf("Check Certificate Error: %s", err.Error())
		log.Printf("Creating new keys pair")

		if err = generateCert(); err != nil {
			log.Fatalf("Generate Certificate Error: %s", err.Error())
		}
	}
}

// генерация открытого и закрытого ключа
func generateCert() error {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Ya Praktikum"},
			Country:      []string{"RU"},
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, 1),

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature & x509.KeyUsageKeyEncipherment,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed generate private key: %w", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed create certificate: %w", err)
	}

	var certPEM bytes.Buffer
	if err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return fmt.Errorf("failed pem encode certificate: %w", err)
	}

	var keyPEM bytes.Buffer
	if err = pem.Encode(&keyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return fmt.Errorf("failed pem encode private key: %w", err)
	}

	keyPath := strings.Split(keysPairDir, "/")
	if len(keyPath) > 1 {
		container := strings.Join(keyPath[:len(keyPath)-1], "/")
		if err = os.MkdirAll(container, 0755); err != nil {
			return fmt.Errorf("failed create container directory: %w", err)
		}
	}

	if err = os.WriteFile(keysPairDir+keyName, keyPEM.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed write tls private key: %w", err)
	}

	if err = os.WriteFile(keysPairDir+certName, certPEM.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed write tls public key: %w", err)
	}

	return nil
}

func checkCert() error {
	_, errKey := os.ReadFile(keysPairDir + keyName)
	if errKey != nil {
		return fmt.Errorf("error reading tls private key: %w", errKey)

	}

	_, errCert := os.ReadFile(keysPairDir + certName)
	if errCert != nil {
		return fmt.Errorf("error reading tls public key: %w", errCert)
	}

	return nil
}
