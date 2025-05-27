#!/bin/bash

# Script to commit and push translation changes

cd /Users/qinqin/workspace/src/github.com/Geek

echo "=== Git Status ==="
git status

echo -e "\n=== Adding files ==="
git add README.md
git add ginTools/pkg/config/k8sconfig.go
git add GenesisGpt/config/api_endpoints.go
git add GenesisGpt/switch-to-production.sh
git add GenesisGpt/.env.production.example

echo -e "\n=== Creating commit ==="
git commit -m "Translate all Chinese text to English for international accessibility

- Updated main README.md from Chinese to English
- Translated all Chinese comments in ginTools/pkg/config/k8sconfig.go
- Added configuration support for switching between mock and production APIs
  - Created config/api_endpoints.go for environment-based API endpoints
  - Added switch-to-production.sh script for easy environment switching
  - Added .env.production.example as template for production configuration

This makes the codebase more accessible to international developers and
maintains consistency across the project.

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

echo -e "\n=== Pushing to remote ==="
git push origin main

echo -e "\n=== Done! ==="