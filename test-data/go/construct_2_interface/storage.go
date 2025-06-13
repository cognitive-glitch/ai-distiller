// storage.go
package storage

// Storer defines a generic interface for key-value storage.
type Storer interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
}

// PostgresStorer is a production implementation using a database.
type PostgresStorer struct {
	ConnectionString string
}

// Get retrieves a value from Postgres.
func (ps *PostgresStorer) Get(key string) ([]byte, error) {
	// db logic...
	return []byte("data from postgres"), nil
}

// Set saves a value to Postgres.
func (ps *PostgresStorer) Set(key string, value []byte) error {
	// db logic...
	return nil
}

// MemoryStorer is an in-memory implementation for testing.
type MemoryStorer struct {
	data map[string][]byte
}

// Get retrieves a value from memory.
func (ms MemoryStorer) Get(key string) ([]byte, error) {
	return ms.data[key], nil
}

// Set saves a value to memory.
func (ms MemoryStorer) Set(key string, value []byte) error {
	if ms.data == nil {
		ms.data = make(map[string][]byte)
	}
	ms.data[key] = value
	return nil
}