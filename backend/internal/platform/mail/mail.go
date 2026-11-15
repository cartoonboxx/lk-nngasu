// Package mail — отправка писем (приглашения задать пароль).
//
// Если SMTP не настроен, письма не отправляются, а пишутся в лог сервера:
// так можно проверять сценарий приглашения локально, без почтового ящика.
package mail

import (
	"encoding/base64"
	"fmt"
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
	// smtp.SendMail сам переходит на TLS через STARTTLS, если сервер его поддерживает.
	return smtp.SendMail(net.JoinHostPort(s.Host, s.Port), auth, s.From, []string{to}, []byte(msg))
}

// Log не отправляет письма, а пишет их в лог. Для разработки и стенда без почты.
type Log struct{}

func (Log) Send(to, subject, body string) error {
	slog.Warn("SMTP is not configured, email was not sent", "to", to, "subject", subject, "body", body)
	return nil
}
func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
