package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

type ResendService struct {
	client    *resend.Client
	fromEmail string
}

func NewResendService(apiKey, fromEmail string) *ResendService {
	client := resend.NewClient(apiKey)
	return &ResendService{
		client:    client,
		fromEmail: fromEmail,
	}
}

func (r *ResendService) SendOTPEmail(ctx context.Context, to, otp, purpose string) error {
	var subject, htmlContent string

	switch purpose {
	case "email_verification":
		subject = "Verify Your Email - Errand Shop"
		htmlContent = fmt.Sprintf(`
			<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #333;">Email Verification</h2>
				<p>Your verification code is:</p>
				<div style="background: #f4f4f4; padding: 20px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0;">
					%s
				</div>
				<p>This code will expire in 10 minutes.</p>
				<p>If you didn't request this, please ignore this email.</p>
			</div>
		`, otp)
	case "password_reset":
		subject = "Reset Your Password - Errand Shop"
		htmlContent = fmt.Sprintf(`
			<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #333;">Password Reset</h2>
				<p>Your password reset code is:</p>
				<div style="background: #f4f4f4; padding: 20px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0;">
					%s
				</div>
				<p>This code will expire in 10 minutes.</p>
				<p>If you didn't request this, please ignore this email.</p>
			</div>
		`, otp)
	default:
		return fmt.Errorf("unknown email purpose: %s", purpose)
	}

	params := &resend.SendEmailRequest{
		From:    r.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    htmlContent,
	}

	_, err := r.client.Emails.Send(params)
	return err
}

func (r *ResendService) SendWelcomeEmail(ctx context.Context, to, name string) error {
	subject := "Welcome to Errand Shop! 🎉"
	htmlContent := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">Welcome to Errand Shop, %s! 🎉</h2>
			<p>Thank you for joining our community. We're excited to have you on board!</p>
			<p>You can now:</p>
			<ul>
				<li>Browse and order from local stores</li>
				<li>Track your deliveries in real-time</li>
				<li>Manage your profile and addresses</li>
			</ul>
			<p>Happy shopping!</p>
			<p>The Errand Shop Team</p>
		</div>
	`, name)

	params := &resend.SendEmailRequest{
		From:    r.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    htmlContent,
	}

	_, err := r.client.Emails.Send(params)
	return err
}
