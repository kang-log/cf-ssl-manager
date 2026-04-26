package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func cfOptions() []cloudflare.Option {
	proxyEnabled := GetSetting("proxyEnabled")
	if proxyEnabled != "true" {
		return nil
	}
	proxyAddr := GetSetting("proxyAddr")
	if proxyAddr == "" {
		return nil
	}
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return nil
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return []cloudflare.Option{cloudflare.HTTPClient(&http.Client{Transport: transport})}
}

func newCFClientWithToken(token string) (*cloudflare.API, error) {
	return cloudflare.NewWithAPIToken(token, cfOptions()...)
}

func newCFClient(email, apiKey string) (*cloudflare.API, error) {
	return cloudflare.New(apiKey, email, cfOptions()...)
}

func VerifyCFAccount(email, apiKey, apiToken string) (string, error) {
	ctx := context.Background()
	if apiToken != "" {
		api, err := newCFClientWithToken(apiToken)
		if err != nil {
			return "", fmt.Errorf("创建 CF 客户端失败: %v", err)
		}
		token, err := api.VerifyAPIToken(ctx)
		if err != nil {
			return "", fmt.Errorf("API Token 验证失败: %v", err)
		}
		if token.Status != "active" {
			return "", fmt.Errorf("API Token 状态异常: %s", token.Status)
		}
		// Get user info
		u, err := api.UserDetails(ctx)
		if err != nil {
			return "", fmt.Errorf("获取用户信息失败: %v", err)
		}
		return u.Email, nil
	}

	if email == "" || apiKey == "" {
		return "", fmt.Errorf("请填写 Email 和 API Key")
	}

	api, err := newCFClient(email, apiKey)
	if err != nil {
		return "", fmt.Errorf("创建 CF 客户端失败: %v", err)
	}
	u, err := api.UserDetails(ctx)
	if err != nil {
		return "", fmt.Errorf("凭证验证失败: %v", err)
	}
	return u.Email, nil
}

func FetchCFZones(email, apiKey, apiToken string) ([]Zone, error) {
	ctx := context.Background()
	var api *cloudflare.API
	var err error

	if apiToken != "" {
		api, err = newCFClientWithToken(apiToken)
	} else {
		api, err = newCFClient(email, apiKey)
	}
	if err != nil {
		return nil, fmt.Errorf("创建 CF 客户端失败: %v", err)
	}

	zones, err := api.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取域名列表失败: %v", err)
	}

	var result []Zone
	for _, z := range zones {
		ns := ""
		if len(z.NameServers) > 0 {
			ns = strings.Join(z.NameServers, ",")
		}
		result = append(result, Zone{
			ID:     z.ID,
			Name:   z.Name,
			Status: string(z.Status),
			Plan:   z.Plan.Name,
			NS:     ns,
		})
	}
	return result, nil
}

func GetCFClientForAccount(account *Account) (*cloudflare.API, error) {
	token, err := Decrypt(account.APIToken)
	if err == nil && token != "" {
		return newCFClientWithToken(token)
	}
	key, err := Decrypt(account.APIKey)
	if err != nil {
		return nil, fmt.Errorf("解密凭证失败: %v", err)
	}
	return newCFClient(account.Email, key)
}

func CreateDNSRecord(account *Account, zoneID, name, content string) (string, error) {
	api, err := GetCFClientForAccount(account)
	if err != nil {
		return "", err
	}
	ctx := context.Background()

	rec := cloudflare.CreateDNSRecordParams{
		Type:    "TXT",
		Name:    name,
		Content: content,
		TTL:     120,
	}
	r, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), rec)
	if err != nil {
		return "", fmt.Errorf("创建 DNS 记录失败: %v", err)
	}
	return r.ID, nil
}

func DeleteDNSRecord(account *Account, zoneID, recordID string) error {
	api, err := GetCFClientForAccount(account)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return api.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), recordID)
}

func TestCloudflareReachability() error {
	client := &http.Client{Timeout: 10 * time.Second}
	if GetSetting("proxyEnabled") == "true" {
		proxyAddr := GetSetting("proxyAddr")
		if proxyAddr != "" {
			if proxyURL, err := url.Parse(proxyAddr); err == nil {
				client.Transport = &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
			}
		}
	}
	resp, err := client.Get("https://api.cloudflare.com/cdn-cgi/trace")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func GetZoneIDByName(account *Account, domain string) (string, error) {
	api, err := GetCFClientForAccount(account)
	if err != nil {
		return "", err
	}
	ctx := context.Background()

	// Extract root domain
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("无效域名: %s", domain)
	}
	rootDomain := strings.Join(parts[len(parts)-2:], ".")

	zones, err := api.ListZones(ctx)
	if err != nil {
		return "", err
	}
	for _, z := range zones {
		if z.Name == rootDomain {
			return z.ID, nil
		}
	}
	return "", fmt.Errorf("未找到域名 %s 对应的 Zone", rootDomain)
}
