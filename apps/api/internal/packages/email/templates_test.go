package email

import (
	"strings"
	"testing"
	"time"
)

func TestPasswordResetEmail(t *testing.T) {
	url := "https://app.example.com/auth/reset-password#token=abc-123_DEF"
	msg, err := PasswordResetEmail("ada@example.com", "Ada", url, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if msg.To != "ada@example.com" {
		t.Errorf("to = %q", msg.To)
	}
	if !strings.Contains(msg.Text, url) {
		t.Errorf("text is missing the link: %s", msg.Text)
	}
	if !strings.Contains(msg.HTML, "#token=abc-123_DEF") {
		t.Errorf("html is missing the link: %s", msg.HTML)
	}
	for _, body := range []string{msg.Text, msg.HTML} {
		if !strings.Contains(body, "60 minutes") {
			t.Errorf("body is missing the expiry: %s", body)
		}
	}
}

func TestVerificationEmail(t *testing.T) {
	url := "https://app.example.com/auth/verify-email#token=abc-123_DEF"
	msg, err := VerificationEmail("ada@example.com", "Ada", url, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if msg.To != "ada@example.com" || !strings.Contains(msg.Subject, "Verify") {
		t.Errorf("to = %q, subject = %q", msg.To, msg.Subject)
	}
	if !strings.Contains(msg.Text, url) {
		t.Errorf("text is missing the link: %s", msg.Text)
	}
	if !strings.Contains(msg.HTML, "#token=abc-123_DEF") {
		t.Errorf("html is missing the link: %s", msg.HTML)
	}
	for _, body := range []string{msg.Text, msg.HTML} {
		if !strings.Contains(body, "24 hours") {
			t.Errorf("body is missing the expiry: %s", body)
		}
	}
}

func TestEmailsEscapeNamesInHTML(t *testing.T) {
	name := `<script>alert(1)</script>`

	reset, err := PasswordResetEmail("a@example.com", name, "https://example.com", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := PasswordChangedEmail("a@example.com", name)
	if err != nil {
		t.Fatal(err)
	}

	for _, html := range []string{reset.HTML, changed.HTML} {
		if strings.Contains(html, "<script>") {
			t.Errorf("name was not escaped: %s", html)
		}
	}
}

func TestPasswordChangedEmail(t *testing.T) {
	msg, err := PasswordChangedEmail("ada@example.com", "Ada")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg.Subject, "changed") || !strings.Contains(msg.Text, "logged out") {
		t.Errorf("unexpected message: %+v", msg)
	}
}
