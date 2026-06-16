# Rules

faas-cli is a REST client for OpenFaaS - CE, Standard, Enterprise and Edge (faasd)

The CLI is written in Go, to see verbose output for HTTP calls run "FAAS_DEBUG=1"

You must never push code, or commit code. The user does that. You may suggest a Chris Beams style commit message.

All go code must be formatted after editing, do not apply formatting to vendor.

## Rules for local models, such as Qwen, etc:

Create and modify files with the write/edit tools, never via cat > file << EOF or bash heredocs. Keep individual tool calls small.

SOTA models like Opus and GPT can ignore the tool calling restrictions.

