package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/strand1/fernwood/pkg/memory"
)

// MulchRecordTool records learnings into mulch domains.
type MulchRecordTool struct {
	mulchMgr *memory.MulchManager
}

func NewMulchRecordTool(mulchMgr *memory.MulchManager) *MulchRecordTool {
	return &MulchRecordTool{
		mulchMgr: mulchMgr,
	}
}

func (t *MulchRecordTool) Name() string {
	return "mulch_record"
}

func (t *MulchRecordTool) Description() string {
	return "Record a learning into the mulch knowledge base. Automatically creates domains if they don't exist."
}

func (t *MulchRecordTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"domain": map[string]any{
				"type":        "string",
				"description": "Domain name (e.g., 'code', 'errors', 'prices')",
			},
			"type": map[string]any{
				"type":        "string",
				"description": "Learning type: convention, pattern, failure, decision, reference, guide",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The learning content to record",
			},
		},
		"required": []string{"domain", "type", "content"},
	}
}

func (t *MulchRecordTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.mulchMgr == nil || !t.mulchMgr.Enabled {
		return &ToolResult{
			ForLLM: "mulch is not enabled",
			IsError: true,
		}
	}

	domain, ok := args["domain"].(string)
	if !ok || strings.TrimSpace(domain) == "" {
		return &ToolResult{
			ForLLM: "domain is required and must be a non-empty string",
			IsError: true,
		}
	}
	domain = strings.TrimSpace(domain)

	recType, ok := args["type"].(string)
	if !ok || strings.TrimSpace(recType) == "" {
		return &ToolResult{
			ForLLM: "type is required and must be one of: convention, pattern, failure, decision, reference, guide",
			IsError: true,
		}
	}
	recType = strings.ToLower(strings.TrimSpace(recType))

	content, ok := args["content"].(string)
	if !ok || strings.TrimSpace(content) == "" {
		return &ToolResult{
			ForLLM: "content is required and must be a non-empty string",
			IsError: true,
		}
	}

	// Ensure domain exists
	if err := t.ensureDomain(domain); err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("failed to ensure domain exists: %v", err),
			IsError: true,
			Err:     err,
		}
	}

	// Build command based on type
	baseArgs := []string{"record", domain}
	switch recType {
	case "convention":
		baseArgs = append(baseArgs, "--type", "convention", "--description", content)
	case "pattern":
		// For pattern, extract name from first line
		name := t.firstLine(content)
		baseArgs = append(baseArgs, "--type", "pattern", "--name", name, "--description", content)
	case "failure":
		baseArgs = append(baseArgs, "--type", "failure", "--description", content, "--resolution", content)
	case "decision":
		title := t.firstLine(content)
		baseArgs = append(baseArgs, "--type", "decision", "--title", title, "--rationale", content)
	case "reference":
		name := t.firstLine(content)
		baseArgs = append(baseArgs, "--type", "reference", "--name", name, "--description", content)
	case "guide":
		baseArgs = append(baseArgs, "--type", "guide", "--description", content)
	default:
		return &ToolResult{
			ForLLM: fmt.Sprintf("invalid type: %s (must be convention, pattern, failure, decision, reference, or guide)", recType),
			IsError: true,
		}
	}

	cmd := exec.Command(t.mulchMgr.BinPath, baseArgs...)
	if t.mulchMgr.WorkingDir != "" {
		cmd.Dir = t.mulchMgr.WorkingDir
	}
	if err := cmd.Run(); err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("mulch record failed: %v", err),
			IsError: true,
			Err:     err,
		}
	}

	return &ToolResult{
		ForLLM: fmt.Sprintf("Recorded %s learning to domain '%s'", recType, domain),
		Silent: false,
	}
}

// ensureDomain creates the domain if it doesn't exist.
func (t *MulchRecordTool) ensureDomain(domain string) error {
	// Check if domain exists by trying to list or by checking data dir.
	// Simpler: just run `mulch add <domain>` and ignore error if domain exists.
	cmd := exec.Command(t.mulchMgr.BinPath, "add", domain)
	if t.mulchMgr.WorkingDir != "" {
		cmd.Dir = t.mulchMgr.WorkingDir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		// If error output contains "already exists", that's fine.
		if strings.Contains(string(out), "already exists") || strings.Contains(string(out), "exists") {
			return nil
		}
		return fmt.Errorf("add domain failed: %v (output: %s)", err, string(out))
	}
	return nil
}

// firstLine returns the first non-empty line of content, truncated to a reasonable length.
func (t *MulchRecordTool) firstLine(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if line != "" {
			if len(line) > 100 {
				line = line[:100] + "..."
			}
			return line
		}
	}
	return "untitled"
}

// Ensure MulchRecordTool implements Tool interface at compile time.
var _ Tool = (*MulchRecordTool)(nil)
