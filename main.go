package main

import (
	"github.com/gliderlabs/ssh"
	"io"
	"log"
)

func authHandler(ctx ssh.Context, password string) bool {
	log.Printf("User: %s connecting from %s with password: %s\n", ctx.User(), ctx.RemoteAddr(), password)
	return true
}

func sessionHandler(s ssh.Session) {
	_, _ = io.WriteString(s, "\nWelcome and goodbye!\n\n")
}

func main() {

	s := &ssh.Server{
		Addr:            ":2222",
		Handler:         sessionHandler,
		PasswordHandler: authHandler,
	}
	
	log.Println("starting ssh server on port 2222...")
	log.Fatal(s.ListenAndServe())
}
