// Package email sends transactional email.
package email

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	mail "github.com/wneessen/go-mail"
)

const sendTimeout = 15 * time.Second

// Message is one email. Text is the fallback for clients that don't render HTML.
type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// Mailer sends email.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

// SMTPOptions configures SMTPMailer.
type SMTPOptions struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	// AllowPlaintext sends without TLS when the server doesn't offer it. It's for
	// local tools such as Mailpit. Otherwise TLS is required.
	AllowPlaintext bool
}

// SMTPMailer sends through any SMTP server, such as Amazon SES, Resend or Mailpit.
type SMTPMailer struct {
	options SMTPOptions
}

// NewSMTPMailer returns a mailer for options. It connects when sending.
func NewSMTPMailer(options SMTPOptions) *SMTPMailer {
	return &SMTPMailer{options: options}
}

// Send opens a connection, sends msg and closes it. Reset emails are infrequent
// enough that keeping a connection open isn't worth the complexity.
func (m *SMTPMailer) Send(ctx context.Context, msg Message) error {
	message := mail.NewMsg()
	if err := message.FromFormat(AppName, m.options.From); err != nil {
		return fmt.Errorf("invalid sender: %w", err)
	}
	if err := message.To(msg.To); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	message.Subject(msg.Subject)
	message.SetBodyString(mail.TypeTextPlain, msg.Text)
	message.AddAlternativeString(mail.TypeTextHTML, msg.HTML)

	client, err := mail.NewClient(m.options.Host, m.clientOptions()...)
	if err != nil {
		return fmt.Errorf("configure smtp client: %w", err)
	}
	if err := client.DialAndSendWithContext(ctx, message); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (m *SMTPMailer) clientOptions() []mail.Option {
	tlsPolicy := mail.TLSMandatory
	if m.options.AllowPlaintext {
		tlsPolicy = mail.TLSOpportunistic
	}

	options := []mail.Option{
		mail.WithTLSPolicy(tlsPolicy),
		mail.WithPort(m.options.Port),
		mail.WithTimeout(sendTimeout),
	}
	if m.options.Username != "" {
		options = append(options,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(m.options.Username),
			mail.WithPassword(m.options.Password),
		)
	}
	return options
}

// LogMailer writes email to the log instead of sending it. It's the fallback for
// local development without SMTP. Messages contain reset links, so config
// validation stops it being used in release mode.
type LogMailer struct{}

// Send logs msg.
func (LogMailer) Send(_ context.Context, msg Message) error {
	slog.Info("email not sent: SMTP_HOST is unset", "to", msg.To, "subject", msg.Subject, "text", msg.Text)
	return nil
}
