package schema

type KubernetesSecret struct {
	Kind       string                   `json:"kind"`
	ApiVersion string                   `json:"apiVersion"`
	Metadata   KubernetesSecretMetadata `json:"metadata"`
	Data       map[string]string        `json:"data"`
}

type KubernetesSecretMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

//Secret secret type
type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}
