package koyeb

func toOpt[T any](v T) *T {
	return &v
}
