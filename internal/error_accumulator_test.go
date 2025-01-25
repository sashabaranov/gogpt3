package openai_test

import (
	"bytes"
	"errors"
	"testing"

	utils "github.com/alejandrojnm/go-openai/internal"
	"github.com/alejandrojnm/go-openai/internal/test"
)

func TestErrorAccumulatorBytes(t *testing.T) {
	accumulator := &utils.DefaultErrorAccumulator{
		Buffer: &bytes.Buffer{},
	}

	errBytes := accumulator.Bytes()
	if len(errBytes) != 0 {
		t.Fatalf("Did not return nil with empty bytes: %s", string(errBytes))
	}

	err := accumulator.Write([]byte("{}"))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	errBytes = accumulator.Bytes()
	if len(errBytes) == 0 {
		t.Fatalf("Did not return error bytes when has error: %s", string(errBytes))
	}
}

func TestErrorByteWriteErrors(t *testing.T) {
	accumulator := &utils.DefaultErrorAccumulator{
		Buffer: &test.FailingErrorBuffer{},
	}
	err := accumulator.Write([]byte("{"))
	if !errors.Is(err, test.ErrTestErrorAccumulatorWriteFailed) {
		t.Fatalf("Did not return error when write failed: %v", err)
	}
}
