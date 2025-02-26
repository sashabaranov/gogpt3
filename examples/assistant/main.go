package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
	ctx := context.Background()

	// Setup the OpenAI client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.OrgID = os.Getenv("OPENAI_ORG_ID")
	client := openai.NewClientWithConfig(config)

	// Create a function that returns the weather in a given location
	f := createWeatherFunction()

	// Create an assistant that uses the function
	assistant, err := createOrModifyAssistant(ctx, client, &f, os.Getenv("OPENAI_ASSISTANT_ID"))
	if err != nil {
		fmt.Printf("Error creating assistant: %v\n", err)
		return
	}

	// Create a thread for this interaction
	fmt.Println("Asking OpenAI 'What is the weather in Boston today?' and providing it a 'get_current_weather()' function...")
	thread, err := client.CreateThread(ctx, openai.ThreadRequest{
		Messages: []openai.ThreadMessage{
			{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"},
		},
	})
	if err != nil {
		fmt.Printf("Error creating a thread: %v\n", err)
		return
	}

	// Create a run for this interaction
	run, err := client.CreateRun(ctx, thread.ID, openai.RunRequest{
		AssistantID:            assistant.ID,
		Model:                  openai.GPT3Dot5Turbo,
		AdditionalInstructions: "Please provide the temperation in Fahrenheit.",
		Tools: []openai.Tool{
			{
				Type:     openai.ToolTypeFunction,
				Function: &f,
			},
		},
	})
	if err != nil {
		fmt.Printf("Error creating a run: %v\n", err)
		return
	}

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	timeout := time.After(time.Second * 60)

	for {
		select {
		case <-timeout:
			fmt.Println("Timeout reached, exiting.")
			return
		case <-ticker.C:
			done, err := pollRun(ctx, client, thread.ID, run.ID, &f)
			if err != nil {
				fmt.Printf("Error polling run: %v\n", err)
				return
			}
			if done {
				printThreadMessages(ctx, client, thread.ID)
				return
			}
		}
	}
}

func createWeatherFunction() openai.FunctionDefinition {
	params := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"location": {
				Type:        jsonschema.String,
				Description: "The city and state, e.g. San Francisco, CA",
			},
			"unit": {
				Type: jsonschema.String,
				Enum: []string{"celsius", "fahrenheit"},
			},
		},
		Required: []string{"location"},
	}
	return openai.FunctionDefinition{
		Name:        "get_current_weather",
		Description: "Get the current weather in a given location",
		Parameters:  params,
	}
}

func createOrModifyAssistant(ctx context.Context, client *openai.Client, f *openai.FunctionDefinition, assistantID string) (openai.Assistant, error) {
	name := "My Weather Assistant"
	description := "Provides the current weather in a given location"
	instructions := "You are a weather app. First determine the location referenced in the user question. If it is unclear, ask for clarification.  Once you have determined the location, use the tool to retrieve the weather and report the weather back to the user."
	request := openai.AssistantRequest{
		Model:        openai.GPT3Dot5Turbo,
		Name:         &name,
		Description:  &description,
		Instructions: &instructions,
		Tools:        []openai.AssistantTool{{Type: openai.AssistantToolTypeFunction, Function: f}},
	}

	if assistantID != "" {
		return client.ModifyAssistant(ctx, assistantID, request)
	} else {
		a, err := client.CreateAssistant(ctx, request)
		if err == nil {
			fmt.Printf("Created assistant %s.\nSave this in the environment variable OPENAI_ASSISTANT_ID for future use.\n", a.ID)
		}
		return a, err
	}
}

func fetchWeather(_ string) (string, error) {
	// This function would normally make an API request to a weather service
	// to get the weather for the given location. For the sake of this example,
	// we'll just return a hardcoded value.
	return "Sunny and 80 degrees.", nil
}

func processToolCalls(toolCalls []openai.ToolCall, f *openai.FunctionDefinition) ([]openai.ToolOutput, error) {
	toolOutputs := []openai.ToolOutput{}
	for _, toolCall := range toolCalls {
		fmt.Printf("OpenAI called us back wanting to invoke our function '%v' with params '%v'\n",
			toolCall.Function.Name, toolCall.Function.Arguments)

		if toolCall.Function.Name == f.Name {
			output, err := fetchWeather(toolCall.Function.Arguments)
			if err != nil {
				return nil, err
			}
			toolOutputs = append(toolOutputs, openai.ToolOutput{
				ToolCallID: toolCall.ID,
				Output:     output,
			})
		}
	}
	return toolOutputs, nil
}

func printThreadMessages(ctx context.Context, client *openai.Client, threadID string) {
	order := "asc"
	mList, err := client.ListMessage(ctx, threadID, nil, &order, nil, nil)
	if err != nil {
		fmt.Printf("Error retrieving thread: %v\n", err)
		return
	}
	for _, m := range mList.Messages {
		for _, c := range m.Content {
			if c.Type == "text" {
				fmt.Printf("%v: %v\n", m.Role, c.Text.Value)
			}
		}
	}
}

func pollRun(ctx context.Context, client *openai.Client, threadID, runID string, f *openai.FunctionDefinition) (bool, error) {
	run, err := client.RetrieveRun(ctx, threadID, runID)
	if err != nil {
		return false, err
	}

	// Check for error statuses
	if run.Status == openai.RunStatusFailed {
		return true, fmt.Errorf("run failed: %v", run.LastError)
	}
	if run.Status == openai.RunStatusCancelled {
		return true, fmt.Errorf("run canceled: %v", run.LastError)
	}
	if run.Status == openai.RunStatusExpired {
		return true, fmt.Errorf("run expired: %v", run.LastError)
	}

	// If OpenAI requires us to submit tool outputs, we should do so
	if run.Status == openai.RunStatusRequiresAction && run.RequiredAction.Type == openai.RequiredActionTypeSubmitToolOutputs {
		toolOutputs, err := processToolCalls(run.RequiredAction.SubmitToolOutputs.ToolCalls, f)
		if err != nil {
			return false, err
		}

		fmt.Println("Sending OpenAI our 'get_current_weather()' function's response")
		_, err = client.SubmitToolOutputs(ctx, threadID, run.ID, openai.SubmitToolOutputsRequest{ToolOutputs: toolOutputs})
		if err != nil {
			return false, err
		}
	}

	if run.Status == openai.RunStatusCompleted {
		return true, nil
	}
	return false, nil
}
