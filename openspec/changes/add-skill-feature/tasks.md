## 1. Core Skill Loading Module

- [ ] 1.1 Create `skill_loader` module with directory structure support
- [ ] 1.2 Implement SKILL.md file parsing with YAML frontmatter (gray-matter or similar)
- [ ] 1.3 Add skill discovery function to scan `~/.nfa/skills/` directory
- [ ] 1.4 Implement skill metadata retrieval (name and description from frontmatter)
- [ ] 1.5 Implement skill content retrieval (full SKILL.md content)
- [ ] 1.6 Add error handling for invalid YAML, missing files, permission errors
- [ ] 1.7 Create `~/.nfa/skills/` directory if it doesn't exist
- [ ] 1.8 Write unit tests for skill loading, parsing, and error handling

## 2. Skill Tool Implementation

- [ ] 2.1 Create `Skill` tool with name parameter schema
- [ ] 2.2 Implement tool function to return skill content by name
- [ ] 2.3 Add tool description for agent understanding
- [ ] 2.4 Handle missing name parameter with error response
- [ ] 2.5 Handle skill not found scenario with error response
- [ ] 2.6 Return success response with full skill content including frontmatter
- [ ] 2.7 Write unit tests for tool success and error cases

## 3. Agent Core Integration

- [ ] 3.1 Modify agent initialization to call skill loading on startup
- [ ] 3.2 Implement system prompt enhancement with available skills list
- [ ] 3.3 Format skills list (name - description) for system prompt
- [ ] 3.4 Handle empty skills list case in system prompt
- [ ] 3.5 Register `Skill` tool in agent's tool registry
- [ ] 3.6 Add graceful error handling during skill loading (log warnings, continue)
- [ ] 3.7 Write integration tests for agent initialization with skills

## 4. Documentation and Examples

- [ ] 4.1 Document skill format in README
- [ ] 4.2 Create example skill `get-price` with proper SKILL.md format
- [ ] 4.3 Document YAML frontmatter required fields (name, description)
- [ ] 4.4 Document skill directory structure (`~/.nfa/skills/<skill-name>/SKILL.md`)
- [ ] 4.5 Add usage examples for the Skill tool
