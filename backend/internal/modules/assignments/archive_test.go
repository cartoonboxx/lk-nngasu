package assignments

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

type memBlobs map[string]string

func (m memBlobs) Open(k string) (io.ReadCloser, error) {
	v, ok := m[k]
	if !ok {
		return nil, io.ErrUnexpectedEOF
	}
	return io.NopCloser(strings.NewReader(v)), nil
}
func (memBlobs) Remove(string) error { return nil }

func TestUniqueName(t *testing.T) {
	used := map[string]int{}
	got := []string{uniqueName(used, "a.pdf"), uniqueName(used, "a.pdf"), uniqueName(used, "a.pdf"), uniqueName(used, "b")}
	want := []string{"a.pdf", "a (2).pdf", "a (3).pdf", "b"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestSafeSegment(t *testing.T) {
	for in, want := range map[string]string{"../x": ".._x", "a/b": "a_b", "": "_", "..": "_", " Иванов И. ": "Иванов И."} {
		if got := safeSegment(in); got != want {
			t.Errorf("safeSegment(%q)=%q want %q", in, got, want)
		}
	}
}

func TestWriteArchive(t *testing.T) {
	text := "ответ"
	svc := &Service{blobs: memBlobs{"1/aa/work.txt": "data"}}
	src := ArchiveSource{Entries: []ArchiveEntry{
		{StudentName: "Иванов И.", AnswerText: &text, Keys: []string{"1/aa/work.txt", "1/aa/missing.txt"}},
		{StudentName: "Иванов И.", Keys: []string{"1/aa/work.txt"}},
	}}
	var buf bytes.Buffer
	if err := svc.WriteArchive(&buf, src); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	want := "Иванов И./answer.txt,Иванов И./work.txt,Иванов И. (2)/work.txt"
	if strings.Join(names, ",") != want {
		t.Fatalf("got %v", names)
	}
	if err := svc.WriteArchive(&buf, ArchiveSource{}); err == nil {
		t.Fatal("empty archive must be an error")
	}
}
