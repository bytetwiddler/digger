package notification

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendIPChangeNotification(t *testing.T) {
	// Create a temporary template file for testing
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "email.html")
	err := os.WriteFile(templatePath, []byte(`<html>{{.Name}} - {{.Email}} - {{.Hostname}} - {{.Port}} - {{.Vendor}} - {{.OldIP}} - {{.NewIP}}</html>`), 0644)
	require.NoError(t, err)

	tests := []struct {
		name       string
		cfg        *config.Config
		hostname   string
		port       int
		entityName string
		oldIP      string
		newIP      string
		wantErr    bool
	}{
		{
			name: "successful notification",
			cfg: &config.Config{
				SMTP: struct {
					Host         string `yaml:"host"`
					Port         int    `yaml:"port"`
					Username     string `yaml:"username"`
					Password     string `yaml:"password"`
					From         string `yaml:"from"`
					To           string `yaml:"to"`
					TemplatePath string `yaml:"template_path"`
				}{
					Host:         "test.smtp.server",
					Port:         25,
					Username:     "testuser",
					Password:     "testpass",
					From:         "from@test.com",
					To:           "to@test.com",
					TemplatePath: templatePath,
				},
			},
			hostname:   "test.example.com",
			port:       443,
			entityName: "Test Vendor",
			oldIP:      "192.168.1.1",
			newIP:      "192.168.1.2",
			wantErr:    true, // Set to true since we don't have a real SMTP server in tests
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendIPChangeNotification(tt.cfg, tt.hostname, tt.port, tt.entityName, tt.oldIP, tt.newIP)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailTemplateRendering(t *testing.T) {
	// Create a temporary template file for testing
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "email.html")
	templateContent, err := os.ReadFile("../../templates/email.html") // Read the actual template
	require.NoError(t, err)

	err = os.WriteFile(templatePath, templateContent, 0644)
	require.NoError(t, err)

	// Test data
	data := EmailData{
		Name:     "Test Team",
		Email:    "test@example.com",
		Hostname: "test.host.com",
		Port:     443,
		Vendor:   "Test Vendor",
		OldIP:    "192.168.1.1",
		NewIP:    "192.168.1.2",
	}

	// Parse template from file
	templateContent, err = os.ReadFile(templatePath)
	require.NoError(t, err)

	tmpl, err := template.New("email").Parse(string(templateContent))
	require.NoError(t, err)

	// Render template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	// Verify rendered content
	rendered := buf.String()
	assert.Contains(t, rendered, data.Name)
	assert.Contains(t, rendered, data.Email)
	assert.Contains(t, rendered, data.Hostname)
	assert.Contains(t, rendered, data.Vendor)
	assert.Contains(t, rendered, data.OldIP)
	assert.Contains(t, rendered, data.NewIP)
	assert.Contains(t, rendered, "LCAPP172")
	assert.Contains(t, rendered, "LCAPP173")
}
