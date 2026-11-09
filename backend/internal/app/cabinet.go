package app

import (
	"context"

	"github.com/gin-gonic/gin"

	"university-api/internal/modules/assignments"
	"university-api/internal/modules/content"
	"university-api/internal/modules/courses"
	"university-api/internal/modules/schedule"
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

// ===== Главная: GET /dashboard =====

// newsItem — элемент ленты новостей. Пока заглушка: лента всегда пустая.
type newsItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// dashboardPage — ответ главной страницы.
type dashboardPage struct {
	Week    schedule.WeekInfo `json:"week"`    // сегодняшняя дата и верхняя/нижняя неделя
	Lessons []schedule.Lesson `json:"lessons"` // пары на сегодня: студенту — группы, преподавателю — свои
	News    []newsItem        `json:"news"`    // заглушка: []
}

// dashboard — главная: пары на сегодня и лента новостей.
// Администратору пар не положено — придёт пустой список (см. schedule.Service.Today).
func dashboard(ss *schedule.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		lessons, err := ss.Today(c.Request.Context(), middleware.Me(c))
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, dashboardPage{Week: ss.Week(nil), Lessons: lessons, News: []newsItem{}})
	}
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
func registerCabinet(protected *gin.RouterGroup, ss *schedule.Service, cs *courses.Service, as *assignments.Service) {
	student := middleware.RequireRole(middleware.RoleStudent)

	protected.GET("/dashboard", dashboard(ss))        // главная
	protected.GET("/grades", student, grades(cs, as)) // оценки студента
}
