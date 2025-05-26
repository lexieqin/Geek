# Model Configuration Guide

GenesisGpt supports multiple AI models. This guide helps you configure the appropriate model for your environment.

## Supported Models

### 1. Qwen-Max (Default)
- **Model ID**: `qwen-max`
- **Provider**: Alibaba Cloud
- **Use Case**: Default model, works well with structured outputs

### 2. Claude 3.5 Sonnet
- **Model ID**: `claude-3-5-sonnet-20241022`
- **Provider**: Anthropic
- **Use Case**: Advanced reasoning, but may need extra guidance to prevent hallucination

### 3. Claude 3.7 Sonnet
- **Model ID**: `claude-3-7-sonnet-20250219`
- **Provider**: Anthropic
- **Use Case**: Latest Claude model with improvements

## Configuration

Set the model using environment variables:

```bash
# For Qwen-Max (default)
export AI_MODEL=qwen-max
export OPENAI_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
export OPENAI_API_KEY=your-dashscope-api-key

# For Claude models
export AI_MODEL=claude-3-5-sonnet-20241022
export OPENAI_BASE_URL=https://api.anthropic.com/v1
export OPENAI_API_KEY=your-anthropic-api-key

# Run GenesisGpt
./genesisgpt chat
```

## Model-Specific Behaviors

### Qwen-Max
- Follows structured output formats closely
- Good at preserving tool output formatting
- Default choice for consistency

### Claude Models
- May summarize or reorganize tool outputs
- Requires stronger prompting to prevent hallucination
- GenesisGpt automatically adds extra safeguards for Claude models

## Troubleshooting

### Issue: Model is hallucinating tool outputs
**Symptoms**: 
- Tool outputs contain dates from wrong years (e.g., 2023 instead of 2025)
- Made-up trace IDs or error messages
- Tool response appears before actual tool execution

**Solution**:
- Ensure you're using the latest version of GenesisGpt
- The system now includes anti-hallucination markers in tool outputs
- Claude models receive additional system prompts to prevent this

### Issue: Output format differs between models
**Solution**:
- This is expected behavior
- Focus on accuracy of information rather than exact formatting
- Both models should identify the same issues

## Environment Setup Examples

### Development (Personal Laptop)
```bash
# .env.development
AI_MODEL=qwen-max
OPENAI_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
OPENAI_API_KEY=sk-xxxxx
```

### Production (Company)
```bash
# .env.production
AI_MODEL=claude-3-5-sonnet-20241022
OPENAI_BASE_URL=https://api.anthropic.com/v1
OPENAI_API_KEY=sk-ant-xxxxx
ENVIRONMENT=production
PROD_JOB_API=https://api.company.com
PROD_TRACE_API=https://api.datadoghq.com
PROD_SANDBOX_API=https://sandbox.company.com
```

## Testing Model Behavior

Test script to verify model behavior:
```bash
#!/bin/bash
# test-model.sh

echo "Testing model: $AI_MODEL"
echo "debug job 81325fc3-b05e-4d9a-ada2-d2399aebe135 under testenv tenant" | ./genesisgpt chat

# Check output for:
# 1. Actual timestamps (2025, not 2023)
# 2. Tool output markers: "=== ACTUAL TOOL OUTPUT START [TS:xxx] ==="
# 3. Correct error messages from your mock data
```

## Best Practices

1. **Always test** after switching models
2. **Monitor** for hallucination, especially with Claude models
3. **Report issues** if you see consistent hallucination despite safeguards
4. **Keep prompts updated** when adding new tools