package translator

func ptrFloat[T ~float32 | ~float64](f *T) *float64 {
	if f == nil {
		return nil
	}
	v := float64(*f)
	return &v
}
