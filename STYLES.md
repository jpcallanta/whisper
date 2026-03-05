# Style guide

Conventions for this codebase.

## Comments

- Add one comment before every function stating its purpose; keep it concise.
- Do not repeat the function or symbol name in the comment; state purpose only (e.g. "Returns...", "Registers...", not "functionName returns...").
- Add one comment before each top-level `if` stating what the condition checks.
- Do not add comments for nested `if` blocks.

## Formatting

- Group assignment statements with a newline before and after the group.
- Group variable declarations with a newline before and after the group.
- Put a newline before every `return`.
- Separate top-level `if` blocks with a newline before and after the block.

## Functions

- Prefer small, unit-testable functions.
- Break this rule only when necessary for correctness.

## Errors

- Return on error; do not continue after a failing call when it is fatal.
- Wrap errors with `fmt.Errorf("context: %w", err)` so callers can use `errors.Is` / `errors.As`.

## Tests

- Name tests like `TestFuncName_Scenario_ExpectedOutcome` or `TestFuncName_Scenario`.
- Use `t.Errorf` or `t.Fatalf` with clear messages: e.g. "want X, got Y" or "expected X".

## Structure

- Use Cobra commands with `RunE` for commands that can fail.
- Declare flags as package-level variables in a `var ()` block.
- Use `init()` to register commands and flags with the root command.

## Imports

- Standard library first, then third-party.
- Use standard grouping (e.g. `goimports`).
