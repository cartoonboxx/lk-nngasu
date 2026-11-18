// Package ratelimit — простое ограничение частоты запросов в памяти процесса.
//
// Окно фиксированной длины: не больше N запросов с одного IP за окно.
// Счётчики живут в памяти одного экземпляра API — для нескольких экземпляров
// за балансировщиком понадобится общее хранилище (Redis), для MVP этого хватает.
package ratelimit

import (
	"math"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
)

type window struct {
	start time.Time
	count int
}

// Limiter считает запросы по ключу (имя правила + IP клиента).
type Limiter struct {
	mu      sync.Mutex
	windows map[string]*window
	now     func() time.Time
}

// New создаёт ограничитель и фоновую очистку устаревших счётчиков.
func New() *Limiter {
	l := &Limiter{windows: map[string]*window{}, now: time.Now}
	go l.cleanup()
	return l
}

// allow учитывает запрос; при превышении возвращает, через сколько секунд можно повторить.
func (l *Limiter) allow(key string, limit int, per time.Duration) (ok bool, retryAfter int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	w := l.windows[key]
	if w == nil || now.Sub(w.start) >= per {
		l.windows[key] = &window{start: now, count: 1}
		return true, 0
	}
	if w.count >= limit {
		wait := per - now.Sub(w.start)
		return false, int(math.Ceil(wait.Seconds()))
	}
	w.count++
	return true, 0
}

// cleanup раз в минуту удаляет счётчики старше часа (больше любого окна правил).
func (l *Limiter) cleanup() {
	for range time.Tick(time.Minute) {
		l.mu.Lock()
		for k, w := range l.windows {
			if l.now().Sub(w.start) > time.Hour {
				delete(l.windows, k)
			}
		}
		l.mu.Unlock()
	}
}

// Middleware — не больше limit запросов с одного IP за per на правило name.
// При превышении — 429 rate_limited и заголовок Retry-After.
//
// identity при необходимости задаёт ключ по пользователю или адресу сокета.
// По умолчанию IP берётся через c.ClientIP(): за nginx нужно указать доверенные прокси
// (engine.SetTrustedProxies), иначе клиент подменит IP заголовком X-Forwarded-For.
func (l *Limiter) Middleware(name string, limit int, per time.Duration, identity ...func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if len(identity) > 0 {
			key = identity[0](c)
		}
		if ok, retry := l.allow(name+"|"+key, limit, per); !ok {
			httpx.Fail(c, apperr.NewRateLimited(max(retry, 1)))
			return
		}
		c.Next()
	}
}
