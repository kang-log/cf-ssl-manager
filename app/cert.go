package app

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
)

type CertDetail struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serialNumber"`
	SigAlgorithm string    `json:"sigAlgorithm"`
	KeyAlgorithm string    `json:"keyAlgorithm"`
	KeyLength    int       `json:"keyLength"`
	NotBefore    time.Time `json:"notBefore"`
	NotAfter     time.Time `json:"notAfter"`
	DaysLeft     int       `json:"daysLeft"`
	SANList      []string  `json:"sanList"`
	SHA1         string    `json:"sha1"`
	SHA256       string    `json:"sha256"`
}

func ParseCertificatePEM(pemStr string) (*CertDetail, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("无法解析 PEM 数据")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %v", err)
	}

	return certToDetail(cert), nil
}

func certToDetail(cert *x509.Certificate) *CertDetail {
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)

	keyAlgo := cert.PublicKeyAlgorithm.String()
	keyLen := 0
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		keyLen = pub.N.BitLen()
		keyAlgo = fmt.Sprintf("RSA %d", keyLen)
	default:
		// For EC keys, determine from the curve
		keyAlgo = cert.PublicKeyAlgorithm.String()
	}

	// Format SAN list
	var sanList []string
	for _, san := range cert.DNSNames {
		if strings.Contains(san, "*") {
			sanList = append(sanList, san+" (泛域名)")
		} else {
			sanList = append(sanList, san+" (DNS)")
		}
	}
	for _, ip := range cert.IPAddresses {
		sanList = append(sanList, ip.String()+" (IP)")
	}

	// Compute fingerprints
	sha1Fp := fmt.Sprintf("%X", sha1.Sum(cert.Raw))
	sha256Fp := fmt.Sprintf("%X", sha256.Sum256(cert.Raw))

	sha1Fp = formatFingerprint(sha1Fp)
	sha256Fp = formatFingerprint(sha256Fp)

	return &CertDetail{
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: formatSerialNumber(cert.SerialNumber.String()),
		SigAlgorithm: cert.SignatureAlgorithm.String(),
		KeyAlgorithm: keyAlgo,
		KeyLength:    keyLen,
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		DaysLeft:     daysLeft,
		SANList:      sanList,
		SHA1:         sha1Fp,
		SHA256:       sha256Fp,
	}
}

func formatFingerprint(fp string) string {
	var parts []string
	for i := 0; i < len(fp); i += 2 {
		if i+2 <= len(fp) {
			parts = append(parts, fp[i:i+2])
		}
	}
	return strings.Join(parts, ":")
}

func formatSerialNumber(serial string) string {
	// Parse as hex
	bi, ok := new(big.Int).SetString(serial, 10)
	if ok {
		serial = fmt.Sprintf("%X", bi)
	}
	if len(serial)%2 != 0 {
		serial = "0" + serial
	}
	var parts []string
	for i := 0; i < len(serial); i += 2 {
		if i+2 <= len(serial) {
			parts = append(parts, strings.ToUpper(serial[i:i+2]))
		}
	}
	return strings.Join(parts, ":")
}

func SaveCertificateFiles(basePath, domainDir string, certRes *certificate.Resource) (string, error) {
	dir := filepath.Join(basePath, domainDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	domain := strings.ReplaceAll(domainDir, "_wildcard", "*")

	files := map[string][]byte{
		domain + ".crt":           certRes.Certificate,
		domain + ".key":           certRes.PrivateKey,
		domain + ".chain.crt":     certRes.IssuerCertificate,
		domain + ".fullchain.pem": append(certRes.Certificate, certRes.IssuerCertificate...),
	}

	for name, data := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0600); err != nil {
			return "", fmt.Errorf("写入文件 %s 失败: %v", name, err)
		}
	}

	return dir, nil
}

func CheckRemoteCert(domain, port string) (*CertDetail, error) {
	if port == "" {
		port = "443"
	}
	addr := net.JoinHostPort(domain, port)

	conn, err := tls.DialWithDialer(nil, "tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		return nil, fmt.Errorf("连接远程服务器失败: %v", err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("远程服务器未返回证书")
	}

	return certToDetail(certs[0]), nil
}
