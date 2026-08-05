package model

type Click struct {
	Code      string `json:"code"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
}

type ClickInfo struct {
	ClickedAt string `json:"clicked_at"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
}

type Analytics struct {
	Code        string      `json:"code"`
	TotalClicks int        `json:"total_clicks"`
	RecentClicks []ClickInfo `json:"recent_clicks"`
}