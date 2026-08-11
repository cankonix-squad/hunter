package notifications

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type Message struct{ To, Subject, Text string }
type EmailSender interface{ Send(Message) error }

type SMTPConfig struct {
	Host                     string
	Port                     int
	Username, Password, From string
	Timeout                  time.Duration
}

func (c SMTPConfig) Validate() error {
	if c.Host == "" || c.Port < 1 || c.Port > 65535 {
		return errors.New("SMTP host and port are required")
	}
	if _, err := mail.ParseAddress(c.From); err != nil {
		return errors.New("SMTP from address is invalid")
	}
	if (c.Username == "") != (c.Password == "") {
		return errors.New("SMTP username and password must be configured together")
	}
	if c.Timeout <= 0 {
		return errors.New("SMTP timeout must be positive")
	}
	return nil
}

type SMTPSender struct{ Config SMTPConfig }

func (s SMTPSender) Send(message Message) error {
	if err := s.Config.Validate(); err != nil {
		return err
	}
	recipient, err := mail.ParseAddress(message.To)
	if err != nil {
		return errors.New("recipient email is invalid")
	}
	address := net.JoinHostPort(s.Config.Host, strconv.Itoa(s.Config.Port))
	conn, err := net.DialTimeout("tcp", address, s.Config.Timeout)
	if err != nil {
		return fmt.Errorf("connect SMTP: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(s.Config.Timeout))
	client, err := smtp.NewClient(conn, s.Config.Host)
	if err != nil {
		return fmt.Errorf("initialize SMTP: %w", err)
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); !ok {
		return errors.New("SMTP server does not support STARTTLS")
	}
	if err = client.StartTLS(&tls.Config{ServerName: s.Config.Host, MinVersion: tls.VersionTLS12}); err != nil {
		return fmt.Errorf("start SMTP TLS: %w", err)
	}
	if s.Config.Username != "" {
		if ok, _ := client.Extension("AUTH"); !ok {
			return errors.New("SMTP server does not support authentication")
		}
		if err = client.Auth(smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.Host)); err != nil {
			return fmt.Errorf("authenticate SMTP: %w", err)
		}
	}
	from, _ := mail.ParseAddress(s.Config.From)
	if err = client.Mail(from.Address); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err = client.Rcpt(recipient.Address); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("open SMTP body: %w", err)
	}
	body := formatMessage(from.String(), recipient.String(), message.Subject, message.Text)
	if _, err = io.WriteString(writer, body); err != nil {
		return fmt.Errorf("write SMTP body: %w", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("close SMTP body: %w", err)
	}
	if err = client.Quit(); err != nil {
		return fmt.Errorf("finish SMTP session: %w", err)
	}
	return nil
}

type DryRunSender struct{}

func (DryRunSender) Send(Message) error { return nil }

func formatMessage(from, to, subject, text string) string {
	subject = strings.ReplaceAll(strings.ReplaceAll(subject, "\r", " "), "\n", " ")
	return "From: " + from + "\r\nTo: " + to + "\r\nSubject: " + subject + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n" + strings.ReplaceAll(text, "\n", "\r\n") + "\r\n"
}
