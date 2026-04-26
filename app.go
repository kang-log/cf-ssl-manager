package main

import (
	"cf-ssl-manager/app"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := app.InitDB(); err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
		runtime.LogErrorf(a.ctx, "数据库初始化失败: %v", err)
	} else {
		fmt.Println("数据库初始化成功, path:", app.GetDBPath())
		runtime.LogInfof(a.ctx, "数据库初始化成功: %s", app.GetDBPath())
	}
}

// ==================== Debug ====================

func (a *App) DebugDB() (map[string]interface{}, error) {
	dbPath := app.GetDBPath()

	result := map[string]interface{}{
		"path": dbPath,
	}

	// Check if file exists
	if info, err := os.Stat(dbPath); err != nil {
		result["fileExists"] = false
		result["fileError"] = err.Error()
	} else {
		result["fileExists"] = true
		result["fileSize"] = info.Size()
	}

	// Test connection via GORM
	if pingErr := app.TestDB(); pingErr != nil {
		result["pingError"] = pingErr.Error()
	} else {
		result["pingOk"] = true
	}

	// Try GORM
	db := app.GetDB()
	if db == nil {
		result["gormStatus"] = "nil"
	} else {
		var count int64
		if err := db.Model(&app.Account{}).Count(&count).Error; err != nil {
			result["countError"] = err.Error()
		}
		result["accountCount"] = count

		var accounts []app.Account
		if err := db.Find(&accounts).Error; err != nil {
			result["findError"] = err.Error()
		}
		accList := make([]map[string]interface{}, 0)
		for _, acc := range accounts {
			accList = append(accList, map[string]interface{}{
				"id":       acc.ID,
				"email":    acc.Email,
				"hasKey":   acc.APIKey != "",
				"hasToken": acc.APIToken != "",
			})
		}
		result["accounts"] = accList

		// SQLite version info
		var version string
		if err := db.Raw("SELECT sqlite_version()").Scan(&version).Error; err == nil {
			result["sqliteVersion"] = version
		}
	}

	// Proxy settings
	proxyEnabled := app.GetSetting("proxyEnabled")
	proxyAddr := app.GetSetting("proxyAddr")
	result["proxyEnabled"] = proxyEnabled
	result["proxyAddr"] = proxyAddr

	// Quick network test to Cloudflare API
	cfReachable := "unknown"
	if tErr := app.TestCloudflareReachability(); tErr != nil {
		cfReachable = fmt.Sprintf("unreachable: %v", tErr)
	} else {
		cfReachable = "ok"
	}
	result["cloudflareApi"] = cfReachable

	return result, nil
}

// ==================== Account Methods ====================

func (a *App) AddAccount(email, apiKey, apiToken string) (map[string]interface{}, error) {
	fmt.Printf("[AddAccount] called with email=%q apiKeyLen=%d apiTokenLen=%d\n", email, len(apiKey), len(apiToken))

	// Validate required fields
	if email == "" || apiKey == "" {
		return nil, fmt.Errorf("邮箱和 API Key 为必填项")
	}

	// Verify credentials first
	verifiedEmail, err := app.VerifyCFAccount(email, apiKey, apiToken)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[AddAccount] verifiedEmail=%q\n", verifiedEmail)

	// Encrypt credentials
	encKey, err := app.Encrypt(apiKey)
	if err != nil {
		return nil, fmt.Errorf("加密 API Key 失败: %v", err)
	}
	encToken, err := app.Encrypt(apiToken)
	if err != nil {
		return nil, fmt.Errorf("加密 API Token 失败: %v", err)
	}
	fmt.Printf("[AddAccount] encKeyLen=%d encTokenLen=%d\n", len(encKey), len(encToken))

	db := app.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	// Use email-based ID for stability
	accountId := fmt.Sprintf("acc_%x", []byte(verifiedEmail))
	fmt.Printf("[AddAccount] accountId=%q\n", accountId)

	var account app.Account
	if err := db.Where("id = ?", accountId).First(&account).Error; err == nil {
		// Update existing — also refresh email in case it was stale
		account.Email = verifiedEmail
		account.APIKey = encKey
		account.APIToken = encToken
		fmt.Printf("[AddAccount] UPDATE path — saving account.ID=%q APIKeyLen=%d APITokenLen=%d\n", account.ID, len(account.APIKey), len(account.APIToken))
		if err := db.Save(&account).Error; err != nil {
			return nil, fmt.Errorf("更新账户失败: %v", err)
		}
	} else {
		account = app.Account{
			ID:       accountId,
			Email:    verifiedEmail,
			APIKey:   encKey,
			APIToken: encToken,
		}
		fmt.Printf("[AddAccount] CREATE path — saving account.ID=%q APIKeyLen=%d APITokenLen=%d\n", account.ID, len(account.APIKey), len(account.APIToken))
		if err := db.Create(&account).Error; err != nil {
			return nil, fmt.Errorf("保存账户失败: %v", err)
		}
	}

	// Immediately read back to verify persistence
	var verify app.Account
	if err := db.First(&verify, "id = ?", accountId).Error; err != nil {
		fmt.Printf("[AddAccount] READ-BACK FAILED: %v\n", err)
	} else {
		fmt.Printf("[AddAccount] READ-BACK OK: id=%q email=%q APIKeyLen=%d APITokenLen=%d\n", verify.ID, verify.Email, len(verify.APIKey), len(verify.APIToken))
		decKey, _ := app.Decrypt(verify.APIKey)
		decToken, _ := app.Decrypt(verify.APIToken)
		fmt.Printf("[AddAccount] READ-BACK decrypted: apiKeyLen=%d apiTokenLen=%d\n", len(decKey), len(decToken))
	}

	return map[string]interface{}{
		"id":        account.ID,
		"email":     verifiedEmail,
		"debugInfo": fmt.Sprintf("encKeyLen=%d encTokenLen=%d savedApiKeyLen=%d savedApiTokenLen=%d", len(encKey), len(encToken), len(account.APIKey), len(account.APIToken)),
	}, nil
}

func (a *App) UpdateAccount(accountId, apiKey, apiToken string) (map[string]interface{}, error) {
	db := app.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var account app.Account
	if err := db.First(&account, "id = ?", accountId).Error; err != nil {
		return nil, fmt.Errorf("账户不存在")
	}

	// Verify new credentials
	verifiedEmail, err := app.VerifyCFAccount(account.Email, apiKey, apiToken)
	if err != nil {
		return nil, fmt.Errorf("凭证验证失败: %v", err)
	}

	encKey, err := app.Encrypt(apiKey)
	if err != nil {
		return nil, fmt.Errorf("加密 API Key 失败: %v", err)
	}
	encToken, err := app.Encrypt(apiToken)
	if err != nil {
		return nil, fmt.Errorf("加密 API Token 失败: %v", err)
	}

	account.Email = verifiedEmail
	account.APIKey = encKey
	account.APIToken = encToken
	db.Save(&account)

	return map[string]interface{}{
		"id":    account.ID,
		"email": verifiedEmail,
	}, nil
}

func (a *App) VerifyAccount(accountId string) (bool, error) {
	db := app.GetDB()
	if db == nil {
		return false, fmt.Errorf("数据库未初始化")
	}

	var account app.Account
	if err := db.First(&account, "id = ?", accountId).Error; err != nil {
		return false, fmt.Errorf("账户不存在")
	}

	apiKey, err := app.Decrypt(account.APIKey)
	if err != nil {
		return false, fmt.Errorf("解密 API Key 失败: %v", err)
	}
	apiToken, err := app.Decrypt(account.APIToken)
	if err != nil {
		return false, fmt.Errorf("解密 API Token 失败: %v", err)
	}

	if apiKey == "" && apiToken == "" {
		return false, fmt.Errorf("账户凭证为空，请编辑账户重新填写 API Key 或 API Token")
	}

	_, err = app.VerifyCFAccount(account.Email, apiKey, apiToken)
	return err == nil, err
}

func (a *App) RemoveAccount(accountId string) error {
	db := app.GetDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	db.Where("account_id = ?", accountId).Delete(&app.Zone{})
	db.Delete(&app.Account{}, "id = ?", accountId)
	return nil
}

func (a *App) LoadAccounts() ([]map[string]interface{}, error) {
	db := app.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var accounts []app.Account
	db.Find(&accounts)

	fmt.Printf("[LoadAccounts] found %d accounts\n", len(accounts))
	for i, acc := range accounts {
		fmt.Printf("[LoadAccounts] acc[%d]: id=%q email=%q apiKeyLen=%d apiTokenLen=%d\n", i, acc.ID, acc.Email, len(acc.APIKey), len(acc.APIToken))
	}

	var result []map[string]interface{}
	for _, acc := range accounts {
		hasKey := false
		if acc.APIKey != "" {
			dec, err := app.Decrypt(acc.APIKey)
			hasKey = err == nil && dec != ""
		}
		hasToken := false
		if acc.APIToken != "" {
			dec, err := app.Decrypt(acc.APIToken)
			hasToken = err == nil && dec != ""
		}
		result = append(result, map[string]interface{}{
			"id":          acc.ID,
			"email":       acc.Email,
			"hasApiKey":   hasKey,
			"hasApiToken": hasToken,
		})
	}
	return result, nil
}

// ==================== Zone Methods ====================

func (a *App) FetchZones(accountId string) ([]map[string]interface{}, error) {
	db := app.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var account app.Account
	if err := db.First(&account, "id = ?", accountId).Error; err != nil {
		return nil, fmt.Errorf("账户不存在")
	}

	fmt.Printf("[FetchZones] accountId=%q found in DB: email=%q storedApiKeyLen=%d storedApiTokenLen=%d\n", accountId, account.Email, len(account.APIKey), len(account.APIToken))

	apiKey, err := app.Decrypt(account.APIKey)
	if err != nil {
		return nil, fmt.Errorf("解密 API Key 失败: %v", err)
	}
	apiToken, err := app.Decrypt(account.APIToken)
	if err != nil {
		return nil, fmt.Errorf("解密 API Token 失败: %v", err)
	}

	fmt.Printf("[FetchZones] after decrypt: apiKeyLen=%d apiTokenLen=%d\n", len(apiKey), len(apiToken))

	if apiKey == "" && apiToken == "" {
		return nil, fmt.Errorf("账户凭证为空，请编辑账户重新填写 API Key 或 API Token")
	}

	zones, err := app.FetchCFZones(account.Email, apiKey, apiToken)
	if err != nil {
		return nil, err
	}

	for _, z := range zones {
		z.AccountID = accountId
		db.Save(&z)
	}

	var result []map[string]interface{}
	for _, z := range zones {
		nsList := strings.Split(z.NS, ",")
		result = append(result, map[string]interface{}{
			"id":        z.ID,
			"accountId": accountId,
			"name":      z.Name,
			"status":    z.Status,
			"plan":      z.Plan,
			"ns":        nsList,
		})
	}
	return result, nil
}

func (a *App) LoadZones() ([]map[string]interface{}, error) {
	db := app.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var zones []app.Zone
	db.Find(&zones)

	var result []map[string]interface{}
	for _, z := range zones {
		nsList := strings.Split(z.NS, ",")
		result = append(result, map[string]interface{}{
			"id":        z.ID,
			"accountId": z.AccountID,
			"name":      z.Name,
			"status":    z.Status,
			"plan":      z.Plan,
			"ns":        nsList,
		})
	}
	return result, nil
}

// ==================== Certificate Methods ====================

func (a *App) ApplyCert(accountId, zoneId, domainsStr, ca, keyAlgo, env, customCertPath string) (map[string]interface{}, error) {
	db := app.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var account app.Account
	if err := db.First(&account, "id = ?", accountId).Error; err != nil {
		return nil, fmt.Errorf("账户不存在")
	}

	domains := parseDomains(domainsStr)
	if len(domains) == 0 {
		return nil, fmt.Errorf("请输入至少一个域名")
	}

	// Override cert path if custom path provided
	if customCertPath != "" {
		app.SaveSetting("certPath", customCertPath)
	}

	cert, err := app.ApplyCertificate(a.ctx, &account, domains, ca, keyAlgo, env)
	if err != nil {
		return nil, err
	}

	return certToMap(cert), nil
}

func (a *App) ListCerts() ([]map[string]interface{}, error) {
	db := app.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var certs []app.Certificate
	db.Order("created_at DESC").Find(&certs)

	var result []map[string]interface{}
	for _, c := range certs {
		result = append(result, certToMap(&c))
	}
	return result, nil
}

func (a *App) GetCertDetail(certId int) (map[string]interface{}, error) {
	db := app.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var cert app.Certificate
	if err := db.First(&cert, certId).Error; err != nil {
		return nil, fmt.Errorf("证书记录不存在")
	}

	result := certToMap(&cert)

	if cert.CertPEM != "" {
		detail, err := app.ParseCertificatePEM(cert.CertPEM)
		if err == nil {
			result["detail"] = detail
		}
	}

	return result, nil
}

func (a *App) DeleteCert(certId int) error {
	db := app.GetDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return db.Delete(&app.Certificate{}, certId).Error
}

func (a *App) CheckRemoteCert(domain, port string) (map[string]interface{}, error) {
	detail, err := app.CheckRemoteCert(domain, port)
	if err != nil {
		return nil, err
	}
	return certDetailToMap(detail), nil
}

func (a *App) ImportCertFromPEM(pemStr string) (map[string]interface{}, error) {
	detail, err := app.ParseCertificatePEM(pemStr)
	if err != nil {
		return nil, err
	}
	return certDetailToMap(detail), nil
}

// ==================== Settings Methods ====================

func (a *App) GetSettings() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"certPath":     app.GetCertPath(),
		"reminderDays": app.GetSetting("reminderDays"),
		"autoRenew":    app.GetSetting("autoRenew"),
		"logLevel":     app.GetSetting("logLevel"),
		"proxyEnabled": app.GetSetting("proxyEnabled"),
		"proxyType":    app.GetSetting("proxyType"),
		"proxyAddr":    app.GetSetting("proxyAddr"),
	}
	return result, nil
}

func (a *App) SaveSettings(settingsJson string) error {
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(settingsJson), &settings); err != nil {
		return fmt.Errorf("解析设置失败: %v", err)
	}
	for k, v := range settings {
		app.SaveSetting(k, fmt.Sprintf("%v", v))
	}
	return nil
}

// ==================== Utility Methods ====================

func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择证书存储目录",
	})
}

func (a *App) OpenPath(path string) error {
	_, err := os.StartProcess("explorer", []string{path}, &os.ProcAttr{})
	return err
}

func (a *App) LoadAllData() (map[string]interface{}, error) {
	accounts, _ := a.LoadAccounts()
	zones, _ := a.LoadZones()
	certs, _ := a.ListCerts()

	return map[string]interface{}{
		"accounts": accounts,
		"zones":    zones,
		"certs":    certs,
	}, nil
}

// ==================== Helpers ====================

func parseDomains(input string) []string {
	lines := strings.Split(input, "\n")
	var domains []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, ",")
		if line != "" {
			parts := strings.Split(line, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					domains = append(domains, p)
				}
			}
		}
	}
	return domains
}

func certToMap(c *app.Certificate) map[string]interface{} {
	status := "valid"
	now := time.Now()
	daysLeft := int(c.ExpiresAt.Sub(now).Hours() / 24)
	if daysLeft <= 0 {
		status = "expired"
	} else if daysLeft <= 30 {
		status = "expiring"
	}

	sanList := strings.Split(c.SANList, ",")

	return map[string]interface{}{
		"id":           c.ID,
		"accountId":    c.AccountID,
		"domain":       c.Domain,
		"sanCount":     c.SANCount,
		"ca":           c.CABrand,
		"keyAlgo":      c.KeyAlgo,
		"issuedAt":     c.IssuedAt.Format("2006-01-02 15:04:05"),
		"expiresAt":    c.ExpiresAt.Format("2006-01-02 15:04:05"),
		"status":       status,
		"sanList":      sanList,
		"filePath":     c.FilePath,
		"certPem":      c.CertPEM,
		"keyPem":       c.KeyPEM,
		"chainPem":     c.ChainPEM,
		"fullChainPem": c.FullChainPEM,
	}
}

func certDetailToMap(d *app.CertDetail) map[string]interface{} {
	return map[string]interface{}{
		"subject":      d.Subject,
		"issuer":       d.Issuer,
		"serialNumber": d.SerialNumber,
		"sigAlgorithm": d.SigAlgorithm,
		"keyAlgorithm": d.KeyAlgorithm,
		"keyLength":    d.KeyLength,
		"notBefore":    d.NotBefore,
		"notAfter":     d.NotAfter,
		"daysLeft":     d.DaysLeft,
		"sanList":      d.SANList,
		"sha1":         d.SHA1,
		"sha256":       d.SHA256,
	}
}
