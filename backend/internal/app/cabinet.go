package app

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"university-api/internal/modules/assignments"
	"university-api/internal/modules/content"
	"university-api/internal/modules/courses"
	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// Страницы личного кабинета, которым нужны данные сразу нескольких модулей.
// Модули друг друга не импортируют, поэтому такие ответы собираются здесь, в app, —
// единственном месте, которое знает обо всех модулях. Своих SQL-запросов здесь нет:
// всё берётся из сервисов модулей, с их же проверками доступа.

// ===== Страница предмета: GET /courses/:id =====

// topicPage — тема курса с материалами (из content) и заданиями (из assignments).
// content.Topic встроена: её поля попадают в JSON на верхний уровень темы.
type topicPage struct {
	content.Topic
	Assignments []assignments.Assignment `json:"assignments"`
}

// courseTopics возвращает загрузчик тем для courses: темы с материалами + задания,
// разложенные по темам. Студент видит только опубликованные задания и свой статус (mine),
// преподаватель — все задания со статистикой (stats): это решает assignments.
func courseTopics(ct *content.Service, as *assignments.Service) courses.TopicsLoader {
	return func(ctx context.Context, p middleware.Principal, courseID int64) (any, error) {
		topics, err := ct.ListTopics(ctx, p, courseID)
		if err != nil {
			return nil, err
		}
		list, err := as.ListAssignments(ctx, p, courseID)
		if err != nil {
			return nil, err
		}

		// Раскладываем задания по темам; порядок внутри темы — как вернул assignments.
		byTopic := make(map[int64][]assignments.Assignment, len(topics))
		for _, a := range list {
			byTopic[a.TopicID] = append(byTopic[a.TopicID], a)
		}

		out := make([]topicPage, 0, len(topics))
		for _, t := range topics {
			items := byTopic[t.ID]
			if items == nil {
				items = []assignments.Assignment{} // в JSON — [], а не null
			}
			out = append(out, topicPage{Topic: t, Assignments: items})
		}
		return out, nil
	}
}

// ===== Новости: GET /news?limit= =====

// Таблицы новостей в БД пока нет, поэтому маршрут — заглушка: всегда пустой список.
// Формат и параметр limit уже рабочие, чтобы фронт мог подключить блок новостей сейчас,
// а при появлении таблицы news поменялся только источник данных.

const (
	newsDefaultLimit = 10 // без ?limit=
	newsMaxLimit     = 50 // больше за раз не отдаём
)

// newsItem — новость в ленте. Поля предварительные: уточнить с фронтом при создании таблицы.
type newsItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Text        string    `json:"text"`
	PublishedAt time.Time `json:"published_at"`
}

// news — последние новости, не больше limit штук (1..50, по умолчанию 10).
func news(c *gin.Context) {
	limit, ok := httpx.OptInt64(c, "limit")
	if !ok {
		return
	}
	n := int64(newsDefaultLimit)
	if limit != nil {
		if *limit < 1 || *limit > newsMaxLimit {
			httpx.Fail(c, apperr.NewInvalid("invalid_limit", "limit must be between 1 and 50"))
			return
		}
		n = *limit
	}

	items := []newsItem{} // заглушка: новостей пока нет
	if int64(len(items)) > n {
		items = items[:n]
	}
	httpx.OK(c, items)
}

// ===== Оценки: GET /grades =====

// subjectGrades — задания одного предмета со статусом и оценкой студента (поле mine).
type subjectGrades struct {
	courses.Course
	Assignments []assignments.Assignment `json:"assignments"`
}

// grades — все задания по предметам текущего семестра с текущим статусом сдачи и оценкой.
// В отличие от /me/grades (только выставленные оценки), здесь есть и несданные задания.
//
// Задания грузятся по курсу за раз: курсов в семестре около десятка,
// а так статусы (в т.ч. missing — просрочено) считает сам assignments, без копии логики.
func grades(cs *courses.Service, as *assignments.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, p := c.Request.Context(), middleware.Me(c)

		// Для студента List сам ограничит курсы его группой.
		list, err := cs.List(ctx, p, courses.CourseFilter{CurrentSemester: true})
		if err != nil {
			httpx.Fail(c, err)
			return
		}

		out := make([]subjectGrades, 0, len(list))
		for _, co := range list {
			items, err := as.ListAssignments(ctx, p, co.ID)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
			if items == nil {
				items = []assignments.Assignment{}
			}
			out = append(out, subjectGrades{Course: co, Assignments: items})
		}
		httpx.OK(c, out)
	}
}

// registerCabinet регистрирует сборные страницы ЛК в группе с проверкой JWT.
// Главная собирается на фронте из отдельных запросов: /user, /schedule, /news, /chats.
func registerCabinet(protected *gin.RouterGroup, cs *courses.Service, as *assignments.Service) {
	student := middleware.RequireRole(middleware.RoleStudent)

	protected.GET("/news", news)                      // лента новостей (пока пустая)
	protected.GET("/grades", student, grades(cs, as)) // оценки студента
}
