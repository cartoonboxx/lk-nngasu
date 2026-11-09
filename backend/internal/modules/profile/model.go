// Package profile — профиль текущего пользователя для личного кабинета (GET /auth/me).
//
// Данные берутся одним запросом из app_user и связанных таблиц: для студента —
// группа и направление подготовки, для преподавателя — кафедра.
// Имена полей в JSON — camelCase по контракту с фронтом.
package profile

// Profile — ответ GET /auth/me.
//
// Student заполнен только у студента, Teacher — только у преподавателя;
// у администратора оба null (если админ не числится студентом или преподавателем).
// Указатели в JSON превращаются в null, поэтому фронт всегда получает все ключи.
type Profile struct {
	ID         int64    `json:"id"`
	Role       string   `json:"role"` // student | teacher | admin — так же, как в токене
	FirstName  string   `json:"firstName"`
	LastName   string   `json:"lastName"`
	MiddleName *string  `json:"middleName"` // nil — отчества нет
	Email      string   `json:"email"`
	AvatarURL  *string  `json:"avatarUrl"` // ссылка на скачивание через модуль files; nil — аватара нет
	IsActive   bool     `json:"isActive"`  // в успешном ответе всегда true: неактивным отвечаем 401
	Student    *Student `json:"student"`   // nil — пользователь не студент
	Teacher    *Teacher `json:"teacher"`   // nil — пользователь не преподаватель
}

// Student — данные студента.
type Student struct {
	RecordBookNumber string `json:"recordBookNumber"` // номер зачётной книжки
	IsHeadman        bool   `json:"isHeadman"`        // староста группы
	Group            *Group `json:"group"`            // nil — студент пока не распределён в группу
}

// Group — учебная группа студента.
type Group struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`           // например "ИС-24-1"
	AdmissionYear  int             `json:"admissionYear"`  // год набора
	Specialization *Specialization `json:"specialization"` // nil — направление не указано
}

// Specialization — направление подготовки (код по ФГОС и название).
type Specialization struct {
	ID   int64  `json:"id"`
	Code string `json:"code"` // например "09.03.02"
	Name string `json:"name"`
}

// Teacher — данные преподавателя.
type Teacher struct {
	Department *Department `json:"department"` // nil — кафедра не указана
}

// Department — кафедра.
type Department struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
