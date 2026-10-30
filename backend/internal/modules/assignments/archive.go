package assignments

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/filekey"
	"university-api/internal/platform/middleware"
)

// ArchiveSource готовит данные для zip: проверяет права и собирает отправленные ответы.
// Сам zip пишется потоком в WriteArchive, чтобы большие архивы не лежали в памяти.
type ArchiveSource struct {
	Name    string // имя файла для скачивания
	Entries []ArchiveEntry
}

func (s *Service) ArchiveOfAssignment(ctx context.Context, p middleware.Principal, assignmentID int64) (ArchiveSource, error) {
	a, err := s.repo.GetAssignment(ctx, assignmentID, false, 0)
	if err != nil {
		return ArchiveSource{}, err
	}
	if err := s.guard.CourseManage(ctx, p, a.CourseID); err != nil {
		return ArchiveSource{}, err
	}
	entries, err := s.repo.ArchiveEntries(ctx, assignmentID, nil)
	return ArchiveSource{Name: fmt.Sprintf("assignment-%d.zip", assignmentID), Entries: entries}, err
}

func (s *Service) ArchiveOfSubmission(ctx context.Context, p middleware.Principal, id int64) (ArchiveSource, error) {
	sub, err := s.repo.GetSubmission(ctx, id)
	if err != nil {
		return ArchiveSource{}, err
	}
	if err := s.canReadSubmission(ctx, p, sub); err != nil {
		return ArchiveSource{}, err
	}
	entries, err := s.repo.ArchiveEntries(ctx, sub.AssignmentID, &id)
	return ArchiveSource{Name: fmt.Sprintf("submission-%d.zip", id), Entries: entries}, err
}

// WriteArchive пишет zip: папка на студента, внутри его файлы и answer.txt с текстом ответа.
// Файл, пропавший с диска, пропускается — архив остаётся пригодным.
func (s *Service) WriteArchive(w io.Writer, src ArchiveSource) error {
	if len(src.Entries) == 0 {
		return apperr.NewNotFound("no_submissions")
	}
	zw := zip.NewWriter(w)
	folders := map[string]int{}
	for _, e := range src.Entries {
		dir := uniqueFolder(folders, e.StudentName)
		if e.AnswerText != nil {
			if err := writeText(zw, path.Join(dir, "answer.txt"), *e.AnswerText); err != nil {
				return err
			}
		}
		names := map[string]int{}
		for _, key := range e.Keys {
			rc, err := s.blobs.Open(key)
			if err != nil {
				continue
			}
			name := uniqueName(names, filekey.Name(key))
			err = copyFile(zw, path.Join(dir, name), rc)
			rc.Close()
			if err != nil {
				return err
			}
		}
	}
	return zw.Close()
}

func writeText(zw *zip.Writer, name, text string) error {
	f, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.WriteString(f, text)
	return err
}

func copyFile(zw *zip.Writer, name string, r io.Reader) error {
	f, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// safeSegment убирает из имени разделители пути, чтобы оно не вышло за пределы папки в архиве.
func safeSegment(s string) string {
	s = strings.NewReplacer("/", "_", "\\", "_").Replace(strings.TrimSpace(s))
	if s == "" || s == "." || s == ".." {
		return "_"
	}
	return s
}

func uniqueFolder(used map[string]int, name string) string {
	name = safeSegment(name)
	n := used[name]
	used[name] = n + 1
	if n == 0 {
		return name
	}
	return fmt.Sprintf("%s (%d)", name, n+1)
}

// uniqueName добавляет суффикс " (2)", " (3)" ... перед расширением, если имя уже занято.
func uniqueName(used map[string]int, name string) string {
	n := used[name]
	used[name] = n + 1
	if n == 0 {
		return name
	}
	ext := path.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s (%d)%s", base, n+1, ext)
}
