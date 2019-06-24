package schema

// StoreItem represents an item of store
type StoreItem struct {
	Icon                   string            `json:"icon"`
	Title                  string            `json:"title"`
	Description            string            `json:"description"`
	Image                  string            `json:"image"`
	Name                   string            `json:"name"`
	Fprocess               string            `json:"fprocess"`
	Network                string            `json:"network"`
	RepoURL                string            `json:"repo_url"`
	Environment            map[string]string `json:"environment"`
	Labels                 map[string]string `json:"labels"`
	Annotations            map[string]string `json:"annotations"`
	ReadOnlyRootFilesystem bool              `json:"readOnlyRootFilesystem"`
}

//StoreFunction represents a multi-arch function in the store
type StoreFunction struct {
	Icon                   string            `json:"icon"`
	Title                  string            `json:"title"`
	Description            string            `json:"description"`
	Name                   string            `json:"name"`
	Fprocess               string            `json:"fprocess"`
	Network                string            `json:"network"`
	RepoURL                string            `json:"repo_url"`
	ReadOnlyRootFilesystem bool              `json:"readOnlyRootFilesystem"`
	Environment            map[string]string `json:"environment"`
	Labels                 map[string]string `json:"labels"`
	Annotations            map[string]string `json:"annotations"`
	Images                 map[string]string `json:"images"`
}

//GetImageName get image name of function for a platform
func (s *StoreFunction) GetImageName(platform string) string {
	imageName, _ := s.Images[platform]
	return imageName
}

// StoreV2 represents an item of store for version 2
type StoreV2 struct {
	Version   string          `json:"version"`
	Functions []StoreFunction `json:"functions"`
}
