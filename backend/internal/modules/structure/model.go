// Package structure владеет справочниками и оргструктурой вуза:
// кафедры, дисциплины, направления, учебные группы, семестры, аудитории.
//
// Архивного состояния нет (контракт 1.2, scheme.dbml): запись можно создать, изменить
// и физически удалить, если на неё ничего не ссылается (иначе 409 catalog_in_use).
package structure

// ListQuery — общие параметры списков справочников.
type ListQuery struct {
	Q             string
	Limit, Offset int
}

// Named — NamedCatalog: кафедра или дисциплина.
type Named struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Specialization — направление подготовки.
type Specialization struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"` // шифр по ФГОС, например «09.03.02»
}

// Group — учебная группа. Название и код направления подтягиваются JOIN-ом.
type Group struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"` // например «24-ИСТ-1»
	SpecializationID   int64  `json:"specialization_id"`
	SpecializationName string `json:"specialization_name"`
	SpecializationCode string `json:"specialization_code"`
	AdmissionYear      int    `json:"admission_year"` // год набора, 2000–2100
}

// GroupInput — данные для создания и полной замены группы (GroupInput).
type GroupInput struct {
	Name             string
	SpecializationID int64
	AdmissionYear    int
}

// GroupFilter — необязательные фильтры списка групп; nil — без фильтра.
type GroupFilter struct {
	ListQuery
	SpecializationID *int64
	AdmissionYear    *int64
}

// Semester — семестр учебного года (не номер семестра обучения студента).
// {AcademicYearStart: 2026, Number: 1} — осенний семестр 2026/27.
type Semester struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"` // вычисляется: «2026/27 · Осенний»
	AcademicYearStart int    `json:"academic_year_start"`
	Number            int    `json:"number"` // 1 — осень, 2 — весна
	// IsCurrent — не более одного текущего семестра; меняется отдельной командой.
	IsCurrent bool `json:"is_current"`
}

// SemesterInput — создание и полная замена семестра; название вычисляет БД,
// флаг текущего меняется отдельной командой.
type SemesterInput struct {
	AcademicYearStart int
	Number            int
}

// Room — аудитория; обозначение хранится в number (до 30 символов).
type Room struct {
	ID     int64  `json:"id"`
	Number string `json:"number"` // например «А-302»
}
