package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fragdanceone/prisma-client-go/logger"
	"github.com/fragdanceone/prisma-client-go/runtime/types"
)

var internalUpdateNotFoundMessage = "Error occurred during query execution:\nInterpretationError(\"Error for binding '0'\", Some(QueryGraphBuilderError(RecordNotFound(\"Record to update not found.\"))))"
var internalDeleteNotFoundMessage = "Error occurred during query execution:\nInterpretationError(\"Error for binding '0'\", Some(QueryGraphBuilderError(RecordNotFound(\"Record to delete does not exist.\"))))"

// Do sends the http Request to the query engine and unmarshals the response
func (e *QueryEngine) Do(ctx context.Context, payload interface{}, v interface{}) error {
	startReq := time.Now()

	body, err := e.Request(ctx, "POST", "/", payload)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	logger.Debug.Printf("[timing] query engine request took %s", time.Since(startReq))

	startParse := time.Now()

	var response GQLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	if len(response.Errors) > 0 {
		first := response.Errors[0]
		if first.Message == internalUpdateNotFoundMessage ||
			first.Message == internalDeleteNotFoundMessage {
			return types.ErrNotFound
		}
		return fmt.Errorf("pql error: %s", first.Message)
	}

	if err := json.Unmarshal(response.Data.Result, v); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	logger.Debug.Printf("[timing] request unmarshaling took %s", time.Since(startParse))

	return nil
}

func (e *QueryEngine) DoJson(ctx context.Context, payload interface{}) (err error, j []byte) {
	startReq := time.Now()

	body, err := e.Request(ctx, "POST", "/", payload)
	if err != nil {
		return fmt.Errorf("request failed: %w", err), nil
	}

	logger.Debug.Printf("[timing] query engine request took %s", time.Since(startReq))

	startParse := time.Now()

	var response GQLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("json unmarshal: %w", err), nil
	}

	if len(response.Errors) > 0 {
		first := response.Errors[0]
		if first.Message == internalUpdateNotFoundMessage ||
			first.Message == internalDeleteNotFoundMessage {
			return types.ErrNotFound, nil
		}
		return fmt.Errorf("pql error: %s", first.Message), nil
	}
	/*
		if err := json.Unmarshal(response.Data.Result, v); err != nil {
			return fmt.Errorf("json unmarshal: %w", err),nil
		}
	*/
	logger.Debug.Printf("[timing] request unmarshaling took %s", time.Since(startParse))

	return nil, response.Data.Result
}

// Do sends the http Request to the query engine and unmarshals the response
func (e *QueryEngine) Batch(ctx context.Context, payload interface{}, v interface{}) error {
	body, err := e.Request(ctx, "POST", "/", payload)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if err := json.Unmarshal(body, &v); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	return nil
}

func (e *QueryEngine) Request(ctx context.Context, method string, path string, payload interface{}) ([]byte, error) {
	if e.disconnected {
		logger.Info.Printf("A query was executed after Disconnect() was called. Make sure to not send any queries after disconnecting the client.")
		return nil, fmt.Errorf("client is disconnected")
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("payload marshal: %w", err)
	}
	// TODO use specific log level
	if logger.Enabled {
		logger.Debug.Printf("prisma engine payload: `%s`", requestBody)
	}
	return request(ctx, e.http, method, e.url+path, requestBody, func(req *http.Request) {
		req.Header.Set("content-type", "application/json")
	})
}

func (e *QueryEngine) GraphQL(ctx context.Context, payload string) ([]byte, error) {
	fmt.Println("Graphql")
	return request(ctx, e.http, "POST", e.url+"/", []byte(payload), func(req *http.Request) {
		req.Header.Set("content-type", "application/json")
	})
}
