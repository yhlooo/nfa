## ADDED Requirements

### Requirement: Model configuration structure
The system SHALL provide a unified `ModelConfig` structure for defining model metadata in configuration files.

#### Scenario: Minimal model configuration
- **WHEN** user configures a model with only the `name` field
- **THEN** system SHALL accept the configuration and use default values for all optional fields

#### Scenario: Full model configuration
- **WHEN** user configures a model with all optional fields (reasoning, vision, cost, contextWindow, maxOutputTokens)
- **THEN** system SHALL store all metadata for future use

### Requirement: Provider models field
Each provider configuration (`OllamaOptions`, `DeepseekOptions`, `OpenAICompatibleOptions`) SHALL include a `Models []ModelConfig` field.

#### Scenario: Ollama provider with models
- **WHEN** user configures `ollama.models` array
- **THEN** system SHALL use the configured models instead of calling Ollama API

#### Scenario: Deepseek provider with models
- **WHEN** user configures `deepseek.models` array
- **THEN** system SHALL use the configured models instead of calling Deepseek API

#### Scenario: OpenAI compatible provider with models
- **WHEN** user configures `openaiCompatible.models` array
- **THEN** system SHALL use the configured models instead of calling provider API

### Requirement: Configuration-driven model registration
The system SHALL register models based on configuration only, without making API calls to discover available models.

#### Scenario: Models configured
- **WHEN** provider has non-empty `models` array in configuration
- **THEN** system SHALL register only the models specified in configuration
- **AND** system SHALL NOT make any API calls to list models

#### Scenario: Models not configured
- **WHEN** provider has empty or missing `models` array in configuration
- **THEN** system SHALL NOT register any models from that provider
- **AND** system SHALL NOT make any API calls to discover models

### Requirement: Model name usage
The system SHALL use the `name` field from `ModelConfig` when registering models with the Genkit framework.

#### Scenario: Model registration with configured name
- **WHEN` model config specifies `name: "qwen3-max"` for provider "aliyun"
- **THEN** system SHALL register model as "aliyun/qwen3-max"

### Requirement: Metadata storage for future use
The system SHALL store all model metadata fields from configuration, even if not currently used in model registration.

#### Scenario: Reasoning field
- **WHEN** model config includes `reasoning: true`
- **THEN** system SHALL store this value in the registered model metadata

#### Scenario: Vision field
- **WHEN** model config includes `vision: false`
- **THEN** system SHALL store this value in the registered model metadata

#### Scenario: Cost information
- **WHEN** model config includes `cost.input` and `cost.output` values
- **THEN** system SHALL store these values for future cost calculation

#### Scenario: Context window
- **WHEN** model config includes `contextWindow: 262144`
- **THEN** system SHALL store this value for future context management

#### Scenario: Max output tokens
- **WHEN** model config includes `maxOutputTokens: 32768`
- **THEN** system SHALL store this value for future output limiting

### Requirement: Cost structure as value type
The `ModelCost` structure SHALL be a value type (not a pointer), with zero values representing free or unknown pricing.

#### Scenario: Empty cost object
- **WHEN** model config includes `cost: {}`
- **THEN** system SHALL interpret this as zero cost (both input and output are 0)

#### Scenario: Omitted cost field
- **WHEN** model config does not include `cost` field
- **THEN** system SHALL use zero values for both input and output costs

### Requirement: Breaking change for Ollama models field
The `OllamaOptions.Models` field SHALL change from `[]string` to `[]ModelConfig`, requiring configuration updates.

#### Scenario: Existing Ollama configuration migration
- **WHEN** user has existing Ollama configuration with `models: ["llama2", "mistral"]`
- **THEN** system SHALL fail to parse the configuration
- **AND** user MUST update to `models: [{"name": "llama2"}, {"name": "mistral"}]`
