package files

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	errTooLarge   = errors.New("file too large")
	errBadKey     = errors.New("invalid storage key")
	errNoSuchFile = errors.New("no such file")
)

// Storage — локальное файловое хранилище. Чтобы перейти на S3/MinIO, замените эту реализацию:
// остальной код видит только Save/Open/Remove/Size.
type Storage struct{ dir string }

func NewStorage(dir string) (*Storage, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &Storage{dir: dir}, nil
}

// resolve превращает ключ в путь внутри хранилища; ключи с ".", ".." и пустыми сегментами
// отвергаются, поэтому выйти за пределы каталога нельзя.
func (s *Storage) resolve(key string) (string, error) {
	if key == "" || strings.ContainsRune(key, 0) {
		return "", errBadKey
	}
	for _, seg := range strings.Split(key, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", errBadKey
		}
	}
	return filepath.Join(s.dir, filepath.FromSlash(key)), nil
}

// Save пишет поток в файл и возвращает размер; больше max байт — errTooLarge.
func (s *Storage) Save(key string, r io.Reader, max int64) (int64, error) {
	p, err := s.resolve(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return 0, err
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(f, io.LimitReader(r, max+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > max {
		err = errTooLarge
	}
	if err != nil {
		_ = os.Remove(p)
		return 0, err
	}
	return n, nil
}

// Path возвращает путь к файлу на диске (для отдачи клиенту); для некорректного ключа — "".
func (s *Storage) Path(key string) string {
	p, _ := s.resolve(key)
	return p
}

// Open открывает файл на чтение.
func (s *Storage) Open(key string) (io.ReadCloser, error) {
	p, err := s.resolve(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errNoSuchFile
	}
	return f, err
}

// Size возвращает размер файла в байтах.
func (s *Storage) Size(key string) (int64, error) {
	p, err := s.resolve(key)
	if err != nil {
		return 0, err
	}
	st, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

// Remove удаляет файл; отсутствующий файл — не ошибка. Пустая папка пользователя убирается.
func (s *Storage) Remove(key string) error {
	p, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_ = os.Remove(filepath.Dir(p)) // папка "<user>/<random>" одна на файл; не пуста — останется
	return nil
}
