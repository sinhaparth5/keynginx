package crypto

import (
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
)

type CertificateRequest struct {
	Domain       string
	KeySize      int
	ValidityDays int
	Country      string
	State        string
	City         string
	Organization string
	Unit         string
	Email        string
}

type KeyPair struct {
	PrivateKey     *rsa.PrivateKey
	Certificate    *x509.Certificate
	PrivateKeyPEM  []byte
	CertificatePEM []byte
}

type CertificateInfo struct {
	Subject         string
	Issuer          string
	NotBefore       time.Time
	NotAfter        time.Time
	DNSNames        []string
	IsExpired       bool
	DaysUntilExpiry int
}

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateKeyPair(req CertificateRequest) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, req.KeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            []string{req.Country},
			Province:           []string{req.State},
			Locality:           []string{req.City},
			Organization:       []string{req.Organization},
			OrganizationalUnit: []string{req.Unit},
			CommonName:         req.Domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(req.ValidityDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	template.DNSNames = []string{req.Domain}

	if req.Domain == "localhost" {
		template.DNSNames = append(template.DNSNames, "127.0.0.1")
		template.IPAddresses = []net.IP{
			net.IPv4(127, 0, 0, 1),
			net.IPv6loopback,
		}
	} else {
		template.DNSNames = append(template.DNSNames, "*."+req.Domain)
	}

	if req.Email != "" {
		template.EmailAddresses = []string{req.Email}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated certificate: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	certificatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return &KeyPair{
		PrivateKey:     privateKey,
		Certificate:    cert,
		PrivateKeyPEM:  privateKeyPEM,
		CertificatePEM: certificatePEM,
	}, nil
}

func (g *Generator) SaveKeyPair(keyPair *KeyPair, privateKeyPath, certificatePath string) error {
	if err := os.MkdirAll(filepath.Dir(privateKeyPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for private key: %w", err)
	}

	if privateKeyPath != certificatePath {
		if err := os.MkdirAll(filepath.Dir(certificatePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for certificate: %w", err)
		}
	}

	if err := os.WriteFile(privateKeyPath, keyPair.PrivateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	if err := os.WriteFile(certificatePath, keyPair.CertificatePEM, 0644); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	return nil
}

func (g *Generator) ValidateCertificate(certPath string) (*CertificateInfo, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from certificate")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("PEM block is not a certificate (type: %s)", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	now := time.Now()
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

	info := &CertificateInfo{
		Subject:         cert.Subject.CommonName,
		Issuer:          cert.Issuer.CommonName,
		NotBefore:       cert.NotBefore,
		NotAfter:        cert.NotAfter,
		DNSNames:        cert.DNSNames,
		IsExpired:       now.After(cert.NotAfter),
		DaysUntilExpiry: daysUntilExpiry,
	}

	return info, nil
}
