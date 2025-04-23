package notification

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"os"
	"text/template"

	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/gophish/gomail"
)

type EmailData struct {
	Name     string
	Email    string
	Hostname string
	Port     int
	Vendor   string
	OldIP    string
	NewIP    string
}

func SendIPChangeNotification(cfg *config.Config, hostname string, port int, entityName string, oldIP, newIP string) error {
	// Read the template file
	templateContent, err := os.ReadFile(cfg.SMTP.TemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read email template file: %w", err)
	}

	// Parse the email template
	t, err := template.New("email").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Prepare dynamic data for the template
	data := EmailData{
		Name:     "Network Security Team",
		Email:    cfg.SMTP.To,
		Hostname: hostname,
		Port:     port,
		Vendor:   entityName,
		OldIP:    oldIP,
		NewIP:    newIP,
	}

	// Render the template into a string
	var renderedEmail bytes.Buffer
	err = t.Execute(&renderedEmail, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Create new email message
	mail := gomail.NewMessage()
	mail.SetHeader("From", cfg.SMTP.From)
	mail.SetHeader("To", cfg.SMTP.To)
	mail.SetHeader("Subject", "IP Address Change Notification")
	mail.SetBody("text/html", renderedEmail.String())

	// Set up the email dialer
	d := gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send the email
	if err := d.DialAndSend(mail); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
