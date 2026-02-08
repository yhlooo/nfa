## 1. Modify SkillLoader structure and constructors

- [x] 1.1 Remove `logger logr.Logger` field from `SkillLoader` struct in `pkg/skills/skill_loader.go`
- [x] 1.2 Change `NewSkillLoader(logger logr.Logger, homeDir string)` to `NewSkillLoader(ctx context.Context, homeDir string)`
- [x] 1.3 Change `NewSkillLoaderWithDir(logger logr.Logger, skillsDir string)` to `NewSkillLoaderWithDir(ctx context.Context, skillsDir string)`
- [x] 1.4 Remove `.WithName("skill_loader")` initialization (no logger to name)

## 2. Modify Load method

- [x] 2.1 Change `Load()` method signature to `Load(ctx context.Context)`
- [x] 2.2 Replace `sl.logger.Info(...)` calls with `logger := logr.FromContextOrDiscard(ctx); logger.Info(...)`

## 3. Update caller in agents

- [x] 3.1 Change `skills.NewSkillLoaderWithDir(a.logger, skillsDir)` to `skills.NewSkillLoaderWithDir(ctx, skillsDir)` in `pkg/agents/genkit.go`
- [x] 3.2 Change `a.skillLoader.Load()` to `a.skillLoader.Load(ctx)` in `pkg/agents/genkit.go`

## 4. Update tests in skill_loader_test.go

- [x] 4.1 Replace `NewSkillLoader(logr.Discard(), tmpDir)` with `NewSkillLoader(t.Context(), tmpDir)` in all test functions
- [x] 4.2 Replace `loader.Load()` with `loader.Load(t.Context())` in all test functions

## 5. Update tests in skill_tool_test.go

- [x] 5.1 Replace `NewSkillLoader(logr.Discard(), tmpDir)` with `NewSkillLoader(t.Context(), tmpDir)` in all test functions
- [x] 5.2 Replace `loader.Load()` with `loader.Load(t.Context())` in all test functions

## 6. Verification

- [x] 6.1 Run `go test ./pkg/skills/...` to verify all tests pass
- [x] 6.2 Run `go test ./pkg/agents/...` to verify all tests pass
- [x] 6.3 Run `go build ./...` to ensure no compilation errors
