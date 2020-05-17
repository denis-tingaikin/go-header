package goheader

type AnalyzerOption interface {
	apply(*analyzer)
}

type applyAnalyzerOptionFunc func(*analyzer)

func (f applyAnalyzerOptionFunc) apply(a *analyzer) {
	f(a)
}

func WithValues(values map[string]Value) AnalyzerOption {
	return applyAnalyzerOptionFunc(func(a *analyzer) {
		a.values = values
	})
}

func WithTemplate(template string) AnalyzerOption {
	return applyAnalyzerOptionFunc(func(a *analyzer) {
		a.template = template
	})
}
