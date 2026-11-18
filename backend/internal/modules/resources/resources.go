// Package resources — страница «Ресурсы»: ссылки на внешние сервисы вуза и контакты.
//
// Данные — серверная конфигурация, а не таблица: редактора в интерфейсе нет, CRUD не нужен.
// По умолчанию берётся встроенный default.json; свой набор задаётся файлом RESOURCES_FILE
// в том же формате. Файл читается и проверяется при старте: ошибка в нём не даёт серверу запуститься.
package resources

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
)

//go:embed default.json
var defaultJSON []byte

// Service — внешний сервис (Moodle, VK, MAX…). icon_key — ключ иконки во фронтенде,
// не SVG и не имя Vue-компонента.
type Service struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	URL         string `json:"url"`
	LinkText    string `json:"link_text"`
	IconKey     string `json:"icon_key"`
}

// Resources — ответ GET /resources. Контакты отдаются как записаны в конфигурации
// (схема ResourceContact в документации пока не прислана).
type Resources struct {
	Services []Service        `json:"services"`
	Contacts []map[string]any `json:"contacts"`
}

var key = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// Load читает конфигурацию: path пустой — встроенный набор.
func Load(path string) (Resources, error) {
	raw := defaultJSON
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return Resources{}, fmt.Errorf("RESOURCES_FILE: %w", err)
		}
		raw = b
	}
	var r Resources
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		return Resources{}, fmt.Errorf("resources: %w", err)
	}
	seen := map[string]bool{}
	for i, s := range r.Services {
		u, err := url.Parse(s.URL)
		switch {
		case !key.MatchString(s.ID) || seen[s.ID]:
			return Resources{}, fmt.Errorf("resources: services[%d].id must be a unique lowercase key", i)
		case s.Label == "" || s.LinkText == "":
			return Resources{}, fmt.Errorf("resources: services[%d] needs label and link_text", i)
		case err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "":
			return Resources{}, fmt.Errorf("resources: services[%d].url must be an absolute http(s) URL", i)
		case !key.MatchString(s.IconKey):
			return Resources{}, fmt.Errorf("resources: services[%d].icon_key must be a lowercase key", i)
		}
		seen[s.ID] = true
	}
	if r.Services == nil {
		r.Services = []Service{}
	}
	if r.Contacts == nil {
		r.Contacts = []map[string]any{}
	}
	return r, nil
}

// Register — GET /resources для авторизованных пользователей.
func Register(protected *gin.RouterGroup, res Resources) {
	protected.GET("/resources", func(c *gin.Context) { httpx.OK(c, res) })
}
