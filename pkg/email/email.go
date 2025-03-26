package email

import (
	"fmt"
	"net/smtp"

	"github.com/bytetwiddler/digger/pkg/config"
)

func SendEmail(cfg *config.Config, subject, body string) error {
	var auth smtp.Auth
	if cfg.SMTP.Username != "" && cfg.SMTP.Password != "" {
		auth = smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)
	}

	to := []string{cfg.SMTP.To}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", cfg.SMTP.To, subject, body))
	addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)

	err := smtp.SendMail(addr, auth, cfg.SMTP.From, to, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
