package text

type Reader interface {
	Peek() rune
	Next() rune
	Done() bool
	Finish() string
	Position() int
	SetPosition(int)
	ReadWhile(func(rune) bool) string
}

func NewReader(text string) Reader {
	return &reader{source: text, position: 0}
}

type reader struct {
	source   string
	position int
}

func (r *reader) Position() int {
	return r.position
}

func (r *reader) Peek() rune {
	if r.Done() {
		return rune(0)
	}
	return rune(r.source[r.position])
}

func (r *reader) Done() bool {
	return r.position >= len(r.source)
}

func (r *reader) Next() rune {
	if r.Done() {
		return rune(0)
	}
	reuslt := r.Peek()
	r.position++
	return reuslt
}

func (r *reader) Finish() string {
	if r.position >= len(r.source) {
		return ""
	}
	defer r.till()
	return r.source[r.position:]
}

func (r *reader) SetPosition(pos int) {
	if pos < 0 {
		r.position = 0
	}
	r.position = pos
}

func (r *reader) ReadWhile(match func(rune) bool) string {
	if match == nil {
		return ""
	}
	start := r.position
	for !r.Done() && match(r.Peek()) {
		r.Next()
	}
	return r.source[start:r.position]
}

func (r *reader) till() {
	r.position = len(r.source)
}
