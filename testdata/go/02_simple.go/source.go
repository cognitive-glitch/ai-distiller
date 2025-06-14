package simple

import "io"

// Writer defines a simple interface for writing data.
type Writer interface {
	// Write accepts a byte slice and returns bytes written and an error.
	Write(p []byte) (n int, err error)
}

// Data represents a container for a piece of data.
// It demonstrates a basic struct with a single field.
type Data struct {
	value string // unexported field
}

// NewData is a constructor-like function, a common Go idiom.
func NewData(v string) *Data {
	return &Data{value: v}
}

// ReadValue returns the current value. This is a value receiver method.
// It can be called on both Data and *Data types.
func (d Data) ReadValue() string {
	return d.value
}

// UpdateValue modifies the value. This is a pointer receiver method.
// It can only be called on a *Data type.
func (d *Data) UpdateValue(v string) {
	d.value = v
}

// Ensure Data does not satisfy the io.Writer interface, but we use it
// to test how the distiller handles imported interfaces.
var _ io.Writer = (*customWriter)(nil) // compile-time check idiom
type customWriter struct{}
func (cw *customWriter) Write(p []byte) (n int, err error) { return 0, nil }