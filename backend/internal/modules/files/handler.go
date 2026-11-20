package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/filekey"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// maxAvatarBytes — потолок для картинок профиля, независимо от общего лимита.
const maxAvatarBytes = 2 << 20 // 2 MiB

// Handler — HTTP-слой модуля files. Отдельного сервиса нет:
// handler сам работает с таблицей file (repo) и файлами на диске (storage).
type Handler struct {
	repo     *Repository
	storage  *Storage
	maxBytes int64 // общий лимит размера загружаемого файла
}

// Info — что клиент узнаёт о файле. Имя и тип выводятся из ключа, размер — с диска.
type Info struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

// formErr — 400 invalid_request с указанием поля multipart-формы.
func formErr(field, problem string) error {
	return apperr.NewInvalid("invalid_request", "request body is invalid").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// errNotFoundFile — 404. Используется и для несуществующих, и для чужих файлов,
// чтобы по ответу нельзя было узнать, существует ли файл с таким id.
func errNotFoundFile() error { return apperr.NewNotFound("file_not_found") }

// errTooBig — файл превышает лимит для выбранного purpose.
func errTooBig() error {
	return apperr.NewTooLarge("file_too_large", "file exceeds the upload limit") // 413
}

// readable проверяет право на чтение; чужой и несуществующий файл неотличимы (оба — 404).
// Читать можно свои файлы (ключ начинается с префикса пользователя), файлы,
// к которым у пользователя есть доступ через привязку (решает repo.CanRead), и любые — админу.
// При успехе возвращает запись о файле; при отказе ответ клиенту уже отправлен.
func (h *Handler) readable(c *gin.Context, id int64) (f stored, ok bool) {
	p := middleware.Me(c)
	allowed, err := h.repo.CanRead(c.Request.Context(), id, p.UserID, p.IsAdmin(), filekey.Prefix(p.UserID))
	if err != nil {
		httpx.Fail(c, err)
		return stored{}, false
	}
	if !allowed {
		httpx.Fail(c, errNotFoundFile())
		return stored{}, false
	}
	f, err = h.repo.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, err)
		return stored{}, false
	}
	return f, true
}

// info — FileInfo из записи; тип и размер берутся из БД, а у старых файлов
// (до миграции 010) — по расширению и с диска. false — файла нет на диске.
func (h *Handler) info(id int64, f stored) (Info, bool) {
	name := filekey.Name(f.key)
	out := Info{ID: id, Name: name, MimeType: mimeFor(name)}
	if f.mime != nil && *f.mime != "" {
		out.MimeType = *f.mime
	}
	if f.size != nil {
		out.SizeBytes = *f.size
		return out, true
	}
	size, err := h.storage.Size(f.key)
	out.SizeBytes = size
	return out, err == nil
}

// POST /files  (multipart/form-data FileUpload: "file" и обязательное "purpose" = attachment | avatar)
//
// Файлы принимаются любых типов. Для аватаров проверяется размер.
// Затем файл пишется на диск, и только после этого создаётся запись в file.
// Ответ — 201 с Info; полученный id клиент передаёт в задание, ответ или профиль.
func (h *Handler) Upload(c *gin.Context) {
	p := middleware.Me(c)

	// Нулевой общий лимит означает отсутствие ограничения для вложений.
	if h.maxBytes > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBytes+(1<<20))
	}
	// Крупные multipart-файлы остаются на диске только до завершения запроса.
	defer func() {
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
	}()
	limit := h.maxBytes
	switch purpose, given := c.GetPostForm("purpose"); {
	case !given:
		httpx.Fail(c, formErr("purpose", "required: attachment or avatar"))
		return
	case purpose == "attachment":
	case purpose == "avatar":
		limit = maxAvatarBytes
		if h.maxBytes > 0 {
			limit = min(limit, h.maxBytes)
		}
	default:
		httpx.Fail(c, formErr("purpose", "must be attachment or avatar"))
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		// Тело длиннее лимита — 413, а не «поле не найдено».
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			httpx.Fail(c, errTooBig())
			return
		}
		httpx.Fail(c, formErr("file", "required"))
		return
	}
	// Быстрая проверка по размеру, заявленному в multipart.
	if limit > 0 && fh.Size > limit {
		httpx.Fail(c, errTooBig())
		return
	}

	name := CleanName(fh.Filename)
	ft := fileType{mime: mimeFor(name)}

	src, err := fh.Open()
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	defer src.Close()

	// Для вложений тип определяется без ограничений на содержимое.
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	head = head[:n]
	if ft.mime == "application/octet-stream" {
		ft.mime = http.DetectContentType(head)
	}

	// Ключ "<user_id>/<random>/<имя>" — владелец и имя хранятся в нём самом.
	key, err := filekey.New(p.UserID, name)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	// Первые байты уже прочитаны — склеиваем их обратно с остатком потока.
	// Save повторно ограничивает размер по фактически записанным байтам:
	// заявленному в multipart fh.Size полностью доверять нельзя.
	size, err := h.storage.Save(key, io.MultiReader(bytes.NewReader(head), src), limit)
	if errors.Is(err, errTooLarge) {
		httpx.Fail(c, errTooBig())
		return
	}
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	// Сначала диск, потом БД. Если запись в БД не удалась — удаляем файл с диска,
	// чтобы не оставлять файлов без записи.
	id, err := h.repo.Insert(c.Request.Context(), key, ft.mime, size)
	if err != nil {
		_ = h.storage.Remove(key)
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, Info{ID: id, Name: name, MimeType: ft.mime, SizeBytes: size})
}

// GET /files/:id — скачивание (всегда как вложение, без исполнения в браузере).
func (h *Handler) Download(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	f, ok := h.readable(c, id)
	if !ok {
		return
	}

	// Path возвращает "" если файла на диске нет (запись в БД есть, файл потерян).
	path := h.storage.Path(f.key)
	if path == "" {
		httpx.Fail(c, errNotFoundFile())
		return
	}
	info, _ := h.info(id, f)

	// nosniff — браузер не пытается угадать тип по содержимому; Content-Type — из FileInfo.mime_type.
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Type", info.MimeType)
	// attachment: файл скачивается, а не открывается в браузере (защита от XSS через
	// загруженный HTML/SVG). filename — ASCII-замена для старых клиентов, filename* — точное имя.
	c.Header("Content-Disposition", contentDisposition(info.Name))
	// ServeFile ставит Content-Length и поддерживает Range; заданный Content-Type не меняет.
	http.ServeFile(c.Writer, c.Request, path)
}

// contentDisposition — attachment с безопасным именем (RFC 6266): ASCII-вариант
// в filename (кавычки и обратные слэши заменены) и UTF-8 в filename*.
func contentDisposition(name string) string {
	ascii := strings.Map(func(r rune) rune {
		if r < 0x20 || r > 0x7e || r == '"' || r == '\\' {
			return '_'
		}
		return r
	}, name)
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, ascii, url.PathEscape(name))
}

// GET /files/:id/meta — информация о файле без содержимого (те же права, что на скачивание).
func (h *Handler) Meta(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	f, ok := h.readable(c, id)
	if !ok {
		return
	}
	info, onDisk := h.info(id, f)
	if !onDisk {
		httpx.Fail(c, errNotFoundFile()) // старый файл без size_bytes и без файла на диске
		return
	}
	httpx.OK(c, info)
}

// DELETE /files/:id — только владелец и только пока файл ни к чему не привязан (иначе 409 file_in_use).
//
// Права строже, чем на чтение: ни доступ через привязку, ни роль администратора права удалять
// не дают. Забытые загрузки через сутки удаляет filegc.Sweeper.
func (h *Handler) Delete(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	p := middleware.Me(c)
	f, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	key := f.key

	// Владелец определяется по ключу. Если ключ не разобрался, owner = 0 и проверка не пройдёт.
	if owner, _ := filekey.Owner(key); owner != p.UserID {
		httpx.Fail(c, errNotFoundFile()) // чужие файлы не раскрываем
		return
	}

	// Сначала БД: если файл привязан, repo.Delete вернёт конфликт (409),
	// и файл на диске останется нетронутым.
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		httpx.Fail(c, err)
		return
	}

	// Ошибку удаления с диска игнорируем: запись уже удалена, клиенту важен результат.
	// В худшем случае на диске останется файл без записи.
	_ = h.storage.Remove(key)
	httpx.NoContent(c)
}
