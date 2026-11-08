package chat

import (
	"context"
	"io"
	"strings"
	"unicode/utf8"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// Guard — проверка, что файл принадлежит пользователю (реализует access.Checker).
type Guard interface {
	OwnsFile(ctx context.Context, userID, fileID int64) error
}

// Blobs — файловое хранилище (реализует files.Storage).
type Blobs interface {
	Open(key string) (io.ReadCloser, error)
	Remove(key string) error
}

type Service struct {
	repo  *Repository
	guard Guard
	blobs Blobs
}

func NewService(repo *Repository, guard Guard, blobs Blobs) *Service {
	return &Service{repo: repo, guard: guard, blobs: blobs}
}

var (
	errChatNotFound    = apperr.NewNotFound("chat_not_found")
	errSelfChat        = apperr.NewInvalid("chat_with_self", "cannot start a chat with yourself")
	errContactNotAllow = apperr.NewForbidden("contact_not_allowed", "you cannot start a chat with this user")
	errPeerInactive    = apperr.NewConflict("peer_inactive", "the other user is deactivated")
	errForbidden       = apperr.NewForbidden("forbidden", "no access to this resource")
	errNotAuthor       = apperr.NewForbidden("not_author", "only the author can edit this message")
)

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
		return "", apperr.NewInvalid("empty_message", "message must have text or files")
	}
	if utf8.RuneCountInString(body) > MaxBodyRunes {
		return "", apperr.NewInvalid("message_too_long", "message is too long")
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

func (s *Service) removeBlobs(keys []string) {
	for _, k := range keys {
		_ = s.blobs.Remove(k)
	}
}

// ---- чаты ----

func (s *Service) List(ctx context.Context, p middleware.Principal) ([]Chat, error) {
	if !p.IsAdmin() {
		if err := s.repo.EnsureGroupChats(ctx, p.UserID); err != nil {
			return nil, err
		}
	}
	return s.repo.ListChats(ctx, p.UserID, p.IsAdmin())
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
		return Chat{}, apperr.NewNotFound("user_not_found")
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
	id, err := s.repo.OpenDirect(ctx, p.UserID, otherID)
	if err != nil {
		return Chat{}, err
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
			return Chat{}, errForbidden
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

func (s *Service) Messages(ctx context.Context, p middleware.Principal, chatID int64, before, after *int64, limit int) ([]Message, error) {
	if before != nil && after != nil {
		return nil, apperr.NewInvalid("invalid_query", "use either before or after, not both")
	}
	if _, err := s.member(ctx, p, chatID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	limit = min(limit, MaxLimit)
	return s.repo.ListMessages(ctx, chatID, before, after, limit)
}

func (s *Service) checkFiles(ctx context.Context, p middleware.Principal, ids []int64) error {
	if len(ids) > MaxFiles {
		return apperr.NewInvalid("too_many_files", "at most 5 files can be attached")
	}
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return apperr.NewInvalid("duplicate_file", "the same file is attached twice")
		}
		seen[id] = true
		if err := s.guard.OwnsFile(ctx, p.UserID, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Send(ctx context.Context, p middleware.Principal, chatID int64, body string, fileIDs []int64) (Message, error) {
	c, err := s.member(ctx, p, chatID)
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
			return Message{}, errPeerInactive
		}
	}
	id, err := s.repo.CreateMessage(ctx, chatID, p.UserID, body, fileIDs)
	if err != nil {
		return Message{}, err
	}
	return s.repo.GetMessage(ctx, id)
}

// ownMessage загружает сообщение и проверяет, что оно из этого чата и не удалено.
func (s *Service) ownMessage(ctx context.Context, chatID, msgID int64) (Message, error) {
	m, err := s.repo.GetMessage(ctx, msgID)
	if err != nil {
		return Message{}, err
	}
	if m.ChatID != chatID || m.Deleted {
		return Message{}, apperr.NewNotFound("message_not_found")
	}
	return m, nil
}

func (s *Service) Edit(ctx context.Context, p middleware.Principal, chatID, msgID int64, body string) (Message, error) {
	if _, err := s.member(ctx, p, chatID); err != nil {
		return Message{}, err
	}
	m, err := s.ownMessage(ctx, chatID, msgID)
	if err != nil {
		return Message{}, err
	}
	if m.SenderID != p.UserID {
		return Message{}, errNotAuthor
	}
	body, err = normalizeBody(body, len(m.Files) > 0)
	if err != nil {
		return Message{}, err
	}
	if err := s.repo.UpdateMessageBody(ctx, msgID, body); err != nil {
		return Message{}, err
	}
	return s.repo.GetMessage(ctx, msgID)
}

// Delete: свои сообщения может удалять автор, любые — администратор.
func (s *Service) Delete(ctx context.Context, p middleware.Principal, chatID, msgID int64) error {
	if _, err := s.member(ctx, p, chatID); err != nil {
		return err
	}
	m, err := s.ownMessage(ctx, chatID, msgID)
	if err != nil {
		return err
	}
	if m.SenderID != p.UserID && !p.IsAdmin() {
		return errForbidden
	}
	keys, err := s.repo.DeleteMessage(ctx, msgID)
	s.removeBlobs(keys)
	return err
}

func (s *Service) MarkRead(ctx context.Context, p middleware.Principal, chatID, messageID int64) error {
	if _, err := s.member(ctx, p, chatID); err != nil {
		return err
	}
	return s.repo.MarkRead(ctx, chatID, p.UserID, messageID)
}
