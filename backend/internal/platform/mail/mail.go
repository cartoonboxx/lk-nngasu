// Package mail — отправка писем (приглашения задать пароль).
//
// Без SMTP доставка откладывается; содержимое писем и токены не выводятся в лог.
package mail

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Sender отправляет текстовое письмо.
type Sender interface {
	Send(to, subject, body string) error
}

// SMTP отправляет письма через почтовый сервер с STARTTLS (обычно порт 587:
// Яндекс, Mail.ru, Gmail). Пароль — пароль приложения почтового ящика.
type SMTP struct {
	Host, Port, User, Password, From string
}

// Send отправляет письмо. Заголовки и тело — UTF-8, тема кодируется по RFC 2047.
func (s SMTP) Send(to, subject, body string) error {
	// Защита от подстановки заголовков через адрес или тему.
	if strings.ContainsAny(to+subject, "\r\n") {
		return fmt.Errorf("mail: invalid header value")
	}
	msg := strings.Join([]string{
		"From: " + s.From,
		"To: " + to,
		"Subject: =?UTF-8?B?" + b64(subject) + "?=",
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.User != "" {
		auth = smtp.PlainAuth("", s.User, s.Password, s.Host)
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(s.Host, s.Port), 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(45 * time.Second)); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	// Ссылки активации нельзя отправлять по незашифрованному SMTP.
	if err := client.StartTLS(&tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}); err != nil {
		return err
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(s.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := io.WriteString(writer, msg); err != nil {
		writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// Log сообщает об отсутствии SMTP, не раскрывая содержимое письма.
type Log struct{}

func (Log) Send(to, subject, body string) error {
	slog.Warn("SMTP is not configured, email was not sent", "to", to, "subject", subject)
	return fmt.Errorf("mail: SMTP is not configured")
}
func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
