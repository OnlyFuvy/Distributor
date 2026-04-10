package service

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"time"
)

// EmailService — сервис для отправки email
type EmailService struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
}

// GenerateCode — генерирует 6-значный код подтверждения
func GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// SendVerificationEmail — отправляет email с кодом подтверждения
func (s *EmailService) SendVerificationEmail(toEmail, code string) error {
	subject := "🎵 Music App — Код подтверждения"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background: #1a1a2e; color: #fff; padding: 20px;">
			<div style="max-width: 500px; margin: 0 auto; background: #16213e; padding: 30px; border-radius: 15px;">
				<h1 style="color: #00d9ff;">Подтверждение регистрации</h1>
				<p>Ваш код подтверждения:</p>
				<div style="background: #00d9ff; color: #1a1a2e; font-size: 32px; font-weight: bold; 
				    padding: 20px; text-align: center; border-radius: 10px; margin: 20px 0;">
					%s
				</div>
				<p>Код действителен 10 минут.</p>
				<p style="color: #ccc; font-size: 12px;">Если вы не регистрировались, просто проигнорируйте это письмо.</p>
			</div>
		</body>
		</html>
	`, code)

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		toEmail, subject, body)

	addr := fmt.Sprintf("%s:%d", s.SMTPHost, s.SMTPPort)
	auth := smtp.PlainAuth("", s.SMTPUser, s.SMTPPassword, s.SMTPHost)

	return smtp.SendMail(addr, auth, s.FromEmail, []string{toEmail}, []byte(msg))
}
