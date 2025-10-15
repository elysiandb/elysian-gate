package forward

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

func ForwardRequest(method string, url string, payload string) (int, string, error) {
	var body io.Reader

	if payload != "" {
		body = bytes.NewBuffer([]byte(payload))
	}

	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("forward error: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(data), nil
}
