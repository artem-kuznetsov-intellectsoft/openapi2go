---
name: go-codestyle
description: Go code style guidelines. Consult whenever writing, editing, generating, or reviewing Go (.go) code. Use this even if the user doesn't explicitly ask for a style check; apply the rules automatically as part of any Go code change.
---

# Go Code Style Guidelines

A running checklist of code style rules for Go code. Apply
these automatically whenever writing, editing, or reviewing Go code. When a
new rule is decided, append it to the checklist below in the same format —
one short rule with a one-line rationale.

## Checklist

- **Use `any` instead of `interface{}`.**
  `any` is the standard alias for `interface{}` since Go 1.18 and is the
  idiomatic, more readable choice going forward. Applies to variable types,
  function signatures, struct fields, type parameters, and generated code.
  Do not use `interface{}` in new or edited code — replace it with `any`
  wherever encountered in code you're touching.

<!-- Append new rules above this line, following the same format. -->
