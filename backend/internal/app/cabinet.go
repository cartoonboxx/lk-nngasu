package app

import (
	"context"

	"github.com/gin-gonic/gin"

	"university-api/internal/modules/assignments"
	"university-api/internal/modules/chat"
	"university-api/internal/modules/content"
	"university-api/internal/modules/courses"
	"university-api/internal/modules/news"
	"university-api/internal/modules/schedule"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// Страницы личного кабинета, которым нужны данные сразу нескольких модулей.
// Модули друг друга не импортируют, поэтому такие ответы собираются здесь, в app, —
// единственном месте, которое знает обо всех модулях. Своих SQL-запросов здесь нет:
// всё берётся из сервисов модулей, с их же проверками доступа.

// ===== Страница предмета: GET /courses/:id =====

// topicPage — TopicDetail: тема курса с материалами (из content) и её заданием
// (из assignments; не более одного на тему) или null.
// content.Topic встроена: её поля попадают в JSON на верхний уровень темы.
type topicPage struct {
	content.Topic
	Assignment *assignments.Assignment `json:"assignment"`
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

		// Раскладываем задания по темам: связь тема — задание один к одному (миграция 012).
		byTopic := make(map[int64]*assignments.Assignment, len(topics))
		for i := range list {
			byTopic[list[i].TopicID] = &list[i]
		}

		out := make([]topicPage, 0, len(topics))
		for _, t := range topics {
			out = append(out, topicPage{Topic: t, Assignment: byTopic[t.ID]})
		}
		return out, nil
	}
}

// ===== Главная страница: GET /dashboard =====

const (
	dashboardChats = 5 // последние чаты
	dashboardNews  = 3 // свежие новости
)

// dashboard — ответ GET /dashboard. Дата — серверная, Europe/Moscow.
// week = null, если не задан WEEK_ONE_DATE; lessons — пары пользователя на сегодня
// (у администратора и без текущего семестра — []). Карточки чатов клиент дальше
// обновляет по событиям WebSocket; счётчик непрочитанных учитывает chat_read.
type dashboard struct {
	Week    *schedule.WeekInfo `json:"week"`
	Lessons []schedule.Lesson  `json:"lessons"`
	Chats   []chat.Chat        `json:"chats"`
	News    []news.NewsSummary `json:"news"`
}

func registerDashboard(protected *gin.RouterGroup, sch *schedule.Service, ch *chat.Service, feed *news.Feed) {
	protected.GET("/dashboard", func(c *gin.Context) {
		ctx, p := c.Request.Context(), middleware.Me(c)
		var d dashboard
		var err error
		if d.Week, d.Lessons, err = sch.Today(ctx, p); err != nil {
			httpx.Fail(c, err)
			return
		}
		if d.Chats, _, err = ch.List(ctx, p, chat.ListQuery{Limit: dashboardChats}); err != nil {
			httpx.Fail(c, err)
			return
		}
		if d.News, err = feed.Latest(ctx, dashboardNews); err != nil {
			httpx.Fail(c, err)
			return
		}
		if d.Chats == nil {
			d.Chats = []chat.Chat{}
		}
		if d.News == nil {
			d.News = []news.NewsSummary{}
		}
		httpx.OK(c, d)
	})
}
