# No specification changes needed

This change is a pure internal refactor that does not introduce new capabilities or modify existing requirements. All behaviors defined in `skill-system` spec remain unchanged.

The change only affects how logger is passed (from constructor parameter to context parameter), which is an implementation detail not captured in specs.
