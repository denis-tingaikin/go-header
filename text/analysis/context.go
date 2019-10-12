package analysis

import "context"

type key string

const (
	templateKey key = "template"
	visitMap    key = "visit map"
	visited     key = "visit"
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

func Visit(ctx context.Context, what string) context.Context {
	if v, ok := ctx.Value(visited).([]string); ok {
		ctx = context.WithValue(ctx, visited, append(v, what))
	} else {
		ctx = context.WithValue(ctx, visited, []string{what})
	}
	if v, ok := ctx.Value(visitMap).(map[string]bool); ok {
		v[what] = true
		return ctx
	} else {
		return context.WithValue(ctx, visitMap, map[string]bool{what: true})
	}
}

func VisitedList(ctx context.Context) []string {
	if v, ok := ctx.Value(visited).([]string); ok {
		return v
	}
	return nil
}

func IsVisited(ctx context.Context, what string) bool {
	if v, ok := ctx.Value(visitMap).(map[string]bool); ok {
		return v[what]
	}
	return false
}

func Leave(ctx context.Context, what string) context.Context {
	if v, ok := ctx.Value(visited).([]string); ok {
		ctx = context.WithValue(ctx, visited, v[0:len(v)])
	}
	if v, ok := ctx.Value(visitMap).(map[string]bool); ok {
		v[what] = false
		return ctx
	} else {
		return context.WithValue(ctx, visitMap, map[string]bool{what: false})
	}
}
