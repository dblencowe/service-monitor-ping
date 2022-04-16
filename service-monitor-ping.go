package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dblencowe/service-monitor-ping/helpers"
)

func main() {
	var wg sync.WaitGroup
	interval := pingInterval()
	queries := buildQueries(*interval)
	if len(*queries) == 0 {
		panic("no ips provided to be monitored")
	}

	doMonitoring := os.Getenv("MONITOR")
	if len(doMonitoring) > 0 {
		for _, query := range *queries {
			wg.Add(1)
			go helpers.MonitorAddress(&query)
		}

		wg.Wait()
	} else {
		setupHttpServer(queries)
	}
}

func buildQueries(interval time.Duration) *[]helpers.Query {
	var queries []helpers.Query
	args := os.Args[1:]
	for _, address := range args {
		queries = append(queries, helpers.Query{
			Address:  address,
			Interval: interval,
			Results:  make([]helpers.Result, 0),
		})
	}

	inputFile := os.Getenv("INPUT_FILE")
	if len(inputFile) > 0 {
		file, err := os.Open(inputFile)
		if err != nil {
			panic(fmt.Sprintf("unable to open input file %s", inputFile))
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			queries = append(queries, helpers.Query{
				Address:  scanner.Text(),
				Interval: interval,
				Results:  make([]helpers.Result, 0),
			})
		}
	}

	return &queries
}

func pingInterval() *time.Duration {
	requestedInterval := os.Getenv("PING_INTERVAL")
	interval := 30 * time.Second
	if len(requestedInterval) > 0 {
		rst, err := strconv.ParseInt(requestedInterval, 6, 12)
		if err != nil {
			log.Printf("unable to parse PING_INTERVAL, setting default value of 30 seconds")
			return &interval
		}
		interval = time.Duration(rst) * time.Second
		return &interval
	}

	return &interval
}

type HttpOutput struct {
	Results *[]helpers.Result
}

func setupHttpServer(queries *[]helpers.Query) {
	enableHttpServer := os.Getenv("HTTP_ENABLE")
	if len(enableHttpServer) == 0 {
		return
	}
	httpServerPort := os.Getenv("HTTP_PORT")
	if len(httpServerPort) == 0 {
		httpServerPort = "8080"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		results := make(chan *[]helpers.Result, len(*queries))
		go queryWrapper(results, queries)
		response := HttpOutput{
			Results: <-results,
		}
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":"+httpServerPort, nil)
	log.Printf("http server listening on :%s", httpServerPort)
}

func queryWrapper(channel chan *[]helpers.Result, queries *[]helpers.Query) {
	var results []helpers.Result
	for _, query := range *queries {
		result, err := helpers.QueryAddress(&query)
		if err != nil {
			panic(err)
		}
		results = append(results, *result)
	}

	channel <- &results
}