# GenesisGpt Runtime Modes Guide

## Overview

GenesisGpt supports two runtime modes:
- **Mock Mode**: Uses local ginTools server for testing and development
- **Production Mode**: Uses real production APIs with authentication

## Mock Mode (Default)

### Prerequisites
- ginTools server running on port 8080

### Step 1: Start Mock Server
```bash
# Terminal 1: Start ginTools mock server
cd ../ginTools
go run main.go

# Verify it's running
curl http://localhost:8080/health || echo "ginTools not responding"
```

### Step 2: Set OpenAI API Key
```bash
# Required for AI functionality
export OPENAI_API_KEY="your-openai-api-key"
```

### Step 3: Run GenesisGpt
```bash
# Terminal 2: Use GenesisGpt in mock mode
cd ../GenesisGpt
./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135 for testenv tenant"
```

### Mock Mode Features
- ✅ No authentication required
- ✅ Uses static test data (job.json, datadogtrace.json, containers.log)
- ✅ Perfect for testing and development
- ✅ Immediate feedback without external dependencies

---

## Production Mode

### Prerequisites
- Valid API tokens for your production systems
- OpenAI API key for AI functionality
- Network access to production APIs
- VPN connection (if required)

### Method 1: Environment Variables (Recommended)

```bash
# Set production mode
export GENESIS_MODE=production

# Configure API endpoints
export GENESIS_JOB_API_URL="https://genesis.company.com/api/v1/tenant/{tenant}/jobs"
export GENESIS_DATADOG_API_URL="https://api.datadoghq.com/api/v2/traces/{traceID}"
export GENESIS_SANDBOX_API_URL="https://sandboxlogs.company.com/api/logs"

# Set authentication tokens
export OPENAI_API_KEY="your-openai-api-key"
export GENESIS_API_TOKEN="your-genesis-api-token"
export DD_API_KEY="your-datadog-api-key"
export DD_APP_KEY="your-datadog-app-key"
export SANDBOX_API_TOKEN="your-sandbox-api-token"

# Run GenesisGpt
./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135 for testenv tenant"
```

### Method 2: Configuration File

Create or edit `config/config.yaml`:
```yaml
mode: production

production:
  job_api_url: "https://genesis.company.com/api/v1/tenant/{tenant}/jobs"
  datadog_api_url: "https://api.datadoghq.com/api/v2/traces/{traceID}"
  sandbox_logs_api_url: "https://sandboxlogs.company.com/api/logs"
  sandbox_smart_logs_api_url: "https://sandboxlogs.company.com/api/logs/smart"
  
  auth:
    job_api:
      type: "bearer"
      token: "${GENESIS_API_TOKEN}"
    datadog:
      type: "api-key"
      api_key: "${DD_API_KEY}"
      app_key: "${DD_APP_KEY}"
    sandbox:
      type: "bearer"
      token: "${SANDBOX_API_TOKEN}"

common:
  timeout: 30s
  retry_count: 3
  retry_delay: 2s
```

Then set your tokens in environment:
```bash
export GENESIS_API_TOKEN="your-genesis-api-token"
export DD_API_KEY="your-datadog-api-key"
export DD_APP_KEY="your-datadog-app-key"
export SANDBOX_API_TOKEN="your-sandbox-api-token"

./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135 for testenv tenant"
```

### Method 3: Custom Config File

```bash
# Use a specific config file
export GENESISGPT_CONFIG=/path/to/custom-config.yaml
./genesisgpt chat "debug job 123"
```

---

## Quick Mode Switching

### Switch to Mock Mode
```bash
export GENESIS_MODE=mock
./genesisgpt chat "debug job 123"
```

### Switch to Production Mode
```bash
export GENESIS_MODE=production
./genesisgpt chat "debug job 123"
```

### Check Current Mode
```bash
# GenesisGpt will show the mode in debug output
export GENESIS_DEBUG=true
./genesisgpt chat "debug job 123"
```

---

## Debug Levels

Both modes support different debug levels:

### Quick Debug (Default)
```bash
./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135"
# Only shows JobError section
```

### Trace Debug
```bash
./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135 with traces"
# Shows JobError + Datadog traces
```

### Full Debug
```bash
./genesisgpt chat "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135 with full analysis"
# Shows JobError + Datadog traces + Sandbox logs
```

---

## Troubleshooting

### Mock Mode Issues

**Problem**: "Job not found" error
```bash
# Check if ginTools is running
lsof -i :8080

# Start ginTools if not running
cd ../ginTools
go run main.go
```

**Problem**: Connection refused
```bash
# Verify ginTools health
curl http://localhost:8080/health

# Check ginTools logs for errors
```

### Production Mode Issues

**Problem**: Authentication errors
```bash
# Check if tokens are set
echo $GENESIS_API_TOKEN
echo $DD_API_KEY

# Test API access manually
curl -H "Authorization: Bearer $GENESIS_API_TOKEN" https://your-api.com/health
```

**Problem**: Network timeouts
```bash
# Check VPN connection
ping genesis.company.com

# Increase timeout in config
export GENESIS_TIMEOUT=60s
```

**Problem**: SSL/TLS errors
```bash
# Check certificates
curl -v https://your-api.com

# Use insecure mode for testing (not recommended for production)
export GENESIS_INSECURE=true
```

---

## Security Best Practices

### 1. Token Management
```bash
# Use environment variables, never hardcode tokens
export GENESIS_API_TOKEN="$(cat ~/.secrets/genesis-token)"

# Rotate tokens regularly
# Use read-only tokens when possible
```

### 2. Config File Security
```bash
# Keep sensitive configs out of git
echo "config/production.yaml" >> .gitignore

# Set proper file permissions
chmod 600 config/production.yaml
```

### 3. Network Security
```bash
# Use VPN when required
# Verify SSL certificates
# Use company-approved endpoints only
```

---

## Examples

### Daily Development Workflow
```bash
# Start your day with mock mode
cd ../ginTools && go run main.go &
cd ../GenesisGpt
./genesisgpt chat "debug job test-job-123"
```

### Production Debugging Session
```bash
# Switch to production for real incident
export GENESIS_MODE=production
export GENESIS_API_TOKEN="$(vault read -field=token secret/genesis)"
./genesisgpt chat "debug job urgent-incident-456 with full analysis"
```

### Testing New Features
```bash
# Test with mock data first
export GENESIS_MODE=mock
./genesisgpt chat "test my new debug feature"

# Then verify with production (if safe)
export GENESIS_MODE=production
./genesisgpt chat "test my new debug feature on non-critical job"
```

---

## Configuration Reference

| Environment Variable | Description | Example |
|---------------------|-------------|---------|
| `GENESIS_MODE` | Runtime mode | `mock` or `production` |
| `GENESIS_JOB_API_URL` | Job service endpoint | `https://api.company.com/jobs` |
| `GENESIS_API_TOKEN` | API authentication token | `abc123...` |
| `GENESIS_DEBUG` | Enable debug logging | `true` or `false` |
| `GENESIS_TIMEOUT` | Request timeout | `30s`, `1m`, `2m` |
| `GENESISGPT_CONFIG` | Custom config file path | `/path/to/config.yaml` |

For more detailed configuration options, see [CONFIG_GUIDE.md](CONFIG_GUIDE.md).