package openai //nolint:testpackage // testing private field

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	utils "github.com/sashabaranov/go-openai/internal"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

var errTestUnmarshalerFailed = errors.New("test unmarshaler failed")

type failingUnMarshaller struct{}

func (*failingUnMarshaller) Unmarshal(_ []byte, _ any) error {
	return errTestUnmarshalerFailed
}

func TestStreamReaderReturnsUnmarshalerErrors(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		errAccumulator: utils.NewErrorAccumulator(),
		unmarshaler:    &failingUnMarshaller{},
	}

	respErr := stream.unmarshalError()
	if respErr != nil {
		t.Fatalf("Did not return nil with empty buffer: %v", respErr)
	}

	err := stream.errAccumulator.Write([]byte("{"))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	respErr = stream.unmarshalError()
	if respErr != nil {
		t.Fatalf("Did not return nil when unmarshaler failed: %v", respErr)
	}
}

func TestStreamReaderReturnsErrTooManyEmptyStreamMessages(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		emptyMessagesLimit: 3,
		reader:             bufio.NewReader(bytes.NewReader([]byte("\n\n\n\n"))),
		errAccumulator:     utils.NewErrorAccumulator(),
		unmarshaler:        &utils.JSONUnmarshaler{},
	}
	_, err := stream.Recv()
	checks.ErrorIs(t, err, ErrTooManyEmptyStreamMessages, "Did not return error when recv failed", err.Error())
}

func TestStreamReaderReturnsErrTestErrorAccumulatorWriteFailed(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader: bufio.NewReader(bytes.NewReader([]byte("\n"))),
		errAccumulator: &utils.DefaultErrorAccumulator{
			Buffer: &test.FailingErrorBuffer{},
		},
		unmarshaler: &utils.JSONUnmarshaler{},
	}
	_, err := stream.Recv()
	checks.ErrorIs(t, err, test.ErrTestErrorAccumulatorWriteFailed, "Did not return error when write failed", err.Error())
}

func TestStreamReaderRecvRaw(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		reader: bufio.NewReader(bytes.NewReader([]byte("data: {\"key\": \"value\"}\n"))),
	}
	rawLine, err := stream.RecvRaw()
	if err != nil {
		t.Fatalf("Did not return raw line: %v", err)
	}
	if !bytes.Equal(rawLine, []byte("{\"key\": \"value\"}")) {
		t.Fatalf("Did not return raw line: %v", string(rawLine))
	}

	streamDataPrefixWithoutSpace := &streamReader[ChatCompletionStreamResponse]{
		reader: bufio.NewReader(bytes.NewReader([]byte("data:{\"key\": \"value\"}\n"))),
	}
	rawLine1, err := streamDataPrefixWithoutSpace.RecvRaw()
	if err != nil {
		t.Fatalf("Did not return raw line: %v", err)
	}
	if !bytes.Equal(rawLine1, []byte("{\"key\": \"value\"}")) {
		t.Fatalf("Did not return raw line: %v", string(rawLine1))
	}
}

func TestStreamReaderParseLine(t *testing.T) {
	testUnits := []struct {
		rawLine []byte
		want    [2][]byte
	}{
		{[]byte("data: value"), [2][]byte{[]byte("data"), []byte("value")}},
		{[]byte("datavalue"), [2][]byte{[]byte("datavalue"), []byte("")}},
		{[]byte("data: "), [2][]byte{[]byte("data"), nil}},
		{[]byte("data:value"), [2][]byte{[]byte("data"), []byte("value")}},
		{[]byte(":"), [2][]byte{[]byte(""), []byte("")}},
		{[]byte(""), [2][]byte{[]byte(""), []byte("")}},
	}

	for _, testUnit := range testUnits {
		stream := &streamReader[ChatCompletionStreamResponse]{}
		name, value := stream.parseLine(testUnit.rawLine)
		if !bytes.Equal(name, testUnit.want[0]) || !bytes.Equal(value, testUnit.want[1]) {
			t.Errorf("parseLine(%q) = %q, %q; want %q, %q",
				testUnit.rawLine, name, value, testUnit.want[0], testUnit.want[1])
		}
	}
}

func TestStreamReaderHandleDataFlagWriteErrAccumulatorError(t *testing.T) {
	stream := &streamReader[ChatCompletionStreamResponse]{
		errAccumulator: &utils.DefaultErrorAccumulator{
			Buffer: &test.FailingErrorBuffer{},
		},
	}
	res, err := stream.handleDataFlag([]byte(`{"error`))
	if res != nil {
		t.Fatalf("Did not return nil when write failed: %v", res)
	}
	checks.ErrorIs(t, err, test.ErrTestErrorAccumulatorWriteFailed,
		"Did not return error when write failed", err.Error())
}
