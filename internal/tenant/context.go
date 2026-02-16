package tenant

import "context"

type contextKey struct{}

func WithInfo(ctx context.Context, info *Info) context.Context {
	return context.WithValue(ctx, contextKey{}, info)
}

func FromContext(ctx context.Context) *Info {
	info, _ := ctx.Value(contextKey{}).(*Info)
	return info
}
