package email

import (
	"testing"

	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestSendEmail(t *testing.T) {
	cfg := &config.Config{
		SMTP: struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			From     string `yaml:"from"`
			To       string `yaml:"to"`
		}{
			Username: "billb",
			Password: "your_password",
			Host:     "localhost",
			Port:     25,
			From:     "billb@localhost",
			To:       "billb@localhost",
		},
	}

	subject := "Test Subject"
	body := "This is a test email body."

	err := SendEmail(cfg, subject, body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
}
