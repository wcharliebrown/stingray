package email

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
	dkimPrivateKey *rsa.PrivateKey
	dkimSelector   string
	dkimDomain     string
}

func NewEmailService(smtpHost, smtpPort, smtpUsername, smtpPassword, fromEmail, fromName, dkimPrivateKeyFile, dkimSelector, dkimDomain string) (*EmailService, error) {
	service := &EmailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		fromName:     fromName,
		dkimSelector: dkimSelector,
		dkimDomain:   dkimDomain,
	}

	// Load DKIM private key from file if provided
	if dkimPrivateKeyFile != "" {
		privateKey, err := loadDKIMPrivateKey(dkimPrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load DKIM private key from %s: %w", dkimPrivateKeyFile, err)
		}

		service.dkimPrivateKey = privateKey
	}

	return service, nil
}

func loadDKIMPrivateKey(filename string) (*rsa.PrivateKey, error) {
	// Read the private key file
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read DKIM private key file: %w", err)
	}

	// Decode the PEM block
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode DKIM private key from file")
	}

	// Try to parse as PKCS1 first
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return privateKey, nil
	}

	// If PKCS1 fails, try PKCS8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DKIM private key (tried both PKCS1 and PKCS8): %w", err)
	}

	// Convert to RSA private key
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not an RSA key")
	}

	return rsaKey, nil
}

func (e *EmailService) SendPasswordResetEmail(toEmail, resetURL string) error {
	subject := "Password Reset Request - Sting Ray CMS"
	
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #007cba;">Password Reset Request</h2>
        <p>Hello,</p>
        <p>You have requested a password reset for your Sting Ray CMS account.</p>
        <p>Click the button below to reset your password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #007cba; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a>
        </div>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #666;">%s</p>
        <p><strong>This link will expire in 1 hour.</strong></p>
        <p>If you didn't request this password reset, please ignore this email.</p>
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="color: #666; font-size: 12px;">This email was sent from Sting Ray CMS. Please do not reply to this email.</p>
    </div>
</body>
</html>`, resetURL, resetURL)

	textBody := fmt.Sprintf(`Password Reset Request

Hello,

You have requested a password reset for your Sting Ray CMS account.

Click the link below to reset your password:
%s

This link will expire in 1 hour.

If you didn't request this password reset, please ignore this email.

---
This email was sent from Sting Ray CMS. Please do not reply to this email.`, resetURL)

	return e.sendEmail(toEmail, subject, textBody, htmlBody)
}

func (e *EmailService) sendEmail(toEmail, subject, textBody, htmlBody string) error {
	// Build email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", e.fromName, e.fromEmail)
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")
	headers["Message-ID"] = fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), e.dkimDomain)

	// Add DKIM signature if private key is available
	if e.dkimPrivateKey != nil {
		dkimSignature, err := e.generateDKIMSignature(headers, textBody)
		if err != nil {
			return fmt.Errorf("failed to generate DKIM signature: %w", err)
		}
		headers["DKIM-Signature"] = dkimSignature
	}

	// Build email message
	var message strings.Builder
	
	// Add headers
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	
	// Add body
	message.WriteString(htmlBody)

	// Send email via SMTP
	auth := smtp.PlainAuth("", e.smtpUsername, e.smtpPassword, e.smtpHost)
	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	
	err := smtp.SendMail(addr, auth, e.fromEmail, []string{toEmail}, []byte(message.String()))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (e *EmailService) generateDKIMSignature(headers map[string]string, body string) (string, error) {
	// Create canonicalized header string
	var headerList []string
	for key, value := range headers {
		if key != "DKIM-Signature" { // Don't include DKIM-Signature in the signature
			headerList = append(headerList, fmt.Sprintf("%s: %s", strings.ToLower(key), strings.TrimSpace(value)))
		}
	}

	// Create canonicalized body
	canonicalizedBody := strings.TrimRight(body, " \t\r\n")
	
	// Create signature data
	signatureData := fmt.Sprintf("v=1; a=rsa-sha256; q=dns/txt; t=%d; c=relaxed/relaxed; h=%s; d=%s; s=%s; bh=%s; b=",
		time.Now().Unix(),
		strings.Join([]string{"from", "to", "subject", "date", "message-id"}, ":"),
		e.dkimDomain,
		e.dkimSelector,
		e.hashBody(canonicalizedBody))

	// Hash the signatureData
	hashed := sha256.Sum256([]byte(signatureData))

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, e.dkimPrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign DKIM data: %w", err)
	}

	// Encode signature
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	
	return signatureData + encodedSignature, nil
}

func (e *EmailService) hashBody(body string) string {
	hash := sha256.Sum256([]byte(body))
	return base64.StdEncoding.EncodeToString(hash[:])
} 