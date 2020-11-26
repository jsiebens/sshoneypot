package main

import (
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
)

func main() {
	cityDb, err := geoip2.Open("/home/jsiebens/Downloads/GeoLite2-City.mmdb")

	if err != nil {
		log.Fatal(err)
	}
	defer cityDb.Close()

	funcName(err, cityDb, net.ParseIP("91.183.51.235"))
	funcName(err, cityDb, net.ParseIP("165.22.195.238"))
	funcName(err, cityDb, net.ParseIP("165.232.108.199"))
}

func funcName(err error, cityDb *geoip2.Reader, ip net.IP) {
	city, err := cityDb.City(ip)

	fmt.Println(city.City.Names)
	fmt.Println(city.Country.Names)
	fmt.Println(city.Location.AccuracyRadius)
	fmt.Println(city.Location.Latitude)
	fmt.Println(city.Location.Longitude)
	fmt.Println("=========================")
}
