package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/strand1/fernwood/pkg/memory"
)

// MulchQueryTool retrieves full expertise for a domain via mulch prime.
type MulchQueryTool struct {
	mulchMgr *memory.MulchManager
}

func NewMulchQueryTool(mulchMgr *memory.MulchManager) *MulchQueryTool {
	return &MulchQueryTool{
		mulchMgr: mulchMgr,
	}
}

func (t *MulchQueryTool) Name() string {
	return "mulch_query"
}

func (t *MulchQueryTool) Description() string {
	return "Retrieve full expertise content for a specific domain from the mulch knowledge base. Use this when you need detailed information on a topic covered by an expertise domain."
}

func (t *MulchQueryTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"domain": map[string]any{
				"type":        "string",
				"description": "The domain name to query (e.g., 'code', 'errors', 'decisions')",
			},
		},
		"required": []string{"domain"},
	}
}

func (t *MulchQueryTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.mulchMgr == nil || !t.mulchMgr.Enabled {
		return &ToolResult{
			ForLLM: "mulch is not enabled",
			IsError: true,
		}
	}

	domain, ok := args["domain"].(string)
	if !ok {
		return &ToolResult{
			ForLLM: "domain parameter is required and must be a string",
			IsError: true,
		}
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return &ToolResult{
			ForLLM: "domain cannot be empty",
			IsError: true,
		}
	}

	// Execute mulch prime for the specific domain
	cmd := t.mulchMgr.BinPath
	argsList := []string{"prime", domain}
	cmdObj := t.mulchMgr

	execCmd := exec.Command(cmd, argsList...)
	if cmdObj.WorkingDir != "" {
		execCmd.Dir = cmdObj.WorkingDir
	}

	out, err := execCmd.Output()
	if err != nil {
		msg := fmt.Sprintf("mulch prime %s failed: %v", domain, err)
		return &ToolResult{
			ForLLM: msg,
			IsError: true,
			Err:     err,
		}
	}

	content := strings.TrimSpace(string(out))
	if content == "" {
		return &ToolResult{
			ForLLM: fmt.Sprintf("mulch prime returned no content for domain '%s'", domain),
			IsError: false,
		}
	}

	// Return the full expertise content
	return &ToolResult{
		ForLLM: fmt.Sprintf("=== Expertise: %s ===\n\n%s", domain, content),
		Silent: false,
	}
}

// Ensure MulchQueryTool implements Tool interface at compile time.
var _ Tool = (*MulchQueryTool)(nil)
