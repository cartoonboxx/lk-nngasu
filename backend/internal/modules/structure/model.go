// Package structure владеет справочниками и оргструктурой вуза:
// кафедры, направления, учебные группы, дисциплины, семестры, аудитории.
// Бизнес-правил здесь почти нет, поэтому сервисного слоя нет: хендлеры ходят в репозиторий.
// Появится логика — вынесите её в service.go.
package structure

// Named — запись справочника «id + название» (кафедры, дисциплины).
// Общая модель для всех таких справочников, их обслуживает catalogHandlers.
type Named struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Specialization — направление подготовки.
type Specialization struct {
	ID   int64  `json:"id"`
	Code string `json:"code"` // шифр направления по ФГОС, например «09.03.04»
	Name string `json:"name"`
}

// Group — учебная группа в том виде, в каком отдаётся наружу.
type Group struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"` // например «ПРИ-23»
	SpecializationID int64  `json:"specialization_id"`
	// SpecializationName подтягивается JOIN-ом для отображения,
	// в таблице группы не хранится.
	SpecializationName string `json:"specialization_name"`
	AdmissionYear      int    `json:"admission_year"` // год набора; курс вычисляется от него и текущего учебного года
}

// GroupInput — данные для создания и изменения группы.
// Поля должны совпадать с groupRequest в handler.go один в один:
// хендлеры конвертируют GroupInput(req) напрямую.
type GroupInput struct {
	Name             string
	SpecializationID int64
	AdmissionYear    int
}

// GroupFilter — необязательные фильтры списка групп; nil — без фильтра.
type GroupFilter struct {
	SpecializationID *int64
	AdmissionYear    *int64 // int64, а не int: так значение приходит из httpx.OptInt64
}

// Semester — семестр учебного года.
// {AcademicYearStart: 2026, Number: 1} — осенний семестр 2026/27,
// {2026, 2} — весенний семестр того же учебного года.
type Semester struct {
	ID                int64 `json:"id"`
	AcademicYearStart int   `json:"academic_year_start"` // год начала учебного года
	Number            int   `json:"number"`              // 1 — осенний, 2 — весенний
	// IsCurrent — семестр, отмеченный текущим; такой может быть только один.
	// Меняется только через SetCurrentSemester, а не через обычное обновление.
	IsCurrent bool `json:"is_current"`
}

// Room — аудитория. Номер строкой: бывают «301а», «Б-12», «спортзал».
type Room struct {
	ID     int64  `json:"id"`
	Number string `json:"number"`
}
