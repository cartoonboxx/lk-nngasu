package profile

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register собирает зависимости модуля profile и регистрирует его маршруты в группе r.
//
// r          — группа С проверкой JWT (в отличие от auth.Register, где логин публичный):
//
//	id пользователя берётся из токена, без токена эндпоинт не имеет смысла;
//
// pool       — общий пул подключений к Postgres, создаётся один раз в main;
// avatarBase — префикс ссылки на файл аватара, например "/api/v1/files/".
//
//	Аватар хранится в модуле files, профиль отдаёт только ссылку на него.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, avatarBase string) {
	// Сборка цепочки: репозиторий (работа с БД) → сервис (сборка профиля) → handler (HTTP).
	h := &Handler{svc: NewService(newRepository(pool), avatarBase)}

	// GET <префикс группы>/auth/me → Handler.Me — профиль текущего пользователя.
	r.GET("/user", h.Me)    // данные текущего пользователя для ЛК
	r.GET("/auth/me", h.Me) // старый путь; удалить, когда фронт перейдёт на /user
}
