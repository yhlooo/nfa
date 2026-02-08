## ADDED Requirements

### Requirement: Project SHALL have a docs/ directory
The project SHALL contain a `docs/` directory at the root level to house all documentation.

#### Scenario: Directory exists
- **WHEN** a user examines the project root
- **THEN** a `docs/` directory MUST be present

### Requirement: docs/ directory SHALL contain three subdirectories
The `docs/` directory SHALL contain three subdirectories: `tutorials/`, `guides/`, and `reference/`.

#### Scenario: Subdirectories exist
- **WHEN** a user lists contents of `docs/`
- **THEN** the following directories MUST be present:
  - `tutorials/`
  - `guides/`
  - `reference/`

### Requirement: tutorials/ SHALL contain step-by-step tutorials
The `tutorials/` directory SHALL contain tutorial files that guide users through the application from beginning to end. Each file represents a tutorial step, and users SHALL read them sequentially from the first to the last.

#### Scenario: Tutorial files are numbered
- **WHEN** a user lists contents of `tutorials/`
- **THEN** tutorial files MUST be named with sequential numbers (e.g., `01-getting-started.md`, `02-basic-usage.md`)

#### Scenario: Tutorials guide users sequentially
- **WHEN** a user reads tutorials
- **THEN** they MUST be able to follow the tutorials in numerical order from start to finish

### Requirement: guides/ SHALL contain feature-specific usage guides
The `guides/` directory SHALL contain detailed usage guides for each feature, organized by feature. Users SHOULD be able to directly index to articles by their titles.

#### Scenario: Guides are feature-specific
- **WHEN** a user needs information about a specific feature
- **THEN** they MUST be able to find a guide file named after that feature (e.g., `skills.md`, `trading.md`)

#### Scenario: Guides are independently accessible
- **WHEN** a user navigates to a guide
- **THEN** they MUST be able to understand the feature without reading other guides

### Requirement: reference/ SHALL contain reference documentation
The `reference/` directory SHALL contain detailed reference information such as API documentation, configuration file structure definitions, and other technical details.

#### Scenario: API documentation is present
- **WHEN** a user needs API information
- **THEN** they MUST find API documentation in `reference/`

#### Scenario: Configuration reference is present
- **WHEN** a user needs configuration information
- **THEN** they MUST find configuration structure documentation in `reference/`

### Requirement: README.md SHALL only contain project overview
The `README.md` file SHALL contain only project introduction and quick start information. Detailed usage instructions SHALL be moved to appropriate files in the `docs/` directory.

#### Scenario: README contains project overview
- **WHEN** a user reads README.md
- **THEN** they MUST find project description, quick start instructions, and links to detailed documentation

#### Scenario: README links to docs
- **WHEN** a user reads README.md
- **THEN** they MUST find links pointing to the `docs/` directory and its subdirectories

### Requirement: Existing skill documentation SHALL migrate to guides/skills.md
The custom skills documentation currently in README.md SHALL be migrated to `docs/guides/skills.md`.

#### Scenario: Skills guide exists
- **WHEN** a user looks for skills documentation
- **THEN** they MUST find it at `docs/guides/skills.md`

#### Scenario: README no longer contains skills documentation
- **WHEN** a user reads README.md
- **THEN** the custom skills section MUST NOT be present
