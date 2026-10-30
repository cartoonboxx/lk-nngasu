package files

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"university-api/internal/platform/filekey"
)

func TestStorageSaveOpenLimitAndRemove(t *testing.T) {
	st, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	key, err := filekey.New(7, "отчёт.pdf")
	if err != nil {
		t.Fatal(err)
	}
	n, err := st.Save(key, strings.NewReader("hello"), 10)
	if err != nil || n != 5 {
		t.Fatalf("save: n=%d err=%v", n, err)
	}
	if size, err := st.Size(key); err != nil || size != 5 {
		t.Fatalf("size: %d %v", size, err)
	}
	rc, err := st.Open(key)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := io.ReadAll(rc); string(b) != "hello" {
		t.Fatalf("content = %q", b)
	}
	rc.Close()

	if err := st.Remove(key); err != nil {
		t.Fatal(err)
	}
	if err := st.Remove(key); err != nil {
		t.Fatalf("removing twice must be a no-op, got %v", err)
	}
	if _, err := st.Open(key); !errors.Is(err, errNoSuchFile) {
		t.Fatalf("open after remove: %v", err)
	}

	key2, _ := filekey.New(7, "big.zip")
	if _, err := st.Save(key2, strings.NewReader("0123456789ABC"), 10); !errors.Is(err, errTooLarge) {
		t.Fatalf("want errTooLarge, got %v", err)
	}
	if _, err := os.Stat(st.Path(key2)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("oversized file must be removed")
	}
}

func TestStorageRejectsTraversalKeys(t *testing.T) {
	st, _ := NewStorage(t.TempDir())
	for _, key := range []string{"", "../x", "1/../../etc/passwd", "/abs", "1//x", "1/./x", "a/..", "x\x00y"} {
		if _, err := st.Save(key, strings.NewReader("x"), 10); !errors.Is(err, errBadKey) {
			t.Errorf("Save(%q): want errBadKey, got %v", key, err)
		}
		if p := st.Path(key); p != "" {
			t.Errorf("Path(%q) = %q, want empty", key, p)
		}
	}
	// двойная точка внутри имени — допустимо
	if _, err := st.resolve("1/abc/report..v2.pdf"); err != nil {
		t.Errorf("name with inner dots must be fine: %v", err)
	}
}

func TestCleanName(t *testing.T) {
	long := strings.Repeat("я", 300) + ".pdf"
	cases := map[string]string{
		"../../etc/passwd":      "passwd",
		`C:\Users\me\Отчёт.pdf`: "Отчёт.pdf",
		"report.pdf":            "report.pdf",
		"a:b*c?.docx":           "a_b_c_.docx",
		"":                      "file",
		".":                     "file",
		"...":                   "file",
		".hidden.pdf":           "hidden.pdf",
		"tab\there.pdf":         "tabhere.pdf",
	}
	for in, want := range cases {
		if got := CleanName(in); got != want {
			t.Errorf("CleanName(%q) = %q, want %q", in, got, want)
		}
	}
	got := CleanName(long)
	if r := []rune(got); len(r) != maxNameRunes || !strings.HasSuffix(got, ".pdf") {
		t.Errorf("long name: %d runes, suffix ok = %v", len(r), strings.HasSuffix(got, ".pdf"))
	}
}

func TestMagicMatchesExtension(t *testing.T) {
	zip := []byte("PK\x03\x04rest")
	cases := []struct {
		ext   string
		types map[string]fileType
		head  []byte
		want  bool
	}{
		{".pdf", attachmentTypes, []byte("%PDF-1.7"), true},
		{".pdf", attachmentTypes, zip, false},
		{".docx", attachmentTypes, zip, true},
		{".docx", attachmentTypes, []byte("%PDF-1.7"), false},
		{".rar", attachmentTypes, []byte("Rar!\x1a\x07\x00"), true},
		{".xls", attachmentTypes, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1, 0}, true},
		{".zip", attachmentTypes, nil, false},
		{".png", avatarTypes, []byte("\x89PNG\r\n\x1a\n"), true},
		{".jpg", avatarTypes, []byte{0xFF, 0xD8, 0xFF, 0xE0}, true},
		{".webp", avatarTypes, []byte("RIFF\x00\x00\x00\x00WEBP"), true},
		{".webp", avatarTypes, []byte("RIFF\x00\x00\x00\x00WAVE"), false},
	}
	for _, tc := range cases {
		if got := tc.types[tc.ext].magic(tc.head); got != tc.want {
			t.Errorf("%s with % x: got %v, want %v", tc.ext, tc.head, got, tc.want)
		}
	}
	if _, ok := attachmentTypes[".exe"]; ok {
		t.Error(".exe must not be allowed")
	}
	if _, ok := attachmentTypes[".png"]; ok {
		t.Error("images are avatar-only")
	}
}

func TestMimeFor(t *testing.T) {
	if got := mimeFor("a.PDF"); got != "application/pdf" {
		t.Errorf("mimeFor(a.PDF) = %q", got)
	}
	if got := mimeFor("noext"); got != "application/octet-stream" {
		t.Errorf("mimeFor(noext) = %q", got)
	}
}
