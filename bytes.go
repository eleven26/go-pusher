package main

import "unsafe"

// StringToBytes 转换 string 到 []byte，不会分配新的内存
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// BytesToString 转换 []byte 到 string，不会分配新的内存
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
