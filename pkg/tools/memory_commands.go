// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/strand1/fernwood/pkg/session"
)

// RegisterMemoryCommands registers all memory commands to the registry.
// These commands wrap the Mulch CLI for all memory operations.
// workspace: directory where .mulch/ lives (mulch commands run from here)
func RegisterMemoryCommands(registry *CommandRegistry, workspace string) {
	// memory store - Record a fact/learning
	registry.Register("memory store", "Store a fact or learning (alias: memory record)", func(args []string, stdin string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("memory store: usage: memory store <domain> <content>")
		}
		domain := args[0]
		content := strings.Join(args[1:], " ")
		if stdin != "" {
			content = stdin
		}
		return cmdMemoryStore(workspace, domain, "convention", content)
	})

	// memory record - Record with explicit type and optional flags
	registry.Register("memory record", "Record a learning with type (memory record <domain> <type> [--title X] [--rationale Y] [--name X] [--description Y] [--resolution Y] <content>)", func(args []string, stdin string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("memory record: usage: memory record <domain> <type> [--title X] [--rationale Y] [--name X] [--description Y] [--resolution Y] <content>")
		}
		domain := args[0]
		recType := args[1]

		// Parse flags and remaining content
		var title, rationale, name, description, resolution, content string
		var remainingArgs []string

		for i := 2; i < len(args); i++ {
			switch args[i] {
			case "--title":
				if i+1 < len(args) {
					title = args[i+1]
					i++
				}
			case "--rationale":
				if i+1 < len(args) {
					rationale = args[i+1]
					i++
				}
			case "--name":
				if i+1 < len(args) {
					name = args[i+1]
					i++
				}
			case "--description":
				if i+1 < len(args) {
					description = args[i+1]
					i++
				}
			case "--resolution":
				if i+1 < len(args) {
					resolution = args[i+1]
					i++
				}
			default:
				remainingArgs = append(remainingArgs, args[i])
			}
		}

		// Join remaining args as content
		if len(remainingArgs) > 0 {
			content = strings.Join(remainingArgs, " ")
		}
		if stdin != "" {
			content = stdin
		}

		return cmdMemoryRecordWithFlags(workspace, domain, recType, title, rationale, name, description, resolution, content)
	})

	// memory facts - List facts in a domain
	registry.Register("memory facts", "List facts/records in a domain", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("memory facts: usage: memory facts <domain>")
		}
		return cmdMemoryFacts(workspace, args[0])
	})

	// memory search - Search across all domains
	registry.Register("memory search", "Search across all memory domains", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("memory search: usage: memory search <query>")
		}
		query := strings.Join(args, " ")
		if stdin != "" {
			query = stdin
		}
		return cmdMemorySearch(workspace, query)
	})

	// memory query - Query a specific domain
	registry.Register("memory query", "Query a specific memory domain", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("memory query: usage: memory query <domain>")
		}
		return cmdMemoryQuery(workspace, args[0])
	})

	// memory forget - Delete a record
	registry.Register("memory forget", "Delete a memory record by ID", func(args []string, stdin string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("memory forget: usage: memory forget <domain> <id>")
		}
		return cmdMemoryForget(workspace, args[0], args[1])
	})

	// memory status - Show mulch status
	registry.Register("memory status", "Show mulch status (domains, record counts)", func(args []string, stdin string) (string, error) {
		return cmdMemoryStatus(workspace)
	})

	// Aliases
	registry.RegisterAlias("mem.store", "memory store")
	registry.RegisterAlias("mem.record", "memory record")
	registry.RegisterAlias("mem.facts", "memory facts")
	registry.RegisterAlias("mem.search", "memory search")
	registry.RegisterAlias("mem.query", "memory query")
	registry.RegisterAlias("mem.forget", "memory forget")
	registry.RegisterAlias("mem.status", "memory status")
}

// cmdMemoryRecordWithFlags records a learning via mulch record with proper flag handling per type
func cmdMemoryRecordWithFlags(workspace, domain, recType, title, rationale, name, description, resolution, content string) (string, error) {
	// Validate record type
	validTypes := []string{"convention", "pattern", "failure", "decision", "reference", "guide"}
	recType = strings.ToLower(recType)
	valid := false
	for _, t := range validTypes {
		if recType == t {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("memory record: invalid type '%s'. Valid types: %s", recType, strings.Join(validTypes, ", "))
	}

	// Build mulch record command based on type requirements
	// Per Mulch docs:
	//   convention: [content] or --description
	//   pattern: --name, --description (or [content])
	//   failure: --description, --resolution
	//   decision: --title, --rationale
	//   reference: --name, --description (or [content])
	//   guide: --name, --description (or [content])

	var cmdArgs []string
	cmdArgs = append(cmdArgs, "record", domain, "--type", recType)

	switch recType {
	case "decision":
		// Decision requires --title and --rationale
		if title == "" || rationale == "" {
			// Try to parse from content if it's JSON
			if parsedTitle, parsedRationale, ok := parseDecisionJSON(content); ok {
				title = parsedTitle
				rationale = parsedRationale
			}
		}
		if title == "" {
			return "", fmt.Errorf("memory record: decision records require --title")
		}
		if rationale == "" {
			return "", fmt.Errorf("memory record: decision records require --rationale")
		}
		cmdArgs = append(cmdArgs, "--title", title, "--rationale", rationale)

	case "failure":
		// Failure requires --description and --resolution
		if description == "" && content != "" {
			description = content
		}
		if resolution == "" {
			// Try to parse from content if it's JSON
			if parsedDesc, parsedRes, ok := parseFailureJSON(content); ok {
				description = parsedDesc
				resolution = parsedRes
			}
		}
		if description == "" {
			return "", fmt.Errorf("memory record: failure records require --description")
		}
		if resolution == "" {
			return "", fmt.Errorf("memory record: failure records require --resolution")
		}
		cmdArgs = append(cmdArgs, "--description", description, "--resolution", resolution)

	case "pattern", "reference", "guide":
		// These require --name and --description (or positional content)
		if name == "" || description == "" {
			// Try to parse from content if it's JSON
			if parsedName, parsedDesc, ok := parseNamedRecordJSON(content); ok {
				if name == "" {
					name = parsedName
				}
				if description == "" {
					description = parsedDesc
				}
			}
		}
		if name == "" && description == "" {
			// Use positional content
			if content == "" {
				return "", fmt.Errorf("memory record: %s records require --name and --description (or positional content)", recType)
			}
			cmdArgs = append(cmdArgs, content)
		} else {
			if name != "" {
				cmdArgs = append(cmdArgs, "--name", name)
			}
			if description != "" {
				cmdArgs = append(cmdArgs, "--description", description)
			}
		}

	case "convention":
		// Convention accepts [content] or --description
		if description != "" {
			cmdArgs = append(cmdArgs, "--description", description)
		} else if content != "" {
			cmdArgs = append(cmdArgs, content)
		} else {
			return "", fmt.Errorf("memory record: convention records require content or --description")
		}
	}

	// Execute the command
	cmd := exec.Command("mulch", cmdArgs...)
	cmd.Dir = workspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mulch record failed: %v\n%s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

// parseDecisionJSON tries to parse title and rationale from JSON content
func parseDecisionJSON(content string) (title, rationale string, ok bool) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "{") {
		return "", "", false
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", "", false
	}
	titleVal, hasTitle := data["title"]
	rationaleVal, hasRationale := data["rationale"]
	if hasTitle && hasRationale {
		if titleStr, ok := titleVal.(string); ok {
			if rationaleStr, ok := rationaleVal.(string); ok {
				return titleStr, rationaleStr, true
			}
		}
	}
	return "", "", false
}

// parseFailureJSON tries to parse description and resolution from JSON content
func parseFailureJSON(content string) (description, resolution string, ok bool) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "{") {
		return "", "", false
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", "", false
	}
	descVal, hasDesc := data["description"]
	resVal, hasRes := data["resolution"]
	if hasDesc && hasRes {
		if descStr, ok := descVal.(string); ok {
			if resStr, ok := resVal.(string); ok {
				return descStr, resStr, true
			}
		}
	}
	return "", "", false
}

// parseNamedRecordJSON tries to parse name and description from JSON content
func parseNamedRecordJSON(content string) (name, description string, ok bool) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "{") {
		return "", "", false
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", "", false
	}
	nameVal, hasName := data["name"]
	descVal, hasDesc := data["description"]
	if hasName && hasDesc {
		if nameStr, ok := nameVal.(string); ok {
			if descStr, ok := descVal.(string); ok {
				return nameStr, descStr, true
			}
		}
	}
	return "", "", false
}

// cmdMemoryStore records a learning via mulch record (simple interface for convention type)
func cmdMemoryStore(workspace, domain, recType, content string) (string, error) {
	return cmdMemoryRecordWithFlags(workspace, domain, recType, "", "", "", "", "", content)
}

// cmdMemoryFacts lists records in a domain via mulch query
func cmdMemoryFacts(workspace, domain string) (string, error) {
	// Execute: mulch query <domain>
	cmd := exec.Command("mulch", "query", domain)
	cmd.Dir = workspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mulch query failed: %v\n%s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

// cmdMemorySearch searches across all domains via mulch search
func cmdMemorySearch(workspace, query string) (string, error) {
	// Execute: mulch search <query>
	cmd := exec.Command("mulch", "search", query)
	cmd.Dir = workspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mulch search failed: %v\n%s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

// cmdMemoryQuery queries a specific domain via mulch prime
func cmdMemoryQuery(workspace, domain string) (string, error) {
	// Execute: mulch prime <domain>
	cmd := exec.Command("mulch", "prime", domain)
	cmd.Dir = workspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mulch prime failed: %v\n%s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

// cmdMemoryForget deletes a record via mulch delete
func cmdMemoryForget(workspace, domain, id string) (string, error) {
	// Execute: mulch delete <domain> <id>
	cmd := exec.Command("mulch", "delete", domain, id)
	cmd.Dir = workspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mulch delete failed: %v\n%s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

// cmdMemoryStatus shows mulch status
func cmdMemoryStatus(workspace string) (string, error) {
	// Execute: mulch status
	cmd := exec.Command("mulch", "status")
	cmd.Dir = workspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mulch status failed: %v\n%s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

// RegisterTopicCommands registers all topic commands to the registry.
// Topics are conversation sessions stored in ~/.fernwood/workspace/sessions/
func RegisterTopicCommands(registry *CommandRegistry, sessionStorage string) {
	// topic list - List recent topics
	registry.Register("topic list", "List topics (newest first, optional limit)", func(args []string, stdin string) (string, error) {
		limit := 10
		if len(args) > 0 {
			var err error
			limit, err = strconv.Atoi(args[0])
			if err != nil {
				return "", fmt.Errorf("topic list: invalid limit '%s'", args[0])
			}
		}
		return cmdTopicList(sessionStorage, limit)
	})

	// topic info - Show topic details
	registry.Register("topic info", "Show topic details and run history", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("topic info: usage: topic info <id>")
		}
		return cmdTopicInfo(sessionStorage, args[0])
	})

	// topic runs - List runs in a topic
	registry.Register("topic runs", "List runs in a topic", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("topic runs: usage: topic runs <id>")
		}
		limit := 10
		if len(args) > 1 {
			var err error
			limit, err = strconv.Atoi(args[1])
			if err != nil {
				return "", fmt.Errorf("topic runs: invalid limit '%s'", args[1])
			}
		}
		return cmdTopicRuns(sessionStorage, args[0], limit)
	})

	// topic run - Show run's full messages
	registry.Register("topic run", "Show run's full messages", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("topic run: usage: topic run <run-id>")
		}
		return cmdTopicRun(sessionStorage, args[0])
	})

	// topic rename - Rename a topic
	registry.Register("topic rename", "Rename a topic", func(args []string, stdin string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("topic rename: usage: topic rename <id> <new-name>")
		}
		return cmdTopicRename(sessionStorage, args[0], strings.Join(args[1:], " "))
	})

	// topic search - Search within a topic
	registry.Register("topic search", "Search within a topic", func(args []string, stdin string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("topic search: usage: topic search <id> <query>")
		}
		return cmdTopicSearch(sessionStorage, args[0], strings.Join(args[1:], " "))
	})

	// topic current - Show current topic ID
	registry.Register("topic current", "Show current topic ID", func(args []string, stdin string) (string, error) {
		return cmdTopicCurrent(sessionStorage)
	})

	// Aliases
	registry.RegisterAlias("topics", "topic list")
	registry.RegisterAlias("topic.info", "topic info")
	registry.RegisterAlias("topic.runs", "topic runs")
}

// cmdTopicList lists recent topics (sessions)
func cmdTopicList(sessionStorage string, limit int) (string, error) {
	// Read session files and list them
	sessions, err := listSessions(sessionStorage)
	if err != nil {
		return "", fmt.Errorf("topic list: %v", err)
	}

	if len(sessions) == 0 {
		return "No topics found", nil
	}

	// Sort by updated time (newest first) and limit
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Recent topics (showing %d of %d):\n\n", len(sessions), len(sessions)))
	for _, s := range sessions {
		result.WriteString(fmt.Sprintf("%s - %s (%d messages)\n", s.Key, s.Updated.Format("2006-01-02 15:04"), len(s.Messages)))
		if s.Summary != "" {
			summary := s.Summary
			if len(summary) > 80 {
				summary = summary[:77] + "..."
			}
			result.WriteString(fmt.Sprintf("  Summary: %s\n", summary))
		}
	}

	return result.String(), nil
}

// cmdTopicInfo shows details for a specific topic
func cmdTopicInfo(sessionStorage, key string) (string, error) {
	session, err := loadSession(sessionStorage, key)
	if err != nil {
		return "", fmt.Errorf("topic info: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Topic: %s\n", session.Key))
	result.WriteString(fmt.Sprintf("Created: %s\n", session.Created.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Updated: %s\n", session.Updated.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Messages: %d\n", len(session.Messages)))
	if session.Summary != "" {
		result.WriteString(fmt.Sprintf("Summary: %s\n", session.Summary))
	}

	return result.String(), nil
}

// cmdTopicRuns lists runs in a topic
func cmdTopicRuns(sessionStorage, key string, limit int) (string, error) {
	session, err := loadSession(sessionStorage, key)
	if err != nil {
		return "", fmt.Errorf("topic runs: %v", err)
	}

	if len(session.Messages) == 0 {
		return "No runs in this topic", nil
	}

	// Group messages into runs (user-assistant pairs)
	type run struct {
		index   int
		summary string
	}
	var runs []run
	for i := 0; i < len(session.Messages); i += 2 {
		summary := "Run"
		if i < len(session.Messages) {
			content := session.Messages[i].Content
			if len(content) > 50 {
				content = content[:47] + "..."
			}
			summary = content
		}
		runs = append(runs, run{index: i, summary: summary})
	}

	if len(runs) > limit {
		runs = runs[:limit]
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Runs in topic %s (showing %d):\n\n", key, len(runs)))
	for i, r := range runs {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.summary))
	}

	return result.String(), nil
}

// cmdTopicRun shows full messages for a run
func cmdTopicRun(sessionStorage, runID string) (string, error) {
	// Parse run ID (format: "topic-key:run-index")
	parts := strings.SplitN(runID, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("topic run: invalid run ID format. Use 'topic-key:run-index'")
	}

	sessionKey := parts[0]
	runIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("topic run: invalid run index '%s'", parts[1])
	}

	session, err := loadSession(sessionStorage, sessionKey)
	if err != nil {
		return "", fmt.Errorf("topic run: %v", err)
	}

	if runIndex < 0 || runIndex >= len(session.Messages) {
		return "", fmt.Errorf("topic run: run index %d out of range (0-%d)", runIndex, len(session.Messages)-1)
	}

	msg := session.Messages[runIndex]
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Run %d from topic %s:\n\n", runIndex, sessionKey))
	result.WriteString(fmt.Sprintf("Role: %s\n", msg.Role))
	result.WriteString(fmt.Sprintf("Content: %s\n", msg.Content))

	return result.String(), nil
}

// cmdTopicRename renames a topic
func cmdTopicRename(sessionStorage, oldKey, newKey string) (string, error) {
	// Load old session
	session, err := loadSession(sessionStorage, oldKey)
	if err != nil {
		return "", fmt.Errorf("topic rename: %v", err)
	}

	// Update key
	session.Key = newKey
	session.Updated = time.Now()

	// Save with new key
	if err := saveSession(sessionStorage, session); err != nil {
		return "", fmt.Errorf("topic rename: failed to save: %v", err)
	}

	// Delete old session file
	if err := deleteSession(sessionStorage, oldKey); err != nil {
		// Log but don't fail - new session is saved
	}

	return fmt.Sprintf("Renamed topic '%s' to '%s'", oldKey, newKey), nil
}

// cmdTopicSearch searches within a topic
func cmdTopicSearch(sessionStorage, key, query string) (string, error) {
	session, err := loadSession(sessionStorage, key)
	if err != nil {
		return "", fmt.Errorf("topic search: %v", err)
	}

	var matches []string
	for i, msg := range session.Messages {
		if strings.Contains(strings.ToLower(msg.Content), strings.ToLower(query)) {
			content := msg.Content
			if len(content) > 100 {
				content = content[:97] + "..."
			}
			matches = append(matches, fmt.Sprintf("Message %d (%s): %s", i, msg.Role, content))
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No matches for '%s' in topic %s", query, key), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d matches for '%s' in topic %s:\n\n", len(matches), query, key))
	for _, m := range matches {
		result.WriteString(m + "\n")
	}

	return result.String(), nil
}

// cmdTopicCurrent shows the current topic ID
func cmdTopicCurrent(sessionStorage string) (string, error) {
	// For now, return the most recently updated session
	sessions, err := listSessions(sessionStorage)
	if err != nil {
		return "", fmt.Errorf("topic current: %v", err)
	}

	if len(sessions) == 0 {
		return "No topics found", nil
	}

	return fmt.Sprintf("Current topic: %s", sessions[0].Key), nil
}

// Helper functions for session management

type sessionInfo struct {
	Key      string
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	Summary string
	Created interface{} // time.Time
	Updated interface{} // time.Time
}

func listSessions(sessionStorage string) ([]*session.Session, error) {
	if sessionStorage == "" {
		return []*session.Session{}, nil
	}

	files, err := os.ReadDir(sessionStorage)
	if err != nil {
		if os.IsNotExist(err) {
			return []*session.Session{}, nil
		}
		return nil, err
	}

	var sessions []*session.Session
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		sess, err := loadSession(sessionStorage, strings.TrimSuffix(file.Name(), ".json"))
		if err != nil {
			continue
		}
		sessions = append(sessions, sess)
	}

	// Sort by updated time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Updated.After(sessions[j].Updated)
	})

	return sessions, nil
}

func loadSession(sessionStorage, key string) (*session.Session, error) {
	if sessionStorage == "" {
		return nil, fmt.Errorf("session storage not configured")
	}

	filename := strings.ReplaceAll(key, ":", "_") + ".json"
	sessionPath := filepath.Join(sessionStorage, filename)

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("topic '%s' not found", key)
		}
		return nil, err
	}

	var sess session.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %v", err)
	}

	return &sess, nil
}

func saveSession(sessionStorage string, sess *session.Session) error {
	if sessionStorage == "" {
		return nil
	}

	// Write the session file directly
	filename := strings.ReplaceAll(sess.Key, ":", "_") + ".json"
	sessionPath := filepath.Join(sessionStorage, filename)

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sessionPath, data, 0644)
}

func deleteSession(sessionStorage, key string) error {
	if sessionStorage == "" {
		return nil
	}

	filename := strings.ReplaceAll(key, ":", "_") + ".json"
	sessionPath := filepath.Join(sessionStorage, filename)
	return os.Remove(sessionPath)
}
