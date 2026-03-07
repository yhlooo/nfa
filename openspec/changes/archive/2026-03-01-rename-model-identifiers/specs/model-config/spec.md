# model-config Specification (Delta)

## MODIFIED Requirements

### Requirement: Breaking change for Ollama models field
The `OllamaOptions.Models` field SHALL change from `[]string` to `[]ModelConfig`, requiring configuration updates.

#### Scenario: Existing Ollama configuration migration
- **WHEN** user has existing Ollama configuration with `models: ["llama2", "mistral"]`
- **THEN** system SHALL fail to parse the configuration
- **AND** user MUST update to `models: [{"name": "llama2"}, {"name": "mistral"}]`

### Requirement: Breaking change for defaultModels field naming
The `defaultModels` configuration object SHALL rename fields from `main`/`fast` to `primary`/`light`, requiring configuration updates.

#### Scenario: Existing defaultModels configuration migration
- **WHEN** user has existing configuration with `defaultModels.main` and `defaultModels.fast` fields
- **THEN** system SHALL fail to recognize the old field names
- **AND** user MUST update configuration to use `defaultModels.primary` and `defaultModels.light`
- **AND** configuration values remain the same, only field names change

#### Scenario: Valid defaultModels configuration after migration
- **WHEN** user updates configuration to use new field names
- **THEN** system SHALL successfully parse the configuration
- **AND** `defaultModels.primary` specifies the primary model for complex tasks
- **AND** `defaultModels.light` specifies the light model for simple tasks
- **AND** `defaultModels.vision` specifies the vision model for image understanding tasks
