// Package users владеет таблицами app_user, teacher, student.
package users

type User struct {
	ID               int64   `json:"id"`
	LastName         string  `json:"last_name"`
	FirstName        string  `json:"first_name"`
	MiddleNames      *string `json:"middle_names"`
	FullName         string  `json:"full_name"`
	Email            string  `json:"email"`
	AvatarFileID     *int64  `json:"avatar_file_id"`
	IsActive         bool    `json:"is_active"`
	Role             string  `json:"role"` // admin | teacher | student
	GroupID          *int64  `json:"group_id,omitempty"`
	GroupName        *string `json:"group_name,omitempty"`
	RecordBookNumber *string `json:"record_book_number,omitempty"`
	IsHeadman        *bool   `json:"is_headman,omitempty"`
	DepartmentID     *int64  `json:"department_id,omitempty"`
	DepartmentName   *string `json:"department_name,omitempty"`
}

type CreateUser struct {
	Role                  string
	LastName, FirstName   string
	MiddleNames           *string
	Email, Password       string
	GroupID, DepartmentID *int64
	RecordBookNumber      *string
	IsHeadman             bool
}

// UpdateProfile — что пользователь может менять о себе.
type UpdateProfile struct {
	LastName, FirstName, MiddleNames *string
	AvatarFileID                     *int64
}

// AdminUpdate — что может менять администратор; nil означает «не трогать».
type AdminUpdate struct {
	LastName, FirstName, MiddleNames *string
	Email, Password                  *string
	IsActive, IsAdmin                *bool
	GroupID, DepartmentID            *int64
	IsHeadman                        *bool
	RecordBookNumber                 *string
}

type ListFilter struct {
	Role         string // student | teacher | admin | "" (все)
	GroupID      *int64
	DepartmentID *int64
	Q            string
	Limit        int
	Offset       int
}
