package notifications

import (
	"strings"
	"testing"
	"time"
)

func TestSMTPConfigRequiresTLSCapableConfiguration(t *testing.T) {
	valid := SMTPConfig{Host: "smtp.example.test", Port: 587, From: "CANKONIX <alerts@example.test>", Timeout: 10 * time.Second}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	valid.Password = "secret"
	if valid.Validate() == nil {
		t.Fatal("expected incomplete credentials to fail")
	}
}
func TestFormatMessagePreventsSubjectHeaderInjection(t *testing.T) {
	message := formatMessage("alerts@example.test", "user@example.test", "Alert\r\nBcc: attacker@example.test", "Body\nLine")
	if strings.Contains(message, "\r\nBcc:") {
		t.Fatal("subject header injection was not removed")
	}
	if !strings.Contains(message, "Body\r\nLine") {
		t.Fatal("body must use SMTP line endings")
	}
}
