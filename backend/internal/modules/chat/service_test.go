package chat

import (
	"strings"
	"testing"

	"university-api/internal/platform/apperr"
)

func TestPair(t *testing.T) {
	for _, c := range [][2]int64{{3, 7}, {7, 3}} {
		low, high := pair(c[0], c[1])
		if low != 3 || high != 7 {
			t.Fatalf("pair(%d,%d) = %d,%d", c[0], c[1], low, high)
		}
	}
}

func TestNormalizeBody(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		hasFiles bool
		want     string
		code     string
	}{
		{"trim", "  привет \n", false, "привет", ""},
		{"empty", "   ", false, "", "empty_message"},
		{"empty with files", "", true, "", ""},
		{"max runes ok", strings.Repeat("я", MaxBodyRunes), false, strings.Repeat("я", MaxBodyRunes), ""},
		{"too long", strings.Repeat("я", MaxBodyRunes+1), false, "", "message_too_long"},
	}
	for _, c := range cases {
		got, err := normalizeBody(c.in, c.hasFiles)
		if c.code == "" {
			if err != nil || got != c.want {
				t.Errorf("%s: got %q, %v", c.name, got, err)
			}
			continue
		}
		ae, ok := err.(*apperr.Error)
		if !ok || ae.Code != c.code {
			t.Errorf("%s: want code %s, got %v", c.name, c.code, err)
		}
	}
}

func ptr(v int64) *int64 { return &v }

func TestDecideContact(t *testing.T) {
	admin := userInfo{ID: 1, Admin: true}
	teacher := userInfo{ID: 2, Teacher: true}
	teacher2 := userInfo{ID: 3, Teacher: true}
	s1 := userInfo{ID: 10, GroupID: ptr(1)}
	s1b := userInfo{ID: 11, GroupID: ptr(1)}
	s2 := userInfo{ID: 12, GroupID: ptr(2)}

	// преподаватель 2 ведёт только группу 1
	teaches := func(teacherID, groupID int64) (bool, error) { return teacherID == 2 && groupID == 1, nil }
	never := func(int64, int64) (bool, error) { t.Fatal("teaches must not be called"); return false, nil }

	cases := []struct {
		name    string
		a, b    userInfo
		teaches func(int64, int64) (bool, error)
		want    bool
	}{
		{"admin with student", admin, s1, never, true},
		{"student with admin", s1, admin, never, true},
		{"teacher with teacher", teacher, teacher2, never, true},
		{"teacher with own student", teacher, s1, teaches, true},
		{"student with own teacher", s1, teacher, teaches, true},
		{"teacher with foreign student", teacher, s2, teaches, false},
		{"foreign student with teacher", s2, teacher, teaches, false},
		{"groupmates", s1, s1b, never, true},
		{"other groups", s1, s2, never, false},
		{"student without group", userInfo{ID: 20}, s1, never, false},
	}
	for _, c := range cases {
		got, err := decideContact(c.a, c.b, c.teaches)
		if err != nil || got != c.want {
			t.Errorf("%s: got %v, %v; want %v", c.name, got, err, c.want)
		}
	}
}
