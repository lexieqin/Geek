# Configuration Guide for Production APIs

## Quick Start

### 1. Mock Mode (Default)
By default, GenesisGpt runs in mock mode using the local ginTools server:
```bash
# Start ginTools mock server
cd ../ginTools
go run main.go

# In another terminal, run GenesisGpt
cd ../GenesisGpt
./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135"
```

### 2. Production Mode
To use real production APIs:

#### Option A: Environment Variables
```bash
# Set environment variables for production APIs
export GENESIS_MODE=production
export GENESIS_JOB_API_URL="https://genesis.company.com/api/v1/tenant/{tenant}/jobs"
export GENESIS_DATADOG_API_URL="https://api.datadoghq.com/api/v2/traces/{traceID}"
export GENESIS_SANDBOX_API_URL="https://sandboxlogs.company.com/api/logs"

# Authentication tokens
export GENESIS_API_TOKEN="your-genesis-api-token"
export DD_API_KEY="your-datadog-api-key"
export DD_APP_KEY="your-datadog-app-key"
export SANDBOX_API_TOKEN="your-sandbox-api-token"

# Run GenesisGpt
./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135"
```

#### Option B: Configuration File
Create `config/config.yaml`:
```yaml
mode: production  # Change from "mock" to "production"

production:
  job_api_url: "https://genesis.company.com/api/v1/tenant/{tenant}/jobs"
  datadog_api_url: "https://api.datadoghq.com/api/v2/traces/{traceID}"
  sandbox_logs_api_url: "https://sandboxlogs.company.com/api/logs"
  sandbox_smart_logs_api_url: "https://sandboxlogs.company.com/api/logs/smart"
  
  auth:
    job_api:
      type: "bearer"
      token: "${GENESIS_API_TOKEN}"  # Will read from env var
    datadog:
      type: "api-key"
      api_key: "${DD_API_KEY}"
      app_key: "${DD_APP_KEY}"
    sandbox:
      type: "bearer"
      token: "${SANDBOX_API_TOKEN}"
```

Then run:
```bash
./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135"
```

## API Endpoints Reference

### Job Service API
- Mock: `http://localhost:8080/tenant/{tenant}/jobs?requuid={jobid}&trace=true`
- Production: Update in config.yaml with your actual endpoint

### Datadog API
- Mock: `http://localhost:8080/api/datadog/trace/{traceID}`
- Production: `https://api.datadoghq.com/api/v2/traces/{traceID}`
- Docs: https://docs.datadoghq.com/api/latest/tracing/

### Sandbox Logs API
- Mock: `http://localhost:8080/api/sandbox/logs?path={path}&file={file}`
- Production: Update in config.yaml with your actual endpoint

## Switching Between Modes

### Quick Switch via Environment
```bash
# For mock mode
export GENESIS_MODE=mock
./genesisgpt chat "debug job 123"

# For production mode
export GENESIS_MODE=production
./genesisgpt chat "debug job 123"
```

### Using Different Config Files
```bash
# Use custom config
export GENESISGPT_CONFIG=/path/to/custom-config.yaml
./genesisgpt chat "debug job 123"
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Check your API tokens are set correctly
   - Verify token format (Bearer vs API Key)
   - Check token permissions

2. **Network Errors**
   - Verify VPN connection if required
   - Check firewall rules
   - Test API endpoints with curl

3. **Mock Server Not Running**
   - Ensure ginTools is running on port 8080
   - Check `lsof -i :8080` to see if port is in use

### Debug Mode
```bash
# Enable debug logging
export GENESIS_DEBUG=true
./genesisgpt chat "debug job 123"
```

## Security Notes

1. **Never commit tokens** to git
2. Use environment variables for sensitive data
3. Keep config.yaml in .gitignore if it contains secrets
4. Use read-only API tokens when possible
5. Rotate tokens regularly