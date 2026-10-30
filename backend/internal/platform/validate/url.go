package validate

import "net/url"

// HTTPURL разрешает только абсолютные http/https-ссылки: ссылка из БД попадёт на фронтенд
// как href, и схемы вроде javascript: там недопустимы.
func HTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
