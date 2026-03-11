# Test Failures Report

**Date:** March 11, 2026  
**Go Version:** go1.24.2  
**Focus:** testing, go

## Summary

Running `make test` produced **3 test failures** across 3 different packages:

1. `cmd/fernwood` - TestNewFernwoodCommand
2. `pkg/config` - TestFullConfig_JSON_BackwardCompat
3. `pkg/tools` - TestFilesystemTool_ReadFile_MissingPath

---

## Failure Details

### 1. cmd/fernwood - TestNewFernwoodCommand

**File:** `cmd/fernwood/main_test.go:22`

**Error:**
```
Not equal:
expected: "🌲 fernwood - Personal AI Assistant vdev\n\n"
actual  : "🌲 fernwood - Agentic Coding Harness vdev\n\n"
```

**Root Cause:** The test expectation hardcodes "Personal AI Assistant" but the actual binary description has been changed to "Agentic Coding Harness". This is likely a branding/product name change that wasn't reflected in the test.

**Suggested Fix:** Update the test to expect "Agentic Coding Harness" instead of "Personal AI Assistant", or refactor to use a constant that can be updated in one place.

---

### 2. pkg/config - TestFullConfig_JSON_BackwardCompat

**File:** `pkg/config/model_config_test.go:227`

**Error:**
```
Unmarshal error: invalid character 'a' looking for beginning of object key string
```

**Root Cause:** The test contains invalid JSON in the `newFormat` string. Specifically, line with `api_key` is missing an opening quote:

```json
api_key": "test-key"  // missing opening quote
```

Should be:
```json
"api_key": "test-key"
```

**Suggested Fix:** Add the missing opening quote to make the JSON valid.

---

### 3. pkg/tools - TestFilesystemTool_ReadFile_MissingPath

**File:** `pkg/tools/filesystem_test.go:73`

**Error:**
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x97f91e]
```

**Stack Trace:**
```
github.com/strand1/fernwood/pkg/tools.(*ReadFileTool).Execute
    /home/dstrand/workspace/fernwood/pkg/tools/filesystem.go:137 +0x7e
```

**Root Cause:** The test creates the tool with `&ReadFileTool{}` which leaves the `fs` field uninitialized (nil). When `Execute` is called, it attempts to use `t.fs.ReadFile(path)` causing a nil pointer dereference.

All other tests in the same file properly use `NewReadFileTool` to create an initialized tool with a working filesystem. This test was written incorrectly.

**Suggested Fix:**
- Option 1: Use `NewReadFileTool("", false)` to match other tests that need a basic filesystem.
- Option 2: Create a mock filesystem implementation for this test.
- Option 3: Add a nil check in `Execute` and return an appropriate error if `fs` is nil (defensive programming).

---

## Notes

- All other packages passed their tests successfully.
- The test suite runs `go generate ./...` before testing.
- The `pkg/channels` package had the longest test duration (~13 seconds).
