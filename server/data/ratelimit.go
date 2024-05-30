package data

type RateLimit struct {
	Day    RateLimitType `json:"day"`
	Hour   RateLimitType `json:"hour"`
	Minute RateLimitType `json:"minute"`
}

type RateLimitType struct {
	LastUpdate string `json:"lastUpdate"`
	Request    int64  `json:"request"`
	Token      int64  `json:"token"`
}
