// Package auth — вход, сессии и активация учётных записей (раздел «Авторизация» документации API).
//
// Как устроена сессия:
//   - POST /auth/login проверяет email и пароль, создаёт сессию на 1 или 30 суток
//     (remember_me) и выдаёт короткий access-токен (JWT) в JSON и refresh-токен
//     в HttpOnly-cookie. JS cookie не видит, поэтому украсть её через XSS нельзя.
//   - Access-токен передаётся в Authorization: Bearer и живёт недолго (ACCESS_TTL).
//     Когда он истёк (401 token_expired), фронт вызывает POST /auth/refresh:
//     по cookie выдаются новый access-токен и новый refresh-токен (ротация).
//   - Refresh и logout защищены от CSRF: заголовок X-CSRF-Token (GET /auth/csrf)
//     и проверка Origin. Сам cookie браузер отправляет автоматически, поэтому
//     одного cookie для таких запросов недостаточно.
//   - Повторное использование старого refresh-токена означает, что его украли:
//     вся сессия отзывается (401 refresh_reused).
//   - Access-токен содержит id сессии; middleware проверяет сессию по БД на каждом
//     запросе, поэтому logout и блокировка действуют сразу.
//
// Приглашения: администратор создаёт пользователя без пароля, тот получает письмо
// со ссылкой и задаёт пароль через POST /auth/activate.
package auth

import "time"

// Token — ответ входа и обновления сессии ({"data": Token}).
// Роли в ответе нет: права определяются по независимым student, teacher и is_admin в user.
type Token struct {
	AccessToken string    `json:"access_token"` // JWT для заголовка Authorization: Bearer
	TokenType   string    `json:"token_type"`   // всегда "Bearer"
	ExpiresAt   time.Time `json:"expires_at"`   // когда истечёт access-токен
	User        UserInfo  `json:"user"`
}

// UserInfo — кто вошёл. student, teacher и is_admin независимы: у пользователя может
// быть несколько ролей сразу, и фронт не должен выбирать права по одной из них.
type UserInfo struct {
	ID          int64        `json:"id"`
	LastName    string       `json:"last_name"`
	FirstName   string       `json:"first_name"`
	MiddleNames *string      `json:"middle_names"`
	FullName    string       `json:"full_name"` // "Фамилия Имя Отчество"
	Email       string       `json:"email"`
	Avatar      *FileRef     `json:"avatar"` // null — аватара нет
	IsActive    bool         `json:"is_active"`
	IsAdmin     bool         `json:"is_admin"`
	Student     *StudentInfo `json:"student"` // null — не студент
	Teacher     *TeacherInfo `json:"teacher"` // null — не преподаватель
}

// FileRef — файл аватара: id и адрес скачивания через модуль files.
type FileRef struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

// StudentInfo — данные студента в профиле.
type StudentInfo struct {
	GroupID          int64   `json:"group_id"`
	GroupName        string  `json:"group_name"`
	RecordBookNumber *string `json:"record_book_number"`
	IsHeadman        bool    `json:"is_headman"` // староста группы
}

// TeacherInfo — данные преподавателя в профиле.
type TeacherInfo struct {
	DepartmentID   int64  `json:"department_id"`
	DepartmentName string `json:"department_name"`
}

// CsrfToken — ответ GET /auth/csrf.
type CsrfToken struct {
	CsrfToken string `json:"csrf_token"` // передавать в заголовке X-CSRF-Token
}

// session — refresh-токен для cookie и срок его жизни; в JSON не попадает.
type session struct {
	refreshToken string
	expiresAt    time.Time
}
