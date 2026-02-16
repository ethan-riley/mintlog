package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var retryDelays = []time.Duration{1 * time.Second, 5 * time.Second, 30 * time.Second}

type WebhookSender struct {
	client *http.Client
}

func NewWebhookSender() *WebhookSender {
	return &WebhookSender{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (ws *WebhookSender) Send(cfg WebhookConfig, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= len(retryDelays); attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelays[attempt-1])
		}

		req, err := http.NewRequest(http.MethodPost, cfg.URL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "mintlog-notifier/1.0")

		for k, v := range cfg.Headers {
			req.Header.Set(k, v)
		}

		if cfg.Secret != "" {
			sig := computeHMAC(body, []byte(cfg.Secret))
			req.Header.Set("X-Mintlog-Signature", "sha256="+sig)
		}

		resp, err := ws.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			slog.Warn("webhook delivery failed, retrying", "attempt", attempt+1, "error", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		slog.Warn("webhook non-2xx response, retrying", "attempt", attempt+1, "status", resp.StatusCode)
	}

	return lastErr
}

func computeHMAC(data, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
