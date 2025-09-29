package poller

// RSSFeed represents the App Store RSS feed structure
type RSSFeed struct {
	Feed struct {
		Entry []RSSEntry `json:"entry"`
	} `json:"feed"`
}

type RSSEntry struct {
	Author struct {
		Name struct {
			Label string `json:"label"`
		} `json:"name"`
	} `json:"author"`
	Content struct {
		Label string `json:"label"`
	} `json:"content"`
	Rating struct {
		Label string `json:"label"` // "1" to "5"
	} `json:"im:rating"`
	Updated struct {
		Label string `json:"label"` // ISO 8601 timestamp
	} `json:"updated"`
	ID struct {
		Label string `json:"label"`
	} `json:"id"`
}
