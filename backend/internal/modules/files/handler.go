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
const maxAvatarBytes = 5 << 20 // 5 МБ

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

// errNotFoundFile — 404. Используется и для несуществующих, и для чужих файлов,
// чтобы по ответу нельзя было узнать, существует ли файл с таким id.
func errNotFoundFile() error { return apperr.NewNotFound("file_not_found") }

// errTooBig — файл превышает лимит для выбранного purpose.
func errTooBig() error {
	return apperr.NewInvalid("file_too_large", "file exceeds the upload limit")
}

// readable проверяет право на чтение; чужой и несуществующий файл неотличимы (оба — 404).
// Читать можно свои файлы (ключ начинается с префикса пользователя), файлы,
// к которым у пользователя есть доступ через привязку (решает repo.CanRead), и любые — админу.
// При успехе возвращает ключ файла; при отказе ответ клиенту уже отправлен.
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
//
// Проверки по порядку: размер → расширение из белого списка → сигнатура содержимого.
// Затем файл пишется на диск, и только после этого создаётся запись в file.
// Ответ — 201 с Info; полученный id клиент передаёт в задание, ответ или профиль.
func (h *Handler) Upload(c *gin.Context) {
	p := middleware.Me(c)

	// purpose определяет допустимые типы и лимит: для аватара — только картинки
	// и не больше maxAvatarBytes (но и не больше общего лимита).
	types, limit := attachmentTypes, h.maxBytes
	switch c.PostForm("purpose") {
	case "", "attachment":
	case "avatar":
		types, limit = avatarTypes, min(h.maxBytes, maxAvatarBytes)
	default:
		httpx.Fail(c, apperr.NewInvalid("invalid_purpose", "purpose must be attachment or avatar"))
		return
	}

	// Ограничиваем тело запроса. +1 МБ — запас на служебную часть multipart
	// (границы, заголовки частей, другие поля формы), точный лимит проверяется ниже.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit+(1<<20))
	fh, err := c.FormFile("file")
	if err != nil {
		httpx.BadRequest(c, errors.New(`multipart field "file" is required (or the file is too large)`))
		return
	}
	// Быстрая проверка по размеру, заявленному в multipart.
	if fh.Size > limit {
		httpx.Fail(c, errTooBig())
		return
	}

	// Очищаем имя от путей и опасных символов (защита от path traversal)
	// и проверяем расширение по белому списку для данного purpose.
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

	// Проверяем сигнатуру (magic bytes): содержимое должно соответствовать расширению,
	// чтобы нельзя было загрузить, например, исполняемый файл под видом .pdf.
	// Ошибку ReadFull игнорируем: файл короче 16 байт — не ошибка, берём сколько прочитали.
	head := make([]byte, 16)
	n, _ := io.ReadFull(src, head)
	head = head[:n]
	if !ft.magic(head) {
		httpx.Fail(c, apperr.NewInvalid("invalid_file_content", "file content does not match its extension"))
		return
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
	id, err := h.repo.Insert(c.Request.Context(), key)
	if err != nil {
		_ = h.storage.Remove(key)
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, Info{ID: id, Name: name, MimeType: ft.mime, SizeBytes: size})
}

// allowedList возвращает список разрешённых расширений через запятую — для текста ошибки.
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

	// Path возвращает "" если файла на диске нет (запись в БД есть, файл потерян).
	path := h.storage.Path(key)
	if path == "" {
		httpx.Fail(c, errNotFoundFile())
		return
	}

	name := filekey.Name(key)
	// nosniff — браузер не пытается угадать тип по содержимому;
	// Content-Type берём по расширению, а не из файла.
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Type", mimeFor(name))
	// Content-Disposition: attachment с исходным именем — файл скачивается,
	// а не открывается в браузере (защита от XSS через загруженный HTML/SVG).
	c.FileAttachment(path, name)
}

// GET /files/:id/meta — информация о файле без содержимого (те же права, что на скачивание).
func (h *Handler) Meta(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	key, ok := h.readable(c, id)
	if !ok {
		return
	}

	// Размер берём с диска. Если файла там нет — отвечаем 404, как и при скачивании.
	size, err := h.storage.Size(key)
	if err != nil {
		httpx.Fail(c, errNotFoundFile())
		return
	}
	name := filekey.Name(key)
	httpx.OK(c, Info{ID: id, Name: name, MimeType: mimeFor(name), SizeBytes: size})
}

// DELETE /files/:id — владелец или админ; файл, который к чему-то привязан, удалить нельзя (409).
//
// Права проверяются строже, чем на чтение: доступ через привязку не даёт права удалять.
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

	// Владелец определяется по ключу. Если ключ не разобрался, owner = 0
	// и проверка не пройдёт — удалить такой файл может только админ.
	if owner, _ := filekey.Owner(key); !p.IsAdmin() && owner != p.UserID {
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
