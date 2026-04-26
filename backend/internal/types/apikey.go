package types

type CreateAPIKeyReq struct {
	Name      string  `json:"name" binding:"required"`
	Quota     float64 `json:"quota"`
	Unlimited bool    `json:"unlimited"`
	Models    string  `json:"models"`
	ExpiresAt string  `json:"expires_at"` // 可选，格式：2025-12-31
}

type CreateAPIKeyResp struct {
	ID        int64   `json:"id"`
	Key       string  `json:"key"`
	Name      string  `json:"name"`
	Quota     float64 `json:"quota"`
	Unlimited bool    `json:"unlimited"`
	Models    string  `json:"models"`
	ExpiresAt string  `json:"expires_at"`
}
