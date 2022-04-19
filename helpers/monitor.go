package helpers

import (
	"errors"
	"fmt"
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

type MonitorResult struct {
	Result Result
	Error  error
}

func QueryAddress(resultChan chan MonitorResult, query *Query) {
	dest, duration, err := Ping(query.Address)
	if err != nil {
		resultChan <- MonitorResult{
			Error: err,
		}
		return
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
			resultChan <- MonitorResult{
				Error: err,
			}
			return
		}
		result.LocalTime = localTime
	}

	resultChan <- MonitorResult{
		Result: result,
		Error:  nil,
	}
}

func geolocate(address *net.IPAddr) (*geoip2.City, error) {
	geomindDbPath := os.Getenv("GEOMIND_DATABASE")
	if _, err := os.Stat(geomindDbPath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("skipping location lookup as no GEOMIND_DATABASE supplied")
	}
	data, err := geoip2.Open(geomindDbPath)
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
