package helpers

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type Result struct {
	PingedAt  time.Time
	Address   *net.IPAddr
	Duration  time.Duration
	City      string
	LocalTime *time.Time
}

type Query struct {
	Address  string
	Interval time.Duration
	Results  []Result
}

func QueryAddress(query *Query) (*Result, error) {
	dest, duration, err := Ping(query.Address)
	if err != nil {
		panic(err)
	}
	result := Result{
		PingedAt: time.Now(),
		Address:  dest,
		Duration: duration,
	}
	location, err := geolocate(dest)
	if err == nil {
		result.City = location.City.Names["en"]
		localTime, err := getLocalTime(location.Location.TimeZone)
		if err != nil {
			return nil, err
		}
		result.LocalTime = localTime
	}

	return &result, nil
}

func MonitorAddress(query *Query) error {
	log.Printf("starting monitor of %s with interval %v\n", query.Address, query.Interval)
	for {
		result, err := QueryAddress(query)
		if err != nil {
			return err
		}
		query.Results = append(query.Results, *result)
		displayTime := fmt.Sprintf("%02d:%02d", result.LocalTime.Hour(), result.LocalTime.Minute())
		log.Printf("Ping %s (%s @ %s): %s, average: %v\n", query.Address, result.City, displayTime, result.Duration, averageResponseTime(query.Results))
		time.Sleep(query.Interval)
	}
}

func averageResponseTime(results []Result) float64 {
	total := 0 * time.Second
	for _, result := range results {
		total += result.Duration
	}

	return (float64(total) / float64(len(results))) / float64(time.Millisecond)
}

func geolocate(address *net.IPAddr) (*geoip2.City, error) {
	geomindDbPath := os.Getenv("GEOMIND_DATABASE")
	if _, err := os.Stat(geomindDbPath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("skipping location lookup as no GEOMIND_DATABASE supplied")
	}
	data, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		return nil, err
	}
	defer data.Close()
	ip := net.ParseIP(address.String())
	record, err := data.City(ip)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func getLocalTime(timezone string) (*time.Time, error) {
	locat, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	now := time.Now().In(locat)

	return &now, nil
}
