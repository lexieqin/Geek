#!/bin/bash

echo "Testing sandbox log fetching in mock mode..."
echo "=========================================="

# Set to mock mode (uses development config by default)
export ENVIRONMENT=development

# Start ginTools in background
echo "Starting ginTools mock server..."
cd ../ginTools && go run . &
GINTOOLS_PID=$!
sleep 2

# Run a test query
cd ../GenesisGpt
echo -e "\nTesting with mock data..."
echo "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135 under testenv tenant" | go run main.go chat

# Cleanup
kill $GINTOOLS_PID 2>/dev/null

echo -e "\nDone!"