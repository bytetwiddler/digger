package email

import (
	"fmt"
	"net/smtp"

	"github.com/bytetwiddler/digger/pkg/config"
)

func SendEmail(cfg *config.Config, subject, body string) error {
	auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)
	to := []string{cfg.SMTP.To}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", cfg.SMTP.To, subject, body))
	addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)

	return fmt.Errorf("failed to send email: %w", smtp.SendMail(addr, auth, cfg.SMTP.From, to, msg))
}
