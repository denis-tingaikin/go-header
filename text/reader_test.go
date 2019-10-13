package text

import "testing"

func TestLocation1(t *testing.T) {
	r := NewReader("abc\n")
	r.Next()
	l := r.Location()
	if l.Position != 1 {
		t.FailNow()
	}
	r.Next()
	l = r.Location()
	if l.Position != 2 {
		t.FailNow()
	}
	r.Next()
	r.Next()
	l = r.Location()
	if l.Position != 0 {
		println(1)
		t.FailNow()
	}
	if l.Line != 1 {
		t.FailNow()
	}
}

func TestLocation2(t *testing.T) {
	r := NewReader("abc\nabc\nabc")
	r.SetPosition(4)
	l := r.Location()
	if l.Line != 1 || l.Position != 0 {
		t.FailNow()
	}
	r.SetPosition(0)
	l = r.Location()
	if l.Line != 0 || l.Position != 0 {
		t.FailNow()
	}
	r.SetPosition(8)
	l = r.Location()
	if l.Line != 2 || l.Position != 0 {
		t.FailNow()
	}
	r.SetPosition(9)
	l = r.Location()
	if l.Line != 2 || l.Position != 1 {
		t.FailNow()
	}
}
