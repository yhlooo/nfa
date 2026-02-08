## Context

nfa is a LLM-based investment advisor agent currently with fixed capabilities. The agent has a tool system that allows it to call external functions. The Skill feature aims to extend this by allowing users to define custom skills in their home directory, which can be loaded dynamically at runtime.

Key constraints:
- Skills are stored in `~/.nfa/skills/` (user-specific, not version-controlled)
- Each skill is a directory with a `SKILL.md` file containing YAML frontmatter and content
- The agent needs to know about available skills via system prompt
- The `Skill` tool must return skill content for the agent to use

## Goals / Non-Goals

**Goals:**
- Enable users to define custom skills without modifying nfa codebase
- Load skills from `~/.nfa/skills/` at runtime
- Parse `SKILL.md` files with YAML frontmatter (name, description)
- Provide a `Skill` tool that returns skill content by name
- Dynamically update system prompt with available skills
- Maintain backwards compatibility with existing nfa functionality

**Non-Goals:**
- Skill validation or execution (skills are descriptive only)
- Skill marketplace or sharing
- Skill dependencies or relationships
- Skill versioning
- Skill hot-reloading during agent execution

## Decisions

### 1. Skill Directory Structure

**Decision**: Each skill is a directory in `~/.nfa/skills/<skill-name>/` containing `SKILL.md`

**Rationale**: Directory-based structure allows for future expansion (e.g., additional files like prompts, templates, or helper scripts). Using the skill name as directory name provides intuitive organization.

**Alternatives Considered**:
- Flat file structure (`~/.nfa/skills/<skill-name>.md`) - rejected, less extensible
- JSON/YAML files instead of Markdown - rejected, Markdown is more user-friendly for editing

### 2. YAML Frontmatter for Metadata

**Decision**: Use YAML frontmatter in `SKILL.md` for `name` and `description` fields

**Rationale**: Standard format for markdown metadata, easy to parse with existing libraries (e.g., `gray-matter`). Keeps metadata and content together.

**Alternatives Considered**:
- Separate `metadata.yaml` file - rejected, increases file count
- JSON block in markdown - rejected, YAML is more readable

### 3. Skill Tool Interface

**Decision**: `Skill` tool accepts `{"name": "skill-name"}` and returns `{"content": "skill content"}`

**Rationale**: Simple, straightforward interface that matches common tool patterns. Returns full skill content (including frontmatter) so the agent has complete information.

**Alternatives Considered**:
- Return only content without frontmatter - rejected, frontmatter may contain useful context
- Return structured JSON with separate metadata fields - rejected, adds unnecessary parsing complexity

### 4. System Prompt Enhancement

**Decision**: Append a formatted list of available skills to the system prompt

**Rationale**: Simple integration point. The agent can see all available skills at startup and decide which to use. Format: "Available skills: get-price, analyze-trend, etc."

**Alternatives Considered**:
- Separate skills prompt - rejected, adds complexity
- Dynamic tool descriptions - rejected, tool interface is simple enough

### 5. Skill Discovery Timing

**Decision**: Scan skills directory once at agent initialization

**Rationale**: Skills are file-system based, hot-reloading during execution is not needed (out of scope). Scanning once is efficient and simpler.

**Alternatives Considered**:
- Watch directory for changes - rejected, out of scope and adds complexity
- Lazy loading on first use - rejected, system prompt needs full skill list

### 6. Error Handling

**Decision**: Return error message from `Skill` tool if skill not found

**Rationale**: Tool should be robust. Agent can handle error gracefully (e.g., inform user). Return format: `{"error": "Skill 'xxx' not found"}`

**Alternatives Considered**:
- Throw exception - rejected, tools should return structured responses
- Return null - rejected, agent needs to know what went wrong

## Risks / Trade-offs

**Risk**: Malicious skills could prompt agent to perform unintended actions
→ **Mitigation**: Skills are local and user-controlled. Document that users should review skills from untrusted sources.

**Risk**: Skills with same name in system vs user directory could cause confusion
→ **Mitigation**: Currently only user directory (`~/.nfa/skills/`) is used. If system skills are added later, implement namespace/priority logic.

**Risk**: Invalid YAML frontmatter breaks parsing
→ **Mitigation**: Handle parsing errors gracefully, log warnings, skip malformed skills.

**Trade-off**: Skills are descriptive only, not executable
→ This limits capabilities but keeps implementation simple and avoids security concerns.

**Trade-off**: No skill validation
→ Trusts users to write correct skill descriptions. Future enhancement could add schema validation.

## Migration Plan

1. **Phase 1**: Implement core skill loading and parsing
   - Create `skill_loader` module
   - Add `Skill` tool
   - Write tests

2. **Phase 2**: Integrate with agent system
   - Modify agent initialization to load skills
   - Update system prompt generation
   - Register `Skill` tool

3. **Phase 3**: Documentation and examples
   - Document skill format
   - Create example skills
   - Update README

**Rollback**: If issues arise, simply remove skill loading code and the `Skill` tool registration. No data migration needed as skills are user files.

## Open Questions

None - requirements are well-defined and design is straightforward.
