# Non-Interactive Mode

## Purpose

Non-interactive command-line mode for single-question queries with optional auto-exit.

## Requirements

### Requirement: Support positional argument for initial question
The system SHALL accept a positional argument to specify the initial question.

#### Scenario: Interactive mode with initial question
- **WHEN** user runs `nfa 'DXYZ 今天股价如何'`
- **THEN** system SHALL start interactive mode AND automatically send the question as the first prompt

#### Scenario: Interactive mode without argument
- **WHEN** user runs `nfa`
- **THEN** system SHALL start interactive mode without sending any initial prompt

#### Scenario: Multiple arguments
- **WHEN** user runs `nfa '问题1' '问题2'`
- **THEN** system SHALL use only the first argument as the initial question

### Requirement: Support -p flag for print-and-exit mode
The system SHALL support `-p` / `--print` flag to enable non-interactive single-query mode.

#### Scenario: Print and exit with question
- **WHEN** user runs `nfa 'DXYZ 今天股价如何' -p`
- **THEN** system SHALL send the question, display the complete response, and exit immediately after the response is complete

#### Scenario: Print and exit without question
- **WHEN** user runs `nfa -p`
- **THEN** system SHALL start in interactive mode (no auto-exit since no question)

### Requirement: Auto-exit after response completion
The system SHALL automatically exit after the agent completes its response when in print-and-exit mode.

#### Scenario: Exit after successful response
- **WHEN** user runs `nfa '问题' -p` and agent returns a response
- **THEN** system SHALL exit immediately after the PromptResponse is received

#### Scenario: Exit after error response
- **WHEN** user runs `nfa '问题' -p` and agent returns an error
- **THEN** system SHALL display the error and exit

### Requirement: Output consistency between modes
The system SHALL produce identical output in both interactive and non-interactive modes for the same question.

#### Scenario: Output includes reasoning
- **WHEN** agent response includes reasoning
- **THEN** both interactive and non-interactive modes SHALL display the reasoning content

#### Scenario: Output includes tool calls
- **WHEN** agent makes tool calls during response
- **THEN** both interactive and non-interactive modes SHALL display tool call information

#### Scenario: Output includes errors
- **WHEN** an error occurs during agent processing
- **THEN** both interactive and non-interactive modes SHALL display the error

### Requirement: Initial prompt timing
The system SHALL send the initial prompt only after session initialization is complete.

#### Scenario: Initial prompt after session creation
- **WHEN** user runs `nfa '问题'`
- **THEN** system SHALL create a session first, then send the initial prompt
