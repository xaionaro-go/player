package libavv4l2

func ptr[T any](v T) *T {
	return &v
}
