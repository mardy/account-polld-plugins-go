package main

// #cgo CXXFLAGS: -std=c++0x -pedantic-errors -Wall -fno-strict-aliasing -I/usr/share/c++/4.8
// #cgo LDFLAGS: -lstdc++
// #cgo pkg-config: Qt5Core Qt5Contacts
// #include "qtcontacts.h"
import "C"

import "fmt"

//export AvatarPath
func AvatarPath(path *C.char) {
	fmt.Println("email", C.GoString(path))
}

func main() {
	C.getAvatar(C.CString("sergiusens@gmail.com"))
	C.getAvatar(C.CString("sergiusens@gmail.com"))
}
