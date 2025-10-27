//go:build ignore

// Test Pattern 4: Blank Imports and Side Effects
// Tests blank imports that execute init() functions for side effects

package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"    // MySQL driver registration
	_ "github.com/lib/pq"                 // PostgreSQL driver registration
	_ "github.com/mattn/go-sqlite3"       // SQLite driver registration
	"net/http"
	_ "net/http/pprof"                    // Profiling endpoints registration
	_ "image/png"                         // PNG format registration
	_ "image/jpeg"                        // JPEG format registration
	"image"
	_ "time/tzdata"                       // Timezone data embedding
	"crypto/tls"
	"log"
)

// Not using: crypto/tls, log

func main() {
	// Using fmt
	fmt.Println("Application with multiple drivers")

	// Using database/sql - the blank imports register the drivers
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		fmt.Printf("MySQL connection error: %v\n", err)
	}
	defer db.Close()

	// Try PostgreSQL
	pgDB, err := sql.Open("postgres", "postgresql://user:password@localhost/dbname")
	if err != nil {
		fmt.Printf("PostgreSQL connection error: %v\n", err)
	}
	defer pgDB.Close()

	// Using image package - blank imports register decoders
	var img image.Image
	fmt.Printf("Image variable type: %T\n", img)

	// The pprof import adds profiling endpoints
	// The image format imports add decoders
	// The tzdata import embeds timezone data
	// All blank imports should be kept!

	// Using http
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server with side-effect imports"))
	})

	fmt.Println("Server ready on :8082")
	http.ListenAndServe(":8082", nil)
}