package assignments

import "testing"

func TestValidGrade(t *testing.T) {
	ok := []float64{0, 0.1, 2.5, 3.7, 4.9, 5, 0.3, 1.1}
	bad := []float64{-0.1, 5.1, 2.55, 0.05, 100}
	for _, g := range ok {
		if !ValidGrade(g) {
			t.Errorf("%v must be valid", g)
		}
	}
	for _, g := range bad {
		if ValidGrade(g) {
			t.Errorf("%v must be invalid", g)
		}
	}
}

func TestValidateReviewTrimsComment(t *testing.T) {
	blank := "   "
	in := ReviewInput{Grade: 4, Comment: &blank}
	if err := validateReview(&in); err != nil || in.Comment != nil {
		t.Fatalf("blank comment must become nil, err=%v", err)
	}
	text := " ok "
	in = ReviewInput{Grade: 4, Comment: &text}
	if err := validateReview(&in); err != nil || *in.Comment != "ok" {
		t.Fatalf("comment must be trimmed, err=%v", err)
	}
	in = ReviewInput{Grade: 7}
	if validateReview(&in) == nil {
		t.Fatal("grade 7 must be rejected")
	}
}
