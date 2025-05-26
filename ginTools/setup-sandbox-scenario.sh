#!/bin/bash

# Script to set up different sandbox log scenarios for testing

SANDBOX_DIR="pkg/staticfile/sandboxlink"

function setup_scenario() {
    local scenario=$1
    
    # Clear existing files and directories (except README.md)
    find $SANDBOX_DIR -mindepth 1 ! -name "README.md" -delete
    
    case $scenario in
        "dns-error")
            echo "Setting up DNS error scenario..."
            cp pkg/staticfile/containers.log $SANDBOX_DIR/
            ;;
            
        "oom")
            echo "Setting up Out of Memory scenario..."
            cat > $SANDBOX_DIR/containers.log << EOF
[podbase] 2025-05-23T00:00:00.000Z stdout F Starting application...
[podbase] 2025-05-23T00:00:10.000Z stdout F Allocating memory...
[podbase] 2025-05-23T00:00:20.000Z stderr F Error: java.lang.OutOfMemoryError: Java heap space
[podbase] 2025-05-23T00:00:21.000Z stderr F Container killed due to OOMKilled
EOF
            ;;
            
        "multi-log")
            echo "Setting up multi-log scenario..."
            cp pkg/staticfile/containers.log $SANDBOX_DIR/
            
            # Create std.out
            cat > $SANDBOX_DIR/std.out << EOF
Application starting...
Loading configuration...
Starting web server on port 8080...
EOF
            
            # Create std.err
            cat > $SANDBOX_DIR/std.err << EOF
ERROR: Failed to connect to database
ERROR: Connection timeout after 30 seconds
ERROR: Application shutting down
EOF
            
            # Create applog directory
            mkdir -p $SANDBOX_DIR/applog
            
            # Create deploy.log in applog subdirectory
            cat > $SANDBOX_DIR/applog/deploy.log << EOF
Deployment started at 2025-05-23T00:00:00Z
Pulling image...
Image pulled successfully
Starting container...
ERROR: Failed to mount volume
ERROR: ConfigMap "app-config" not found
EOF
            
            # Create app-specific log
            cat > $SANDBOX_DIR/applog/application.log << EOF
2025-05-23T00:00:00Z [INFO] Application version 1.2.3 starting
2025-05-23T00:00:01Z [INFO] Loading configuration from environment
2025-05-23T00:00:02Z [ERROR] Required environment variable DB_HOST not set
2025-05-23T00:00:03Z [FATAL] Application failed to start
EOF
            ;;
            
        "empty")
            echo "Setting up empty sandbox scenario (no logs)..."
            # Don't copy any files
            ;;
            
        *)
            echo "Unknown scenario: $scenario"
            echo "Available scenarios: dns-error, oom, multi-log, empty"
            exit 1
            ;;
    esac
    
    echo "Scenario '$scenario' set up successfully!"
    echo "Files in sandbox:"
    ls -la $SANDBOX_DIR | grep -v "^d"
}

# Check if scenario is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <scenario>"
    echo "Available scenarios: dns-error, oom, multi-log, empty"
    exit 1
fi

setup_scenario $1