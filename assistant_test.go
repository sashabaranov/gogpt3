package openai_test

import (
	"context"

	openai "github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test/checks"

	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TestAssistant Tests the assistant endpoint of the API using the mocked server.
func TestAssistant(t *testing.T) {
	assistantID := "asst_abc123"
	assistantName := "Ambrogio"
	assistantDescription := "Ambrogio is a friendly assistant."
	assistantInstructions := `You are a personal math tutor. 
When asked a question, write and run Python code to answer the question.`
	assistantFileID := "file-wB6RM6wHdA49HfS2DJ9fEyrH"
	limit := 20
	order := "desc"
	after := "asst_abc122"
	before := "asst_abc124"

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/assistants/"+assistantID+"/files/"+assistantFileID,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.AssistantFile{
					ID:          assistantFileID,
					Object:      "assistant.file",
					CreatedAt:   1234567890,
					AssistantID: assistantID,
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodDelete {
				fmt.Fprintln(w, `{
					id: "file-wB6RM6wHdA49HfS2DJ9fEyrH",
					object: "assistant.file.deleted",
					deleted: true
				  }`)
			}
		},
	)

	server.RegisterHandler(
		"/v1/assistants/"+assistantID+"/files",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.AssistantFilesList{
					AssistantFiles: []openai.AssistantFile{
						{
							ID:          assistantFileID,
							Object:      "assistant.file",
							CreatedAt:   1234567890,
							AssistantID: assistantID,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodPost {
				var request openai.AssistantFileRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.AssistantFile{
					ID:          request.FileID,
					Object:      "assistant.file",
					CreatedAt:   1234567890,
					AssistantID: assistantID,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/assistants/"+assistantID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.Assistant{
					ID:           assistantID,
					Object:       "assistant",
					CreatedAt:    1234567890,
					Name:         &assistantName,
					Model:        openai.GPT4TurboPreview,
					Description:  &assistantDescription,
					Instructions: &assistantInstructions,
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodPost:
				var request openai.Assistant
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.Assistant{
					ID:           assistantID,
					Object:       "assistant",
					CreatedAt:    1234567890,
					Name:         request.Name,
					Model:        request.Model,
					Description:  request.Description,
					Instructions: request.Instructions,
					Tools:        request.Tools,
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodDelete:
				fmt.Fprintln(w, `{
					"id": "asst_abc123",
					"object": "assistant.deleted",
					"deleted": true
				  }`)
			}
		},
	)

	server.RegisterHandler(
		"/v1/assistants",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				var request openai.AssistantRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.Assistant{
					ID:           assistantID,
					Object:       "assistant",
					CreatedAt:    1234567890,
					Name:         request.Name,
					Model:        request.Model,
					Description:  request.Description,
					Instructions: request.Instructions,
					Tools:        request.Tools,
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.AssistantsList{
					LastID:  &assistantID,
					FirstID: &assistantID,
					Assistants: []openai.Assistant{
						{
							ID:           assistantID,
							Object:       "assistant",
							CreatedAt:    1234567890,
							Name:         &assistantName,
							Model:        openai.GPT4TurboPreview,
							Description:  &assistantDescription,
							Instructions: &assistantInstructions,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	t.Run("create_assistant", func(t *testing.T) {
		_, err := client.CreateAssistant(ctx, openai.AssistantRequest{
			Name:         &assistantName,
			Description:  &assistantDescription,
			Model:        openai.GPT4TurboPreview,
			Instructions: &assistantInstructions,
		})
		checks.NoError(t, err, "CreateAssistant error")
	})

	t.Run("retrieve_assistant", func(t *testing.T) {
		_, err := client.RetrieveAssistant(ctx, assistantID)
		checks.NoError(t, err, "RetrieveAssistant error")
	})

	t.Run("delete_assistant", func(t *testing.T) {
		_, err := client.DeleteAssistant(ctx, assistantID)
		checks.NoError(t, err, "DeleteAssistant error")
	})

	t.Run("list_assistant", func(t *testing.T) {
		_, err := client.ListAssistants(ctx, &limit, &order, &after, &before)
		checks.NoError(t, err, "ListAssistants error")
	})

	t.Run("create_assistant_file", func(t *testing.T) {
		_, err := client.CreateAssistantFile(ctx, assistantID, openai.AssistantFileRequest{
			FileID: assistantFileID,
		})
		checks.NoError(t, err, "CreateAssistantFile error")
	})

	t.Run("list_assistant_files", func(t *testing.T) {
		_, err := client.ListAssistantFiles(ctx, assistantID, &limit, &order, &after, &before)
		checks.NoError(t, err, "ListAssistantFiles error")
	})

	t.Run("retrieve_assistant_file", func(t *testing.T) {
		_, err := client.RetrieveAssistantFile(ctx, assistantID, assistantFileID)
		checks.NoError(t, err, "RetrieveAssistantFile error")
	})

	t.Run("delete_assistant_file", func(t *testing.T) {
		err := client.DeleteAssistantFile(ctx, assistantID, assistantFileID)
		checks.NoError(t, err, "DeleteAssistantFile error")
	})

	t.Run("modify_assistant_no_tools", func(t *testing.T) {
		assistant, err := client.ModifyAssistant(ctx, assistantID, openai.AssistantRequest{
			Name:         &assistantName,
			Description:  &assistantDescription,
			Model:        openai.GPT4TurboPreview,
			Instructions: &assistantInstructions,
		})
		checks.NoError(t, err, "ModifyAssistant error")

		if assistant.Tools != nil {
			t.Errorf("expected nil got %v", assistant.Tools)
		}
	})

	t.Run("modify_assistant_with_tools", func(t *testing.T) {
		assistant, err := client.ModifyAssistant(ctx, assistantID, openai.AssistantRequest{
			Name:         &assistantName,
			Description:  &assistantDescription,
			Model:        openai.GPT4TurboPreview,
			Instructions: &assistantInstructions,
			Tools:        []openai.AssistantTool{{Type: openai.AssistantToolTypeFunction}},
		})
		checks.NoError(t, err, "ModifyAssistant error")

		if assistant.Tools == nil || len(assistant.Tools) != 1 {
			t.Errorf("expected a slice got %v", assistant.Tools)
		}
	})

	t.Run("modify_assistant_empty_tools", func(t *testing.T) {
		assistant, err := client.ModifyAssistant(ctx, assistantID, openai.AssistantRequest{
			Name:         &assistantName,
			Description:  &assistantDescription,
			Model:        openai.GPT4TurboPreview,
			Instructions: &assistantInstructions,
			Tools:        make([]openai.AssistantTool, 0),
		})

		checks.NoError(t, err, "ModifyAssistant error")

		if assistant.Tools == nil {
			t.Errorf("expected a slice got %v", assistant.Tools)
		}
	})
}

func TestAzureAssistant(t *testing.T) {
	assistantID := "asst_abc123"
	assistantName := "Ambrogio"
	assistantDescription := "Ambrogio is a friendly assistant."
	assistantInstructions := `You are a personal math tutor. 
When asked a question, write and run Python code to answer the question.`
	assistantFileID := "file-wB6RM6wHdA49HfS2DJ9fEyrH"
	limit := 20
	order := "desc"
	after := "asst_abc122"
	before := "asst_abc124"

	client, server, teardown := setupAzureTestServer()
	defer teardown()

	server.RegisterHandler(
		"/openai/assistants/"+assistantID+"/files/"+assistantFileID,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.AssistantFile{
					ID:          assistantFileID,
					Object:      "assistant.file",
					CreatedAt:   1234567890,
					AssistantID: assistantID,
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodDelete {
				fmt.Fprintln(w, `{
					id: "file-wB6RM6wHdA49HfS2DJ9fEyrH",
					object: "assistant.file.deleted",
					deleted: true
				  }`)
			}
		},
	)

	server.RegisterHandler(
		"/openai/assistants/"+assistantID+"/files",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.AssistantFilesList{
					AssistantFiles: []openai.AssistantFile{
						{
							ID:          assistantFileID,
							Object:      "assistant.file",
							CreatedAt:   1234567890,
							AssistantID: assistantID,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodPost {
				var request openai.AssistantFileRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.AssistantFile{
					ID:          request.FileID,
					Object:      "assistant.file",
					CreatedAt:   1234567890,
					AssistantID: assistantID,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/openai/assistants/"+assistantID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.Assistant{
					ID:           assistantID,
					Object:       "assistant",
					CreatedAt:    1234567890,
					Name:         &assistantName,
					Model:        openai.GPT4TurboPreview,
					Description:  &assistantDescription,
					Instructions: &assistantInstructions,
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodPost:
				var request openai.AssistantRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.Assistant{
					ID:           assistantID,
					Object:       "assistant",
					CreatedAt:    1234567890,
					Name:         request.Name,
					Model:        request.Model,
					Description:  request.Description,
					Instructions: request.Instructions,
					Tools:        request.Tools,
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodDelete:
				fmt.Fprintln(w, `{
					"id": "asst_abc123",
					"object": "assistant.deleted",
					"deleted": true
				  }`)
			}
		},
	)

	server.RegisterHandler(
		"/openai/assistants",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				var request openai.AssistantRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.Assistant{
					ID:           assistantID,
					Object:       "assistant",
					CreatedAt:    1234567890,
					Name:         request.Name,
					Model:        request.Model,
					Description:  request.Description,
					Instructions: request.Instructions,
					Tools:        request.Tools,
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.AssistantsList{
					LastID:  &assistantID,
					FirstID: &assistantID,
					Assistants: []openai.Assistant{
						{
							ID:           assistantID,
							Object:       "assistant",
							CreatedAt:    1234567890,
							Name:         &assistantName,
							Model:        openai.GPT4TurboPreview,
							Description:  &assistantDescription,
							Instructions: &assistantInstructions,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	_, err := client.CreateAssistant(ctx, openai.AssistantRequest{
		Name:         &assistantName,
		Description:  &assistantDescription,
		Model:        openai.GPT4TurboPreview,
		Instructions: &assistantInstructions,
	})
	checks.NoError(t, err, "CreateAssistant error")

	_, err = client.RetrieveAssistant(ctx, assistantID)
	checks.NoError(t, err, "RetrieveAssistant error")

	_, err = client.ModifyAssistant(ctx, assistantID, openai.AssistantRequest{
		Name:         &assistantName,
		Description:  &assistantDescription,
		Model:        openai.GPT4TurboPreview,
		Instructions: &assistantInstructions,
	})
	checks.NoError(t, err, "ModifyAssistant error")

	_, err = client.DeleteAssistant(ctx, assistantID)
	checks.NoError(t, err, "DeleteAssistant error")

	_, err = client.ListAssistants(ctx, &limit, &order, &after, &before)
	checks.NoError(t, err, "ListAssistants error")

	_, err = client.CreateAssistantFile(ctx, assistantID, openai.AssistantFileRequest{
		FileID: assistantFileID,
	})
	checks.NoError(t, err, "CreateAssistantFile error")

	_, err = client.ListAssistantFiles(ctx, assistantID, &limit, &order, &after, &before)
	checks.NoError(t, err, "ListAssistantFiles error")

	_, err = client.RetrieveAssistantFile(ctx, assistantID, assistantFileID)
	checks.NoError(t, err, "RetrieveAssistantFile error")

	err = client.DeleteAssistantFile(ctx, assistantID, assistantFileID)
	checks.NoError(t, err, "DeleteAssistantFile error")
}
