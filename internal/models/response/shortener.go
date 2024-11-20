package response

type Shortener struct {
	Result string `json:"result"`
}

type Batch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
