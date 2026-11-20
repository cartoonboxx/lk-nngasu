package chat

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// Guard — проверка, что файл принадлежит пользователю (реализует access.Checker).
type Guard interface {
	OwnsFile(ctx context.Context, userID, fileID int64) error
}

type Service struct {
	repo  *Repository
	guard Guard
	// publish — рассылка события WebSocket пользователям (realtime.Hub.Publish); nil — без событий.
	publish func(userIDs []int64, typ string, data any)
}

func NewService(repo *Repository, guard Guard) *Service {
	return &Service{repo: repo, guard: guard}
}

var (
	errChatNotFound    = apperr.NewNotFound("chat_not_found")
	errSelfChat        = apperr.NewInvalid("self_chat", "cannot start a chat with yourself")
	errUserNotFound    = apperr.NewNotFound("user_not_found")
	errGroupForbidden  = apperr.NewForbidden("group_forbidden", "you are not a member of this group chat")
	errContactNotAllow = apperr.NewForbidden("contact_not_allowed", "you cannot start a chat with this user")
	errChatForbidden   = apperr.NewForbidden("chat_forbidden", "you are not a participant of this chat")
	errFileNotFound    = apperr.NewNotFound("file_not_found")
)

// payloadErr — invalid_payload с полем (формат details.fields, как в HTTP).
func payloadErr(field, problem string) error {
	return apperr.NewInvalid("invalid_payload", "Некорректный кадр.").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// pair упорядочивает пару пользователей: в БД личный чат хранится как (меньший, больший).
func pair(a, b int64) (low, high int64) {
	if a < b {
		return a, b
	}
	return b, a
}

// normalizeBody обрезает пробелы и проверяет длину. Пустой текст допустим только с вложениями.
func normalizeBody(body string, hasFiles bool) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" && !hasFiles {
		return "", payloadErr("body", "must not be empty without files")
	}
	if utf8.RuneCountInString(body) > MaxBodyRunes {
		return "", payloadErr("body", "must be at most 4000 characters")
	}
	return body, nil
}

// decideContact — можно ли начать личный чат. Администратор — с кем угодно; преподаватели
// между собой; студент с преподавателем — если тот ведёт курс у группы студента;
// студенты — только одногруппники. teaches вызывается лениво, только когда нужен запрос в БД.
func decideContact(me, other userInfo, teaches func(teacherID, groupID int64) (bool, error)) (bool, error) {
	if me.Admin || other.Admin {
		return true, nil
	}
	switch {
	case me.Teacher && other.Teacher:
		return true, nil
	case me.Teacher:
		if other.GroupID == nil {
			return false, nil
		}
		return teaches(me.ID, *other.GroupID)
	case other.Teacher:
		if me.GroupID == nil {
			return false, nil
		}
		return teaches(other.ID, *me.GroupID)
	}
	return me.GroupID != nil && other.GroupID != nil && *me.GroupID == *other.GroupID, nil
}

// ---- чаты ----

func (s *Service) List(ctx context.Context, p middleware.Principal, q ListQuery) ([]Chat, int, error) {
	if !p.IsAdmin() {
		if err := s.repo.EnsureGroupChats(ctx, p.UserID); err != nil {
			return nil, 0, err
		}
	}
	return s.repo.ListChats(ctx, p.UserID, p.IsAdmin(), q)
}

func (s *Service) Contacts(ctx context.Context, p middleware.Principal, q ListQuery) ([]Contact, int, error) {
	return s.repo.ListContacts(ctx, p.UserID, q)
}

func (s *Service) OpenDirect(ctx context.Context, p middleware.Principal, otherID int64) (Chat, error) {
	if otherID == p.UserID {
		return Chat{}, errSelfChat
	}
	other, err := s.repo.UserInfo(ctx, otherID)
	if err != nil {
		return Chat{}, err
	}
	if !other.Active {
		return Chat{}, errUserNotFound
	}
	me, err := s.repo.UserInfo(ctx, p.UserID)
	if err != nil {
		return Chat{}, err
	}
	ok, err := decideContact(me, other, func(t, g int64) (bool, error) { return s.repo.TeachesGroup(ctx, t, g) })
	if err != nil {
		return Chat{}, err
	}
	if !ok {
		return Chat{}, errContactNotAllow
	}
	id, created, err := s.repo.OpenDirect(ctx, p.UserID, otherID)
	if err != nil {
		return Chat{}, err
	}
	// Новый личный чат — событие chat.created обоим участникам, каждому со своим
	// заголовком (имя собеседника).
	if created && s.publish != nil {
		for _, u := range []int64{p.UserID, otherID} {
			if view, err := s.repo.ChatView(ctx, id, u); err == nil {
				s.publish([]int64{u}, "chat.created", map[string]any{"chat": view})
			}
		}
	}
	return s.repo.ChatView(ctx, id, p.UserID)
}

func (s *Service) OpenGroup(ctx context.Context, p middleware.Principal, groupID int64) (Chat, error) {
	if !p.IsAdmin() {
		ok, err := s.repo.CanAccessGroup(ctx, p.UserID, groupID)
		if err != nil {
			return Chat{}, err
		}
		if !ok {
			return Chat{}, errGroupForbidden
		}
	}
	id, err := s.repo.OpenGroup(ctx, groupID, p.UserID)
	if err != nil {
		return Chat{}, err
	}
	return s.repo.ChatView(ctx, id, p.UserID)
}

// member возвращает чат, если пользователь в нём состоит. Чужой и несуществующий чат
// неотличимы (оба — 404), чтобы не раскрывать чужие переписки.
func (s *Service) member(ctx context.Context, p middleware.Principal, chatID int64) (chatRow, error) {
	c, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return chatRow{}, err
	}
	switch c.Kind {
	case KindDirect:
		if (c.UserLow != nil && *c.UserLow == p.UserID) || (c.UserHigh != nil && *c.UserHigh == p.UserID) {
			return c, nil
		}
	case KindGroup:
		if p.IsAdmin() {
			return c, nil
		}
		if c.GroupID != nil {
			ok, err := s.repo.CanAccessGroup(ctx, p.UserID, *c.GroupID)
			if err != nil {
				return chatRow{}, err
			}
			if ok {
				return c, nil
			}
		}
	}
	return chatRow{}, errChatNotFound
}

func (s *Service) Get(ctx context.Context, p middleware.Principal, chatID int64) (Chat, error) {
	if _, err := s.member(ctx, p, chatID); err != nil {
		return Chat{}, err
	}
	return s.repo.ChatView(ctx, chatID, p.UserID)
}

// ---- сообщения ----

// Messages — страница истории; limit уже проверен хендлером (1–MaxLimit).
func (s *Service) Messages(ctx context.Context, p middleware.Principal, chatID int64, before *int64, limit int) (MessagePage, error) {
	if _, err := s.member(ctx, p, chatID); err != nil {
		return MessagePage{}, err
	}
	return s.repo.ListMessages(ctx, chatID, before, limit)
}

// participant — пользователь состоит в чате (те же правила, что member), иначе deny.
func (s *Service) participant(ctx context.Context, p middleware.Principal, chatID int64, deny error) (chatRow, error) {
	c, err := s.member(ctx, p, chatID)
	if apperr.IsNotFound(err) {
		return chatRow{}, deny
	}
	return c, err
}

func (s *Service) checkFiles(ctx context.Context, p middleware.Principal, ids []int64) error {
	if len(ids) > MaxFiles {
		return payloadErr("file_ids", "at most 5 files")
	}
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return payloadErr("file_ids", "must be unique positive ids")
		}
		seen[id] = true
		if err := s.guard.OwnsFile(ctx, p.UserID, id); err != nil {
			var ae *apperr.Error
			if errors.As(err, &ae) && ae.Kind != apperr.Internal {
				return errFileNotFound
			}
			return err
		}
	}
	return nil
}

// Send — message.send: участник чата, текст и/или до 5 своих файлов.
func (s *Service) Send(ctx context.Context, p middleware.Principal, chatID int64, body string, fileIDs []int64) (Message, error) {
	c, err := s.participant(ctx, p, chatID, errChatForbidden)
	if err != nil {
		return Message{}, err
	}
	body, err = normalizeBody(body, len(fileIDs) > 0)
	if err != nil {
		return Message{}, err
	}
	if err := s.checkFiles(ctx, p, fileIDs); err != nil {
		return Message{}, err
	}
	if c.Kind == KindDirect {
		peer := *c.UserLow
		if peer == p.UserID {
			peer = *c.UserHigh
		}
		u, err := s.repo.UserInfo(ctx, peer)
		if err != nil {
			return Message{}, err
		}
		if !u.Active {
			return Message{}, errChatForbidden // собеседник заблокирован
		}
	}
	id, err := s.repo.CreateMessage(ctx, chatID, p.UserID, body, fileIDs)
	if apperr.IsNotFound(err) || isCode(err, "file_not_found") {
		return Message{}, errFileNotFound
	}
	if err != nil {
		return Message{}, err
	}
	return s.repo.GetMessage(ctx, id)
}

func isCode(err error, code string) bool {
	var ae *apperr.Error
	return errors.As(err, &ae) && ae.Code == code
}

// Participants доступен только пользователям, имеющим доступ к чату.
func (s *Service) Participants(ctx context.Context, p middleware.Principal, chatID int64) ([]Person, error) {
	if _, err := s.member(ctx, p, chatID); err != nil {
		return nil, err
	}
	return s.repo.Participants(ctx, chatID)
}
