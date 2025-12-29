// package elasticsearch provides elasticsearch client management
package elasticsearch

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jessewkun/gocommon/logger"
)

const maxLogBodySize = 1024 // For logging, only read up to 1KB of the body

// readCloser is a helper struct to combine a Reader and a Closer.
type readCloser struct {
	io.Reader
	io.Closer
}

type loggingTransport struct {
	transport     http.RoundTripper
	slowThreshold time.Duration
}

func newLoggingTransport(slowThreshold time.Duration) *loggingTransport {
	return &loggingTransport{
		transport:     http.DefaultTransport,
		slowThreshold: slowThreshold,
	}
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()
	ctx := req.Context()

	// Safely read request body for logging.
	// This reads the entire body into memory, which can be an issue for very large
	// requests like bulk indexing.
	var reqBodyBytes []byte
	if req.Body != nil {
		var err error
		reqBodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			logger.Warn(ctx, TAG, "failed to read request body for logging: %v", err)
		}
		// Restore req.Body for the transport to read.
		req.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
	}

	resp, err := t.transport.RoundTrip(req)

	duration := time.Since(startTime)
	fields := map[string]interface{}{
		"method":   req.Method,
		"url":      req.URL.String(),
		"duration": duration,
		"req_body": string(reqBodyBytes),
	}

	if err != nil {
		fields["error"] = err.Error()
		fields["status_code"] = -1
		logger.ErrorWithField(ctx, TAG, "ES_REQUEST_ERROR", fields)
		return nil, err
	}

	fields["status_code"] = resp.StatusCode

	if resp.StatusCode >= 400 { // IsError()
		var respBodyBytes []byte
		if resp.Body != nil {
			// Read a limited chunk of the response body for logging.
			lr := io.LimitReader(resp.Body, maxLogBodySize)
			var readErr error
			respBodyBytes, readErr = io.ReadAll(lr)
			if readErr != nil {
				logger.Error(ctx, TAG, fmt.Errorf("failed to read limited response body for logging: %w", readErr))
			}

			// IMPORTANT: Restore the response body so the caller can read it from the beginning.
			// We combine the part we read with the rest of the original stream.
			originalBodyCloser := resp.Body
			resp.Body = &readCloser{
				Reader: io.MultiReader(bytes.NewReader(respBodyBytes), originalBodyCloser),
				Closer: originalBodyCloser,
			}
		}
		fields["resp_body"] = string(respBodyBytes)
		logger.ErrorWithField(ctx, TAG, "ES_REQUEST_FAILED", fields)
	} else if duration > t.slowThreshold {
		logger.WarnWithField(ctx, TAG, "ES_SLOW_QUERY", fields)
	} else {
		// Only log Info for successful, non-slow queries to avoid excessive logging.
		logger.InfoWithField(ctx, TAG, "ES_QUERY", fields)
	}

	return resp, nil
}
