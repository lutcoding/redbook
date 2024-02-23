package main

import "github.com/lutcoding/redbook/internal"

func main() {
	server, err := internal.NewServer()
	if err != nil {
		panic(err)
	}
	if err := server.Serve(":8080"); err != nil {
		panic(err)
	}
}
