package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Add these fields to your existing Config struct
type Config struct {
	Env              string
	Port             string
	DatabaseUrl      string
	ResendAPIKey     string
	FromEmail        string
	AllowedOrigins   string
	Version          uint
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromPhone  string
	JWTSecret        string

	// New fields for scaling
	JWTAccessSecret  string
	JWTRefreshSecret string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration

	// Payment & Notifications
	FCMServerKey             string
	UploadBucketURL          string

	// Paystack Configuration
	PaystackSecretKey        string
	PaystackPublicKey        string
	PaystackWebhookSecret    string
	AppBaseURL               string
	CallbackURL              string
}

// Add to LoadConfig() function
func LoadConfig() *Config {
	// Load .env files in order of priority (.env.local overrides .env)
	err := godotenv.Load(".env.local", ".env")
	if err != nil {
		log.Printf("Warning: Error loading .env files: %v", err)
	}

	databaseUrl := os.Getenv("DATABASE")
	if databaseUrl == "" {
		log.Fatal("DATABASE environment variable is missing!")
	}

	resendApiKey := os.Getenv("RESEND_API_KEY")
	if resendApiKey == "" {
		log.Fatal("RESEND_API_KEY is missing!")
	}

	allowedOrigins := os.Getenv("AllowedOrigins")
	if allowedOrigins == "" {
		log.Fatal("AllowedOrigins url is missing")
	}

	versionStr := os.Getenv("VERSION")
	if versionStr == "" {
		log.Fatal("version is missing")
	}

	version, err := strconv.ParseUint(versionStr, 10, 32)
	if err != nil {
		log.Fatalf("Invalid version format: %v", err)
	}

	twilioSID := os.Getenv("TWILIO_ACCOUNT_SID")
	twilioToken := os.Getenv("TWILIO_AUTH_TOKEN")
	twilioFrom := os.Getenv("TWILIO_FROM_PHONE")

	if twilioSID == "" || twilioToken == "" || twilioFrom == "" {
		log.Fatal("Twilio SMS configs are missing!")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is missing!")
	}

	// Remove this line: config.FromEmail = getEnv("FROM_EMAIL", "noreply@errandshop.com")

	return &Config{
		Version:          uint(version),
		Port:             getEnv("PORT", "9090"),
		DatabaseUrl:      databaseUrl,
		ResendAPIKey:     resendApiKey,
		FromEmail:        getEnv("FROM_EMAIL", "noreply@errandshop.com"), // Add this line
		AllowedOrigins:   allowedOrigins,
		TwilioAccountSID: twilioSID,
		TwilioAuthToken:  twilioToken,
		TwilioFromPhone:  twilioFrom,
		JWTSecret:        jwtSecret,

		// Notifications
		FCMServerKey:             getEnv("FCM_SERVER_KEY", ""),
		UploadBucketURL:          getEnv("UPLOAD_BUCKET_URL", ""),

		// Paystack Configuration
		PaystackSecretKey:        getEnv("PAYSTACK_SECRET_KEY", ""),
		PaystackPublicKey:        getEnv("PAYSTACK_PUBLIC_KEY", ""),
		PaystackWebhookSecret:    getEnv("PAYSTACK_WEBHOOK_SECRET", ""),
		AppBaseURL:               getEnv("APP_BASE_URL", "http://localhost:9090"),
		CallbackURL:              getEnv("CALLBACK_URL", ""),
	}
}

// getEnv tries to get the value of the key from the environment variables
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt tries to get the integer value of the key from the environment variables
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}
