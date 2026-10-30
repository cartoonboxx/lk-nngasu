// Package structure владеет справочниками и оргструктурой вуза:
// кафедры, направления, учебные группы, дисциплины, семестры, аудитории.
// Бизнес-правил здесь почти нет, поэтому сервисного слоя нет: хендлеры ходят в репозиторий.
// Появится логика — вынесите её в service.go.
package structure

// Named — запись справочника «id + название» (кафедры, дисциплины).
type Named struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Specialization struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Group struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	SpecializationID   int64  `json:"specialization_id"`
	SpecializationName string `json:"specialization_name"`
	AdmissionYear      int    `json:"admission_year"`
}

type GroupInput struct {
	Name             string
	SpecializationID int64
	AdmissionYear    int
}

type GroupFilter struct {
	SpecializationID *int64
	AdmissionYear    *int64
}

type Semester struct {
	ID                int64 `json:"id"`
	AcademicYearStart int   `json:"academic_year_start"`
	Number            int   `json:"number"`
	IsCurrent         bool  `json:"is_current"`
}

type Room struct {
	ID     int64  `json:"id"`
	Number string `json:"number"`
}
