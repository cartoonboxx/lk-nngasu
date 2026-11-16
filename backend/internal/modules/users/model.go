// Package users владеет таблицами app_user, teacher, student.
//
// Пользователь хранится в app_user. Учебные роли независимы: у пользователя может
// быть строка в student, в teacher, в обеих или ни в одной; права администратора —
// отдельный флаг is_admin.
package users

// FileRef — аватар: id файла и адрес для скачивания (с авторизацией).
type FileRef struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

// StudentProfile — данные учебной роли студента (StudentProfile в документации).
type StudentProfile struct {
	GroupID          int64   `json:"group_id"`
	GroupName        string  `json:"group_name"`
	RecordBookNumber *string `json:"record_book_number"` // у старых записей может отсутствовать
	IsHeadman        bool    `json:"is_headman"`         // староста; не самостоятельная роль
}

// TeacherProfile — данные учебной роли преподавателя (TeacherProfile).
type TeacherProfile struct {
	DepartmentID   int64  `json:"department_id"`
	DepartmentName string `json:"department_name"`
}

// User — Profile / UserSummary из документации: профиль без документов и хеша пароля.
// Отдаётся в GET /me, в списках пользователей и преподавателю в карточке.
// student и teacher всегда присутствуют в JSON: null — роли нет.
type User struct {
	ID          int64           `json:"id"`
	LastName    string          `json:"last_name"`
	FirstName   string          `json:"first_name"`
	MiddleNames *string         `json:"middle_names"`
	FullName    string          `json:"full_name"` // «Фамилия Имя Отчество», вычисляет сервер
	Email       string          `json:"email"`
	Avatar      *FileRef        `json:"avatar"`
	IsActive    bool            `json:"is_active"`
	IsAdmin     bool            `json:"is_admin"`
	Student     *StudentProfile `json:"student"`
	Teacher     *TeacherProfile `json:"teacher"`
}

// AdminUser — полная карточка пользователя, только для администратора.
// Поля User встраиваются и в JSON идут на одном уровне с документами.
type AdminUser struct {
	User
	Passport *string `json:"passport"` // 10 цифр; у пользователей, заведённых до миграции 006, может быть null
	Snils    *string `json:"snils"`    // 11 цифр; то же
	// CredentialState — "active": пароль задан; "invited": ждём активации по приглашению.
	CredentialState string `json:"credential_state"`
}

// StudentRole — роль студента при создании или замене (StudentWrite).
// Названия группы здесь нет: их вычисляет сервер.
type StudentRole struct {
	GroupID          int64
	RecordBookNumber *string
	IsHeadman        bool
}

// TeacherRole — роль преподавателя при создании или замене (TeacherWrite).
type TeacherRole struct {
	DepartmentID int64
}

// CreateUser — создание пользователя (UserCreateInput) или одобрение заявки.
// Пароля нет: администратор создаёт учётку без пароля, пользователь задаёт его
// по приглашению. При одобрении заявки хеш пароля берётся из заявки.
type CreateUser struct {
	LastName, FirstName string
	MiddleNames         *string
	Email               string
	Passport, Snils     string // уже нормализованные цифры
	Student             *StudentRole
	Teacher             *TeacherRole
	AvatarFileID        *int64

	// CreatedBy — администратор, который создаёт пользователя (для приглашения).
	CreatedBy int64
}

// UpdateProfile — изменения своего профиля (ProfilePatchInput).
// Учебные роли и документы меняет только администратор.
type UpdateProfile struct {
	LastName, FirstName Field[string] // null запрещён
	MiddleNames         Field[string] // null очищает отчество
	AvatarFileID        Field[int64]  // null убирает аватар
}

// AdminUpdate — частичная правка администратором (UserPatchInput).
// Field различает «ключа нет» (не трогать) и «передан null».
type AdminUpdate struct {
	LastName, FirstName Field[string]      // null запрещён
	MiddleNames         Field[string]      // null очищает отчество
	Email               Field[string]      // null запрещён
	Passport, Snils     Field[string]      // null запрещён
	Student             Field[StudentRole] // null снимает роль, объект — задаёт или заменяет
	Teacher             Field[TeacherRole] // то же
	AvatarFileID        Field[int64]       // null убирает аватар
	IsActive            Field[bool]        // блокировка / восстановление
	// is_admin и пароль здесь не меняются: права администратора назначаются
	// отдельной процедурой вне панели (cmd/createadmin).
}

// CourseStudent — строка состава группы или курса: только то, что нужно
// для учебной работы (без email, документов и полной карточки).
type CourseStudent struct {
	StudentID        int64   `json:"student_id"`
	FullName         string  `json:"full_name"`
	RecordBookNumber *string `json:"record_book_number"`
	IsHeadman        bool    `json:"is_headman"`
}

// InvitationReceipt — ответ на повторную отправку приглашения.
type InvitationReceipt struct {
	UserID         int64  `json:"user_id"`
	DeliveryStatus string `json:"delivery_status"` // "queued"
}

// ListFilter — параметры выборки списка пользователей.
// Пустые/нулевые значения означают «без фильтра».
//
// Все фильтры объединяются через AND. «Староста» во фронте — это
// role=student&is_headman=true: староста — признак студента, а не роль.
type ListFilter struct {
	Role         string // student | teacher | "" (все); совмещение ролей допустимо
	IsHeadman    *bool  // true — только старосты, false — студенты без признака
	IsActive     *bool  // активен / заблокирован
	IsAdmin      *bool  // администраторы / не администраторы
	GroupID      *int64 // только студенты этой группы
	DepartmentID *int64 // только преподаватели этой кафедры
	Q            string // подстрока ФИО или email, без учёта регистра
	Desc         bool   // sort=-full_name
	Limit        int    // размер страницы, 1–200
	Offset       int    // сколько записей пропустить
}
