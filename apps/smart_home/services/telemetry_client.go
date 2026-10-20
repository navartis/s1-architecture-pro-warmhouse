package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type TelemetryClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type telemetryPoint struct {
	DeviceID string  `json:"deviceId"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
}

func NewTelemetryClient(baseURL string) *TelemetryClient {
	return &TelemetryClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (t *TelemetryClient) Send(ctx context.Context, deviceID string, value float64, unit string) error {
	body, err := json.Marshal(telemetryPoint{DeviceID: deviceID, Value: value, Unit: unit})
	if err != nil {
		return fmt.Errorf("encode telemetry: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.BaseURL+"/telemetry", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build telemetry request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send telemetry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status from telemetry service: %d", resp.StatusCode)
	}

	return nil
}
