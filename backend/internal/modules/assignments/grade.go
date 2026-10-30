package assignments

import (
	"math"
	"strings"

	"university-api/internal/platform/apperr"
)

// MaxGrade — высшая оценка. Шкала: от 0.0 до 5.0 с шагом 0.1 (numeric(2,1) в БД).
const MaxGrade = 5.0

// ValidGrade проверяет диапазон и шаг 0.1 (с поправкой на погрешность float).
func ValidGrade(g float64) bool {
	if math.IsNaN(g) || g < 0 || g > MaxGrade {
		return false
	}
	return math.Abs(g*10-math.Round(g*10)) < 1e-9
}

func validateReview(in *ReviewInput) error {
	if !ValidGrade(in.Grade) {
		return apperr.NewInvalid("bad_grade", "grade must be between 0.0 and 5.0 with step 0.1")
	}
	if in.Comment != nil {
		c := strings.TrimSpace(*in.Comment)
		if c == "" {
			in.Comment = nil
		} else {
			in.Comment = &c
		}
	}
	return nil
}
