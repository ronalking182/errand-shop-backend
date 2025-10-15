package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMService handles Firebase Cloud Messaging operations
type FCMService struct {
	client *messaging.Client
}

// FCMMessage represents a push notification message
type FCMMessage struct {
	Token    string            `json:"token"`
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data,omitempty"`
	ImageURL string            `json:"image_url,omitempty"`
}

// NewFCMService creates a new FCM service instance
func NewFCMService() (*FCMService, error) {
	// Get Firebase credentials from environment variable
	credentialsPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credentialsPath == "" {
		// Try to get credentials from JSON string
		credentialsJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON")
		if credentialsJSON == "" {
			return nil, fmt.Errorf("Firebase credentials not found. Set FIREBASE_CREDENTIALS_PATH or FIREBASE_CREDENTIALS_JSON environment variable")
		}

		// Parse JSON credentials
		var creds map[string]interface{}
		if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
			return nil, fmt.Errorf("invalid Firebase credentials JSON: %v", err)
		}

		// Initialize Firebase app with JSON credentials
		opt := option.WithCredentialsJSON([]byte(credentialsJSON))
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			return nil, fmt.Errorf("error initializing Firebase app: %v", err)
		}

		// Get messaging client
		client, err := app.Messaging(context.Background())
		if err != nil {
			return nil, fmt.Errorf("error getting messaging client: %v", err)
		}

		return &FCMService{client: client}, nil
	}

	// Initialize Firebase app with credentials file
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	// Get messaging client
	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %v", err)
	}

	return &FCMService{client: client}, nil
}

// SendMessage sends a single push notification
func (f *FCMService) SendMessage(ctx context.Context, msg *FCMMessage) (string, error) {
	if f.client == nil {
		return "", fmt.Errorf("FCM client not initialized")
	}

	// Build the message
	message := &messaging.Message{
		Token: msg.Token,
		Notification: &messaging.Notification{
			Title: msg.Title,
			Body:  msg.Body,
		},
		Data: msg.Data,
	}

	// Add image if provided
	if msg.ImageURL != "" {
		message.Notification.ImageURL = msg.ImageURL
	}

	// Send the message
	response, err := f.client.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending FCM message: %v", err)
		return "", fmt.Errorf("failed to send FCM message: %v", err)
	}

	log.Printf("Successfully sent FCM message: %s", response)
	return response, nil
}

// SendMulticast sends push notifications to multiple tokens
func (f *FCMService) SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string, imageURL string) (*messaging.BatchResponse, error) {
	if f.client == nil {
		return nil, fmt.Errorf("FCM client not initialized")
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens provided")
	}

	// Build the multicast message
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	// Add image if provided
	if imageURL != "" {
		message.Notification.ImageURL = imageURL
	}

	// Send the multicast message
	response, err := f.client.SendMulticast(ctx, message)
	if err != nil {
		log.Printf("Error sending FCM multicast: %v", err)
		return nil, fmt.Errorf("failed to send FCM multicast: %v", err)
	}

	log.Printf("Successfully sent FCM multicast. Success: %d, Failure: %d", response.SuccessCount, response.FailureCount)

	// Log any failures
	for i, resp := range response.Responses {
		if !resp.Success {
			log.Printf("Failed to send to token %s: %v", tokens[i], resp.Error)
		}
	}

	return response, nil
}

// SendToTopic sends a push notification to a topic
func (f *FCMService) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string, imageURL string) (string, error) {
	if f.client == nil {
		return "", fmt.Errorf("FCM client not initialized")
	}

	// Build the message
	message := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	// Add image if provided
	if imageURL != "" {
		message.Notification.ImageURL = imageURL
	}

	// Send the message
	response, err := f.client.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending FCM topic message: %v", err)
		return "", fmt.Errorf("failed to send FCM topic message: %v", err)
	}

	log.Printf("Successfully sent FCM topic message: %s", response)
	return response, nil
}

// SubscribeToTopic subscribes tokens to a topic
func (f *FCMService) SubscribeToTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	if f.client == nil {
		return nil, fmt.Errorf("FCM client not initialized")
	}

	response, err := f.client.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		log.Printf("Error subscribing to topic %s: %v", topic, err)
		return nil, fmt.Errorf("failed to subscribe to topic: %v", err)
	}

	log.Printf("Successfully subscribed %d tokens to topic %s", len(tokens), topic)
	return response, nil
}

// UnsubscribeFromTopic unsubscribes tokens from a topic
func (f *FCMService) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	if f.client == nil {
		return nil, fmt.Errorf("FCM client not initialized")
	}

	response, err := f.client.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		log.Printf("Error unsubscribing from topic %s: %v", topic, err)
		return nil, fmt.Errorf("failed to unsubscribe from topic: %v", err)
	}

	log.Printf("Successfully unsubscribed %d tokens from topic %s", len(tokens), topic)
	return response, nil
}