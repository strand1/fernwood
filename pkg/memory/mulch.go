package memory

import (
	"log"
	"os/exec"
	"strings"
)

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
