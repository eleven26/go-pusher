package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	// 注：下面这行代码只是为了演示
	// 实际使用的时候，我们应该将 JWT_SECRET 设置到环境变量中
	err := os.Setenv("JWT_SECRET", "secret")
	if err != nil {
		panic(err)
	}

	addr := flag.String("addr", ":8181", "http service address")
	flag.Parse()

	err = server(*addr)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
