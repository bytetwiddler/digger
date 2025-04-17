package notification

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"text/template"

	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/gophish/gomail"
)

const emailTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Third Party IP Change Alert</title>
    <style>
        table {
            width: 80%;
            border-collapse: collapse;
            margin: 20px 0;
            font-size: 14px;
            text-align: left;
        }
        table, th, td {
            border: 1px solid #dddddd;
        }
        th {
            padding: 10px;
            text-align: center;
        }
        td {
            padding: 10px;
            text-align: center;
        }
        th {
            background-color: #f2f2f2;
        }
    </style>
</head>
<body>
    <h2>Greetings, {{.Name}}!</h2>
    <p>DNS resolution for a 3rd party vendor file transfer site has changed.</br></br>
                Please alter outbound rules so that the following servers can access the new IP address 
				at the desired ports:</br></br>

				<b>LCAPP172</b></br>
				<b>LCAPP173</b></br>
				<b>LCISBATCH01</b></br>
				<b>LCISBATCH02</b></br>
				<b>LCMBD01</b></br></br>
                
				Failure to do so before the next batch run will result in production delays.</p>
    <table>
        <tr><td>Contact DL</td><td>{{.Email}}</td></tr>
        <tr><td>Site hostname</td><td>{{.Hostname}}</td></tr>
        <tr><td>Port</td><td>{{.Port}}</td></tr>
        <tr><td>Vendor</td><td>{{.Vendor}}</td></tr>
        <tr><td>Old IP</td><td>{{.OldIP}}</td></tr>
        <tr><td>New IP</td><td>{{.NewIP}}</td></tr>
    </table>
</body>
</html>`

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
	// Parse the email template
	t, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Prepare dynamic data for the template
	data := EmailData{
		Name:     "Security Team",
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
