package ip2geo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Query struct {
	Query string `json:"query"`
}

type GeoInfo struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
}

func Lookup(q []Query, counter *prometheus.CounterVec) ([]GeoInfo, error) {
	req, _ := json.Marshal(q)

	resp, err := http.Post("http://ip-api.com/batch", "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}

	counter.WithLabelValues(
		"/batch",
		strconv.Itoa(resp.StatusCode),
	).Inc()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error getting geo information [%d]", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []GeoInfo
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
