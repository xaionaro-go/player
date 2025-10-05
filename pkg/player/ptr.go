package player

func ptr[T any](v T) *T {
	return &v
}
