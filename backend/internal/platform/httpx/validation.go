package httpx

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"university-api/internal/platform/apperr"
)

// Ошибки разбора тела запроса отдаются по документации API:
//
//	400 invalid_request, details.fields = {"имя_поля": ["описание", ...]}
//
// Имена полей — как в JSON (email, password), а не как в Go-структуре.

func init() {
	// Контракт: лишние поля в теле отклонять 400, а не молча игнорировать.
	binding.EnableDecoderDisallowUnknownFields = true

	// Валидатор gin по умолчанию называет поля по Go-структуре (Email);
	// просим его брать имя из json-тега (email).
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "" {
				return f.Name
			}
			return name
		})
	}
}

// bindError переводит ошибку ShouldBindJSON в 400 invalid_request с details.fields там,
// где можно указать конкретное поле.
func bindError(err error) *apperr.Error {
	base := apperr.NewInvalid("invalid_request", "request body is invalid")

	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := map[string][]string{}
		for _, fe := range verrs {
			// Namespace = "decisionBody.user.last_name"; без имени корневой структуры
			// получается путь в JSON: "user.last_name".
			name := fe.Namespace()
			if i := strings.Index(name, "."); i >= 0 {
				name = name[i+1:]
			}
			// Поля PATCH-запросов обёрнуты в Field{Set, Value}: убираем служебный
			// сегмент, чтобы получилось "student.group_id", а не "student.Value.group_id".
			name = strings.ReplaceAll(name, ".Value", "")
			fields[name] = append(fields[name], describe(fe))
		}
		return base.WithDetails(map[string]any{"fields": fields})
	}

	// Значение не того типа: строка вместо числа и т.п.
	var terr *json.UnmarshalTypeError
	if errors.As(err, &terr) && terr.Field != "" {
		return base.WithDetails(map[string]any{
			"fields": map[string][]string{terr.Field: {"must be " + terr.Type.String()}},
		})
	}

	// Неизвестный ключ: encoding/json пишет `json: unknown field "x"`.
	if msg := err.Error(); strings.HasPrefix(msg, `json: unknown field "`) {
		name := strings.TrimSuffix(strings.TrimPrefix(msg, `json: unknown field "`), `"`)
		return base.WithDetails(map[string]any{
			"fields": map[string][]string{name: {"unknown field"}},
		})
	}

	// Пустое тело или неразбираемый JSON — без подробностей по полям.
	return apperr.NewInvalid("invalid_request", "request body must be valid JSON")
}

// describe — короткое описание нарушенного правила для details.fields.
func describe(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "required"
	case "email":
		return "must be a valid email"
	// Для строк min/max — длина, для чисел — значение.
	case "min":
		if isNumber(fe.Kind()) {
			return "must be at least " + fe.Param()
		}
		return "must be at least " + fe.Param() + " characters"
	case "max":
		if isNumber(fe.Kind()) {
			return "must be at most " + fe.Param()
		}
		return "must be at most " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	case "gt":
		return "must be greater than " + fe.Param()
	default:
		return "invalid value"
	}
}

// isNumber — числовое поле (для текста ошибок min/max).
func isNumber(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}
