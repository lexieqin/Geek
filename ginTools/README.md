# Kubernetes API Gateway

A lightweight API gateway built with Go and Gin framework that provides a RESTful interface to interact with Kubernetes clusters. This project is designed to support AI agents and other tools in managing Kubernetes resources.

## Features

- Generic resource management (CRUD operations)
- Pod logs and events retrieval
- Dynamic resource type handling
- RESTful API design following Kubernetes conventions

## Prerequisites

- Go 1.22 or later
- Access to a Kubernetes cluster
- Valid kubeconfig file (usually at `~/.kube/config`)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/lexieqin/Geek/tree/main/ginTools
cd ginTools
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the server:
```bash
go run main.go
```

The server will start on port 8080.

## API Documentation

### Resource Management

#### List Resources
```http
GET /:resource?ns=<namespace>
```
Lists all resources of the specified type in the given namespace.

Example:
```bash
# List all deployments in kube-system namespace
curl http://localhost:8080/deployments?ns=kube-system
```

#### Get Resource Details
```http
GET /:resource?ns=<namespace>&name=<resource-name>
```
Gets details of a specific resource.

Example:
```bash
# Get details of a specific deployment
curl http://localhost:8080/deployments?ns=default&name=my-deployment
```

#### Create Resource
```http
POST /:resource
Content-Type: application/json

{
    "yaml": "<yaml-content>"
}
```
Creates a new resource from YAML.

Example:
```bash
curl -X POST http://localhost:8080/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "yaml": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: example-deployment\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:latest"
  }'
```

#### Update Resource
```http
PUT /:resource?ns=<namespace>&name=<resource-name>
Content-Type: application/json

{
    "yaml": "<yaml-content>"
}
```
Updates an existing resource.

Example:
```bash
curl -X PUT http://localhost:8080/deployments?ns=default&name=example-deployment \
  -H "Content-Type: application/json" \
  -d '{
    "yaml": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: example-deployment\nspec:\n  replicas: 5\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.19"
  }'
```

#### Patch Resource
```http
PATCH /:resource?ns=<namespace>&name=<resource-name>
Content-Type: application/json

{
    "patch": "<json-patch>"
}
```
Partially updates a resource using JSON patch.

Example:
```bash
curl -X PATCH http://localhost:8080/deployments?ns=default&name=example-deployment \
  -H "Content-Type: application/json" \
  -d '{
    "patch": "[{\"op\": \"replace\", \"path\": \"/spec/replicas\", \"value\": 3}]"
  }'
```

#### Delete Resource
```http
DELETE /:resource?ns=<namespace>&name=<resource-name>
```
Deletes a resource.

Example:
```bash
curl -X DELETE http://localhost:8080/deployments?ns=default&name=example-deployment
```

### Pod Operations

#### List Pods
```http
GET /namespaces/:namespace/pods
```
Lists all pods in the specified namespace.

Example:
```bash
# List all pods in kube-system namespace
curl http://localhost:8080/namespaces/kube-system/pods

# List all pods in default namespace
curl http://localhost:8080/namespaces/default/pods
```

#### Get Pod Details
```http
GET /namespaces/:namespace/pods/:podName
```
Gets details of a specific pod.

Example:
```bash
curl http://localhost:8080/namespaces/default/pods/my-pod
```

#### Get Pod Logs
```http
GET /namespaces/:namespace/pods/:podName/logs?container=<container-name>&tail=<number-of-lines>
```
Gets logs for a specific pod.

Example:
```bash
# Get last 100 lines of logs
curl http://localhost:8080/namespaces/default/pods/my-pod/logs

# Get last 50 lines of logs from a specific container
curl http://localhost:8080/namespaces/default/pods/my-pod/logs?container=nginx&tail=50
```

#### Get Pod Events
```http
GET /namespaces/:namespace/pods/:podName/events?type=<event-type>
```
Gets events related to a specific pod.

Example:
```bash
# Get all events
curl http://localhost:8080/namespaces/default/pods/my-pod/events

# Get only warning events
curl http://localhost:8080/namespaces/default/pods/my-pod/events?type=Warning
```

### Resource Type Information

#### Get GVR (GroupVersionResource)
```http
GET /get/gvr?resource=<resource-type>
```
Gets the GroupVersionResource information for a resource type.

Example:
```bash
curl http://localhost:8080/get/gvr?resource=pods
```

#### Get Resource List
```http
GET /get/resource?resource=<resource-type>
```
Gets a list of resources of the specified type.

Example:
```bash
curl http://localhost:8080/get/resource?resource=pods
```

#### Get Resources by Type
```http
POST /get/resource?resource=<resource-type>&type=<type>
```
Gets resources filtered by type.

Example:
```bash
curl -X POST http://localhost:8080/get/resource?resource=pods&type=app
```

## Response Format

All API responses follow a consistent format:

```json
{
    "data": <response-data>,
    "meta": {
        // Additional metadata
    }
}
```

For errors:
```json
{
    "error": "Error message"
}
```

## Error Handling

The API uses standard HTTP status codes:
- 200: Success
- 400: Bad Request
- 404: Not Found
- 500: Internal Server Error

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 