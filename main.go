package main

import (
	"github.com/gliderlabs/ssh"
	"github.com/mmcloughlin/geohash"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"sync"
)

func authHandler(c *prometheus.CounterVec) ssh.PasswordHandler {
	return func(ctx ssh.Context, password string) bool {

		geo, err := ipToGeo(ctx.RemoteAddr())

		if err == nil {
			c.WithLabelValues(
				geo.CountryCode,
				geo.City,
				geohash.Encode(geo.Lat, geo.Lon),
			).Inc()
		}

		return false
	}
}

func sessionHandler(s ssh.Session) {
	_, _ = io.WriteString(s, "\nWelcome and goodbye!\n\n")
}

func main() {

	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ssh",
			Subsystem: "honeypot",
			Name:      "sessions_total",
			Help:      "A counter for sessions to the honeypot.",
		},
		[]string{
			"country",
			"city",
			"geohash",
		},
	)

	s := &ssh.Server{
		Addr:            ":2222",
		Handler:         sessionHandler,
		PasswordHandler: authHandler(counter),
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("starting ssh server on port 2222...")
		if err := s.ListenAndServe(); err != nil {
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

	wg.Wait()
}
