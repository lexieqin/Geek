# GenesisGpt - AI-Powered Kubernetes Assistant

An intelligent CLI tool that enables natural language interactions with Kubernetes clusters through AI-powered conversations.

## Overview

GenesisGpt is a conversational AI assistant for Kubernetes that understands natural language queries and performs cluster operations. Built using the ReAct (Reasoning and Acting) pattern, it provides an intuitive interface for users to manage Kubernetes resources without memorizing complex kubectl commands.

## Features

- **Natural Language Interface**: Interact with your Kubernetes cluster using plain English
- **ReAct Agent Architecture**: Implements reasoning and acting patterns for intelligent decision-making
- **Comprehensive Kubernetes Operations**: Create, list, delete, and manage various Kubernetes resources
- **Pod Management**: View logs, events, and debug pod issues
- **Context-Aware Conversations**: Maintains conversation history for multi-turn interactions
- **Human-in-the-Loop**: Requests confirmation for critical operations
- **Extensible Tool System**: Modular architecture for adding new capabilities

## Prerequisites

- Go 1.19 or higher
- Access to a Kubernetes cluster
- [ginTools](../ginTools) API server running (default: localhost:8080)
- OpenAI-compatible API key (configured for Alibaba DashScope)

## Installation

```bash
go build -o genesisgpt main.go
```

## Configuration

Set your API key as an environment variable:

```bash
export OPENAI_API_KEY="your-api-key"
```

Ensure ginTools is running:

```bash
cd ../ginTools && ./gintools
```

## Usage

Start an interactive chat session:

```bash
./genesisgpt chat
```

### Example Interactions

```
You: Create a nginx deployment with 3 replicas
GenesisGpt: I'll create an nginx deployment with 3 replicas for you...

You: Show me all pods in the default namespace
GenesisGpt: Let me list the pods in the default namespace...

You: Get logs from the nginx pod
GenesisGpt: I'll retrieve the logs from the nginx pod...

You: Delete the nginx deployment
GenesisGpt: I need to confirm this action. Do you want to delete the deployment "nginx" in namespace "default"?

You: Debug job with ID 81325fc3-b05e-4d9a-ada2-d2399aebe135 for tenant testenv
GenesisGpt: I'll debug the job following the standard workflow...
```

## Available Tools

### 1. CreateTool
Creates Kubernetes resources from natural language descriptions:
- Generates appropriate YAML configurations
- Supports all standard Kubernetes resource types
- Validates resource definitions before creation

### 2. ListTool
Lists and retrieves Kubernetes resources:
- Filter by namespace and resource type
- Get detailed information about specific resources
- Support for all Kubernetes resource types

### 3. DeleteTool
Safely deletes Kubernetes resources:
- Requires confirmation for critical resources
- Supports cascading deletes
- Provides clear feedback on deletion status

### 4. PodTool
Specialized pod operations:
- Retrieve pod logs (with tail and container selection)
- View pod events
- Debug pod issues

### 5. ClusterTool
Cluster information and discovery:
- List available clusters
- Show cluster configuration
- Display cluster status

### 6. ResourceInfoTool
Resource type discovery:
- Get GroupVersionResource (GVR) information
- List available resource types
- Show resource schemas

### 7. HumanTool
Human interaction for confirmations:
- Requests user confirmation for destructive operations
- Provides safety checks for critical actions

## Architecture

```
GenesisGpt/
├── main.go                     # Entry point
├── cmd/
│   ├── root.go                # Root command setup
│   ├── chat.go                # Chat command implementation
│   ├── ai/
│   │   └── message.go         # AI message handling
│   ├── promptTpl/
│   │   └── prompt.go          # ReAct prompt templates
│   ├── tools/                 # Tool implementations
│   │   ├── clustersTool.go
│   │   ├── createTool.go
│   │   ├── deleteTool.go
│   │   ├── humanTool.go
│   │   ├── listTool.go
│   │   ├── podTool.go
│   │   └── resourceInfoTool.go
│   └── utils/
│       └── httpUtils.go       # HTTP communication utilities
```

## How It Works

1. **User Input**: Natural language query from the user
2. **AI Thinking**: The AI analyzes the query and determines the appropriate action
3. **Tool Selection**: Based on the analysis, the AI selects the right tool
4. **Tool Execution**: The selected tool communicates with ginTools API
5. **Result Processing**: The AI interprets the results
6. **Response Generation**: A natural language response is provided to the user

### ReAct Loop

The system follows the ReAct pattern:
```
Thought → Action → Action Input → PAUSE → Observation → (repeat until complete) → Final Answer
```

## Integration with ginTools

GenesisGpt relies on ginTools as its backend API server for actual Kubernetes operations:

- **HTTP Communication**: All Kubernetes operations go through ginTools REST API
- **Resource Management**: Create, read, update, delete operations via ginTools
- **Pod Operations**: Logs and events retrieval through specialized endpoints
- **Error Handling**: Graceful error propagation from ginTools to user

## Extending GenesisGpt

### Adding New Tools

1. Create a new tool file in `cmd/tools/`
2. Implement the tool interface with:
   - Name and description
   - Parameter schema
   - Run method
3. Register the tool in the chat command
4. Update prompt templates if needed

### Customizing AI Behavior

- Modify `cmd/promptTpl/prompt.go` to adjust the ReAct prompt
- Update system prompts for different interaction styles
- Configure AI model parameters in `cmd/chat.go`

## Troubleshooting

### Common Issues

1. **"Connection refused" errors**
   - Ensure ginTools is running on the correct port
   - Check firewall settings

2. **"Unauthorized" errors**
   - Verify your Kubernetes credentials
   - Check RBAC permissions

3. **AI not understanding queries**
   - Try rephrasing with more specific terminology
   - Check API key configuration

## Security Considerations

- API keys are never logged or stored
- All Kubernetes operations respect RBAC permissions
- Confirmation required for destructive operations
- Secure communication with backend services

## Future Enhancements

- Support for custom resource definitions (CRDs)
- Multi-cluster management
- Advanced troubleshooting capabilities
- Integration with monitoring tools
- Voice interface support

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]