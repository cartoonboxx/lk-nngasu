// Package users владеет таблицами app_user, teacher, student.
//
// Пользователь хранится в app_user; если он преподаватель или студент,
// у него есть ещё строка в teacher или student с полями своей роли.
package users

// User — пользователь в том виде, в каком он отдаётся наружу (API, ответы списков).
// Собирается из app_user и, в зависимости от роли, из student или teacher.
// Поля с omitempty заполняются только для своей роли, у остальных они nil
// и в JSON не попадают.
type User struct {
	ID          int64   `json:"id"`
	LastName    string  `json:"last_name"`
	FirstName   string  `json:"first_name"`
	MiddleNames *string `json:"middle_names"` // отчество (может отсутствовать)
	// FullName — «Фамилия Имя Отчество», собирается из частей ФИО
	// (в БД или при чтении); напрямую не редактируется.
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	AvatarFileID *int64 `json:"avatar_file_id"` // ссылка на загруженный файл аватара; nil — аватара нет
	// IsActive — false для заблокированных/отключённых учёток: вход запрещён,
	// но запись и связанные данные сохраняются.
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"` // admin | teacher | student

	// Только для студентов (строка в student).
	GroupID          *int64  `json:"group_id,omitempty"`
	GroupName        *string `json:"group_name,omitempty"`         // название группы, подтягивается JOIN-ом для отображения
	RecordBookNumber *string `json:"record_book_number,omitempty"` // номер зачётной книжки
	IsHeadman        *bool   `json:"is_headman,omitempty"`         // староста группы

	// Только для преподавателей (строка в teacher).
	DepartmentID   *int64  `json:"department_id,omitempty"`
	DepartmentName *string `json:"department_name,omitempty"` // название кафедры, подтягивается JOIN-ом
}

// CreateUser — данные для создания пользователя (делает администратор).
// В зависимости от Role создаётся ещё строка в student или teacher.
type CreateUser struct {
	Role                string // admin | teacher | student
	LastName, FirstName string
	MiddleNames         *string
	Email               string
	Password            string // открытый пароль; хешируется перед записью, в БД не хранится

	// GroupID обязателен для студента, DepartmentID — для преподавателя;
	// для других ролей игнорируются.
	GroupID, DepartmentID *int64

	// Только для студента.
	RecordBookNumber *string
	IsHeadman        bool
}

// UpdateProfile — что пользователь может менять о себе сам.
// nil означает «не трогать». Email, пароль, роль и привязки к группе/кафедре
// здесь намеренно отсутствуют — их меняет только администратор.
type UpdateProfile struct {
	LastName, FirstName, MiddleNames *string
	AvatarFileID                     *int64
}

// AdminUpdate — что может менять администратор; nil означает «не трогать».
type AdminUpdate struct {
	LastName, FirstName, MiddleNames *string
	Email                            *string
	Password                         *string // новый открытый пароль; хешируется перед записью
	IsActive                         *bool   // блокировка/разблокировка учётки
	IsAdmin                          *bool   // выдать/снять права администратора

	// Применяются только если у пользователя соответствующая роль:
	// GroupID, IsHeadman, RecordBookNumber — студенту, DepartmentID — преподавателю.
	GroupID, DepartmentID *int64
	IsHeadman             *bool
	RecordBookNumber      *string
}

// ListFilter — параметры выборки списка пользователей.
// Пустые/нулевые значения означают «без фильтра».
type ListFilter struct {
	Role         string // student | teacher | admin | "" (все)
	GroupID      *int64 // только студенты этой группы
	DepartmentID *int64 // только преподаватели этой кафедры
	Q            string // поиск по подстроке в ФИО и email
	Limit        int    // размер страницы
	Offset       int    // сколько записей пропустить (пагинация)
}
