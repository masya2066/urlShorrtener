package response

type Shortener struct {
	Result string `json:"result"`
}

type Batch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
