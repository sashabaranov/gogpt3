package openai //nolint:testpackage // testing private field

import (
	"github.com/alejandrojnm/go-openai/internal/test"
	"github.com/alejandrojnm/go-openai/internal/test/checks"

	"bytes"
	"errors"
	"os"
	"testing"
)

type failingWriter struct {
}

var errMockFailingWriterError = errors.New("mock writer failed")

func (*failingWriter) Write([]byte) (int, error) {
	return 0, errMockFailingWriterError
}

func TestFormBuilderWithFailingWriter(t *testing.T) {
	dir, cleanup := test.CreateTestDirectory(t)
	defer cleanup()

	file, err := os.CreateTemp(dir, "")
	if err != nil {
		t.Errorf("Error creating tmp file: %v", err)
	}
	defer file.Close()
	defer os.Remove(file.Name())

	builder := NewFormBuilder(&failingWriter{})
	err = builder.CreateFormFile("file", file)
	checks.ErrorIs(t, err, errMockFailingWriterError, "formbuilder should return error if writer fails")
}

func TestFormBuilderWithClosedFile(t *testing.T) {
	dir, cleanup := test.CreateTestDirectory(t)
	defer cleanup()

	file, err := os.CreateTemp(dir, "")
	if err != nil {
		t.Errorf("Error creating tmp file: %v", err)
	}
	file.Close()
	defer os.Remove(file.Name())

	body := &bytes.Buffer{}
	builder := NewFormBuilder(body)
	err = builder.CreateFormFile("file", file)
	checks.HasError(t, err, "formbuilder should return error if file is closed")
	checks.ErrorIs(t, err, os.ErrClosed, "formbuilder should return error if file is closed")
}
