#!/bin/bash

# Test script for the new sandbox URL pattern
# This demonstrates how GenesisGpt will use the sandbox URL directly

BASE_URL="http://localhost:8080"
SANDBOX_URL="${BASE_URL}/sandboxlogs/#/katbox/browse?path=/csi-data-dir/7d1f4a89-b6ec-44e4-b047-d34d6d3f9704&hostip=0.0.0.0"

echo "Testing new sandbox URL pattern..."
echo "================================"

echo -e "\n1. List files in sandbox:"
curl -s "${SANDBOX_URL}&action=list" | jq .

echo -e "\n2. Read containers.log:"
curl -s "${SANDBOX_URL}&action=read&file=containers.log" | head -20

echo -e "\n3. Read specific log file (applog/deploy.log):"
curl -s "${SANDBOX_URL}&action=read&file=applog/deploy.log" | head -10

echo -e "\n4. Analyze logs (smart analysis):"
curl -s "${SANDBOX_URL}&action=analyze" | jq .

echo -e "\nDone!"