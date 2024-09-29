package shared

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Url   string `json:"url"`
}
