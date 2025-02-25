package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

var emptyStr = ""

func setupServerForTestMessage(t *testing.T, server *test.ServerTest) {
	threadID := "thread_abc123"
	messageID := "msg_abc123"
	fileID := "file_abc123"

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/messages/"+messageID+"/files/"+fileID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(
					openai.MessageFile{
						ID:        fileID,
						Object:    "thread.message.file",
						CreatedAt: 1699061776,
						MessageID: messageID,
					})
				fmt.Fprintln(w, string(resBytes))
			default:
				t.Fatalf("unsupported messages http method: %s", r.Method)
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/messages/"+messageID+"/files",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(
					openai.MessageFilesList{MessageFiles: []openai.MessageFile{{
						ID:        fileID,
						Object:    "thread.message.file",
						CreatedAt: 0,
						MessageID: messageID,
					}}})
				fmt.Fprintln(w, string(resBytes))
			default:
				t.Fatalf("unsupported messages http method: %s", r.Method)
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/messages/"+messageID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				metadata := map[string]any{}
				err := json.NewDecoder(r.Body).Decode(&metadata)
				checks.NoError(t, err, "unable to decode metadata in modify message call")
				payload, ok := metadata["metadata"].(map[string]any)
				if !ok {
					t.Fatalf("metadata payload improperly wrapped %+v", metadata)
				}

				resBytes, _ := json.Marshal(
					openai.Message{
						ID:        messageID,
						Object:    "thread.message",
						CreatedAt: 1234567890,
						ThreadID:  threadID,
						Role:      "user",
						Content: []openai.MessageContent{{
							Type: "text",
							Text: &openai.MessageText{
								Value:       "How does AI work?",
								Annotations: nil,
							},
						}},
						FileIds:     nil,
						AssistantID: &emptyStr,
						RunID:       &emptyStr,
						Metadata:    payload,
					})

				fmt.Fprintln(w, string(resBytes))
			case http.MethodGet:
				resBytes, _ := json.Marshal(
					openai.Message{
						ID:        messageID,
						Object:    "thread.message",
						CreatedAt: 1234567890,
						ThreadID:  threadID,
						Role:      "user",
						Content: []openai.MessageContent{{
							Type: "text",
							Text: &openai.MessageText{
								Value:       "How does AI work?",
								Annotations: nil,
							},
						}},
						FileIds:     nil,
						AssistantID: &emptyStr,
						RunID:       &emptyStr,
						Metadata:    nil,
					})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodDelete:
				resBytes, _ := json.Marshal(openai.MessageDeletionStatus{
					ID:      messageID,
					Object:  "thread.message.deleted",
					Deleted: true,
				})
				fmt.Fprintln(w, string(resBytes))
			default:
				t.Fatalf("unsupported messages http method: %s", r.Method)
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/messages",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.Message{
					ID:        messageID,
					Object:    "thread.message",
					CreatedAt: 1234567890,
					ThreadID:  threadID,
					Role:      "user",
					Content: []openai.MessageContent{{
						Type: "text",
						Text: &openai.MessageText{
							Value:       "How does AI work?",
							Annotations: nil,
						},
					}},
					FileIds:     nil,
					AssistantID: &emptyStr,
					RunID:       &emptyStr,
					Metadata:    nil,
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.MessagesList{
					Object: "list",
					Messages: []openai.Message{{
						ID:        messageID,
						Object:    "thread.message",
						CreatedAt: 1234567890,
						ThreadID:  threadID,
						Role:      "user",
						Content: []openai.MessageContent{{
							Type: "text",
							Text: &openai.MessageText{
								Value:       "How does AI work?",
								Annotations: nil,
							},
						}},
						FileIds:     nil,
						AssistantID: &emptyStr,
						RunID:       &emptyStr,
						Metadata:    nil,
					}},
					FirstID: &messageID,
					LastID:  &messageID,
					HasMore: false,
				})
				fmt.Fprintln(w, string(resBytes))
			default:
				t.Fatalf("unsupported messages http method: %s", r.Method)
			}
		},
	)
}

// TestMessages Tests the messages endpoint of the API using the mocked server.
func TestMessages(t *testing.T) {
	threadID := "thread_abc123"
	messageID := "msg_abc123"
	fileID := "file_abc123"

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	setupServerForTestMessage(t, server)
	ctx := context.Background()

	// static assertion of return type
	var msg openai.Message
	msg, err := client.CreateMessage(ctx, threadID, openai.MessageRequest{
		Role:     "user",
		Content:  "How does AI work?",
		FileIds:  nil,
		Metadata: nil,
	})
	checks.NoError(t, err, "CreateMessage error")
	if msg.ID != messageID {
		t.Fatalf("unexpected message id: '%s'", msg.ID)
	}

	var msgs openai.MessagesList
	msgs, err = client.ListMessage(ctx, threadID, nil, nil, nil, nil, nil)
	checks.NoError(t, err, "ListMessages error")
	if len(msgs.Messages) != 1 {
		t.Fatalf("unexpected length of fetched messages")
	}

	// with pagination options set
	limit := 1
	order := "desc"
	after := "obj_foo"
	before := "obj_bar"
	runID := "run_abc123"
	msgs, err = client.ListMessage(ctx, threadID, &limit, &order, &after, &before, &runID)
	checks.NoError(t, err, "ListMessages error")
	if len(msgs.Messages) != 1 {
		t.Fatalf("unexpected length of fetched messages")
	}

	msg, err = client.RetrieveMessage(ctx, threadID, messageID)
	checks.NoError(t, err, "RetrieveMessage error")
	if msg.ID != messageID {
		t.Fatalf("unexpected message id: '%s'", msg.ID)
	}

	msg, err = client.ModifyMessage(ctx, threadID, messageID,
		map[string]any{
			"foo": "bar",
		})
	checks.NoError(t, err, "ModifyMessage error")
	if msg.Metadata["foo"] != "bar" {
		t.Fatalf("expected message metadata to get modified")
	}

	msgDel, err := client.DeleteMessage(ctx, threadID, messageID)
	checks.NoError(t, err, "DeleteMessage error")
	if msgDel.ID != messageID {
		t.Fatalf("unexpected message id: '%s'", msg.ID)
	}
	if !msgDel.Deleted {
		t.Fatalf("expected deleted is true")
	}
	_, err = client.DeleteMessage(ctx, threadID, "not_exist_id")
	checks.HasError(t, err, "DeleteMessage error")

	// message files
	var msgFile openai.MessageFile
	msgFile, err = client.RetrieveMessageFile(ctx, threadID, messageID, fileID)
	checks.NoError(t, err, "RetrieveMessageFile error")
	if msgFile.ID != fileID {
		t.Fatalf("unexpected message file id: '%s'", msgFile.ID)
	}

	var msgFiles openai.MessageFilesList
	msgFiles, err = client.ListMessageFiles(ctx, threadID, messageID)
	checks.NoError(t, err, "RetrieveMessageFile error")
	if len(msgFiles.MessageFiles) != 1 {
		t.Fatalf("unexpected count of message files: %d", len(msgFiles.MessageFiles))
	}
	if msgFiles.MessageFiles[0].ID != fileID {
		t.Fatalf("unexpected message file id: '%s' in list message files", msgFiles.MessageFiles[0].ID)
	}
}
