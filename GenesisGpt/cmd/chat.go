/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/lexieqin/Geek/K8sGpt/cmd/ai"
	"github.com/lexieqin/Geek/K8sGpt/cmd/promptTpl"
	"github.com/lexieqin/Geek/K8sGpt/cmd/tools"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		scanner := bufio.NewScanner(cmd.InOrStdin())
		fmt.Println("Hello, I am your K8s assistant. How can I help you? (Type 'exit' to quit):")
		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break
			}
			input := scanner.Text()
			if input == "exit" {
				fmt.Println("Goodbye!")
				return
			}

			prompt := buildPrompt(createTool, listTool, deleteTool, humanTool, clustersTool, podTool, resourceInfoTool, jobDebugTool, sandboxLogTool, intelligentDebugTool, input)
			ai.MessageStore.AddForUser(prompt)
			i := 1
			for {
				first_response := ai.NormalChat(ai.MessageStore.ToMessage())
				fmt.Printf("========Round %d Response========\n", i)
				fmt.Println(first_response.Content)

				regexPattern := regexp.MustCompile(`Final Answer:\s*(.*)`)
				finalAnswer := regexPattern.FindStringSubmatch(first_response.Content)
				if len(finalAnswer) > 1 {
					fmt.Println("========Final GPT Response========")
					fmt.Println(first_response.Content)
					break
				}

				ai.MessageStore.AddForAssistant(first_response.Content)

				regexAction := regexp.MustCompile(`Action:\s*(.*?)[\n]`)
				regexActionInput := regexp.MustCompile(`Action Input:\s*({[\s\S]*?})`)

				action := regexAction.FindStringSubmatch(first_response.Content)
				actionInput := regexActionInput.FindStringSubmatch(first_response.Content)

				if len(action) > 1 && len(actionInput) > 1 {
					i++
					Observation := "Observation: %s"
					if action[1] == createTool.Name {
						var param tools.CreateToolParam
						_ = json.Unmarshal([]byte(actionInput[1]), &param)

						output := createTool.Run(param.Prompt, param.Resource)
						Observation = fmt.Sprintf(Observation, output)
					} else if action[1] == listTool.Name {
						var param tools.ListToolParam
						_ = json.Unmarshal([]byte(actionInput[1]), &param)

						output, _ := listTool.Run(param.Resource, param.Namespace, param.Name, param.Type)
						Observation = fmt.Sprintf(Observation, output)
					} else if action[1] == deleteTool.Name {
						var param tools.DeleteToolParam
						_ = json.Unmarshal([]byte(actionInput[1]), &param)

						err := deleteTool.Run(param.Resource, param.Name, param.Namespace)
						if err != nil {
							Observation = fmt.Sprintf(Observation, "Deletion failed")
						} else {
							Observation = fmt.Sprintf(Observation, "Deletion successful")
						}
					} else if action[1] == humanTool.Name {
						var param tools.HumanToolParam
						_ = json.Unmarshal([]byte(actionInput[1]), &param)

						output := humanTool.Run(param.Prompt)
						Observation = fmt.Sprintf(Observation, output)
					} else if action[1] == clustersTool.Name {
						output, _ := clustersTool.Run()
						Observation = fmt.Sprintf(Observation, output)
					} else if action[1] == podTool.Name {
						var param tools.PodToolParam
						_ = json.Unmarshal([]byte(actionInput[1]), &param)

						output, err := podTool.Run(param)
						if err != nil {
							Observation = fmt.Sprintf(Observation, "Error: "+err.Error())
						} else {
							Observation = fmt.Sprintf(Observation, output)
						}
					} else if action[1] == resourceInfoTool.Name {
						var param tools.ResourceInfoToolParam
						_ = json.Unmarshal([]byte(actionInput[1]), &param)

						output, err := resourceInfoTool.Run(param)
						if err != nil {
							Observation = fmt.Sprintf(Observation, "Error: "+err.Error())
						} else {
							Observation = fmt.Sprintf(Observation, output)
						}
					} else if action[1] == jobDebugTool.Name() {
						output, err := jobDebugTool.Run(actionInput[1])
						if err != nil {
							Observation = fmt.Sprintf(Observation, "Error: "+err.Error())
						} else {
							Observation = fmt.Sprintf(Observation, output)
						}
					} else if action[1] == sandboxLogTool.Name() {
						output, err := sandboxLogTool.Run(actionInput[1])
						if err != nil {
							Observation = fmt.Sprintf(Observation, "Error: "+err.Error())
						} else {
							Observation = fmt.Sprintf(Observation, output)
						}
					} else if action[1] == intelligentDebugTool.Name() {
						output, err := intelligentDebugTool.Run(actionInput[1])
						if err != nil {
							Observation = fmt.Sprintf(Observation, "Error: "+err.Error())
						} else {
							Observation = fmt.Sprintf(Observation, output)
						}
					}

					prompt = first_response.Content + Observation
					fmt.Printf("========Round %d Prompt========\n", i)
					fmt.Println(prompt)
					ai.MessageStore.AddForUser(prompt)
				}
			}
		}
	},
}

func buildPrompt(createTool *tools.CreateTool, listTool *tools.ListTool, deleteTool *tools.DeleteTool, humanTool *tools.HumanTool, clustersTool *tools.ClusterTool, podTool *tools.PodTool, resourceInfoTool *tools.ResourceInfoTool, jobDebugTool *tools.JobDebugTool, sandboxLogTool *tools.SandboxLogTool, intelligentDebugTool *tools.IntelligentDebugTool, query string) string {
	createToolDef := "Name: " + createTool.Name + "\nDescription: " + createTool.Description + "\nArgsSchema: " + createTool.ArgsSchema + "\n"
	listToolDef := "Name: " + listTool.Name + "\nDescription: " + listTool.Description + "\nArgsSchema: " + listTool.ArgsSchema + "\n"
	deleteToolDef := "Name: " + deleteTool.Name + "\nDescription: " + deleteTool.Description + "\nArgsSchema: " + deleteTool.ArgsSchema + "\n"
	humanToolDef := "Name: " + humanTool.Name + "\nDescription: " + humanTool.Description + "\nArgsSchema: " + humanTool.ArgsSchema + "\n"
	clusterToolDef := "Name: " + clustersTool.Name + "\nDescription: " + clustersTool.Description + "\n"
	podToolDef := "Name: " + podTool.Name + "\nDescription: " + podTool.Description + "\nArgsSchema: " + podTool.ArgsSchema + "\n"
	resourceInfoToolDef := "Name: " + resourceInfoTool.Name + "\nDescription: " + resourceInfoTool.Description + "\nArgsSchema: " + resourceInfoTool.ArgsSchema + "\n"
	jobDebugToolDef := "Name: " + jobDebugTool.Name() + "\nDescription: " + jobDebugTool.Description() + "\nArgsSchema: " + jobDebugTool.ArgsSchema() + "\n"
	sandboxLogToolDef := "Name: " + sandboxLogTool.Name() + "\nDescription: " + sandboxLogTool.Description() + "\nArgsSchema: " + sandboxLogTool.ArgsSchema() + "\n"
	intelligentDebugToolDef := "Name: " + intelligentDebugTool.Name() + "\nDescription: " + intelligentDebugTool.Description() + "\nArgsSchema: " + intelligentDebugTool.ArgsSchema() + "\n"

	toolsList := make([]string, 0)
	toolsList = append(toolsList, createToolDef, listToolDef, deleteToolDef, humanToolDef, clusterToolDef, podToolDef, resourceInfoToolDef, jobDebugToolDef, sandboxLogToolDef, intelligentDebugToolDef)

	tool_names := make([]string, 0)
	tool_names = append(tool_names, createTool.Name, listTool.Name, deleteTool.Name, humanTool.Name, clustersTool.Name, podTool.Name, resourceInfoTool.Name, jobDebugTool.Name(), sandboxLogTool.Name(), intelligentDebugTool.Name())

	prompt := fmt.Sprintf(promptTpl.Template, toolsList, tool_names, "", query)

	return prompt
}

func init() {
	rootCmd.AddCommand(chatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
