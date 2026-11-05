package email

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

// AppName appears in the sender name and in email content.
const AppName = "Turgorepo"

var passwordResetHTML = template.Must(template.New("password_reset").Parse(`<p>Hi {{.Name}},</p>
<p>We received a request to reset your {{.AppName}} password.</p>
<p><a href="{{.URL}}">Reset your password</a></p>
<p>This link works once and expires in {{.Minutes}} minutes. If you didn't ask for this, you can ignore this email and your password won't change.</p>
`))

var passwordChangedHTML = template.Must(template.New("password_changed").Parse(`<p>Hi {{.Name}},</p>
<p>Your {{.AppName}} password was just changed, and you've been logged out on all devices.</p>
<p>If this wasn't you, reset your password right away and contact support.</p>
`))

type templateData struct {
	AppName string
	Name    string
	URL     string
	Minutes int
}

// PasswordResetEmail builds the email containing a reset link.
func PasswordResetEmail(to, name, resetURL string, validFor time.Duration) (Message, error) {
	minutes := int(validFor.Minutes())
	html, err := render(passwordResetHTML, templateData{AppName: AppName, Name: name, URL: resetURL, Minutes: minutes})
	if err != nil {
		return Message{}, err
	}
	return Message{
		To:      to,
		Subject: fmt.Sprintf("Reset your %s password", AppName),
		Text: fmt.Sprintf(
			"Hi %s,\n\nWe received a request to reset your %s password. Open this link to continue:\n\n%s\n\nThis link works once and expires in %d minutes. If you didn't ask for this, you can ignore this email and your password won't change.\n",
			name, AppName, resetURL, minutes,
		),
		HTML: html,
	}, nil
}

// PasswordChangedEmail builds the notice sent after a successful reset, so the
// owner of the account finds out if someone else did it.
func PasswordChangedEmail(to, name string) (Message, error) {
	html, err := render(passwordChangedHTML, templateData{AppName: AppName, Name: name})
	if err != nil {
		return Message{}, err
	}
	return Message{
		To:      to,
		Subject: fmt.Sprintf("Your %s password was changed", AppName),
		Text: fmt.Sprintf(
			"Hi %s,\n\nYour %s password was just changed, and you've been logged out on all devices.\n\nIf this wasn't you, reset your password right away and contact support.\n",
			name, AppName,
		),
		HTML: html,
	}, nil
}

func render(tmpl *template.Template, data templateData) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render %s email: %w", tmpl.Name(), err)
	}
	return buf.String(), nil
}
