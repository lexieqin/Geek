package tools

import (
	"github.com/lexieqin/Geek/GenesisGpt/cmd/utils"
)

// ClusterTool represents a tool for listing k8s cluster commands.
type ClusterTool struct {
	Name        string
	Description string
}

// NewClusterTool creates a new ClusterTool instance.
func NewClusterTool() *ClusterTool {
	return &ClusterTool{
		Name:        "ClusterTool",
		Description: "Used to list cluster information",
	}
}

// Run executes the command and returns the output.
func (l *ClusterTool) Run() (string, error) {

	url := "http://localhost:8081/clusters"

	s, err := utils.GetHTTP(url)

	return s, err
}
