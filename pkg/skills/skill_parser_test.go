package skills

import (
	"testing"
)

func TestParseSkillContent_Valid(t *testing.T) {
	content := `---
name: test-skill
description: A test skill for testing
---

This is the skill content.
It can have multiple lines.
`

	parsed, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent() error = %v", err)
	}

	if parsed.Metadata.Name != "test-skill" {
		t.Errorf("expected name 'test-skill', got '%s'", parsed.Metadata.Name)
	}

	if parsed.Metadata.Description != "A test skill for testing" {
		t.Errorf("expected description 'A test skill for testing', got '%s'", parsed.Metadata.Description)
	}

	expectedContent := "This is the skill content.\nIt can have multiple lines.\n"
	if parsed.Content != expectedContent {
		t.Errorf("expected content '%s', got '%s'", expectedContent, parsed.Content)
	}
}

func TestParseSkillContent_MissingFrontmatterStart(t *testing.T) {
	content := `name: test-skill
description: A test skill
---

Content here.
`

	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for missing frontmatter start, got nil")
	}

	expectedErrMsg := "skill file must start with YAML frontmatter"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestParseSkillContent_MissingFrontmatterEnd(t *testing.T) {
	content := `---
name: test-skill
description: A test skill

Content here.
`

	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for missing frontmatter end, got nil")
	}

	expectedErrMsg := "frontmatter not closed"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestParseSkillContent_MissingName(t *testing.T) {
	content := `---
description: A test skill
---

Content here.
`

	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for missing name field, got nil")
	}

	expectedErrMsg := "missing required field: name"
	if err.Error() != expectedErrMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestParseSkillContent_MissingDescription(t *testing.T) {
	content := `---
name: test-skill
---

Content here.
`

	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for missing description field, got nil")
	}

	expectedErrMsg := "missing required field: description"
	if err.Error() != expectedErrMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestParseSkillContent_InvalidYAML(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
invalid yaml: [unclosed bracket
---

Content here.
`

	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}

	expectedErrMsg := "parse YAML frontmatter error"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestParseSkillContent_EmptyContent(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
---
`

	parsed, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent() error = %v", err)
	}

	if parsed.Content != "" {
		t.Errorf("expected empty content, got '%s'", parsed.Content)
	}
}

func TestParseSkillContent_CRLF(t *testing.T) {
	content := "---\r\nname: test-skill\r\ndescription: A test skill\r\n---\r\n\r\nContent here.\r\n"

	parsed, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent() error = %v", err)
	}

	if parsed.Metadata.Name != "test-skill" {
		t.Errorf("expected name 'test-skill', got '%s'", parsed.Metadata.Name)
	}

	if parsed.Metadata.Description != "A test skill" {
		t.Errorf("expected description 'A test skill', got '%s'", parsed.Metadata.Description)
	}
}

func TestFormatSkillContent(t *testing.T) {
	content := &SkillContent{
		Metadata: SkillMetadata{
			Name:        "test-skill",
			Description: "A test skill",
		},
		Content: "This is the content.\nMultiple lines.",
	}

	formatted := FormatSkillContent(content)

	expected := `---
name: test-skill
description: A test skill
---

This is the content.
Multiple lines.
`

	if formatted != expected {
		t.Errorf("expected formatted content:\n%s\ngot:\n%s", expected, formatted)
	}
}
