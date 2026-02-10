## 1. Data Structures

- [x] 1.1 Add `ModelConfig` struct to `pkg/models/providers.go` with fields: Name, Reasoning, Vision, Cost, ContextWindow, MaxOutputTokens
- [x] 1.2 Add `ModelCost` struct to `pkg/models/providers.go` with fields: Input, Output
- [x] 1.3 Update `OllamaOptions.Models` field from `[]string` to `[]ModelConfig` in `pkg/models/ollama.go`
- [x] 1.4 Add `Models []ModelConfig` field to `DeepseekOptions` in `pkg/models/deepseek.go`
- [x] 1.5 Add `Models []ModelConfig` field to `OpenAICompatibleOptions` in `pkg/models/openai_compatible.go`

## 2. Ollama Provider

- [x] 2.1 Remove `ListModels()` method from `OllamaOptions` in `pkg/models/ollama.go`
- [x] 2.2 Remove `ListOllamaTagsResponse` and `OllamaModel` types (no longer needed) in `pkg/models/ollama.go`
- [x] 2.3 Update `RegisterModels()` to iterate over configured `opts.Models` instead of calling API
- [x] 2.4 Update `RegisterModels()` to use `ModelConfig.Name` for model registration

## 3. Deepseek Provider

- [x] 3.1 Update `RegisterModels()` method signature to accept `[]ModelConfig` instead of calling API
- [x] 3.2 Remove `client.Models.List()` API call from `RegisterModels()` in `pkg/genkitplugins/deepseek/deepseek.go`
- [x] 3.3 Update `RegisterModels()` to iterate over configured models and use their names
- [x] 3.4 Keep hardcoded thinking model detection logic (`m.ID == "deepseek-reasoner"`) for now

## 4. OpenAI Compatible Provider

- [x] 4.1 Remove `ListOpenAICompatibleModels()` function from `pkg/models/openai_compatible.go`
- [x] 4.2 Remove `ListOpenAICompatibleModelsResponse` and `OpenAICompatibleModel` types (no longer needed)
- [x] 4.3 Update model registration logic in `pkg/agents/genkit.go` to use configured models only
- [x] 4.4 Remove API discovery call to `GET /models` endpoint in `NewGenkitWithModels()`

## 5. Genkit Integration

- [x] 5.1 Update `NewGenkitWithModels()` in `pkg/agents/genkit.go` to handle empty `models` arrays (no model registration, no API calls)
- [x] 5.2 Add warning log when all providers have empty `models` arrays
- [x] 5.3 Verify model registration uses correct naming format (`provider/model-name`)

## 6. Documentation

- [x] 6.1 Update `docs/guides/model-config.md` with new `ModelConfig` structure examples
- [x] 6.2 Add BREAKING CHANGE notice for Ollama configuration format
- [x] 6.3 Add migration guide for existing Ollama users
- [x] 6.4 Update configuration examples to show all available `ModelConfig` fields

## 7. Testing

- [x] 7.1 Test configuration parsing with minimal model config (name only)
- [x] 7.2 Test configuration parsing with full model config (all fields)
- [x] 7.3 Test empty `models` array behavior (no models registered)
- [x] 7.4 Test Ollama model registration with new config format
- [x] 7.5 Test Deepseek model registration with new config format
- [x] 7.6 Test OpenAI Compatible model registration with new config format
- [x] 7.7 Verify `nfa models list` shows configured models correctly
- [x] 7.8 Verify startup time improvement (no API discovery calls)
