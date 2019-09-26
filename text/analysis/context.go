package analysis

import "context"

type key string

const (
	templateKey key = "template"
)

func Template(ctx context.Context) string {
	if v, ok := ctx.Value(templateKey).(string); ok {
		return v
	}
	return ""
}

func WithTemplate(ctx context.Context, template string) context.Context {
	return context.WithValue(ctx, templateKey, template)
}
