package app

import "time"

type Account struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email"`
	APIKey    string    `json:"apiKey"` // encrypted
	APIToken  string    `json:"apiToken"` // encrypted
	CreatedAt time.Time `json:"createdAt"`
}

type Zone struct {
	ID        string `json:"id" gorm:"primaryKey"`
	AccountID string `json:"accountId"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Plan      string `json:"plan"`
	NS        string `json:"ns"` // comma-separated
}

type Certificate struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID   string    `json:"accountId"`
	Domain      string    `json:"domain"`
	SANList     string    `json:"sanList"` // comma-separated
	SANCount    int       `json:"sanCount"`
	CABrand     string    `json:"ca"`
	KeyAlgo     string    `json:"keyAlgo"`
	IssuedAt    time.Time `json:"issuedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	FilePath    string    `json:"filePath"`
	CertPEM     string    `json:"certPem"`
	KeyPEM      string    `json:"keyPem"`
	ChainPEM    string    `json:"chainPem"`
	FullChainPEM string   `json:"fullChainPem"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Setting struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value string `json:"value"`
}
