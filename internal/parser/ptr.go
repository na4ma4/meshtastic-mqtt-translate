package parser

func ptr[T any](v T) *T {
	return &v
}
