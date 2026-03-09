package memory

import (
	"os/exec"
	"strings"
)

type MulchManager struct {
	BinPath   string
	Enabled   bool
	Domains   []string
}

func NewMulchManager(binPath string, enabled bool, domains []string) *MulchManager {
	if binPath == "" {
		binPath = "mulch"
	}
	if len(domains) == 0 {
		domains = []string{"code", "errors", "decisions"}
	}
	return &MulchManager{BinPath: binPath, Enabled: enabled, Domains: domains}
}

// Init ensures .mulch/ exists in the current directory. Safe to call multiple times.
func (m *MulchManager) Init() error {
	if !m.Enabled {
		return nil
	}
	return exec.Command(m.BinPath, "init").Run()
}

// Prime returns expertise context to inject into the system prompt.
// Returns empty string silently if mulch is not initialized or not enabled.
func (m *MulchManager) Prime() string {
	if !m.Enabled {
		return ""
	}
	out, err := exec.Command(m.BinPath, "prime").Output()
	if err != nil {
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
	return exec.Command(m.BinPath, "record", domain, "--type", recordType, content).Run()
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
