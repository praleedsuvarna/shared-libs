package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// GenerateEmailVerificationToken creates a secure random token
func GenerateEmailVerificationToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// SendVerificationEmail sends an email with verification link
func SendVerificationEmail(email, verificationToken string) error {
	from := mail.NewEmail("Your App Name", os.Getenv("SENDER_EMAIL"))
	subject := "Verify Your Email"
	to := mail.NewEmail("", email)

	// Construct verification link
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s",
		os.Getenv("FRONTEND_URL"),
		verificationToken,
	)

	// HTML Content
	htmlContent := fmt.Sprintf(`
        <h1>Verify Your Email</h1>
        <p>Click the link below to verify your email address:</p>
        <a href="%s">Verify Email</a>
        <p>If you did not create an account, please ignore this email.</p>
    `, verificationLink)

	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)

	// Create a new SendGrid client
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))

	// Send the email
	_, err := client.Send(message)
	return err
}
