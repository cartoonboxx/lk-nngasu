// Package structure владеет справочниками и оргструктурой вуза:
// кафедры, дисциплины, направления, учебные группы, семестры, аудитории.
//
// Все справочники можно архивировать: архивная запись остаётся в списках
// (с archived = true), на неё продолжают ссылаться старые данные, но новые
// связи (группа → направление, студент → группа, курс → дисциплина…) с ней
// создать нельзя. Восстановления из архива пока нет.
package structure

// ListQuery — общие параметры списков справочников.
// Archived: false — только действующие (по умолчанию), true — только архив,
// nil — оба состояния (include_archived=true).
type ListQuery struct {
	Q             string
	Archived      *bool
	Limit, Offset int
}

// Named — NamedCatalog: кафедра или дисциплина.
type Named struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Archived bool   `json:"archived"`
}

// Specialization — направление подготовки.
type Specialization struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"` // шифр по ФГОС, например «09.03.02»
	Archived bool   `json:"archived"`
}

// Group — учебная группа. Название и код направления подтягиваются JOIN-ом.
type Group struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"` // например «24-ИСТ-1»
	Archived           bool   `json:"archived"`
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
	Name              string `json:"name"` // например «2026/27 · Осенний»
	Archived          bool   `json:"archived"`
	AcademicYearStart int    `json:"academic_year_start"`
	Number            int    `json:"number"` // 1 — осень, 2 — весна
	// IsCurrent — не более одного текущего семестра; меняется отдельной командой.
	IsCurrent bool `json:"is_current"`
}

// SemesterInput — создание и полная замена семестра; флаг текущего здесь не меняется.
type SemesterInput struct {
	Name              string
	AcademicYearStart int
	Number            int
}

// Room — аудитория; name — обозначение, которое видят в справочнике и расписании.
type Room struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"` // например «Корпус А, ауд. 302»
	Archived bool   `json:"archived"`
}
