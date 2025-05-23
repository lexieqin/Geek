package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/lexieqin/Geek/K8sGpt/cmd/ai"
	"github.com/lexieqin/Geek/K8sGpt/cmd/tools"
)

// Session management
var (
	sessions = make(map[string]*Session)
	sessionsMutex sync.RWMutex
)

type Session struct {
	ID                    string
	MessageStore          ai.ChatMessages
	LastAccessed          time.Time
	PendingConfirmation   bool
	ConfirmationPrompt    string
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run K8sGpt as an HTTP server",
	Long:  `Run K8sGpt as an HTTP server that accepts queries via POST requests`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set server mode
		os.Setenv("K8SGPT_SERVER_MODE", "true")
		
		// Initialize tools
		createTool := tools.NewCreateTool()
		listTool := tools.NewListTool()
		deleteTool := tools.NewDeleteTool()
		humanTool := tools.NewHumanTool()
		clustersTool := tools.NewClusterTool()
		podTool := tools.NewPodTool()
		resourceInfoTool := tools.NewResourceInfoTool()
		jobDebugTool := tools.NewJobDebugTool()
		sandboxLogTool := tools.NewSandboxLogTool()
		intelligentDebugTool := tools.NewIntelligentDebugTool()

		http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			var request struct {
				Query               string `json:"query"`
				ShowThinkingProcess bool   `json:"showThinkingProcess"`
				SessionID           string `json:"sessionId"`
			}

			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Process the query
			fmt.Printf("Received query: %s (session: %s, show thinking: %v)\n", request.Query, request.SessionID, request.ShowThinkingProcess)
			response, sessionID := processQueryWithSession(request.Query, request.SessionID, request.ShowThinkingProcess, 
				createTool, listTool, deleteTool, humanTool, clustersTool, podTool, resourceInfoTool, 
				jobDebugTool, sandboxLogTool, intelligentDebugTool)
			fmt.Printf("Sending response: %s\n", response)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"response": response,
				"sessionId": sessionID,
			})
		})

		port := os.Getenv("PORT")
		if port == "" {
			port = "8090"
		}

		fmt.Printf("K8sGpt server listening on port %s\n", port)
		http.ListenAndServe(":"+port, nil)
	},
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func getOrCreateSession(sessionID string) *Session {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	
	// Clean up old sessions (older than 30 minutes)
	for id, session := range sessions {
		if time.Since(session.LastAccessed) > 30*time.Minute {
			delete(sessions, id)
		}
	}
	
	// Get or create session
	if sessionID != "" {
		if session, exists := sessions[sessionID]; exists {
			session.LastAccessed = time.Now()
			return session
		}
	}
	
	// Create new session
	newID := sessionID
	if newID == "" {
		newID = generateSessionID()
	}
	
	messageStore := make(ai.ChatMessages, 0)
	messageStore.Clear() // Initialize with system prompt
	
	session := &Session{
		ID:           newID,
		MessageStore: messageStore,
		LastAccessed: time.Now(),
	}
	sessions[newID] = session
	return session
}

func processQueryWithSession(query, sessionID string, showThinkingProcess bool, 
	createTool *tools.CreateTool, listTool *tools.ListTool, deleteTool *tools.DeleteTool, 
	humanTool *tools.HumanTool, clustersTool *tools.ClusterTool, podTool *tools.PodTool, 
	resourceInfoTool *tools.ResourceInfoTool, jobDebugTool *tools.JobDebugTool,
	sandboxLogTool *tools.SandboxLogTool, intelligentDebugTool *tools.IntelligentDebugTool) (string, string) {
	
	// Get or create session
	session := getOrCreateSession(sessionID)
	
	// Check if this is a response to a pending confirmation
	if session.PendingConfirmation && (strings.ToLower(strings.TrimSpace(query)) == "yes" || strings.ToLower(strings.TrimSpace(query)) == "no") {
		// Add the human confirmation response to the conversation
		session.MessageStore.AddForUser(fmt.Sprintf("Observation: %s", strings.TrimSpace(query)))
		session.PendingConfirmation = false
		session.ConfirmationPrompt = ""
		
		// Continue processing from where we left off
		response := processQueryWithSessionObj("", showThinkingProcess, session, createTool, listTool, 
			deleteTool, humanTool, clustersTool, podTool, resourceInfoTool, jobDebugTool, 
			sandboxLogTool, intelligentDebugTool)
		
		return response, session.ID
	}
	
	// Process query with session's message store
	response := processQueryWithSessionObj(query, showThinkingProcess, session, createTool, listTool, 
		deleteTool, humanTool, clustersTool, podTool, resourceInfoTool, jobDebugTool, 
		sandboxLogTool, intelligentDebugTool)
	
	return response, session.ID
}

func processQueryWithSessionObj(query string, showThinkingProcess bool, session *Session,
	createTool *tools.CreateTool, listTool *tools.ListTool, 
	deleteTool *tools.DeleteTool, humanTool *tools.HumanTool, clustersTool *tools.ClusterTool, 
	podTool *tools.PodTool, resourceInfoTool *tools.ResourceInfoTool, jobDebugTool *tools.JobDebugTool,
	sandboxLogTool *tools.SandboxLogTool, intelligentDebugTool *tools.IntelligentDebugTool) string {
	
	// Build prompt
	if query != "" {
		prompt := buildServerPrompt(createTool, listTool, deleteTool, humanTool, clustersTool, 
			podTool, resourceInfoTool, jobDebugTool, sandboxLogTool, intelligentDebugTool, query)
		
		// Use the session's messageStore to maintain context
		session.MessageStore.AddForUser(prompt)
	}
	
	var fullConversation strings.Builder
	
	// Process with AI
	maxRounds := 10
	for i := 1; i <= maxRounds; i++ {
		fmt.Printf("Round %d - Calling AI...\n", i)
		response := ai.NormalChat(session.MessageStore.ToMessage())
		fmt.Printf("AI Response: %s\n", response.Content)
		
		// Add to full conversation if showing thinking process
		if showThinkingProcess {
			fullConversation.WriteString(fmt.Sprintf("**Round %d:**\n", i))
			fullConversation.WriteString(response.Content)
			fullConversation.WriteString("\n\n")
		}
		
		// Check for final answer (capture everything after "Final Answer:")
		if strings.Contains(response.Content, "Final Answer:") {
			parts := strings.SplitN(response.Content, "Final Answer:", 2)
			if len(parts) == 2 {
				finalAnswer := strings.TrimSpace(parts[1])
				if showThinkingProcess {
					fullConversation.WriteString("---\n\n**Final Answer:**\n")
					fullConversation.WriteString(finalAnswer)
					return fullConversation.String()
				}
				return finalAnswer
			}
		}
		
		session.MessageStore.AddForAssistant(response.Content)
		
		// Extract and execute action
		actionRe := regexp.MustCompile(`Action:\s*(.*?)[\n]`)
		actionInputRe := regexp.MustCompile(`Action Input:\s*({[\s\S]*?})`)
		
		action := actionRe.FindStringSubmatch(response.Content)
		actionInput := actionInputRe.FindStringSubmatch(response.Content)
		
		if len(action) > 1 && len(actionInput) > 1 {
			observation := executeAction(action[1], actionInput[1], createTool, listTool, 
				deleteTool, humanTool, clustersTool, podTool, resourceInfoTool, 
				jobDebugTool, sandboxLogTool, intelligentDebugTool)
			
			// Check if human confirmation is required
			if strings.Contains(observation, "[HUMAN_CONFIRMATION_REQUIRED]") {
				// Extract the confirmation prompt
				confirmationRe := regexp.MustCompile(`\[HUMAN_CONFIRMATION_REQUIRED\]: (.+)`)
				matches := confirmationRe.FindStringSubmatch(observation)
				if len(matches) > 1 {
					confirmPrompt := matches[1]
					session.PendingConfirmation = true
					session.ConfirmationPrompt = confirmPrompt
					
					if showThinkingProcess {
						fullConversation.WriteString("\n")
						fullConversation.WriteString(observation)
						fullConversation.WriteString("\n\n---\n\n**Human Confirmation Required:**\n")
						fullConversation.WriteString(confirmPrompt)
						fullConversation.WriteString("\n\nPlease respond with 'yes' or 'no' to continue.")
						return fullConversation.String()
					}
					return fmt.Sprintf("**Confirmation Required:** %s\n\nPlease respond with 'yes' or 'no' to continue.", confirmPrompt)
				}
			}
			
			if showThinkingProcess {
				fullConversation.WriteString("\n")
				fullConversation.WriteString(observation)
				fullConversation.WriteString("\n\n")
			}
			
			prompt := response.Content + observation
			session.MessageStore.AddForUser(prompt)
		} else {
			// No valid action, return current response
			if showThinkingProcess {
				return fullConversation.String() + "\n\n**Note:** Process ended without a clear final answer."
			}
			return response.Content
		}
	}
	
	return "I couldn't complete the task within the allowed steps. Please try a simpler query."
}


func executeAction(actionName, actionInput string, createTool *tools.CreateTool, listTool *tools.ListTool,
	deleteTool *tools.DeleteTool, humanTool *tools.HumanTool, clustersTool *tools.ClusterTool,
	podTool *tools.PodTool, resourceInfoTool *tools.ResourceInfoTool, jobDebugTool *tools.JobDebugTool,
	sandboxLogTool *tools.SandboxLogTool, intelligentDebugTool *tools.IntelligentDebugTool) string {
	
	observation := "Observation: "
	
	switch actionName {
	case createTool.Name:
		var param tools.CreateToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output := createTool.Run(param.Prompt, param.Resource)
		observation += output
		
	case listTool.Name:
		var param tools.ListToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output, _ := listTool.Run(param.Resource, param.Namespace, param.Name, param.Type)
		observation += output
		
	case deleteTool.Name:
		var param tools.DeleteToolParam
		json.Unmarshal([]byte(actionInput), &param)
		err := deleteTool.Run(param.Resource, param.Name, param.Namespace)
		if err != nil {
			observation += "Deletion failed: " + err.Error()
		} else {
			observation += "Deletion successful"
		}
		
	case humanTool.Name:
		var param tools.HumanToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output := humanTool.Run(param.Prompt)
		observation += output
		
	case clustersTool.Name:
		output, _ := clustersTool.Run()
		observation += output
		
	case podTool.Name:
		var param tools.PodToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output, err := podTool.Run(param)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
		
	case resourceInfoTool.Name:
		var param tools.ResourceInfoToolParam
		json.Unmarshal([]byte(actionInput), &param)
		output, err := resourceInfoTool.Run(param)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
		
	case jobDebugTool.Name():
		output, err := jobDebugTool.Run(actionInput)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
		
	case sandboxLogTool.Name():
		output, err := sandboxLogTool.Run(actionInput)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
		
	case intelligentDebugTool.Name():
		output, err := intelligentDebugTool.Run(actionInput)
		if err != nil {
			observation += "Error: " + err.Error()
		} else {
			observation += output
		}
		
	default:
		observation += fmt.Sprintf("Unknown action: %s", actionName)
	}
	
	return observation
}

func buildServerPrompt(createTool *tools.CreateTool, listTool *tools.ListTool, deleteTool *tools.DeleteTool, 
	humanTool *tools.HumanTool, clustersTool *tools.ClusterTool, podTool *tools.PodTool, 
	resourceInfoTool *tools.ResourceInfoTool, jobDebugTool *tools.JobDebugTool, 
	sandboxLogTool *tools.SandboxLogTool, intelligentDebugTool *tools.IntelligentDebugTool, query string) string {
	// For now, use the same logic as chat - we could refactor this into a shared package
	return buildPrompt(createTool, listTool, deleteTool, humanTool, clustersTool, 
		podTool, resourceInfoTool, jobDebugTool, sandboxLogTool, intelligentDebugTool, query)
}

func init() {
	rootCmd.AddCommand(serverCmd)
}