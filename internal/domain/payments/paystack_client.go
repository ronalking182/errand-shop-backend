package payments

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Paystack API response structures are defined in dto.go

type PaystackWebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		ID              int64  `json:"id"`
		Domain          string `json:"domain"`
		Status          string `json:"status"`
		Reference       string `json:"reference"`
		Amount          int64  `json:"amount"`
		Message         string `json:"message"`
		GatewayResponse string `json:"gateway_response"`
		PaidAt          string `json:"paid_at"`
		CreatedAt       string `json:"created_at"`
		Channel         string `json:"channel"`
		Currency        string `json:"currency"`
		IPAddress       string `json:"ip_address"`
		Metadata        interface{} `json:"metadata"`
		Customer        struct {
			ID           int64  `json:"id"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
			Phone        string `json:"phone"`
			Metadata     interface{} `json:"metadata"`
			RiskAction   string `json:"risk_action"`
		} `json:"customer"`
	} `json:"data"`
}

// PaystackClient handles Paystack API interactions
type PaystackClient struct {
	secretKey     string
	webhookSecret string
	baseURL       string
	appBaseURL    string
	callbackURL   string
	client        *http.Client
}

// NewPaystackClient creates a new Paystack client
func NewPaystackClient(secretKey, webhookSecret, appBaseURL, callbackURL string) *PaystackClient {
	return &PaystackClient{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		baseURL:       "https://api.paystack.co",
		appBaseURL:    appBaseURL,
		callbackURL:   callbackURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InitializeTransaction initializes a payment transaction
func (p *PaystackClient) InitializeTransaction(email string, amount int64, reference string, metadata map[string]interface{}) (*PaystackInitializeResponse, error) {
	payload := map[string]interface{}{
		"email":        email,
		"amount":       amount,
		"reference":    reference,
		"callback_url": p.callbackURL,
	}

	if metadata != nil {
		payload["metadata"] = metadata
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.baseURL+"/transaction/initialize", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response PaystackInitializeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Status {
		return nil, fmt.Errorf("paystack error: %s", response.Message)
	}

	return &response, nil
}

// VerifyTransaction verifies a payment transaction
func (p *PaystackClient) VerifyTransaction(reference string) (*PaystackVerifyResponse, error) {
	req, err := http.NewRequest("GET", p.baseURL+"/transaction/verify/"+reference, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.secretKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response PaystackVerifyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Status {
		return nil, fmt.Errorf("paystack error: %s", response.Message)
	}

	return &response, nil
}

// ValidateWebhookSignature validates the Paystack webhook signature
func (p *PaystackClient) ValidateWebhookSignature(payload []byte, signature string) bool {
	h := hmac.New(sha512.New, []byte(p.webhookSecret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ParseWebhookEvent parses a Paystack webhook event
func (p *PaystackClient) ParseWebhookEvent(payload []byte) (*PaystackWebhookEvent, error) {
	var event PaystackWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook event: %w", err)
	}
	return &event, nil
}