#!/bin/bash

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Match target filenames openapi.json or swagger.json
if [[ "$FILE_PATH" == *"openapi.json" || "$FILE_PATH" == *"swagger.json" ]]; then
  # Exit 0 and print the JSON structured output to return the deny decision
  jq -n '{
    "hookSpecificOutput": {
      "hookEventName": "PreToolUse",
      "permissionDecision": "deny",
      "permissionDecisionReason": "Reading openapi.json or swagger.json directly is forbidden to prevent context window bloat. Please use `jq -c` with the Bash tool to query and navigate specific fields of this spec file instead."
    }
  }'
  exit 0
fi

# Allow all other reads to proceed normally through the standard flow
exit 0