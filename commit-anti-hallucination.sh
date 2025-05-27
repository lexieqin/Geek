#!/bin/bash

# Script to commit anti-hallucination improvements

cd /Users/qinqin/workspace/src/github.com/Geek

echo "=== Git Status ==="
git status

echo -e "\n=== Adding files ==="
# Add the anti-hallucination changes
git add GenesisGpt/cmd/promptTpl/prompt.go
git add GenesisGpt/cmd/tools/intelligentDebugTool.go
git add GenesisGpt/cmd/ai/message.go
git add GenesisGpt/MODEL_CONFIGURATION.md

# Also add the translation changes if not yet committed
git add README.md
git add ginTools/pkg/config/k8sconfig.go
git add GenesisGpt/config/api_endpoints.go
git add GenesisGpt/switch-to-production.sh
git add GenesisGpt/.env.production.example

echo -e "\n=== Creating commit ==="
git commit -m "Add anti-hallucination measures and model configuration support

Major improvements to prevent AI models from hallucinating tool outputs:

1. Enhanced prompt template with strict rules against generating fake data
2. Added unique timestamp markers to tool outputs (ACTUAL TOOL OUTPUT START/END)
3. Model-specific handling for Claude to prevent hallucination
4. Made AI model configurable via AI_MODEL environment variable
5. Added explicit examples of wrong vs correct tool usage
6. Created MODEL_CONFIGURATION.md guide for multi-model support

Also includes previous translation work:
- Translated Chinese text to English throughout codebase
- Added production API configuration support

These changes ensure consistent behavior across different AI models
(Qwen-Max, Claude) and prevent models from generating fictional
debug data instead of using actual tool responses.

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

echo -e "\n=== Pushing to remote ==="
git push origin main

echo -e "\n=== Done! ==="