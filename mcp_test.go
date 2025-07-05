package taskcp

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

func TestRegisterMCPTools(t *testing.T) {
	service := New()

	s := server.NewMCPServer("Test Server", "1.0.0")

	err := service.RegisterMCPTools(s)
	if err != nil {
		t.Fatalf("Failed to register MCP tools: %v", err)
	}

}