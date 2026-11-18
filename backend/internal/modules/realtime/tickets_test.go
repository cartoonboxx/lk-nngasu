package realtime

import (
	"errors"
	"testing"
	"time"

	"university-api/internal/platform/middleware"
)

func TestTicketsOneTime(t *testing.T) {
	now := time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)
	s := newTickets()
	s.now = func() time.Time { return now }

	id, exp, err := s.issue(middleware.Principal{UserID: 7, Sid: "s"}, "http://a")
	if err != nil || !exp.Equal(now.Add(TicketTTL)) || len(id) < 40 {
		t.Fatalf("issue: %q %v %v", id, exp, err)
	}
	got, err := s.redeem(id)
	if err != nil || got.principal.UserID != 7 || got.origin != "http://a" {
		t.Fatalf("redeem: %+v %v", got, err)
	}
	if _, err := s.redeem(id); !errors.Is(err, errTicketUsed) {
		t.Fatalf("second redeem: %v", err)
	}
	if _, err := s.redeem("nope"); !errors.Is(err, errTicketExpired) {
		t.Fatalf("unknown: %v", err)
	}

	id2, _, _ := s.issue(middleware.Principal{UserID: 7}, "")
	now = now.Add(TicketTTL + time.Second)
	if _, err := s.redeem(id2); !errors.Is(err, errTicketExpired) {
		t.Fatalf("expired: %v", err)
	}
}

func TestDecodeStrict(t *testing.T) {
	var v struct {
		ChatID *int64 `json:"chat_id"`
	}
	for _, raw := range []string{`{"chat_id":1,"x":2}`, `{"chat_id":"a"}`, `{`} {
		err := Decode([]byte(raw), &v)
		if err == nil || toWSError(err).Code != "invalid_payload" {
			t.Errorf("%s: %v", raw, err)
		}
	}
	if err := Decode([]byte(`{"chat_id":5}`), &v); err != nil || *v.ChatID != 5 {
		t.Fatalf("ok: %v", err)
	}
}
