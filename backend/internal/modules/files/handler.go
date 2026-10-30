package files

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/filekey"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// maxAvatarBytes — потолок для картинок профиля, независимо от общего лимита.
const maxAvatarBytes = 5 << 20

type Handler struct {
	repo     *Repository
	storage  *Storage
	maxBytes int64
}

// Info — что клиент узнаёт о файле. Имя и тип выводятся из ключа, размер — с диска.
type Info struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

func errNotFoundFile() error { return apperr.NewNotFound("file_not_found") }

func errTooBig() error {
	return apperr.NewInvalid("file_too_large", "file exceeds the upload limit")
}

// readable проверяет право на чтение; чужой и несуществующий файл неотличимы (оба — 404).
func (h *Handler) readable(c *gin.Context, id int64) (key string, ok bool) {
	p := middleware.Me(c)
	allowed, err := h.repo.CanRead(c.Request.Context(), id, p.UserID, p.IsAdmin(), filekey.Prefix(p.UserID))
	if err != nil {
		httpx.Fail(c, err)
		return "", false
	}
	if !allowed {
		httpx.Fail(c, errNotFoundFile())
		return "", false
	}
	key, err = h.repo.Key(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, err)
		return "", false
	}
	return key, true
}

// POST /files  (multipart/form-data: поле "file", необязательное "purpose" = attachment | avatar)
func (h *Handler) Upload(c *gin.Context) {
	p := middleware.Me(c)

	types, limit := attachmentTypes, h.maxBytes
	switch c.PostForm("purpose") {
	case "", "attachment":
	case "avatar":
		types, limit = avatarTypes, min(h.maxBytes, maxAvatarBytes)
	default:
		httpx.Fail(c, apperr.NewInvalid("invalid_purpose", "purpose must be attachment or avatar"))
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit+(1<<20))
	fh, err := c.FormFile("file")
	if err != nil {
		httpx.BadRequest(c, errors.New(`multipart field "file" is required (or the file is too large)`))
		return
	}
	if fh.Size > limit {
		httpx.Fail(c, errTooBig())
		return
	}

	name := CleanName(fh.Filename)
	ft, ok := types[extOf(name)]
	if !ok {
		httpx.Fail(c, apperr.NewInvalid("unsupported_file_type", "file type is not allowed: "+allowedList(types)))
		return
	}

	src, err := fh.Open()
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	defer src.Close()

	head := make([]byte, 16)
	n, _ := io.ReadFull(src, head)
	head = head[:n]
	if !ft.magic(head) {
		httpx.Fail(c, apperr.NewInvalid("invalid_file_content", "file content does not match its extension"))
		return
	}

	key, err := filekey.New(p.UserID, name)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	size, err := h.storage.Save(key, io.MultiReader(bytes.NewReader(head), src), limit)
	if errors.Is(err, errTooLarge) {
		httpx.Fail(c, errTooBig())
		return
	}
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	id, err := h.repo.Insert(c.Request.Context(), key)
	if err != nil {
		_ = h.storage.Remove(key)
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, Info{ID: id, Name: name, MimeType: ft.mime, SizeBytes: size})
}

func allowedList(types map[string]fileType) string {
	exts := make([]string, 0, len(types))
	for ext := range types {
		exts = append(exts, ext)
	}
	sort.Strings(exts) // порядок карты случаен — сортируем, чтобы сообщение было стабильным
	return strings.Join(exts, ", ")
}

// GET /files/:id — скачивание (всегда как вложение, без исполнения в браузере).
func (h *Handler) Download(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	key, ok := h.readable(c, id)
	if !ok {
		return
	}
	path := h.storage.Path(key)
	if path == "" {
		httpx.Fail(c, errNotFoundFile())
		return
	}
	name := filekey.Name(key)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Type", mimeFor(name))
	c.FileAttachment(path, name)
}

// GET /files/:id/meta
func (h *Handler) Meta(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	key, ok := h.readable(c, id)
	if !ok {
		return
	}
	size, err := h.storage.Size(key)
	if err != nil {
		httpx.Fail(c, errNotFoundFile())
		return
	}
	name := filekey.Name(key)
	httpx.OK(c, Info{ID: id, Name: name, MimeType: mimeFor(name), SizeBytes: size})
}

// DELETE /files/:id — владелец или админ; файл, который к чему-то привязан, удалить нельзя (409).
func (h *Handler) Delete(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	p := middleware.Me(c)
	key, err := h.repo.Key(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if owner, _ := filekey.Owner(key); !p.IsAdmin() && owner != p.UserID {
		httpx.Fail(c, errNotFoundFile()) // чужие файлы не раскрываем
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		httpx.Fail(c, err)
		return
	}
	_ = h.storage.Remove(key)
	httpx.NoContent(c)
}
