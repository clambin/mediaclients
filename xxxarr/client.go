package xxxarr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func call[T any](ctx context.Context, client *http.Client, target string) (T, error) {
	var response T
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return response, fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("unexpected http status: %s", resp.Status)
	}

	var body bytes.Buffer
	if err = json.NewDecoder(io.TeeReader(resp.Body, &body)).Decode(&response); err != nil {
		err = &ErrInvalidJSON{
			Err:  err,
			Body: body.Bytes(),
		}
	}
	return response, err
}
