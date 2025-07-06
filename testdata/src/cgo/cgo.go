/*MY TITLE1*/ //want `template doesn't match`

//golangcitest:args -Egoheader
//golangcitest:config_path testdata/goheader.yml
//golangcitest:expected_exitcode 1

package cgo

/*
 #include <stdio.h>
 #include <stdlib.h>

 void myprint(char* s) {
 	printf("%s\n", s);
 }
*/
import "C"

import (
	"unsafe"
)

func _() {
	cs := C.CString("Hello from stdio\n")
	C.myprint(cs)
	C.free(unsafe.Pointer(cs))
}
