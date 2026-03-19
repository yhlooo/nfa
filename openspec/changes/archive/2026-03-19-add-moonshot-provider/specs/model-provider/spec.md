## ADDED Requirements

### Requirement: Moonshot AI provider configuration
The system SHALL support Moonshot AI as a model provider through the `moonshotai` configuration key.

#### Scenario: Moonshot AI provider with API key only
- **WHEN** user configures `moonshotai.apiKey` without `baseURL`
- **THEN** system SHALL use the default base URL `https://api.moonshot.cn/v1`

#### Scenario: Moonshot AI provider with custom base URL
- **WHEN** user configures both `moonshotai.apiKey` and `moonshotai.baseURL`
- **THEN** system SHALL use the configured base URL

#### Scenario: Moonshot AI provider with custom models
- **WHEN** user configures `moonshotai.models` array
- **THEN** system SHALL use the configured models instead of default suggested models

### Requirement: Kimi K2.5 model registration
The system SHALL register the `kimi-k2.5` model with appropriate capabilities.

#### Scenario: Model registration with capabilities
- **WHEN** Moonshot provider is configured
- **THEN** system SHALL register `kimi-k2.5` model with:
  - `reasoning: true` (supports reasoning/thinking mode)
  - `vision: true` (supports image understanding)
  - `contextWindow: 256000`
  - `maxOutputTokens: 256000`

#### Scenario: Model registration with pricing
- **WHEN** Kimi K2.5 is registered
- **THEN** system SHALL store cost information:
  - `cost.input: 0.004` (per 1K input tokens)
  - `cost.output: 0.021` (per 1K output tokens)

### Requirement: Moonshot AI provider integration
The system SHALL integrate Moonshot AI provider following the same pattern as ZAI provider.

#### Scenario: Plugin creation
- **WHEN** Moonshot AI provider is configured
- **THEN** system SHALL create an `oai.OpenAICompatible` plugin with:
  - `Provider: "moonshotai"`
  - `BaseURL: "https://api.moonshot.cn/v1"` (or configured URL)
  - `APIKey: <configured-api-key>`

#### Scenario: Model registration flow
- **WHEN** Moonshot AI provider initializes
- **THEN** system SHALL register models via `OpenAICompatibleOptions.RegisterModels`
- **AND** model names SHALL be prefixed with `moonshotai/`

### Requirement: i18n support for Moonshot AI models
The system SHALL provide internationalized descriptions for Moonshot AI models.

#### Scenario: Model description in Chinese
- **WHEN** UI language is Chinese
- **THEN** system SHALL display Chinese description for Kimi K2.5

#### Scenario: Model description in English
- **WHEN** UI language is English
- **THEN** system SHALL display English description for Kimi K2.5
