package email

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/domain"
	"html/template"

	"gopkg.in/gomail.v2"
)

type SMTPEmailSender struct {
	host     string
	port     int
	user     string
	password string
}

func NewSMTPEmailSender(host string, port int, user, password string) *SMTPEmailSender {
	return &SMTPEmailSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
	}
}

const emailTemplateHTML = `
<!DOCTYPE html>
<html>
<body>
	<h2>Thank you for returning {{.AssetID}}</h2>
	<p>Booking ID: {{.BookingID}}</p>
	<p>Total duration: {{.DurationMinutes}} minutes</p>
	<br/>
	<p>We hope to see you again!</p>
</body>
</html>
`

func (s *SMTPEmailSender) SendThankYouEmail(ctx context.Context, emailAddress string, session *domain.UsageSession) error {
	t, err := template.New("email").Parse(emailTemplateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, session); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.user)
	m.SetHeader("To", emailAddress)
	m.SetHeader("Subject", fmt.Sprintf("Thank you for returning %s", session.AssetID))
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(s.host, s.port, s.user, s.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
