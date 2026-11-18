package realtime

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"university-api/internal/platform/middleware"
)

// TicketTTL — срок жизни билета на подключение.
const TicketTTL = 30 * time.Second

// ticket — одноразовый билет: кто, из какой сессии и с какого Origin его получил.
type ticket struct {
	principal middleware.Principal
	origin    string // "" — билет выдан без Origin (Postman, curl): подключаться можно с любого допустимого
	expires   time.Time
	used      bool
}

// tickets — хранилище билетов в памяти процесса. Билет живёт 30 секунд, поэтому
// после рестарта клиент просто получает новый. При нескольких экземплярах API
// за балансировщиком хранилище нужно вынести в общую БД/Redis.
type tickets struct {
	mu   sync.Mutex
	byID map[string]*ticket
	now  func() time.Time
}

func newTickets() *tickets { return &tickets{byID: map[string]*ticket{}, now: time.Now} }

// issue выдаёт новый билет. Сам билет — 32 случайных байта в base64url.
func (s *tickets) issue(p middleware.Principal, origin string) (string, time.Time, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", time.Time{}, err
	}
	id := base64.RawURLEncoding.EncodeToString(b[:])
	now := s.now()
	t := &ticket{principal: p, origin: origin, expires: now.Add(TicketTTL)}

	s.mu.Lock()
	defer s.mu.Unlock()
	// Уборка просроченных: использованные держим до конца срока, чтобы отличать ticket_used.
	for k, v := range s.byID {
		if now.After(v.expires) {
			delete(s.byID, k)
		}
	}
	s.byID[id] = t
	return id, t.expires, nil
}

// redeem погашает билет: ok — билет действителен и теперь использован.
func (s *tickets) redeem(id string) (ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, found := s.byID[id]
	switch {
	case !found || s.now().After(t.expires):
		delete(s.byID, id)
		return ticket{}, errTicketExpired
	case t.used:
		return ticket{}, errTicketUsed
	}
	t.used = true
	return *t, nil
}
