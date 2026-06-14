package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// List of realistic Indian pincodes representing different states and regions
var pincodes = []uint32{
	110001, // New Delhi
	400001, // Mumbai
	560001, // Bengaluru
	600001, // Chennai
	700001, // Kolkata
	500001, // Hyderabad
	380001, // Ahmedabad
	411001, // Pune
	522412, // Narasaraopet
	500081, // Madhapur
	560037, // Marathahalli
	560103, // Outer Ring Road
	560068, // Bommanahalli
	600096, // Perungudi
	110092, // Nirman Vihar
	400051, // Bandra East
	201301, // Noida
	122001, // Gurgaon
	302001, // Jaipur
	682001, // Kochi
	570001, // Mysuru
	580001, // Dharwad
	590001, // Belagavi
	530001, // Visakhapatnam
	520001, // Vijayawada
	524001, // Nellore
	517501, // Tirupati
	751001, // Bhubaneswar
	800001, // Patna
	226001, // Lucknow
	452001, // Indore
	395001, // Surat
	641001, // Coimbatore
	695001, // Trivandrum
	141001, // Ludhiana
	143001, // Amritsar
	160017, // Chandigarh
	171001, // Shimla
	190001, // Srinagar
	781001, // Guwahati
	
}

type Stats struct {
	TotalReqs    int64
	SuccessReqs  int64
	FailedReqs   int64
	Status200    int64
	Status404    int64
	StatusOther  int64
	TotalLatency int64 // in microseconds
	MinLatency   int64 // in microseconds
	MaxLatency   int64 // in microseconds
}

func main() {
	targetURL := flag.String("url", "https://go-pincode-service.onrender.com", "Target base URL of the pincode service")
	concurrency := flag.Int("c", 5, "Number of concurrent worker goroutines")
	duration := flag.Duration("d", 10*time.Second, "Duration of the test (e.g. 5s, 10s, 30s)")
	flag.Parse()

	fmt.Printf("==================================================\n")
	fmt.Printf("Starting Load Test\n")
	fmt.Printf("Target:      %s\n", *targetURL)
	fmt.Printf("Concurrency: %d workers\n", *concurrency)
	fmt.Printf("Duration:    %v\n", *duration)
	fmt.Printf("==================================================\n")

	stats := &Stats{
		MinLatency: 9999999999,
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	// Set random seed
	rand.Seed(time.Now().UnixNano())

	startTime := time.Now()

	// Launch worker goroutines
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					pin := pincodes[rand.Intn(len(pincodes))]
					url := fmt.Sprintf("%s/pincode/%d", *targetURL, pin)

					reqStart := time.Now()
					resp, err := client.Get(url)
					latency := time.Since(reqStart).Microseconds()

					atomic.AddInt64(&stats.TotalReqs, 1)

					// Update latency stats
					if latency < atomic.LoadInt64(&stats.MinLatency) {
						atomic.StoreInt64(&stats.MinLatency, latency)
					}
					if latency > atomic.LoadInt64(&stats.MaxLatency) {
						atomic.StoreInt64(&stats.MaxLatency, latency)
					}
					atomic.AddInt64(&stats.TotalLatency, latency)

					if err != nil {
						atomic.AddInt64(&stats.FailedReqs, 1)
						continue
					}

					atomic.AddInt64(&stats.SuccessReqs, 1)
					if resp.StatusCode == http.StatusOK {
						atomic.AddInt64(&stats.Status200, 1)
					} else if resp.StatusCode == http.StatusNotFound {
						atomic.AddInt64(&stats.Status404, 1)
					} else {
						atomic.AddInt64(&stats.StatusOther, 1)
					}
					resp.Body.Close()
				}
			}
		}(i)
	}

	// Run for the specified duration
	time.Sleep(*duration)
	close(stopChan)
	wg.Wait()

	totalTime := time.Since(startTime)

	// Display results
	totalReqs := atomic.LoadInt64(&stats.TotalReqs)
	successReqs := atomic.LoadInt64(&stats.SuccessReqs)
	failedReqs := atomic.LoadInt64(&stats.FailedReqs)
	status200 := atomic.LoadInt64(&stats.Status200)
	status404 := atomic.LoadInt64(&stats.Status404)
	statusOther := atomic.LoadInt64(&stats.StatusOther)
	totalLatency := atomic.LoadInt64(&stats.TotalLatency)
	minLatency := atomic.LoadInt64(&stats.MinLatency)
	maxLatency := atomic.LoadInt64(&stats.MaxLatency)

	if minLatency == 9999999999 {
		minLatency = 0
	}

	avgLatency := float64(0)
	if totalReqs > 0 {
		avgLatency = float64(totalLatency) / float64(totalReqs)
	}

	rps := float64(totalReqs) / totalTime.Seconds()

	fmt.Printf("\nTest Results:\n")
	fmt.Printf("--------------------------------------------------\n")
	fmt.Printf("Elapsed Time:         %v\n", totalTime)
	fmt.Printf("Total Requests:       %d\n", totalReqs)
	if totalReqs > 0 {
		fmt.Printf("Successful Requests:  %d (%.2f%%)\n", successReqs, float64(successReqs)/float64(totalReqs)*100)
		fmt.Printf("Failed Requests:      %d (%.2f%%)\n", failedReqs, float64(failedReqs)/float64(totalReqs)*100)
	} else {
		fmt.Printf("Successful Requests:  0 (0.00%)\n")
		fmt.Printf("Failed Requests:      0 (0.00%)\n")
	}
	fmt.Printf("Requests/Sec (RPS):   %.2f reqs/sec\n", rps)
	fmt.Printf("\nLatency Metrics:\n")
	fmt.Printf("--------------------------------------------------\n")
	fmt.Printf("Average Latency:      %.2f ms\n", avgLatency/1000.0)
	fmt.Printf("Min Latency:          %.2f ms\n", float64(minLatency)/1000.0)
	fmt.Printf("Max Latency:          %.2f ms\n", float64(maxLatency)/1000.0)
	fmt.Printf("\nResponse Codes:\n")
	fmt.Printf("--------------------------------------------------\n")
	fmt.Printf("200 OK:               %d\n", status200)
	fmt.Printf("404 Not Found:        %d\n", status404)
	fmt.Printf("Other Status Codes:   %d\n", statusOther)
	fmt.Printf("==================================================\n")
}
