package relay

func ptr[T any](v T) *T {
	return &v
}
