package memory

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/strand1/fernwood/pkg/providers"
)

// LLMSummarizer defines the interface for summarizing domain content.
type LLMSummarizer interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// MulchManager manages mulch operations for expertise domains.
type MulchManager struct {
	BinPath    string
	Enabled    bool
	Domains    []string
	WorkingDir string
}

func NewMulchManager(workingDir, binPath string, enabled bool, domains []string) *MulchManager {
	if binPath == "" {
		binPath = "mulch"
	}
	if len(domains) == 0 {
		domains = []string{"code", "errors", "decisions"}
	}
	return &MulchManager{
		BinPath:    binPath,
		Enabled:    enabled,
		Domains:    domains,
		WorkingDir: workingDir,
	}
}

// Init ensures .mulch/ exists in the workspace. Safe to call multiple times.
func (m *MulchManager) Init() error {
	if !m.Enabled {
		return nil
	}
	cmd := exec.Command(m.BinPath, "init")
	if m.WorkingDir != "" {
		cmd.Dir = m.WorkingDir
	}
	if err := cmd.Run(); err != nil {
		log.Printf("[mulch] init failed: %v", err)
		return err
	}
	return nil
}

// Prime returns expertise context to inject into the system prompt.
// It primes only the configured domains. Returns empty string on error or if no output.
func (m *MulchManager) Prime() string {
	if !m.Enabled {
		return ""
	}
	args := []string{"prime"}
	// Append domains if configured
	if len(m.Domains) > 0 {
		args = append(args, m.Domains...)
	}
	cmd := exec.Command(m.BinPath, args...)
	if m.WorkingDir != "" {
		cmd.Dir = m.WorkingDir
	}
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[mulch] prime failed: %v", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Record writes a structured learning back to mulch.
type MulchRecord struct {
	Domain  string `json:"domain"`
	Type    string `json:"type"`    // convention, failure, decision, pattern
	Content string `json:"content"`
}

func (m *MulchManager) Record(domain, recordType, content string) error {
	if !m.Enabled || content == "" {
		return nil
	}
	cmd := exec.Command(m.BinPath, "record", domain, "--type", recordType, content)
	if m.WorkingDir != "" {
		cmd.Dir = m.WorkingDir
	}
	if err := cmd.Run(); err != nil {
		log.Printf("[mulch] record failed: %v", err)
		return err
	}
	return nil
}

// RecordBatch records multiple learnings, typically called at session end.
func (m *MulchManager) RecordBatch(records []MulchRecord) {
	for _, r := range records {
		_ = m.Record(r.Domain, r.Type, r.Content)
	}
}

// IsAvailable checks if the mulch binary exists and is callable.
func (m *MulchManager) IsAvailable() bool {
	_, err := exec.LookPath(m.BinPath)
	return err == nil
}

// ListDomains returns a list of all available domains from mulch status.
func (m *MulchManager) ListDomains() ([]string, error) {
	if !m.Enabled {
		return []string{}, nil
	}
	cmd := exec.Command(m.BinPath, "status")
	if m.WorkingDir != "" {
		cmd.Dir = m.WorkingDir
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("mulch status failed: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	domains := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Only consider lines that look like domain entries: "<name>: <N> records"
		// Skip headers, separators (===), empty lines, etc.
		if !strings.Contains(line, " records") {
			continue
		}
		// Extract the part before the colon, then trim space and any trailing colon
		idx := strings.IndexByte(line, ':')
		if idx <= 0 {
			continue
		}
		name := strings.TrimSpace(line[:idx])
		// Remove trailing colon if present (defensive)
		name = strings.TrimSuffix(name, ":")
		if name == "" {
			continue
		}
		domains = append(domains, name)
	}
	return domains, nil
}

// summariesDir returns the path to .mulch/summaries directory.
func (m *MulchManager) summariesDir() string {
	if m.WorkingDir != "" {
		return filepath.Join(m.WorkingDir, ".mulch", "summaries")
	}
	return filepath.Join(".mulch", "summaries")
}

// summaryPath returns the path to the summary file for a domain.
func (m *MulchManager) summaryPath(domain string) string {
	return filepath.Join(m.summariesDir(), domain+".md")
}

// domainDataPath returns the path to the domain JSONL data file.
func (m *MulchManager) domainDataPath(domain string) string {
	if m.WorkingDir != "" {
		return filepath.Join(m.WorkingDir, ".mulch", "domains", domain+".jsonl")
	}
	return filepath.Join(".mulch", "domains", domain+".jsonl")
}

// IsSummaryStale checks if the summary for a domain needs refreshing.
// A summary is stale if:
// - The summary file doesn't exist
// - The domain JSONL mtime is newer than the summary mtime
func (m *MulchManager) IsSummaryStale(domain string) (bool, error) {
	summaryPath := m.summaryPath(domain)
	info, err := os.Stat(summaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil // No summary exists, needs to be created
		}
		return true, fmt.Errorf("error checking summary: %w", err)
	}

	dataPath := m.domainDataPath(domain)
	dataInfo, err := os.Stat(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("domain data file doesn't exist: %s", dataPath)
		}
		return true, fmt.Errorf("error checking domain data: %w", err)
	}

	// Summary is stale if data file is newer than summary
	return dataInfo.ModTime().After(info.ModTime()), nil
}

// SummarizeDomain calls mulch prime for the domain and sends the output to the LLM
// to generate a concise 1-2 sentence summary. The summary is written to .mulch/summaries/<domain>.md.
func (m *MulchManager) SummarizeDomain(ctx context.Context, domain string, llmClient LLMSummarizer) error {
	if !m.Enabled {
		return nil
	}

	// Prime the domain to get full expertise
	args := []string{"prime", domain}
	cmd := exec.Command(m.BinPath, args...)
	if m.WorkingDir != "" {
		cmd.Dir = m.WorkingDir
	}
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("mulch prime %s failed: %w", domain, err)
	}

	content := strings.TrimSpace(string(out))
	if content == "" {
		return fmt.Errorf("mulch prime returned empty output for domain %s", domain)
	}

	// Build prompt for LLM
	prompt := fmt.Sprintf(`Summarize this expertise domain in 1-2 sentences for use as a compact context index. Be specific about what topics are covered.

Domain: %s

Expertise content:
%s
`, domain, content)

	// Get summary from LLM
	summary, err := llmClient.Complete(ctx, prompt)
	if err != nil {
		return fmt.Errorf("LLM summarization failed: %w", err)
	}

	summary = strings.TrimSpace(summary)
	if summary == "" {
		return fmt.Errorf("LLM returned empty summary for domain %s", domain)
	}

	// Ensure summaries directory exists
	summariesDir := m.summariesDir()
	if err := os.MkdirAll(summariesDir, 0o755); err != nil {
		return fmt.Errorf("failed to create summaries directory: %w", err)
	}

	// Write summary file
	summaryPath := m.summaryPath(domain)
	if err := os.WriteFile(summaryPath, []byte(summary), 0o644); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	log.Printf("[mulch] Summarized domain %s -> %s", domain, summaryPath)
	return nil
}

// SummarizeDomains iterates all domains, checks which summaries are stale,
// and only refreshes stale ones. Returns lists of refreshed and skipped domains.
func (m *MulchManager) SummarizeDomains(ctx context.Context, llmClient LLMSummarizer) (refreshed []string, skipped []string, err error) {
	if !m.Enabled {
		return refreshed, skipped, nil
	}

	// Get list of domains (use configured domains if mulch status fails)
	domains := m.Domains
	list, err := m.ListDomains()
	if err != nil {
		log.Printf("[mulch] Warning: failed to list domains, using configured list: %v", err)
	} else if len(list) > 0 {
		domains = list
	}

	for _, domain := range domains {
		stale, err := m.IsSummaryStale(domain)
		if err != nil {
			// If we can't determine staleness (e.g., data file missing), skip
			log.Printf("[mulch] Skipping domain %s: %v", domain, err)
			skipped = append(skipped, domain)
			continue
		}
		if !stale {
			skipped = append(skipped, domain)
			continue
		}

		// Refresh stale summary
		if err := m.SummarizeDomain(ctx, domain, llmClient); err != nil {
			log.Printf("[mulch] Failed to summarize domain %s: %v", domain, err)
			skipped = append(skipped, domain)
			continue
		}
		refreshed = append(refreshed, domain)
	}

	return refreshed, skipped, nil
}

// LoadDomainIndex reads all summary .md files from .mulch/summaries/ and returns
// a formatted string listing all domains and their summaries.
func (m *MulchManager) LoadDomainIndex() string {
	if !m.Enabled {
		return ""
	}

	summariesDir := m.summariesDir()
	entries, err := os.ReadDir(summariesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		log.Printf("[mulch] Error reading summaries directory: %v", err)
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Expertise Domains\n\n")

	// Read each .md file and format as "- domain: summary text"
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			domain := strings.TrimSuffix(entry.Name(), ".md")
			dataPath := filepath.Join(summariesDir, entry.Name())
			content, err := os.ReadFile(dataPath)
			if err != nil {
				continue
			}
			summary := strings.TrimSpace(string(content))
			if summary != "" {
				fmt.Fprintf(&sb, "- %s: %s\n", domain, summary)
			}
		}
	}

	index := sb.String()
	if index == "## Expertise Domains\n\n" {
		return ""
	}
	return index
}

// providerSummarizer adapts a providers.LLMProvider to the LLMSummarizer interface.
type providerSummarizer struct {
	provider providers.LLMProvider
	model    string
}

// NewProviderSummarizer creates an LLMSummarizer that uses the given LLM provider.
func NewProviderSummarizer(provider providers.LLMProvider, model string) LLMSummarizer {
	return &providerSummarizer{
		provider: provider,
		model:    model,
	}
}

func (s *providerSummarizer) Complete(ctx context.Context, prompt string) (string, error) {
	// Use same pattern as autoRecordLearnings in loop.go
	messages := []providers.Message{{Role: "user", Content: prompt}}
	opts := map[string]any{
		"max_tokens":       2000,
		"temperature":      0.3,
		"prompt_cache_key": "mulch_summarize",
	}
	resp, err := s.provider.Chat(ctx, messages, nil, s.model, opts)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
