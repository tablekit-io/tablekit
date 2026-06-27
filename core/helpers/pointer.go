package helpers

// Pointer returns a pointer to v, for the optional *bool annotation hints.
func Pointer[T any](v T) *T { return &v }
