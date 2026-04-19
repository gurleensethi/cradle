---
name: code-docs-sync
description: Ensure code documentation is accurate, coherent, and synchronized with implementation
---

## Objective
Review all code documentation across the codebase and verify alignment with actual implementation.

## Documentation styles
Concise sentences that focus on the behaviour of the function, instead of documenting implementation details. What the function is for and what is does is more important that how it does it.

## Tasks
1. **Identify gaps**: Find undocumented code sections, missing parameter descriptions, or outdated examples
2. **Verify accuracy**: Confirm that documentation reflects current behavior and API signatures
3. **Check coherence**: Ensure documentation style, terminology, and clarity are consistent
4. **Update**: Correct inaccurate, incomplete, or misleading documentation
5. **Prioritize**: Flag critical documentation issues for high-impact code paths

## Hard Rules (do not break these rules)
- No package doc required
- Don't document simple and obvious code and functions.
- Don't change any code.

## Scope
- Code comments and docstrings