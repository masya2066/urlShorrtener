package request

type Shortener struct {
	URL string `json:"url"`
}

type Batch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
