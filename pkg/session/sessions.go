package session

import (
	"github.com/jsiebens/sshoneypot/pkg/ip2geo"
	"github.com/mmcloughlin/geohash"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sync"
)

type Sessions struct {
	sync.RWMutex

	sessions    *prometheus.CounterVec
	apiRequests *prometheus.CounterVec
	values      map[string]int
}

func NewSessions() *Sessions {
	return &Sessions{
		sessions: promauto.NewCounterVec(
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
		),
		apiRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "honeypot",
				Subsystem: "api_geoip",
				Name:      "total_requests",
			},
			[]string{
				"path",
				"status",
			},
		),
		values: map[string]int{},
	}
}

func (s *Sessions) Inc(ip string) {
	s.Lock()

	defer s.Unlock()

	if val, ok := s.values[ip]; ok {
		s.values[ip] = val + 1
	} else {
		s.values[ip] = 1
	}

}

func (s *Sessions) Flush() {
	v := s.copy()

	if len(v) == 0 {
		return
	}

	var q []ip2geo.Query
	for k2, _ := range v {
		q = append(q, ip2geo.Query{Query: k2})
	}

	results, _ := ip2geo.Lookup(q, s.apiRequests)

	for _, geo := range results {
		s.sessions.WithLabelValues(
			geo.Country,
			geo.CountryCode,
			geo.City,
			geohash.Encode(geo.Lat, geo.Lon),
		).Add(float64(v[geo.Query]))
	}
}

func (s *Sessions) copy() map[string]int {
	s.Lock()
	defer s.Unlock()

	v := make(map[string]int)
	for k2, v2 := range s.values {
		v[k2] = v2
		delete(s.values, k2)
	}
	return v
}
