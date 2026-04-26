package app

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	LEProduction = "https://acme-v02.api.letsencrypt.org/directory"
	LEStaging    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	LiteSSLProduction = "https://acme.litessl.com/directory"
)

// legoUser implements registration.User
type legoUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *legoUser) GetEmail() string                        { return u.Email }
func (u *legoUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *legoUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

// cfDNSProvider implements the lego DNS provider interface using Cloudflare API
type cfDNSProvider struct {
	account *Account
	zoneID  string
	records []string // created record IDs
	ctx     context.Context
	wailsCtx context.Context
}

func (p *cfDNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	if p.wailsCtx != nil {
		runtime.EventsEmit(p.wailsCtx, "acme-log", map[string]interface{}{
			"level": "info",
			"msg":   fmt.Sprintf("创建 DNS TXT 记录: %s", fqdn),
		})
	}

	recordID, err := CreateDNSRecord(p.account, p.zoneID, fqdn, value)
	if err != nil {
		return err
	}
	p.records = append(p.records, recordID)
	return nil
}

func (p *cfDNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	if p.wailsCtx != nil {
		runtime.EventsEmit(p.wailsCtx, "acme-log", map[string]interface{}{
			"level": "info",
			"msg":   fmt.Sprintf("清理 DNS 记录: %s", fqdn),
		})
	}

	for _, id := range p.records {
		DeleteDNSRecord(p.account, p.zoneID, id)
	}
	p.records = nil
	return nil
}

func (p *cfDNSProvider) Timeout() (timeout, interval time.Duration) {
	return 5 * time.Minute, 10 * time.Second
}

func generatePrivateKey(algo string) (crypto.PrivateKey, error) {
	switch algo {
	case "RSA 2048":
		return rsa.GenerateKey(rand.Reader, 2048)
	case "RSA 4096":
		return rsa.GenerateKey(rand.Reader, 4096)
	case "ECDSA P-384":
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	default: // ECDSA P-256
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}
}

func keyTypeFromAlgo(algo string) certcrypto.KeyType {
	switch algo {
	case "RSA 2048":
		return certcrypto.RSA2048
	case "RSA 4096":
		return certcrypto.RSA4096
	case "ECDSA P-384":
		return certcrypto.EC384
	default:
		return certcrypto.EC256
	}
}

func ApplyCertificate(wailsCtx context.Context, account *Account, domains []string, ca, keyAlgo, env string) (*Certificate, error) {
	emitLog := func(level, msg string) {
		runtime.EventsEmit(wailsCtx, "acme-log", map[string]interface{}{
			"level": level,
			"msg":   msg,
		})
	}

	emitLog("info", "初始化 ACME 客户端...")

	// Determine ACME directory URL
	directoryURL := LEProduction
	caName := "Let's Encrypt"
	if ca == "LiteSSL" {
		directoryURL = LiteSSLProduction
		caName = "LiteSSL"
	} else if env == "staging" {
		directoryURL = LEStaging
	}
	emitLog("info", fmt.Sprintf("连接 %s (%s)", caName, directoryURL))

	// Generate account key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成账户密钥失败: %v", err)
	}

	user := &legoUser{
		Email: account.Email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)
	config.CADirURL = directoryURL
	config.Certificate.KeyType = keyTypeFromAlgo(keyAlgo)

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建 ACME 客户端失败: %v", err)
	}

	// Get zone ID for the first domain
	zoneID, err := GetZoneIDByName(account, domains[0])
	if err != nil {
		return nil, fmt.Errorf("获取 Zone ID 失败: %v", err)
	}

	// Set DNS provider
	dnsProvider := &cfDNSProvider{
		account: account,
		zoneID:  zoneID,
		ctx:     context.Background(),
		wailsCtx: wailsCtx,
	}
	err = client.Challenge.SetDNS01Provider(dnsProvider)
	if err != nil {
		return nil, fmt.Errorf("设置 DNS 验证器失败: %v", err)
	}

	// Register account
	emitLog("info", "创建/获取 ACME Account...")
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("注册 ACME 账户失败: %v", err)
	}
	user.Registration = reg

	// Request certificate
	domainStr := strings.Join(domains, ", ")
	emitLog("info", fmt.Sprintf("创建证书订单: %s", domainStr))

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	emitLog("info", "获取 DNS-01 Challenge...")
	emitLog("info", "等待 DNS 记录传播...")
	emitLog("info", "通知 ACME 服务器验证...")

	certRes, err := client.Certificate.Obtain(request)
	if err != nil {
		emitLog("error", fmt.Sprintf("证书申请失败: %v", err))
		return nil, fmt.Errorf("申请证书失败: %v", err)
	}

	emitLog("info", "域名验证通过")
	emitLog("info", fmt.Sprintf("生成 %s 密钥对...", keyAlgo))
	emitLog("info", "证书签发成功，下载证书...")

	// Parse the certificate to get expiration
	emitLog("info", "解析证书信息...")
	certInfo, err := ParseCertificatePEM(string(certRes.Certificate))
	if err != nil {
		emitLog("warn", fmt.Sprintf("解析证书失败: %v", err))
	}

	// Save files
	certPath := GetCertPath()
	domainDir := strings.ReplaceAll(domains[0], "*", "_wildcard")
	emitLog("info", fmt.Sprintf("清理 DNS 验证记录..."))

	// Save certificate files
	filePath, err := SaveCertificateFiles(certPath, domainDir, certRes)
	if err != nil {
		emitLog("warn", fmt.Sprintf("保存证书文件失败: %v", err))
	}

	emitLog("info", fmt.Sprintf("证书已保存: %s", filePath))

	// Build SAN list
	sanList := strings.Join(domains, ",")

	now := time.Now()
	expiresAt := now.AddDate(0, 0, 90) // default 90 days
	if certInfo != nil {
		expiresAt = certInfo.NotAfter
	}

	cert := &Certificate{
		AccountID:    account.ID,
		Domain:       domains[0],
		SANList:      sanList,
		SANCount:     len(domains),
		CABrand:      caName,
		KeyAlgo:      keyAlgo,
		IssuedAt:     now,
		ExpiresAt:    expiresAt,
		FilePath:     filePath,
		CertPEM:      string(certRes.Certificate),
		KeyPEM:       string(certRes.PrivateKey),
		ChainPEM:     string(certRes.IssuerCertificate),
		FullChainPEM: string(certRes.Certificate) + "\n" + string(certRes.IssuerCertificate),
		CreatedAt:    now,
	}

	// Save to database
	if gormDB != nil {
		gormDB.Create(cert)
	}

	emitLog("info", "证书申请完成！")
	return cert, nil
}
