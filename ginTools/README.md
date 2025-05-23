# ginTools - Kubernetes API Gateway

A lightweight HTTP API gateway for Kubernetes operations, designed to provide RESTful endpoints for AI agents and other tools to interact with Kubernetes clusters.

## Overview

ginTools serves as a bridge between HTTP clients and the Kubernetes API, offering both generic resource management capabilities and specialized pod operations. Built with Go and the Gin web framework, it provides a clean REST API interface for Kubernetes resource manipulation.

## Features

- **Generic Resource Management**: Create, read, update, and delete any Kubernetes resource type
- **Dynamic Resource Discovery**: Automatically discovers and handles all resource types in your cluster
- **Pod Operations**: Specialized endpoints for pod logs and events
- **JSON Patch Support**: Apply partial updates to resources using JSON Patch
- **Namespace Support**: All operations are namespace-aware
- **Clean REST API**: Intuitive HTTP endpoints following REST conventions

## Prerequisites

- Go 1.19 or higher
- Access to a Kubernetes cluster
- Valid kubeconfig file at `~/.kube/config`

## Installation

```bash
go build -o gintools main.go
```

## Usage

Start the server:

```bash
./gintools
```

The server will start on port 8080 by default.

## API Endpoints

### Generic Resource Operations

- **List Resources**
  ```
  GET /:resource?namespace=<namespace>
  ```

- **Create Resource**
  ```
  POST /:resource
  Content-Type: application/x-yaml
  Body: <YAML content>
  ```

- **Update Resource**
  ```
  PUT /:resource?name=<name>&namespace=<namespace>
  Content-Type: application/x-yaml
  Body: <YAML content>
  ```

- **Patch Resource**
  ```
  PATCH /:resource?name=<name>&namespace=<namespace>
  Content-Type: application/json-patch+json
  Body: <JSON Patch array>
  ```

- **Delete Resource**
  ```
  DELETE /:resource?name=<name>&namespace=<namespace>
  ```

- **Get Resource Status**
  ```
  GET /:resource/status?name=<name>&namespace=<namespace>
  ```

### Resource Discovery

- **Get GroupVersionResource Info**
  ```
  GET /get/gvr?resource=<resource>
  ```

- **Get All Resources of a Type**
  ```
  GET /get/resource?resource=<resource>&namespace=<namespace>
  ```

- **Get Filtered Resources**
  ```
  POST /get/resource
  Body: { "resource": "<resource>", "namespace": "<namespace>" }
  ```

### Pod-Specific Operations

- **List Pods**
  ```
  GET /namespaces/:namespace/pods
  ```

- **Get Pod Details**
  ```
  GET /namespaces/:namespace/pods/:podName
  ```

- **Get Pod Logs**
  ```
  GET /namespaces/:namespace/pods/:podName/logs?container=<container>&previous=<bool>&tail=<lines>
  ```

- **Get Pod Events**
  ```
  GET /namespaces/:namespace/pods/:podName/events
  ```

## Example Usage

### Create a Deployment

```bash
curl -X POST http://localhost:8080/deployments \
  -H "Content-Type: application/x-yaml" \
  -d @deployment.yaml
```

### Get Pod Logs

```bash
curl http://localhost:8080/namespaces/default/pods/nginx-pod/logs?tail=100
```

### Apply JSON Patch

```bash
curl -X PATCH http://localhost:8080/deployments?name=nginx&namespace=default \
  -H "Content-Type: application/json-patch+json" \
  -d '[{"op": "replace", "path": "/spec/replicas", "value": 3}]'
```

## Architecture

```
├── main.go                 # Application entry point and route definitions
├── pkg/
│   ├── config/
│   │   └── k8sconfig.go   # Kubernetes client configuration
│   ├── controllers/
│   │   ├── resourceCtl.go      # Generic resource controller
│   │   └── podLogEventCtl.go   # Pod-specific operations controller
│   └── services/
│       ├── resourceService.go      # Generic resource business logic
│       └── podLogEventService.go   # Pod operations business logic
```

## Configuration

The application uses the standard Kubernetes configuration from `~/.kube/config`. It supports both in-cluster and out-of-cluster configurations.

### Environment Variables

- `KUBECONFIG`: Path to kubeconfig file (default: `~/.kube/config`)
- `PORT`: Server port (default: 8080)

## Error Handling

All endpoints return appropriate HTTP status codes:
- `200 OK`: Successful operation
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server-side error

Error responses include detailed messages in JSON format:
```json
{
  "error": "error description"
}
```

## Integration with AI Agents

ginTools is designed to work seamlessly with AI agents like K8sGpt, providing a simple HTTP interface that AI models can easily understand and use for Kubernetes operations.

## Security Considerations

- Ensure proper RBAC permissions for the service account
- Consider implementing authentication/authorization middleware
- Use TLS for production deployments
- Validate all input parameters

## License

[Add your license information here]