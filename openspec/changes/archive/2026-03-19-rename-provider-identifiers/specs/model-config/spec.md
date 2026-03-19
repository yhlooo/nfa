## MODIFIED Requirements

### Requirement: Model name usage

The system SHALL use the `name` field from `ModelConfig` when registering models with the Genkit framework.

#### Scenario: Model registration with configured name for Qwen
- **WHEN** model config specifies `name: "qwen3-max"` for provider "qwen"
- **THEN** system SHALL register model as "qwen/qwen3-max"

#### Scenario: Model registration with configured name for ZAI
- **WHEN** model config specifies `name: "glm-5"` for provider "z-ai"
- **THEN** system SHALL register model as "z-ai/glm-5"
