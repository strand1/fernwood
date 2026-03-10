# Fernwood TODO

## Refactor: Remove OAuth/Authentication Package

**Status:** Not started  
**Priority:** Low (non-blocking)  
**Created:** 2026-03-09

### Context

Fernwood does not require cloud login/authentication. The `pkg/auth/` package and related auth commands are vestigial code from the original PicoClaw codebase. They add complexity and maintenance burden without providing value to Fernwood's core use case.

### Scope

Remove all OAuth/token-based authentication infrastructure:

- Delete `pkg/auth/` directory entirely
- Delete `cmd/fernwood/internal/auth/` command
- Remove `auth.NewAuthCommand()` from `cmd/fernwood/main.go`
- Remove auth-related status display from `cmd/fernwood/internal/status/helpers.go`
- Strip OAuth branches from `pkg/providers/factory_provider.go`:
  - `createClaudeAuthProvider()` (Anthropic OAuth)
  - `createCodexAuthProvider()` (OpenAI OAuth)
  - `antigravity_provider.go` (Google Antigravity OAuth-only provider)
- Remove `getCredential = auth.GetCredential` var in factory
- Remove any `auth.GetCredential`, `auth.RefreshAccessToken`, `auth.SetCredential` calls

### Impact

- `fernwood auth` command will be removed
- OAuth-based provider auth (OpenAI/Anthropic/OAuth) will no longer work
- Antigravity provider will be removed
- Build will no longer depend on `pkg/auth`

### Note

API key-based authentication remains unaffected. Fernwood will continue to work with all providers that use API keys (OpenAI, Anthropic via API key, Gemini, Groq, etc.).

### Checklist

- [ ] Search for all imports of `pkg/auth` across codebase
- [ ] Remove `pkg/auth` directory
- [ ] Remove `cmd/fernwood/internal/auth/`
- [ ] Update `main.go` to remove auth command
- [ ] Clean up status command
- [ ] Clean up `pkg/providers/factory_provider.go` (remove OAuth branches)
- [ ] Remove `antigravity_provider.go`
- [ ] Update `pkg/providers/factory.go` to remove `getCredential` variable and any auth-related logic
- [ ] Update documentation (if any)
- [ ] Test build and basic agent functionality with API-key providers
- [ ] Run tests
