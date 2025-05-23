# Job Debugging Guide for GenesisGpt

This guide explains how to use the enhanced GenesisGpt for debugging failed Kubernetes jobs with integrated Datadog traces, error analysis, and sandbox logs.

## Overview

The job debugging feature provides comprehensive troubleshooting capabilities for failed Kubernetes jobs by:
- Finding jobs by UUID or name
- Extracting Datadog trace links
- Analyzing job errors and pod failures
- Accessing sandbox logs
- Correlating events and pod information

## Prerequisites

1. **Job Annotations/Labels**: Your Kubernetes jobs should include:
   - UUID label: `job-uuid: <uuid>` or annotation `job-uuid: <uuid>`
   - Datadog trace annotations:
     - `datadog.trace.url: <url>`
     - `datadog.trace.id: <trace-id>`
     - `dd.trace.link: <link>`
   - Sandbox URL annotation: `sandbox.url: <url>`

2. **Services Running**:
   - ginTools API server on port 8080
   - GenesisGpt with JobDebugTool enabled

## Usage Examples

### 1. Debug a Failed Job by UUID

```
You: Debug the failed job with UUID abc-123-def-456
GenesisGpt: I'll help you debug the failed job with UUID abc-123-def-456...

=== Job Summary ===
Name: default/data-processing-job-xyz
UUID: abc-123-def-456
Status: Failed

=== Trace Information ===
Datadog URL: https://app.datadoghq.com/trace/1234567890
Trace ID: abc123def456
Trace Link: https://app.datadoghq.com/apm/trace/1234567890

=== Error Details ===
Reason: BackoffLimitExceeded
Message: Job has reached the specified backoff limit

Pod Errors:
  - Pod: data-processing-job-xyz-abcde, Container: main
    Reason: Error
    Message: Container failed with exit code 1

=== Log Information ===
Sandbox URL: https://sandbox.example.com/logs/abc-123-def-456

=== Events ===
  - [Warning] FailedCreate: Error creating pod
  - [Warning] BackoffLimitExceeded: Job has reached the specified backoff limit
```

### 2. Debug a Job by Name

```
You: Debug the job named etl-job in namespace production
K8sGpt: I'll retrieve debug information for the job etl-job in namespace production...
```

### 3. Get Only Trace Information

```
You: Get the Datadog traces for job process-data in namespace default
K8sGpt: I'll fetch the trace information for the job...
```

### 4. Analyze Job Errors

```
You: What errors occurred in job batch-processing with UUID xyz-789?
K8sGpt: Let me check the errors for that job...
```

## How It Works

### Architecture

```
User Query
    ↓
K8sGpt (JobDebugTool)
    ↓
ginTools API (/jobs/debug endpoints)
    ↓
Kubernetes API
    ↓
Job + Pods + Events
    ↓
Formatted Debug Info
```

### Data Flow

1. **UUID Lookup**: If UUID provided, searches across all namespaces using label selectors
2. **Job Retrieval**: Gets job details including status, labels, and annotations
3. **Trace Extraction**: Parses Datadog-related annotations
4. **Error Analysis**: Examines job conditions and pod termination states
5. **Log Collection**: Identifies sandbox URLs and container log locations
6. **Event Correlation**: Fetches related Kubernetes events
7. **Pod Association**: Lists all pods created by the job

## Setting Up Jobs for Debugging

### Example Job with Debug Annotations

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: data-processing-job
  namespace: default
  labels:
    job-uuid: "abc-123-def-456"
  annotations:
    datadog.trace.url: "https://app.datadoghq.com/trace/1234567890"
    datadog.trace.id: "abc123def456"
    sandbox.url: "https://sandbox.example.com/logs/abc-123-def-456"
spec:
  template:
    metadata:
      labels:
        job-uuid: "abc-123-def-456"
    spec:
      containers:
      - name: main
        image: myapp:latest
        env:
        - name: DD_TRACE_ID
          value: "abc123def456"
```

## API Endpoints (ginTools)

The following endpoints are available for direct API access:

- `GET /jobs/:namespace/:name/debug` - Complete debug information
- `GET /jobs/:namespace/:name/traces` - Datadog trace links only
- `GET /jobs/:namespace/:name/errors` - Error details only
- `GET /jobs/:namespace/:name/sandbox` - Sandbox log information
- `GET /jobs/:namespace/:name/pods` - Associated pods
- `GET /jobs/uuid/:uuid` - Find job by UUID

## Extending the Debug Capabilities

### Adding Custom Annotations

You can extend the debugging information by adding custom annotations to your jobs:

```yaml
annotations:
  # Application-specific
  app.version: "1.2.3"
  app.commit: "abc123"
  
  # Monitoring
  prometheus.io/query: "job_processing_duration"
  grafana.dashboard: "https://grafana.example.com/d/abc123"
  
  # CI/CD
  jenkins.build.url: "https://jenkins.example.com/job/123"
  gitlab.pipeline.url: "https://gitlab.example.com/pipeline/456"
```

### Integrating with External Tools

The JobDebugTool can be extended to:
1. Fetch actual Datadog traces using the WebFetch tool
2. Query Prometheus metrics for the job
3. Access external logging systems
4. Integrate with CI/CD pipelines

## Troubleshooting

### Common Issues

1. **Job Not Found by UUID**
   - Ensure the job has the correct label: `job-uuid: <uuid>`
   - Check if the job exists in the specified namespace
   - Try searching without namespace filter

2. **Missing Trace Information**
   - Verify annotations are properly set on the job
   - Check annotation keys match expected format
   - Ensure values are valid URLs

3. **Empty Error Details**
   - Job might still be running (not failed yet)
   - Check pod status for more details
   - Review events for additional information

### Debug Commands

```bash
# Check if ginTools is running
curl http://localhost:8080/jobs/default/test-job/debug

# Test UUID lookup
curl http://localhost:8080/jobs/uuid/abc-123

# Get raw job YAML to check annotations
kubectl get job <job-name> -o yaml
```

## Best Practices

1. **Consistent UUID Format**: Use a standard UUID format across all jobs
2. **Annotation Standards**: Define organization-wide annotation keys
3. **Error Handling**: Always include meaningful error messages in jobs
4. **Log Retention**: Ensure sandbox logs are retained for debugging
5. **Trace Correlation**: Use the same trace ID across all systems

## Future Enhancements

- Automatic trace fetching from Datadog API
- Integration with log aggregation systems
- ML-based error pattern recognition
- Automated remediation suggestions
- Cross-cluster job debugging