//go:build !windows && amd64

package verycomplex

/*
#include <stdio.h>
#include <stdlib.h>

static void myprint(char* s) {
  printf("%s\n", s);
}
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

//go:embed README.md
var readmeContent []byte

type UnsafeStruct struct {
	// Multi-line and complex struct tags are common in ORMs/serializers.
	ID   int64 `json:"id,omitempty"
	            xml:"id,attr"`
	data uintptr
}

// SetData uses unsafe to store a pointer as a uintptr.
//go:noinline
func (s *UnsafeStruct) SetData(p *[1024]byte) {
	s.data = uintptr(unsafe.Pointer(p))
}

// Inspect uses reflection to analyze an interface{}.
func Inspect(v any) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		// Using CGO
		cMsg := C.CString(fmt.Sprintf("Field: %s, Tag: %s", field.Name, field.Tag))
		defer C.free(unsafe.Pointer(cMsg))
		C.myprint(cMsg)
	}
}