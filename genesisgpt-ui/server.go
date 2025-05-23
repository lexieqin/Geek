package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/lexieqin/Geek/K8sGpt/cmd/ai"
	"github.com/lexieqin/Geek/K8sGpt/cmd/promptTpl"
	"github.com/lexieqin/Geek/K8sGpt/cmd/tools"
)

type QueryRequest struct {
	Query string `json:"query"`
}

type QueryResponse struct {
	Response string `json:"response"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Global tools
var (
	createTool           *tools.CreateTool
	listTool             *tools.ListTool
	deleteTool           *tools.DeleteTool
	humanTool            *tools.HumanTool
	clustersTool         *tools.ClusterTool
	podTool              *tools.PodTool
	resourceInfoTool     *tools.ResourceInfoTool
	jobDebugTool         *tools.JobDebugTool
	sandboxLogTool       *tools.SandboxLogTool
	intelligentDebugTool *tools.IntelligentDebugTool
)

func initTools() {
	createTool = tools.NewCreateTool()
	listTool = tools.NewListTool()
	deleteTool = tools.NewDeleteTool()
	humanTool = tools.NewHumanTool()
	clustersTool = tools.NewClusterTool()
	podTool = tools.NewPodTool()
	resourceInfoTool = tools.NewResourceInfoTool()
	jobDebugTool = tools.NewJobDebugTool()
	sandboxLogTool = tools.NewSandboxLogTool()
	intelligentDebugTool = tools.NewIntelligentDebugTool()
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Process the query
	response := processQuery(req.Query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QueryResponse{Response: response})
}

func processQuery(query string) string {
	// Build prompt
	prompt := buildPrompt(createTool, listTool, deleteTool, humanTool, clustersTool, 
		podTool, resourceInfoTool, jobDebugTool, sandboxLogTool, intelligentDebugTool, query)
	
	// Initialize message store for this query
	messageStore := &ai.MessageStore{}
	messageStore.AddForUser(prompt)
	
	maxRounds := 10
	for i := 1; i <= maxRounds; i++ {
		response := ai.NormalChat(messageStore.ToMessage())
		
		// Check for final answer
		finalAnswerPattern := regexp.MustCompile(`Final Answer:\s*(.*)`)
		if matches := finalAnswerPattern.FindStringSubmatch(response.Content); len(matches) > 1 {
			return matches[1]
		}
		
		messageStore.AddForAssistant(response.Content)
		
		// Extract action and input
		actionPattern := regexp.MustCompile(`Action:\s*(.*?)[\n]`)
		actionInputPattern := regexp.MustCompile(`Action Input:\s*({[\s\S]*?})`)
		
		action := actionPattern.FindStringSubmatch(response.Content)
		actionInput := actionInputPattern.FindStringSubmatch(response.Content)
		
		if len(action) > 1 && len(actionInput) > 1 {
			observation := executeAction(action[1], actionInput[1])
			messageStore.AddForUser(response.Content + observation)
		} else {
			// If no valid action found, return the response
			return response.Content
		}
	}
	
	return "I couldn't complete the task within the maximum number of steps. Please try rephrasing your query."
}

func executeAction(actionName, actionInput string) string {
	observation := "Observation: "
	
	switch actionName {
	case "IntelligentDebugTool":
		output, err := intelligentDebugTool.Run(actionInput)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
	case "JobDebugTool":
		output, err := jobDebugTool.Run(actionInput)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
	case "SandboxLogTool":
		output, err := sandboxLogTool.Run(actionInput)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
	case "CreateTool":
		var param tools.CreateToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output := createTool.Run(param.Prompt, param.Resource)
		observation += output
	case "ListTool":
		var param tools.ListToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output, _ := listTool.Run(param.Resource, param.Namespace, param.Name, param.Type)
		observation += output
	case "DeleteTool":
		var param tools.DeleteToolParam
		json.Unmarshal([]byte(actionInput), &param)
		err := deleteTool.Run(param.Resource, param.Name, param.Namespace)
		if err != nil {
			observation += "Deletion failed: " + err.Error()
		} else {
			observation += "Deletion successful"
		}
	case "PodTool":
		var param tools.PodToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output, err := podTool.Run(param)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
	case "ClusterTool":
		output := clustersTool.Run()
		observation += output
	case "ResourceInfoTool":
		var param tools.ResourceInfoToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output, err := resourceInfoTool.Run(param)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
	case "HumanTool":
		observation += "Human confirmation required. Please confirm the action."
	default:
		observation += fmt.Sprintf("Unknown action: %s", actionName)
	}
	
	return observation
}

func buildPrompt(createTool *tools.CreateTool, listTool *tools.ListTool, deleteTool *tools.DeleteTool, 
	humanTool *tools.HumanTool, clustersTool *tools.ClusterTool, podTool *tools.PodTool, 
	resourceInfoTool *tools.ResourceInfoTool, jobDebugTool *tools.JobDebugTool, 
	sandboxLogTool *tools.SandboxLogTool, intelligentDebugTool *tools.IntelligentDebugTool, query string) string {
	
	// Build tool definitions
	toolDefs := []string{
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", createTool.Name, createTool.Description, createTool.ArgsSchema),
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", listTool.Name, listTool.Description, listTool.ArgsSchema),
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", deleteTool.Name, deleteTool.Description, deleteTool.ArgsSchema),
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", humanTool.Name, humanTool.Description, humanTool.ArgsSchema),
		fmt.Sprintf("Name: %s\nDescription: %s\n", clustersTool.Name, clustersTool.Description),
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", podTool.Name, podTool.Description, podTool.ArgsSchema),
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", resourceInfoTool.Name, resourceInfoTool.Description, resourceInfoTool.ArgsSchema),
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", jobDebugTool.Name(), jobDebugTool.Description(), jobDebugTool.ArgsSchema()),
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", sandboxLogTool.Name(), sandboxLogTool.Description(), sandboxLogTool.ArgsSchema()),
		fmt.Sprintf("Name: %s\nDescription: %s\nArgsSchema: %s\n", intelligentDebugTool.Name(), intelligentDebugTool.Description(), intelligentDebugTool.ArgsSchema()),
	}
	
	toolNames := []string{
		createTool.Name, listTool.Name, deleteTool.Name, humanTool.Name, 
		clustersTool.Name, podTool.Name, resourceInfoTool.Name, 
		jobDebugTool.Name(), sandboxLogTool.Name(), intelligentDebugTool.Name(),
	}
	
	return fmt.Sprintf(promptTpl.Template, toolDefs, toolNames, "", query)
}

func main() {
	// Initialize tools
	initTools()
	
	// Serve static files
	http.Handle("/", http.FileServer(http.Dir(".")))
	
	// API endpoint
	http.HandleFunc("/api/query", handleQuery)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	fmt.Printf("K8sGPT UI Server running on port %s\n", port)
	fmt.Printf("Open http://localhost:%s in your browser\n", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}