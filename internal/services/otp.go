package services

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

type OTPService struct {
	db           *sql.DB
	emailService *EmailService
}

func NewOTPService(db *sql.DB) *OTPService {
	return &OTPService{
		db:           db,
		emailService: NewEmailService(),
	}
}

// SendOTP generates and sends OTP to email
func (s *OTPService) SendOTP(email string) error {
	_, err := s.SendOTPAndGetCode(email)
	return err
}

// SendOTPAndGetCode generates and sends OTP to email, returns the code
func (s *OTPService) SendOTPAndGetCode(email string) (string, error) {
	// Generate OTP
	code, err := GenerateOTP()
	if err != nil {
		return "", err
	}

	// Store OTP in database (expires in 10 minutes)
	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	
	result, err := s.db.Exec(`
		INSERT INTO yt_otp_codes (email, code, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (email) DO UPDATE
		SET code = EXCLUDED.code,
		    expires_at = EXCLUDED.expires_at,
		    created_at = EXCLUDED.created_at,
		    attempts = 0
	`, email, code, expiresAt)
	
	if err != nil {
		log.Printf("[ERROR] Failed to insert OTP into database for %s: %v", email, err)
		return "", err
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("[DEBUG] OTP stored in database for %s (rows affected: %d)", email, rowsAffected)

	// Send email
	err = s.emailService.SendOTP(email, code)
	return code, err
}

// VerifyOTP checks if the provided OTP is valid
func (s *OTPService) VerifyOTP(email, code string) error {
	var storedCode string
	var expiresAt time.Time
	var attempts int

	err := s.db.QueryRow(`
		SELECT code, expires_at, attempts
		FROM yt_otp_codes
		WHERE email = $1
	`, email).Scan(&storedCode, &expiresAt, &attempts)

	if err == sql.ErrNoRows {
		return errors.New("no OTP found for this email")
	}
	if err != nil {
		return err
	}

	// Check if too many attempts
	if attempts >= 5 {
		return errors.New("too many failed attempts, please request a new code")
	}

	// Check if expired
	if time.Now().UTC().After(expiresAt) {
		return errors.New("OTP has expired, please request a new code")
	}

	// Check if code matches
	if storedCode != code {
		// Increment attempts
		s.db.Exec(`
			UPDATE yt_otp_codes 
			SET attempts = attempts + 1 
			WHERE email = $1
		`, email)
		return errors.New("invalid OTP code")
	}

	// OTP is valid - mark as used
	_, err = s.db.Exec(`
		DELETE FROM yt_otp_codes WHERE email = $1
	`, email)

	return err
}

// CleanupExpiredOTPs removes expired OTP codes (run periodically)
func (s *OTPService) CleanupExpiredOTPs() error {
	_, err := s.db.Exec(`
		DELETE FROM yt_otp_codes 
		WHERE expires_at < NOW()
	`)
	return err
}