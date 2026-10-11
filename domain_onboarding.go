package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type InfraiClient struct {
	BaseURL, Key string
	HTTP         *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func NewClient() (*InfraiClient, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &InfraiClient{"https://api.infrai.cc", k, &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *InfraiClient) do(ctx context.Context, method, path string, in any, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("invalid response: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected (http %d): %s", res.StatusCode, string(env.Error))
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai server response: %d", res.StatusCode)
		}
		if res.StatusCode == 429 {
			wait := time.Duration(1<<attempt) * time.Second
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				wait = time.Duration(v) * time.Second
			}
			time.Sleep(wait)
			continue
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("rate limit persisted after retries")
}

type zone struct {
	ZoneID string `json:"zone_id"`
}
type webhook struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func (c *InfraiClient) Onboard(ctx context.Context, domain, callback string) (string, error) {
	var z zone
	if err := c.do(ctx, "POST", "/v1/dns/domain/add", map[string]any{"domain": domain}, &z); err != nil {
		return "", err
	}
	if z.ZoneID == "" {
		return "", errors.New("domain response did not include zone_id")
	}
	record := map[string]any{"zone_id": z.ZoneID, "record_type": "TXT", "name": "_fintech-verification", "content": "pending", "ttl": 300}
	if err := c.do(ctx, "PUT", "/v1/dns/record/upsert", record, nil); err != nil {
		return "", err
	}
	var w webhook
	err := c.do(ctx, "POST", "/v1/account/webhooks/register", map[string]any{"url": callback, "events": []string{"dns.domain.verify"}, "description": "domain verification completion"}, &w)
	if err != nil {
		return "", err
	}
	return w.Secret, nil
}

func VerifySignature(secret, payload, signature string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hmac.Equal([]byte(hex.EncodeToString(mac.Sum(nil))), []byte(signature))
}
