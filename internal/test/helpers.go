package test

import (
	"github.com/alejandrojnm/go-openai/internal/test/checks"

	"net/http"
	"os"
	"testing"
)

// CreateTestFile creates a fake file with "hello" as the content.
func CreateTestFile(t *testing.T, path string) {
	file, err := os.Create(path)
	checks.NoError(t, err, "failed to create file")

	if _, err = file.WriteString("hello"); err != nil {
		t.Fatalf("failed to write to file %v", err)
	}
	file.Close()
}

// CreateTestDirectory creates a temporary folder which will be deleted when cleanup is called.
func CreateTestDirectory(t *testing.T) (path string, cleanup func()) {
	t.Helper()

	path, err := os.MkdirTemp(os.TempDir(), "")
	checks.NoError(t, err)

	return path, func() { os.RemoveAll(path) }
}

// TokenRoundTripper is a struct that implements the RoundTripper
// interface, specifically to handle the authentication token by adding a token
// to the request header. We need this because the API requires that each
// request include a valid API token in the headers for authentication and
// authorization.
type TokenRoundTripper struct {
	Token    string
	Fallback http.RoundTripper
}

// RoundTrip takes an *http.Request as input and returns an
// *http.Response and an error.
//
// It is expected to use the provided request to create a connection to an HTTP
// server and return the response, or an error if one occurred. The returned
// Response should have its Body closed. If the RoundTrip method returns an
// error, the Client's Get, Head, Post, and PostForm methods return the same
// error.
func (t *TokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return t.Fallback.RoundTrip(req)
}
