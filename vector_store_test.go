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

// TestVectorStore Tests the vector store endpoint of the API using the mocked server.
func TestVectorStore(t *testing.T) {
	vectorStoreID := "vs_abc123"
	vectorStoreName := "TestStore"
	vectorStoreFileID := "file-wB6RM6wHdA49HfS2DJ9fEyrH"
	vectorStoreFileBatchID := "vsfb_abc123"
	limit := 20
	order := "desc"
	after := "vs_abc122"
	before := "vs_abc123"

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/vector_stores/"+vectorStoreID+"/files/"+vectorStoreFileID,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.VectorStoreFile{
					ID:            vectorStoreFileID,
					Object:        "vector_store.file",
					CreatedAt:     1234567890,
					VectorStoreID: vectorStoreID,
					Status:        "completed",
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodDelete {
				fmt.Fprintln(w, `{
					id: "file-wB6RM6wHdA49HfS2DJ9fEyrH",
					object: "vector_store.file.deleted",
					deleted: true
				  }`)
			}
		},
	)

	server.RegisterHandler(
		"/v1/vector_stores/"+vectorStoreID+"/files",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.VectorStoreFilesList{
					VectorStoreFiles: []openai.VectorStoreFile{
						{
							ID:            vectorStoreFileID,
							Object:        "vector_store.file",
							CreatedAt:     1234567890,
							VectorStoreID: vectorStoreID,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodPost {
				var request openai.VectorStoreFileRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.VectorStoreFile{
					ID:            request.FileID,
					Object:        "vector_store.file",
					CreatedAt:     1234567890,
					VectorStoreID: vectorStoreID,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/vector_stores/"+vectorStoreID+"/file_batches/"+vectorStoreFileBatchID+"/files",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.VectorStoreFilesList{
					VectorStoreFiles: []openai.VectorStoreFile{
						{
							ID:            vectorStoreFileID,
							Object:        "vector_store.file",
							CreatedAt:     1234567890,
							VectorStoreID: vectorStoreID,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/vector_stores/"+vectorStoreID+"/file_batches/"+vectorStoreFileBatchID+"/cancel",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				resBytes, _ := json.Marshal(openai.VectorStoreFileBatch{
					ID:            vectorStoreFileBatchID,
					Object:        "vector_store.file_batch",
					CreatedAt:     1234567890,
					VectorStoreID: vectorStoreID,
					Status:        "cancelling",
					FileCounts: openai.VectorStoreFileCount{
						InProgress: 0,
						Completed:  1,
						Failed:     0,
						Cancelled:  0,
						Total:      0,
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/vector_stores/"+vectorStoreID+"/file_batches/"+vectorStoreFileBatchID,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.VectorStoreFileBatch{
					ID:            vectorStoreFileBatchID,
					Object:        "vector_store.file_batch",
					CreatedAt:     1234567890,
					VectorStoreID: vectorStoreID,
					Status:        "completed",
					FileCounts: openai.VectorStoreFileCount{
						Completed: 1,
					},
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodPost {
				resBytes, _ := json.Marshal(openai.VectorStoreFileBatch{
					ID:            vectorStoreFileBatchID,
					Object:        "vector_store.file_batch",
					CreatedAt:     1234567890,
					VectorStoreID: vectorStoreID,
					Status:        "cancelling",
					FileCounts: openai.VectorStoreFileCount{
						Completed: 1,
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/vector_stores/"+vectorStoreID+"/file_batches",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				var request openai.VectorStoreFileBatchRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.VectorStoreFileBatch{
					ID:            vectorStoreFileBatchID,
					Object:        "vector_store.file_batch",
					CreatedAt:     1234567890,
					VectorStoreID: vectorStoreID,
					Status:        "completed",
					FileCounts: openai.VectorStoreFileCount{
						InProgress: 0,
						Completed:  len(request.FileIDs),
						Failed:     0,
						Cancelled:  0,
						Total:      0,
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/vector_stores/"+vectorStoreID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.VectorStore{
					ID:        vectorStoreID,
					Object:    "vector_store",
					CreatedAt: 1234567890,
					Name:      vectorStoreName,
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodPost:
				var request openai.VectorStore
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.VectorStore{
					ID:        vectorStoreID,
					Object:    "vector_store",
					CreatedAt: 1234567890,
					Name:      request.Name,
				})
				fmt.Fprintln(w, string(resBytes))
			case http.MethodDelete:
				fmt.Fprintln(w, `{
					"id": "vectorstore_abc123",
					"object": "vector_store.deleted",
					"deleted": true
				  }`)
			}
		},
	)

	server.RegisterHandler(
		"/v1/vector_stores",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				var request openai.VectorStoreRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.VectorStore{
					ID:        vectorStoreID,
					Object:    "vector_store",
					CreatedAt: 1234567890,
					Name:      request.Name,
					FileCounts: openai.VectorStoreFileCount{
						InProgress: 0,
						Completed:  0,
						Failed:     0,
						Cancelled:  0,
						Total:      0,
					},
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.VectorStoresList{
					LastID:  &vectorStoreID,
					FirstID: &vectorStoreID,
					VectorStores: []openai.VectorStore{
						{
							ID:        vectorStoreID,
							Object:    "vector_store",
							CreatedAt: 1234567890,
							Name:      vectorStoreName,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	t.Run("create_vector_store", func(t *testing.T) {
		_, err := client.CreateVectorStore(ctx, openai.VectorStoreRequest{
			Name: vectorStoreName,
		})
		checks.NoError(t, err, "CreateVectorStore error")
	})

	t.Run("retrieve_vector_store", func(t *testing.T) {
		_, err := client.RetrieveVectorStore(ctx, vectorStoreID)
		checks.NoError(t, err, "RetrieveVectorStore error")
	})

	t.Run("delete_vector_store", func(t *testing.T) {
		_, err := client.DeleteVectorStore(ctx, vectorStoreID)
		checks.NoError(t, err, "DeleteVectorStore error")
	})

	t.Run("list_vector_store", func(t *testing.T) {
		_, err := client.ListVectorStores(context.TODO(), openai.Pagination{
			Limit:  &limit,
			Order:  &order,
			After:  &after,
			Before: &before,
		})
		checks.NoError(t, err, "ListVectorStores error")
	})

	t.Run("create_vector_store_file", func(t *testing.T) {
		_, err := client.CreateVectorStoreFile(context.TODO(), vectorStoreID, openai.VectorStoreFileRequest{
			FileID: vectorStoreFileID,
		})
		checks.NoError(t, err, "CreateVectorStoreFile error")
	})

	t.Run("list_vector_store_files", func(t *testing.T) {
		_, err := client.ListVectorStoreFiles(ctx, vectorStoreID, openai.Pagination{
			Limit:  &limit,
			Order:  &order,
			After:  &after,
			Before: &before,
		})
		checks.NoError(t, err, "ListVectorStoreFiles error")
	})

	t.Run("retrieve_vector_store_file", func(t *testing.T) {
		_, err := client.RetrieveVectorStoreFile(ctx, vectorStoreID, vectorStoreFileID)
		checks.NoError(t, err, "RetrieveVectorStoreFile error")
	})

	t.Run("delete_vector_store_file", func(t *testing.T) {
		err := client.DeleteVectorStoreFile(ctx, vectorStoreID, vectorStoreFileID)
		checks.NoError(t, err, "DeleteVectorStoreFile error")
	})

	t.Run("modify_vector_store", func(t *testing.T) {
		_, err := client.ModifyVectorStore(ctx, vectorStoreID, openai.VectorStoreRequest{
			Name: vectorStoreName,
		})
		checks.NoError(t, err, "ModifyVectorStore error")
	})

	t.Run("create_vector_store_file_batch", func(t *testing.T) {
		_, err := client.CreateVectorStoreFileBatch(ctx, vectorStoreID, openai.VectorStoreFileBatchRequest{
			FileIDs: []string{vectorStoreFileID},
		})
		checks.NoError(t, err, "CreateVectorStoreFileBatch error")
	})

	t.Run("retrieve_vector_store_file_batch", func(t *testing.T) {
		_, err := client.RetrieveVectorStoreFileBatch(ctx, vectorStoreID, vectorStoreFileBatchID)
		checks.NoError(t, err, "RetrieveVectorStoreFileBatch error")
	})

	t.Run("list_vector_store_files_in_batch", func(t *testing.T) {
		_, err := client.ListVectorStoreFilesInBatch(
			ctx,
			vectorStoreID,
			vectorStoreFileBatchID,
			openai.Pagination{
				Limit:  &limit,
				Order:  &order,
				After:  &after,
				Before: &before,
			})
		checks.NoError(t, err, "ListVectorStoreFilesInBatch error")
	})

	t.Run("cancel_vector_store_file_batch", func(t *testing.T) {
		_, err := client.CancelVectorStoreFileBatch(ctx, vectorStoreID, vectorStoreFileBatchID)
		checks.NoError(t, err, "CancelVectorStoreFileBatch error")
	})
}
