package main

import (
	"github.com/gliderlabs/ssh"
	"github.com/jasonlvhit/gocron"
	"github.com/jsiebens/sshoneypot/pkg/session"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func authHandler(s *session.Sessions) ssh.PasswordHandler {
	return func(ctx ssh.Context, password string) bool {
		//log.Printf("User: %s connecting from %s with password: %s\n", ctx.User(), ctx.RemoteAddr(), password)

		s.Inc(strings.Split(ctx.RemoteAddr().String(), ":")[0])

		return false
	}
}

func sessionHandler(s ssh.Session) {
	_, _ = io.WriteString(s, "\nWelcome and goodbye!\n\n")
}

func main() {

	port := getenv("PORT", "2222")

	sessions := session.NewSessions()

	sshServer := &ssh.Server{
		Addr:            ":" + port,
		Handler:         sessionHandler,
		PasswordHandler: authHandler(sessions),
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("starting ssh server on port %s...\n", port)
		if err := sshServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		promServer := http.NewServeMux()
		promServer.Handle("/metrics", promhttp.Handler())

		log.Println("starting metrics server on port 2112...")
		if err := http.ListenAndServe(":2112", promServer); err != nil {
			log.Fatal(err)
		}
	}()

	_ = gocron.Every(4).Seconds().Do(sessions.Flush)
	gocron.Start()

	wg.Wait()
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
