package schema

// StoreItem represents an item of store
type StoreItem struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Image       string            `json:"image"`
	Name        string            `json:"name"`
	Fprocess    string            `json:"fprocess"`
	Network     string            `json:"network"`
	RepoURL     string            `json:"repo_url"`
	Environment map[string]string `json:"environment"`
}
