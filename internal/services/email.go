package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type EmailService struct {
	sesClient *ses.Client
	FromEmail string
	FromName  string
}

type OTPData struct {
	Code      string
	ExpiresAt time.Time
	Email     string
}

func NewEmailService() *EmailService {
	log.Printf("[INFO] Initializing Email Service...")
	
	fromEmail := getEnvOrDefault("FROM_EMAIL", "noreply@yukti.io")
	fromName := getEnvOrDefault("FROM_NAME", "Yukti FinOps")
	
	log.Printf("[INFO] Email config: FROM_EMAIL=%s, FROM_NAME=%s", fromEmail, fromName)
	
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("[WARN] Failed to load AWS config for SES: %v. Using dev mode.", err)
		return &EmailService{
			sesClient: nil,
			FromEmail: fromEmail,
			FromName:  fromName,
		}
	}

	log.Printf("[INFO] AWS SES client initialized successfully")
	return &EmailService{
		sesClient: ses.NewFromConfig(cfg),
		FromEmail: fromEmail,
		FromName:  fromName,
	}
}

func (e *EmailService) SendOTP(email, code string) error {
	log.Printf("[INFO] SendOTP called for user: %s", email)
	
	// Print OTP prominently for testing
	fmt.Printf("\n🔐 OTP CODE FOR TESTING 🔐\n")
	fmt.Printf("================================\n")
	fmt.Printf("User Email: %s\n", email)
	fmt.Printf("OTP Code: %s\n", code)
	fmt.Printf("Valid for: 10 minutes\n")
	fmt.Printf("================================\n\n")
	
	subject := fmt.Sprintf("Yukti OTP for %s", email)
	body := e.generateOTPEmailBody(code)
	
	// In sandbox mode, send to verified email (FROM_EMAIL)
	// In production, send to user's email
	recipient := e.FromEmail
	if e.sesClient == nil {
		log.Printf("[INFO] Dev mode: sending to user email %s", email)
		recipient = email // Dev mode - use user email
	} else {
		log.Printf("[INFO] Sandbox mode: sending to verified email %s (user: %s)", recipient, email)
	}
	
	log.Printf("[INFO] Calling sendEmail with recipient: %s, subject: %s", recipient, subject)
	err := e.sendEmail(recipient, subject, body)
	if err != nil {
		log.Printf("[ERROR] SendOTP failed: %v", err)
		return err
	}
	
	log.Printf("[INFO] SendOTP completed successfully for %s", email)
	return nil
}

func (e *EmailService) SendWelcomeEmail(email, companyName string) error {
	subject := "Welcome to Yukti FinOps!"
	body := e.generateWelcomeEmailBody(companyName)
	
	return e.sendEmail(email, subject, body)
}

func (e *EmailService) SendInvitationEmail(email, tenantName, inviteURL string) error {
	subject := fmt.Sprintf("You've been invited to join %s on Yukti", tenantName)
	body := e.generateInvitationEmailBody(tenantName, inviteURL)
	
	return e.sendEmail(email, subject, body)
}

func (e *EmailService) sendEmail(to, subject, body string) error {
	log.Printf("[INFO] sendEmail called: to=%s, subject=%s", to, subject)
	
	// Dev mode - print to console if SES not configured
	if e.sesClient == nil {
		log.Printf("[INFO] SES client is nil - using dev mode")
		fmt.Printf("\n=== EMAIL (DEV MODE) ===\n")
		fmt.Printf("To: %s\n", to)
		fmt.Printf("Subject: %s\n", subject)
		fmt.Printf("Body:\n%s\n", body)
		fmt.Printf("=====================\n\n")
		log.Printf("[INFO] Dev mode email sent successfully")
		return nil
	}

	log.Printf("[INFO] Preparing AWS SES email...")
	source := fmt.Sprintf("%s <%s>", e.FromName, e.FromEmail)
	log.Printf("[INFO] Email source: %s", source)
	log.Printf("[INFO] Email destination: %s", to)
	
	// Send via AWS SES
	input := &ses.SendEmailInput{
		Source: aws.String(source),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{
				Html: &types.Content{
					Data:    aws.String(body),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}

	log.Printf("[INFO] Calling AWS SES SendEmail API...")
	result, err := e.sesClient.SendEmail(context.Background(), input)
	if err != nil {
		log.Printf("[ERROR] AWS SES SendEmail failed: %v", err)
		log.Printf("[ERROR] Failed to send email via SES to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[INFO] AWS SES SendEmail succeeded. MessageId: %v", result.MessageId)
	log.Printf("[INFO] Email sent successfully via SES to %s", to)
	return nil
}

func (e *EmailService) generateOTPEmailBody(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Yukti Verification Code</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="text-align: center; margin-bottom: 30px;">
        <h1 style="color: #2563eb;">Yukti FinOps</h1>
        <p style="color: #6b7280;">Cloud Cost Optimization Platform</p>
    </div>
    
    <div style="background: #f8fafc; padding: 30px; border-radius: 8px; text-align: center;">
        <h2 style="color: #1f2937; margin-bottom: 20px;">Verify Your Email Address</h2>
        
        <p style="color: #4b5563; margin-bottom: 30px;">
            Enter this verification code to complete your account setup:
        </p>
        
        <div style="background: white; padding: 20px; border-radius: 6px; border: 2px solid #e5e7eb; margin: 20px 0;">
            <span style="font-size: 32px; font-weight: bold; color: #2563eb; letter-spacing: 4px;">%s</span>
        </div>
        
        <p style="color: #6b7280; font-size: 14px; margin-top: 20px;">
            This code expires in 10 minutes for security reasons.
        </p>
        
        <p style="color: #6b7280; font-size: 14px;">
            If you didn't request this code, please ignore this email.
        </p>
    </div>
    
    <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb;">
        <p style="color: #9ca3af; font-size: 12px;">
            🔒 Your financial data is protected with enterprise-grade security
        </p>
    </div>
</body>
</html>`, code)
}

func (e *EmailService) generateWelcomeEmailBody(companyName string) string {
	greeting := "Welcome to Yukti!"
	if companyName != "" {
		greeting = fmt.Sprintf("Welcome %s to Yukti!", companyName)
	}
	
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to Yukti FinOps</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="text-align: center; margin-bottom: 30px;">
        <h1 style="color: #2563eb;">%s</h1>
        <p style="color: #6b7280;">Start optimizing your cloud costs today</p>
    </div>
    
    <div style="background: #f8fafc; padding: 30px; border-radius: 8px;">
        <h2 style="color: #1f2937;">Your account is ready!</h2>
        
        <p style="color: #4b5563; line-height: 1.6;">
            Thank you for joining Yukti FinOps. You're now ready to:
        </p>
        
        <ul style="color: #4b5563; line-height: 1.8;">
            <li>Connect your AWS account securely</li>
            <li>Discover hidden cost optimization opportunities</li>
            <li>Generate Infrastructure as Code for fixes</li>
            <li>Track your savings with real-time dashboards</li>
        </ul>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="http://localhost:3000/onboarding" 
               style="background: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold;">
                Complete Setup →
            </a>
        </div>
        
        <p style="color: #6b7280; font-size: 14px;">
            Need help? Reply to this email or visit our documentation.
        </p>
    </div>
    
    <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb;">
        <p style="color: #9ca3af; font-size: 12px;">
            Yukti FinOps - Enterprise Cloud Cost Optimization
        </p>
    </div>
</body>
</html>`, greeting)
}

// GenerateOTP creates a 6-digit OTP code
func GenerateOTP() (string, error) {
	max := big.NewInt(999999)
	min := big.NewInt(100000)
	
	n, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%06d", n.Add(n, min).Int64()), nil
}

func (e *EmailService) generateInvitationEmailBody(tenantName, inviteURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Team Invitation</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="text-align: center; margin-bottom: 30px;">
        <h1 style="color: #2563eb;">Yukti FinOps</h1>
        <p style="color: #6b7280;">Cloud Cost Optimization Platform</p>
    </div>
    
    <div style="background: #f8fafc; padding: 30px; border-radius: 8px;">
        <h2 style="color: #1f2937;">You've been invited!</h2>
        
        <p style="color: #4b5563; line-height: 1.6;">
            You've been invited to join <strong>%s</strong> on Yukti FinOps.
        </p>
        
        <p style="color: #4b5563; line-height: 1.6;">
            Click the button below to accept the invitation and join the team:
        </p>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" 
               style="background: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold;">
                Accept Invitation →
            </a>
        </div>
        
        <p style="color: #6b7280; font-size: 14px;">
            This invitation expires in 7 days.
        </p>
        
        <p style="color: #6b7280; font-size: 14px;">
            If you didn't expect this invitation, you can safely ignore this email.
        </p>
    </div>
    
    <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb;">
        <p style="color: #9ca3af; font-size: 12px;">
            Yukti FinOps - Enterprise Cloud Cost Optimization
        </p>
    </div>
</body>
</html>`, tenantName, inviteURL)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}