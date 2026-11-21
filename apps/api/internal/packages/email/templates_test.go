package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordResetEmail(t *testing.T) {
	url := "https://app.example.com/auth/reset-password#token=abc-123_DEF"
	msg, err := PasswordResetEmail("ada@example.com", "Ada", url, time.Hour)
	require.NoError(t, err)

	assert.Equal(t, "ada@example.com", msg.To)
	assert.Contains(t, msg.Text, url, "text is missing the link")
	assert.Contains(t, msg.HTML, "#token=abc-123_DEF", "html is missing the link")
	for _, body := range []string{msg.Text, msg.HTML} {
		assert.Contains(t, body, "60 minutes", "body is missing the expiry")
	}
}

func TestVerificationEmail(t *testing.T) {
	url := "https://app.example.com/auth/verify-email#token=abc-123_DEF"
	msg, err := VerificationEmail("ada@example.com", "Ada", url, 24*time.Hour)
	require.NoError(t, err)

	assert.Equal(t, "ada@example.com", msg.To)
	assert.Contains(t, msg.Subject, "Verify")
	assert.Contains(t, msg.Text, url, "text is missing the link")
	assert.Contains(t, msg.HTML, "#token=abc-123_DEF", "html is missing the link")
	for _, body := range []string{msg.Text, msg.HTML} {
		assert.Contains(t, body, "24 hours", "body is missing the expiry")
	}
}

func TestEmailsEscapeNamesInHTML(t *testing.T) {
	name := `<script>alert(1)</script>`

	reset, err := PasswordResetEmail("a@example.com", name, "https://example.com", time.Hour)
	require.NoError(t, err)
	changed, err := PasswordChangedEmail("a@example.com", name)
	require.NoError(t, err)

	for _, html := range []string{reset.HTML, changed.HTML} {
		assert.NotContains(t, html, "<script>", "name was not escaped")
	}
}

func TestPasswordChangedEmail(t *testing.T) {
	msg, err := PasswordChangedEmail("ada@example.com", "Ada")
	require.NoError(t, err)
	assert.Contains(t, msg.Subject, "changed")
	assert.Contains(t, msg.Text, "logged out")
}
