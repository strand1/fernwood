package tools

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/strand1/fernwood/pkg/memory"
	"github.com/strand1/fernwood/pkg/providers"
)

// MulchRecordTool records learnings into mulch domains.
type MulchRecordTool struct {
	mulchMgr       *memory.MulchManager
	provider       providers.LLMProvider
	model          string
	forceDomainList []string // test-only override
}

func NewMulchRecordTool(mulchMgr *memory.MulchManager, provider providers.LLMProvider, model string) *MulchRecordTool {
	return &MulchRecordTool{
		mulchMgr: mulchMgr,
		provider: provider,
		model:    model,
	}
}

// SetTestDomainList overrides domain list for testing.
func (t *MulchRecordTool) SetTestDomainList(list []string) {
	t.forceDomainList = list
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

	// Phase 1 & 2: resolve domain
	resolvedDomain, isNew, err := t.resolveDomain(ctx, domain, content, recType)
	if err != nil {
		// Log but continue with original candidate (fail open)
		log.Printf("[mulch] domain resolution error: %v", err)
		resolvedDomain = domain
		isNew = true
	}

	// Ensure domain exists if it's new (or if resolution created a new match)
	if isNew {
		if err := t.ensureDomain(resolvedDomain); err != nil {
			return &ToolResult{
				ForLLM:  fmt.Sprintf("failed to ensure domain exists: %v", err),
				IsError: true,
				Err:     err,
			}
		}
	}

	// Build command based on type
	baseArgs := []string{"record", resolvedDomain}
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
		name := t.firstLine(content)
		baseArgs = append(baseArgs, "--type", "guide", "--name", name, "--description", content)
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
		ForLLM: fmt.Sprintf("Recorded %s learning to domain '%s'", recType, resolvedDomain),
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

// normalizeDomainName normalizes a domain name for comparison.
// - Lowercase
// - Remove '-', '_', spaces
// - Strip common suffixes: errors, patterns, notes, tips, stuff, things
func normalizeDomainName(s string) string {
	s = strings.ToLower(s)
	// Remove separators
	s = strings.NewReplacer("-", "", "_", "", " ", "").Replace(s)
	// Strip suffixes
	suffixes := []string{"errors", "patterns", "notes", "tips", "stuff", "things"}
	for _, suffix := range suffixes {
		s = strings.TrimSuffix(s, suffix)
	}
	return s
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Create two rows for DP
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			// Minimum of deletion, insertion, substitution
			deletion := prev[j] + 1
			insertion := curr[j-1] + 1
			substitution := prev[j-1] + cost
			// Find minimum of three
			m := deletion
			if insertion < m {
				m = insertion
			}
			if substitution < m {
				m = substitution
			}
			curr[j] = m
		}
		// Swap prev and curr
		prev, curr = curr, prev
	}

	return prev[lb]
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// resolveDomain finds the best matching existing domain for the candidate.
// Returns the resolved domain name and whether it's a new domain (true if candidate is new).
func (t *MulchRecordTool) getDomains() ([]string, error) {
	if len(t.forceDomainList) > 0 {
		return t.forceDomainList, nil
	}
	return t.mulchMgr.ListDomains()
}

func (t *MulchRecordTool) resolveDomain(
	ctx context.Context,
	candidate string,
	content string,
	recordType string,
) (string, bool, error) {
	// Get list of existing domains
	domains, err := t.getDomains()
	if err != nil {
		// On error, fail open: use candidate as is
		log.Printf("[mulch] failed to list domains: %v", err)
		return candidate, true, nil
	}

	// Check verbatim match first
	for _, d := range domains {
		if d == candidate {
			return candidate, false, nil // found, not new
		}
	}

	// Phase 1: String similarity
	normCandidate := normalizeDomainName(candidate)
	bestMatch := ""
	bestDist := 999
	for _, d := range domains {
		normD := normalizeDomainName(d)

		// Check if after normalization they are identical
		if normCandidate == normD {
			// Perfect match after normalization
			return d, false, nil
		}

		// Compute edit distance on normalized strings
		dist := levenshtein(normCandidate, normD)
		if dist < bestDist {
			bestDist = dist
			bestMatch = d
		}

		// Prefix match: if one is a prefix of the other and at least 4 chars, count as a direct match
		if len(normCandidate) >= 4 && len(normD) >= 4 {
			if strings.HasPrefix(normCandidate, normD) || strings.HasPrefix(normD, normCandidate) {
				// Strong match: return this domain immediately
				return d, false, nil
			}
		}
	}

	// If we have a close match (distance <= 2), use it
	if bestDist <= 2 && bestMatch != "" {
		log.Printf("[mulch] domain resolution: %q -> %q (phase1, dist=%d)", candidate, bestMatch, bestDist)
		return bestMatch, false, nil
	}

	// Phase 2: LLM judgment (only if we have a provider)
	if t.provider != nil {
		// Build prompt
		prompt := fmt.Sprintf(`You are helping maintain a knowledge base. A new learning needs to be recorded.

Candidate domain: "%s"
Learning type: %s
Learning summary: %s

Existing domains:
%s

Does any existing domain clearly cover this topic? Reply with ONLY one of:
- The exact existing domain name (if it's a clear fit)
- NONE (if no existing domain fits and a new one is warranted)

Do not explain. One word or domain name only.`,
			candidate,
			recordType,
			truncate(content, 100),
			strings.Join(domains, "\n"),
		)

		// Call LLM with short timeout
		llmCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		messages := []providers.Message{{Role: "user", Content: prompt}}
		opts := map[string]any{
			"max_tokens":     20,
			"temperature":    0.0,
			"prompt_cache":   "disable",
		}
		resp, err := t.provider.Chat(llmCtx, messages, nil, t.model, opts)
		if err != nil {
			// Fail open
			log.Printf("[mulch] domain resolution LLM call failed: %v; creating new domain", err)
			return candidate, true, nil
		}

		answer := strings.TrimSpace(resp.Content)
		if answer == "" || answer == "NONE" || answer == "none" {
			// Create new domain
			log.Printf("[mulch] domain resolution: creating new domain %q (llm confirmed)", candidate)
			return candidate, true, nil
		}

		// Validate that the returned domain exists in the index (could be stale but that's ok)
		for _, d := range domains {
			if d == answer {
				log.Printf("[mulch] domain resolution: %q -> %q (phase2)", candidate, answer)
				return answer, false, nil
			}
		}
		// LLM returned something not in the list - treat as NONE
		return candidate, true, nil
	}

	// No provider configured; treat as new domain
	return candidate, true, nil
}

// truncate limits a string to max length, adding ellipsis if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

