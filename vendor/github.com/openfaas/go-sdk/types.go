package sdk

import "github.com/openfaas/faas-provider/types"

type SystemInfo struct {
	Arch     string            `json:"arch,omitempty"`
	Provider Provider          `json:"provider,omitempty"`
	Version  types.VersionInfo `json:"version,omitempty"`
}

type Provider struct {
	Provider      string            `json:"provider,omitempty"`
	Version       types.VersionInfo `json:"version,omitempty"`
	Orchestration string            `json:"orchestration,omitempty"`
}
