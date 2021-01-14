package main

import (
	"github.com/gliderlabs/ssh"
	"github.com/mmcloughlin/geohash"
	"github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

func authHandler(s *prometheus.CounterVec, db *geoip2.Reader) ssh.PasswordHandler {
	return func(ctx ssh.Context, password string) bool {
		addr := ctx.RemoteAddr()
		ip := net.ParseIP(strings.Split(addr.String(), ":")[0])

		if ip != nil {
			record, err := db.City(ip)

			if err == nil && record.City.GeoNameID != 0 {
				s.WithLabelValues(
					record.Country.Names["en"],
					record.Country.IsoCode,
					record.City.Names["en"],
					geohash.Encode(record.Location.Latitude, record.Location.Latitude),
				).Inc()
			}
		}

		return false
	}
}

func sessionHandler(s ssh.Session) {
	_, _ = io.WriteString(s, "\nWelcome and goodbye!\n\n")
}

func main() {
	db, err := geoip2.Open("./GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sessions := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "honeypot",
			Subsystem: "sessions",
			Name:      "total",
			Help:      "A sessions for sessions to the honeypot.",
		},
		[]string{
			"country",
			"code",
			"city",
			"geohash",
		},
	)

	port := getenv("PORT", "2222")

	sshServer := &ssh.Server{
		Addr:            ":" + port,
		Handler:         sessionHandler,
		PasswordHandler: authHandler(sessions, db),
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

	wg.Wait()
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
