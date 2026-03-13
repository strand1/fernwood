// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RegisterFSCommands registers all file system commands to the registry.
// workspace: base directory for relative paths
// restrict: if true, restrict all operations to workspace
func RegisterFSCommands(registry *CommandRegistry, workspace string, restrict bool) {
	// ls - List directory contents
	registry.Register("ls", "List directory contents", func(args []string, stdin string) (string, error) {
		// Check if any args look like flags (start with -)
		hasFlags := false
		path := "."
		for i, arg := range args {
			if strings.HasPrefix(arg, "-") {
				hasFlags = true
			} else if i == 0 || !strings.HasPrefix(args[i-1], "-") {
				// First non-flag argument is the path
				path = arg
			}
		}
		
		// If flags are present, use shell ls for full compatibility
		if hasFlags {
			return cmdShellLs(args, workspace, restrict)
		}
		
		return cmdLs(path, workspace, restrict)
	})

	// fs.ls alias
	registry.RegisterAlias("fs.ls", "ls")

	// cat - Read file content
	registry.Register("cat", "Read file content (auto-detect binary)", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("cat: missing file operand")
		}
		return cmdCat(args[0], workspace, restrict)
	})

	// fs.cat alias
	registry.RegisterAlias("fs.cat", "cat")

	// write - Write file
	registry.Register("write", "Write file (stdin if no content arg)", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("write: missing file operand")
		}
		return cmdWrite(args[0], stdin, workspace, restrict)
	})

	// fs.write alias
	registry.RegisterAlias("fs.write", "write")

	// stat - File metadata
	registry.Register("stat", "File metadata (size, mtime, type)", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("stat: missing file operand")
		}
		return cmdStat(args[0], workspace, restrict)
	})

	// fs.stat alias
	registry.RegisterAlias("fs.stat", "stat")

	// rm - Remove file
	registry.Register("rm", "Remove file (safety checks)", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("rm: missing file operand")
		}
		return cmdRm(args[0], workspace, restrict)
	})

	// fs.rm alias
	registry.RegisterAlias("fs.rm", "rm")

	// cp - Copy file
	registry.Register("cp", "Copy file", func(args []string, stdin string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("cp: missing file operand (usage: cp <src> <dst>)")
		}
		return cmdCp(args[0], args[1], workspace, restrict)
	})

	// fs.cp alias
	registry.RegisterAlias("fs.cp", "cp")

	// mv - Move/rename file
	registry.Register("mv", "Move/rename file", func(args []string, stdin string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("mv: missing file operand (usage: mv <src> <dst>)")
		}
		return cmdMv(args[0], args[1], workspace, restrict)
	})

	// fs.mv alias
	registry.RegisterAlias("fs.mv", "mv")

	// mkdir - Create directory
	registry.Register("mkdir", "Create directory", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("mkdir: missing operand")
		}
		return cmdMkdir(args[0], workspace, restrict)
	})

	// fs.mkdir alias
	registry.RegisterAlias("fs.mkdir", "mkdir")

	// grep - Search text
	registry.Register("grep", "Search text (grep [-i] [-v] [-c] [-n] <pattern> [file])", func(args []string, stdin string) (string, error) {
		return cmdGrep(args, stdin, workspace, restrict)
	})

	// head - First N lines
	registry.Register("head", "First N lines (head [-n N] [file])", func(args []string, stdin string) (string, error) {
		return cmdHead(args, stdin, workspace, restrict)
	})

	// tail - Last N lines
	registry.Register("tail", "Last N lines (tail [-n N] [file])", func(args []string, stdin string) (string, error) {
		return cmdTail(args, stdin, workspace, restrict)
	})

	// wc - Count lines/words/chars
	registry.Register("wc", "Count lines/words/chars (wc [-l] [-w] [-c] [file])", func(args []string, stdin string) (string, error) {
		return cmdWc(args, stdin, workspace, restrict)
	})

	// find - Find files (find [path] -name <pattern>)
	registry.Register("find", "Find files (find [path] -name <pattern>)", func(args []string, stdin string) (string, error) {
		return cmdFind(args, workspace, restrict)
	})
	registry.RegisterAlias("fs.find", "find")

	// which - Find command in PATH
	registry.Register("which", "Find command location in PATH", func(args []string, stdin string) (string, error) {
		return cmdWhich(args)
	})

	// web_fetch - Fetch URL content
	registry.Register("web_fetch", "Fetch URL content (web_fetch <url>)", func(args []string, stdin string) (string, error) {
		return cmdWebFetch(args)
	})

	// sh - Execute shell command
	registry.Register("sh", "Execute shell command (sh <command>)", func(args []string, stdin string) (string, error) {
		return cmdSh(args, stdin)
	})
	registry.RegisterAlias("shell", "sh")
	registry.RegisterAlias("bash", "sh")
	registry.RegisterAlias("python", "sh")
	registry.RegisterAlias("python3", "sh")

	// sed - Stream editor
	registry.Register("sed", "Stream editor (sed [-n] [-e script] [file])", func(args []string, stdin string) (string, error) {
		return cmdSed(args, stdin, workspace, restrict)
	})
}

// cmdLs implements the ls command
// cmdShellLs executes the actual shell ls command with flags for full compatibility
func cmdShellLs(args []string, workspace string, restrict bool) (string, error) {
	// Build the command with proper argument handling
	cmdArgs := append([]string{"ls"}, args...)
	
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = workspace
	
	if restrict {
		// Set up restricted environment
		cmd.Env = append(os.Environ(),
			"PWD="+workspace,
		)
	}
	
	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))
	
	if err != nil {
		// Include error output (ls often outputs errors to stderr)
		if result != "" {
			return result, nil // Return ls output even with non-zero exit
		}
		return "", fmt.Errorf("ls: %v", err)
	}
	
	return result, nil
}

func cmdLs(path, workspace string, restrict bool) (string, error) {
	absPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("ls: cannot access '%s': %v", path, err)
	}

	// Sort entries: directories first, then files
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	var result strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			result.WriteString("DIR:  " + entry.Name() + "/\n")
		} else {
			result.WriteString("FILE: " + entry.Name() + "\n")
		}
	}

	return strings.TrimSuffix(result.String(), "\n"), nil
}

// cmdCat implements the cat command
func cmdCat(path, workspace string, restrict bool) (string, error) {
	absPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("cat: %s: No such file or directory", path)
		}
		return "", fmt.Errorf("cat: %s: %v", path, err)
	}

	// Check for binary content
	if IsBinary(data) {
		return FormatBinaryError(path, int64(len(data))), nil
	}

	return string(data), nil
}

// cmdWrite implements the write command
func cmdWrite(path, content, workspace string, restrict bool) (string, error) {
	absPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return "", fmt.Errorf("write: failed to create directories: %v", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write: %s: %v", path, err)
	}

	return fmt.Sprintf("Written %d bytes to %s", len(content), path), nil
}

// cmdStat implements the stat command
func cmdStat(path, workspace string, restrict bool) (string, error) {
	absPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("stat: cannot stat '%s': No such file or directory", path)
		}
		return "", fmt.Errorf("stat: cannot stat '%s': %v", path, err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("File: %s\n", path))
	result.WriteString(fmt.Sprintf("Size: %d bytes\n", info.Size()))
	result.WriteString(fmt.Sprintf("Type: %s\n", fileType(info)))
	result.WriteString(fmt.Sprintf("Modified: %s\n", info.ModTime().Format(time.RFC3339)))
	result.WriteString(fmt.Sprintf("Mode: %s", info.Mode()))

	return result.String(), nil
}

// cmdRm implements the rm command
func cmdRm(path, workspace string, restrict bool) (string, error) {
	absPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	if err := os.Remove(absPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("rm: cannot remove '%s': No such file or directory", path)
		}
		return "", fmt.Errorf("rm: cannot remove '%s': %v", path, err)
	}

	return fmt.Sprintf("Removed: %s", path), nil
}

// cmdCp implements the cp command
func cmdCp(src, dst, workspace string, restrict bool) (string, error) {
	absSrc, err := validatePath(src, workspace, restrict)
	if err != nil {
		return "", err
	}

	absDst, err := validatePath(dst, workspace, restrict)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(absSrc)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("cp: cannot stat '%s': No such file or directory", src)
		}
		return "", fmt.Errorf("cp: cannot read '%s': %v", src, err)
	}

	if err := os.WriteFile(absDst, data, 0644); err != nil {
		return "", fmt.Errorf("cp: cannot create '%s': %v", dst, err)
	}

	return fmt.Sprintf("Copied: %s → %s", src, dst), nil
}

// cmdMv implements the mv command
func cmdMv(src, dst, workspace string, restrict bool) (string, error) {
	absSrc, err := validatePath(src, workspace, restrict)
	if err != nil {
		return "", err
	}

	absDst, err := validatePath(dst, workspace, restrict)
	if err != nil {
		return "", err
	}

	if err := os.Rename(absSrc, absDst); err != nil {
		return "", fmt.Errorf("mv: cannot move '%s' to '%s': %v", src, dst, err)
	}

	return fmt.Sprintf("Moved: %s → %s", src, dst), nil
}

// cmdMkdir implements the mkdir command
func cmdMkdir(path, workspace string, restrict bool) (string, error) {
	absPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return "", fmt.Errorf("mkdir: cannot create directory '%s': %v", path, err)
	}

	return fmt.Sprintf("Created directory: %s", path), nil
}

// cmdGrep implements the grep command
func cmdGrep(args []string, stdin, workspace string, restrict bool) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("grep: missing pattern")
	}

	// Parse flags
	var (
		ignoreCase   bool
		invert       bool
		count        bool
		lineNumbers  bool
		pattern      string
		file         string
	)

	i := 0
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		switch args[i] {
		case "-i":
			ignoreCase = true
		case "-v":
			invert = true
		case "-c":
			count = true
		case "-n":
			lineNumbers = true
		default:
			return "", fmt.Errorf("grep: unrecognized option '%s'", args[i])
		}
		i++
	}

	if i >= len(args) {
		return "", fmt.Errorf("grep: missing pattern")
	}

	pattern = args[i]
	i++

	if i < len(args) {
		file = args[i]
	}

	// Get input from file or stdin
	var input string
	if file != "" {
		absPath, err := validatePath(file, workspace, restrict)
		if err != nil {
			return "", err
		}
		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("grep: %s: %v", file, err)
		}
		input = string(data)
	} else {
		input = stdin
	}

	if input == "" {
		return "", nil
	}

	// Perform grep
	lines := strings.Split(input, "\n")
	var matches []string
	matchCount := 0

	for lineNum, line := range lines {
		var matched bool
		if ignoreCase {
			matched = strings.Contains(strings.ToLower(line), strings.ToLower(pattern))
		} else {
			matched = strings.Contains(line, pattern)
		}

		if invert {
			matched = !matched
		}

		if matched {
			matchCount++
			if !count {
				if lineNumbers {
					matches = append(matches, fmt.Sprintf("%d:%s", lineNum+1, line))
				} else {
					matches = append(matches, line)
				}
			}
		}
	}

	if count {
		return fmt.Sprintf("%d", matchCount), nil
	}

	return strings.Join(matches, "\n"), nil
}

// cmdHead implements the head command
func cmdHead(args []string, stdin, workspace string, restrict bool) (string, error) {
	n := 10 // default

	// Parse -n flag
	i := 0
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		if args[i] == "-n" {
			i++
			if i >= len(args) {
				return "", fmt.Errorf("head: option requires an argument -- 'n'")
			}
			var err error
			n, err = strconv.Atoi(args[i])
			if err != nil {
				return "", fmt.Errorf("head: invalid number of lines: %s", args[i])
			}
		}
		i++
	}

	// Get input from file or stdin
	var input string
	if i < len(args) {
		absPath, err := validatePath(args[i], workspace, restrict)
		if err != nil {
			return "", fmt.Errorf("head: cannot open '%s': %v", args[i], err)
		}
		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("head: cannot open '%s': %v", args[i], err)
		}
		input = string(data)
	} else {
		input = stdin
	}

	if input == "" {
		return "", nil
	}

	lines := strings.Split(input, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}

	return strings.Join(lines, "\n"), nil
}

// cmdTail implements the tail command
func cmdTail(args []string, stdin, workspace string, restrict bool) (string, error) {
	n := 10 // default

	// Parse -n flag
	i := 0
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		if args[i] == "-n" {
			i++
			if i >= len(args) {
				return "", fmt.Errorf("tail: option requires an argument -- 'n'")
			}
			var err error
			n, err = strconv.Atoi(args[i])
			if err != nil {
				return "", fmt.Errorf("tail: invalid number of lines: %s", args[i])
			}
		}
		i++
	}

	// Get input from file or stdin
	var input string
	if i < len(args) {
		absPath, err := validatePath(args[i], workspace, restrict)
		if err != nil {
			return "", fmt.Errorf("tail: cannot open '%s': %v", args[i], err)
		}
		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("tail: cannot open '%s': %v", args[i], err)
		}
		input = string(data)
	} else {
		input = stdin
	}

	if input == "" {
		return "", nil
	}

	lines := strings.Split(input, "\n")
	start := len(lines) - n
	if start < 0 {
		start = 0
	}
	lines = lines[start:]

	return strings.Join(lines, "\n"), nil
}

// cmdWc implements the wc command
func cmdWc(args []string, stdin, workspace string, restrict bool) (string, error) {
	// Parse flags
	var (
		countLines   bool
		countWords   bool
		countChars   bool
		showDefaults bool
	)

	if len(args) == 0 {
		showDefaults = true
	}

	i := 0
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		switch args[i] {
		case "-l":
			countLines = true
		case "-w":
			countWords = true
		case "-c":
			countChars = true
		default:
			return "", fmt.Errorf("wc: unrecognized option '%s'", args[i])
		}
		i++
	}

	// If no flags specified, show all
	if !countLines && !countWords && !countChars {
		showDefaults = true
	}

	// Get input from file or stdin
	var input string
	var filename string
	if i < len(args) {
		absPath, err := validatePath(args[i], workspace, restrict)
		if err != nil {
			return "", fmt.Errorf("wc: cannot open '%s': %v", args[i], err)
		}
		filename = args[i]
		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("wc: cannot open '%s': %v", filename, err)
		}
		input = string(data)
	} else {
		input = stdin
	}

	// Count
	lines := strings.Count(input, "\n")
	words := len(strings.Fields(input))
	chars := len(input)

	var result strings.Builder
	if countLines || showDefaults {
		if filename != "" {
			result.WriteString(fmt.Sprintf("  %d", lines))
		} else {
			result.WriteString(fmt.Sprintf("%d", lines))
		}
	}
	if countWords || showDefaults {
		result.WriteString(fmt.Sprintf(" %d", words))
	}
	if countChars || showDefaults {
		result.WriteString(fmt.Sprintf(" %d", chars))
	}
	if filename != "" {
		result.WriteString(" " + filename)
	}

	return result.String(), nil
}

// fileType returns a human-readable file type string
func fileType(info fs.FileInfo) string {
	mode := info.Mode()
	switch {
	case mode.IsDir():
		return "directory"
	case mode.IsRegular():
		return "regular file"
	case mode&os.ModeSymlink != 0:
		return "symbolic link"
	case mode&os.ModeDevice != 0:
		return "device"
	case mode&os.ModeNamedPipe != 0:
		return "named pipe"
	case mode&os.ModeSocket != 0:
		return "socket"
	default:
		return "unknown"
	}
}

// cmdFind implements the find command
func cmdFind(args []string, workspace string, restrict bool) (string, error) {
	// Default to workspace root if no path specified
	searchPath := workspace
	namePattern := ""

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-name":
			if i+1 < len(args) {
				namePattern = args[i+1]
				i++
			}
		default:
			// Assume it's a path
			if !strings.HasPrefix(args[i], "-") {
				searchPath = args[i]
			}
		}
	}

	// Validate path if restricted
	if restrict {
		var err error
		searchPath, err = validatePath(searchPath, workspace, restrict)
		if err != nil {
			return "", err
		}
	}

	// Check if path exists
	info, err := os.Stat(searchPath)
	if err != nil {
		return "", fmt.Errorf("find: '%s': %v", searchPath, err)
	}

	var results []string

	if info.IsDir() {
		// Walk directory
		err = filepath.Walk(searchPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Apply name filter if specified
			if namePattern != "" {
				matched, err := filepath.Match(namePattern, info.Name())
				if err != nil {
					return fmt.Errorf("find: invalid pattern '%s': %v", namePattern, err)
				}
				if !matched {
					return nil
				}
			}

			results = append(results, path)
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("find: %v", err)
		}
	} else {
		// Single file - check if it matches pattern
		if namePattern != "" {
			matched, err := filepath.Match(namePattern, info.Name())
			if err != nil {
				return "", fmt.Errorf("find: invalid pattern '%s': %v", namePattern, err)
			}
			if matched {
				results = append(results, searchPath)
			}
		} else {
			results = append(results, searchPath)
		}
	}

	if len(results) == 0 {
		return "", nil
	}

	return strings.Join(results, "\n"), nil
}

// cmdWhich implements the which command
func cmdWhich(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("which: usage: which <command>")
	}

	var results []string
	for _, cmd := range args {
		path, err := exec.LookPath(cmd)
		if err != nil {
			// Command not found - skip silently (like which)
			continue
		}
		results = append(results, path)
	}

	if len(results) == 0 {
		return "", nil
	}

	return strings.Join(results, "\n"), nil
}

// cmdWebFetch implements web_fetch command (simple URL fetch)
func cmdWebFetch(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("web_fetch: usage: web_fetch <url>")
	}

	url := args[0]
	
	// Simple HTTP GET
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("web_fetch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("web_fetch: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("web_fetch: %v", err)
	}

	return string(body), nil
}

// cmdSh executes arbitrary shell commands
func cmdSh(args []string, stdin string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("sh: usage: sh <command>")
	}

	// Join all args into a single command string
	cmdStr := strings.Join(args, " ")
	
	// Execute via shell
	cmd := exec.Command("sh", "-c", cmdStr)
	
	// Provide stdin if available
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	
	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))
	
	if err != nil {
		// Include output even on error (shell commands often output useful info)
		if result != "" {
			return result, nil
		}
		return "", fmt.Errorf("sh: %v", err)
	}
	
	return result, nil
}

// cmdSed implements basic sed functionality
func cmdSed(args []string, stdin string, workspace string, restrict bool) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("sed: usage: sed [-n] [-e script] [-f script-file] [file...]")
	}

	var (
		scripts    []string
		files      []string
		quiet      bool
		i          = 0
	)

	// Parse arguments
	for i < len(args) {
		if args[i] == "-n" {
			quiet = true
			i++
		} else if args[i] == "-e" {
			if i+1 >= len(args) {
				return "", fmt.Errorf("sed: -e requires an argument")
			}
			scripts = append(scripts, args[i+1])
			i += 2
		} else if args[i] == "-f" {
			if i+1 >= len(args) {
				return "", fmt.Errorf("sed: -f requires an argument")
			}
			// Load script from file
			scriptPath := args[i+1]
			absPath, err := validatePath(scriptPath, workspace, restrict)
			if err != nil {
				return "", err
			}
			data, err := os.ReadFile(absPath)
			if err != nil {
				return "", fmt.Errorf("sed: %s: %v", scriptPath, err)
			}
			scripts = append(scripts, string(data))
			i += 2
		} else if strings.HasPrefix(args[i], "-") {
			return "", fmt.Errorf("sed: unrecognized option '%s'", args[i])
		} else {
			files = append(files, args[i])
			i++
		}
	}

	// Get input
	var input string
	if len(files) == 0 {
		input = stdin
	} else if len(files) == 1 {
		absPath, err := validatePath(files[0], workspace, restrict)
		if err != nil {
			return "", err
		}
		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("sed: %s: %v", files[0], err)
		}
		input = string(data)
	} else {
		return "", fmt.Errorf("sed: multiple files not supported")
	}

	if input == "" {
		return "", nil
	}

	// Apply scripts
	lines := strings.Split(input, "\n")
	var output []string

	for _, line := range lines {
		modified := line
		printLine := !quiet

		for _, script := range scripts {
			result, shouldPrint, err := applySedScript(modified, script)
			if err != nil {
				return "", fmt.Errorf("sed: %v", err)
			}
			modified = result
			if shouldPrint {
				printLine = true
			}
		}

		if printLine {
			output = append(output, modified)
		}
	}

	return strings.Join(output, "\n"), nil
}

// applySedScript applies a single sed script to a line
func applySedScript(line, script string) (string, bool, error) {
	script = strings.TrimSpace(script)
	if script == "" {
		return line, true, nil
	}

	// Handle substitution: s/pattern/replacement/[flags]
	if strings.HasPrefix(script, "s/") {
		parts := strings.Split(script[2:], "/")
		if len(parts) < 2 {
			return line, true, fmt.Errorf("invalid substitution: %s", script)
		}
		pattern := parts[0]
		replacement := parts[1]
		
		// Handle escaped slashes in pattern
		if len(parts) > 2 {
			// Check if there are more parts due to escaped slashes
			for i := 2; i < len(parts); i++ {
				if parts[i-1] == "" && i > 2 {
					// This was an escaped slash
					pattern += "/" + parts[i]
				} else if i == len(parts)-1 {
					// This is the flags part
					break
				} else {
					replacement += "/" + parts[i]
				}
			}
		}

		// Convert sed replacement syntax to Go syntax
		replacement = strings.ReplaceAll(replacement, "\\&", "$0")
		replacement = strings.ReplaceAll(replacement, "\\1", "$1")
		replacement = strings.ReplaceAll(replacement, "\\2", "$2")

		// Perform substitution
		re, err := regexp.Compile(pattern)
		if err != nil {
			return line, true, fmt.Errorf("invalid pattern: %v", err)
		}
		
		// Check for global flag
		flags := ""
		if len(parts) >= 2 {
			lastPart := parts[len(parts)-1]
			if strings.Contains(lastPart, "g") {
				flags = "g"
			}
		}
		
		if flags == "g" {
			return re.ReplaceAllString(line, replacement), true, nil
		}
		return re.ReplaceAllString(line, replacement), true, nil
	}

	// Handle delete: /pattern/d
	if strings.HasSuffix(script, "/d") {
		pattern := script[1 : len(script)-2]
		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			return line, true, fmt.Errorf("invalid pattern: %v", err)
		}
		if matched {
			return "", false, nil // Delete = don't print
		}
		return line, true, nil
	}

	// Handle print: /pattern/p
	if strings.HasSuffix(script, "/p") {
		pattern := script[1 : len(script)-2]
		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			return line, true, fmt.Errorf("invalid pattern: %v", err)
		}
		if matched {
			return line, true, nil
		}
		return "", false, nil // Don't print if not matched
	}

	// No recognized command, return line as-is
	return line, true, nil
}
