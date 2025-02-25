package openai

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"

	utils "github.com/sashabaranov/go-openai/internal"
)

var (
	dataFlag        = "data"
	doneFlag        = []byte("[DONE]")
	errorFlag       = []byte(`"error"`)
	errorPrefixFlag = []byte(`{"error`)
)

type streamable interface {
	ChatCompletionStreamResponse | CompletionResponse
}

type streamReader[T streamable] struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator utils.ErrorAccumulator
	unmarshaler    utils.Unmarshaler

	httpHeader
}

func (stream *streamReader[T]) Recv() (response T, err error) {
	rawLine, err := stream.RecvRaw()
	if err != nil {
		return
	}

	err = stream.unmarshaler.Unmarshal(rawLine, &response)
	if err != nil {
		return
	}
	return response, nil
}

func (stream *streamReader[T]) RecvRaw() ([]byte, error) {
	if stream.isFinished {
		return nil, io.EOF
	}

	return stream.processLines()
}

func (stream *streamReader[T]) processLines() ([]byte, error) {
	var (
		emptyMessagesCount uint
		hasErrorPrefix     bool
	)

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')
		if readErr != nil || (hasErrorPrefix && readErr == io.EOF) {
			return stream.handleReadError()
		}

		name, value := stream.parseLine(rawLine)

		switch string(name) {
		case dataFlag:
			return stream.handleDataFlag(value)
		default:
			if err := stream.handleDefaultCase(rawLine, name, &emptyMessagesCount, &hasErrorPrefix); err != nil {
				return nil, err
			}
		}
	}
}

func (stream *streamReader[T]) handleReadError() ([]byte, error) {
	respErr := stream.unmarshalError()
	if respErr != nil {
		return nil, fmt.Errorf("error, %w", respErr.Error)
	}
	return nil, io.EOF
}

func (stream *streamReader[T]) parseLine(rawLine []byte) ([]byte, []byte) {
	name, value, _ := bytes.Cut(rawLine, []byte(":"))
	value = bytes.TrimSpace(value)
	return name, value
}

func (stream *streamReader[T]) handleDataFlag(value []byte) ([]byte, error) {
	if bytes.Equal(value, doneFlag) {
		stream.isFinished = true
		return nil, io.EOF
	}
	if bytes.HasPrefix(value, errorPrefixFlag) {
		if err := stream.writeErrAccumulator(value); err != nil {
			return nil, err
		}
		respErr := stream.unmarshalError()
		if respErr != nil {
			return nil, fmt.Errorf("error, %w", respErr.Error)
		}
	}
	return value, nil
}

func (stream *streamReader[T]) handleDefaultCase(
	rawLine, name []byte,
	emptyMessagesCount *uint,
	hasErrorPrefix *bool,
) error {
	if err := stream.writeErrAccumulator(rawLine); err != nil {
		return err
	}
	if bytes.Equal(name, errorFlag) {
		*hasErrorPrefix = true
		return nil
	}

	*emptyMessagesCount++
	if *emptyMessagesCount > stream.emptyMessagesLimit {
		return ErrTooManyEmptyStreamMessages
	}
	return nil
}

func (stream *streamReader[T]) writeErrAccumulator(p []byte) error {
	writeErr := stream.errAccumulator.Write(p)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

func (stream *streamReader[T]) unmarshalError() (errResp *ErrorResponse) {
	errBytes := stream.errAccumulator.Bytes()
	if len(errBytes) == 0 {
		return
	}

	err := stream.unmarshaler.Unmarshal(errBytes, &errResp)
	if err != nil {
		errResp = nil
	}

	return
}

func (stream *streamReader[T]) Close() error {
	return stream.response.Body.Close()
}
