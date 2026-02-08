## Why

nfa is currently a basic LLM-based investment advisor agent with fixed capabilities. The Skill feature will allow users to extend nfa's abilities by defining custom skills without modifying the core codebase. This enables rapid prototyping of new capabilities and customization for specific use cases while keeping the core agent clean.

## What Changes

- **New Skill System**: Introduce a plugin architecture where users can define skills in `~/.nfa/skills/`
- **Skill Metadata**: Each skill directory contains a `SKILL.md` file with YAML frontmatter (name, description) and implementation details
- **Skill Tool**: Add a new tool `Skill` that reads skill content by name and returns it to the agent
- **System Prompt Enhancement**: List available skills in the system prompt so the agent knows which skills exist
- **Skill Discovery**: Automatically scan `~/.nfa/skills/` for available skills

## Capabilities

### New Capabilities
- `skill-system`: Core skill management including reading skills, parsing SKILL.md files, and skill discovery
- `skill-tool`: The `Skill` tool that exposes skill content to the agent

### Modified Capabilities
- `agent-core`: Requires modification to include available skills in the system prompt and register the Skill tool

## Impact

- **Code Changes**: New modules for skill loading/parsing, tool registration, and system prompt generation
- **User Directory**: Uses `~/.nfa/skills/` for user-provided skills
- **Tool Interface**: Adds `Skill` tool to nfa's tool registry
- **System Prompt**: Dynamically includes skill list (names and descriptions)
- **Backwards Compatible**: No breaking changes to existing functionality
