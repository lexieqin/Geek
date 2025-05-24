#!/bin/bash

# GenesisGpt Configuration Setup Script
# This script helps set up the configuration for different environments

echo "=== GenesisGpt Configuration Setup ==="
echo

# Function to prompt for input with default value
prompt_with_default() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"
    
    read -p "$prompt [$default]: " value
    value="${value:-$default}"
    eval "$var_name='$value'"
}

# Select environment
echo "Select environment:"
echo "1) Development (local/mock)"
echo "2) Staging"
echo "3) Production"
read -p "Enter choice [1-3]: " env_choice

case $env_choice in
    1)
        ENVIRONMENT="development"
        ;;
    2)
        ENVIRONMENT="staging"
        ;;
    3)
        ENVIRONMENT="production"
        ;;
    *)
        echo "Invalid choice. Defaulting to development."
        ENVIRONMENT="development"
        ;;
esac

echo
echo "Setting up configuration for: $ENVIRONMENT"
echo

# Create .env file for the selected environment
ENV_FILE=".env.$ENVIRONMENT"
echo "# GenesisGpt Environment Configuration" > $ENV_FILE
echo "# Generated on $(date)" >> $ENV_FILE
echo >> $ENV_FILE
echo "GENESIS_ENV=$ENVIRONMENT" >> $ENV_FILE

if [ "$ENVIRONMENT" != "development" ]; then
    echo
    echo "=== API Configuration ==="
    
    # Job Service
    prompt_with_default "Job Service Base URL" "https://api.genesis.company.com" JOB_SERVICE_URL
    echo "JOB_SERVICE_URL=$JOB_SERVICE_URL" >> $ENV_FILE
    
    prompt_with_default "Job Service Auth Type (none/bearer/api_key)" "bearer" JOB_SERVICE_AUTH_TYPE
    echo "JOB_SERVICE_AUTH_TYPE=$JOB_SERVICE_AUTH_TYPE" >> $ENV_FILE
    
    if [ "$JOB_SERVICE_AUTH_TYPE" != "none" ]; then
        read -sp "Job Service Token/API Key: " JOB_SERVICE_TOKEN
        echo
        echo "JOB_SERVICE_TOKEN=$JOB_SERVICE_TOKEN" >> $ENV_FILE
    fi
    
    echo
    
    # Datadog
    prompt_with_default "Datadog Site (datadoghq.com/datadoghq.eu)" "datadoghq.com" DATADOG_SITE
    echo "DATADOG_SITE=$DATADOG_SITE" >> $ENV_FILE
    
    read -sp "Datadog API Key: " DD_API_KEY
    echo
    echo "DD_API_KEY=$DD_API_KEY" >> $ENV_FILE
    
    read -sp "Datadog Application Key: " DD_APP_KEY
    echo
    echo "DD_APP_KEY=$DD_APP_KEY" >> $ENV_FILE
    
    echo
    
    # Sandbox Service
    prompt_with_default "Sandbox Service Base URL" "https://sandbox.genesis.company.com" SANDBOX_SERVICE_URL
    echo "SANDBOX_SERVICE_URL=$SANDBOX_SERVICE_URL" >> $ENV_FILE
    
    prompt_with_default "Sandbox Service Auth Type (none/bearer)" "bearer" SANDBOX_AUTH_TYPE
    echo "SANDBOX_AUTH_TYPE=$SANDBOX_AUTH_TYPE" >> $ENV_FILE
    
    if [ "$SANDBOX_AUTH_TYPE" != "none" ]; then
        read -sp "Sandbox Service Token: " SANDBOX_TOKEN
        echo
        echo "SANDBOX_TOKEN=$SANDBOX_TOKEN" >> $ENV_FILE
    fi
    
    echo
    
    # Feature Flags
    echo "=== Feature Configuration ==="
    prompt_with_default "Enable Tracing (true/false)" "true" ENABLE_TRACING
    echo "ENABLE_TRACING=$ENABLE_TRACING" >> $ENV_FILE
    
    prompt_with_default "Enable Smart Logs (true/false)" "true" ENABLE_SMART_LOGS
    echo "ENABLE_SMART_LOGS=$ENABLE_SMART_LOGS" >> $ENV_FILE
    
    prompt_with_default "Verbose Logging (true/false)" "false" VERBOSE_LOGGING
    echo "VERBOSE_LOGGING=$VERBOSE_LOGGING" >> $ENV_FILE
else
    # Development environment - use defaults
    echo "JOB_SERVICE_URL=http://localhost:8080" >> $ENV_FILE
    echo "DATADOG_SERVICE_URL=http://localhost:8080" >> $ENV_FILE
    echo "SANDBOX_SERVICE_URL=http://localhost:8080" >> $ENV_FILE
    echo "GINTOOLS_URL=http://localhost:8080" >> $ENV_FILE
    echo "USE_MOCK_DATA=true" >> $ENV_FILE
    echo "ENABLE_TRACING=true" >> $ENV_FILE
    echo "ENABLE_SMART_LOGS=true" >> $ENV_FILE
    echo "VERBOSE_LOGGING=true" >> $ENV_FILE
fi

echo
echo "=== Configuration Complete ==="
echo
echo "Configuration saved to: $ENV_FILE"
echo
echo "To use this configuration:"
echo "  export \$(cat $ENV_FILE | xargs)"
echo
echo "Or source it directly:"
echo "  source $ENV_FILE"
echo
echo "You can also copy the appropriate config file:"
echo "  cp config/config.$ENVIRONMENT.json config.json"
echo

# Create a convenience script to load the environment
LOAD_SCRIPT="load-env-$ENVIRONMENT.sh"
cat > $LOAD_SCRIPT << EOF
#!/bin/bash
# Load $ENVIRONMENT environment variables
export \$(cat $ENV_FILE | grep -v '^#' | xargs)
echo "Loaded $ENVIRONMENT environment"
EOF

chmod +x $LOAD_SCRIPT
echo "Created helper script: $LOAD_SCRIPT"