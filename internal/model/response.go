package model

type ShortUrlResponse struct {
	//Key      string `json:"key"`
	LongUrl  string `json:"long_url"`
	ShortUrl string `json:"short_url"`
}
