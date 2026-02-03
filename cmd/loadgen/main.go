//use go tools that keep persistent connection so the application is stressed, not the network

package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	Total   uint64
	Success uint64
	Fail    uint64
}

func (s *Stats) Get() (total, success, fail uint64) {
	return atomic.LoadUint64(&s.Total), atomic.LoadUint64(&s.Success), atomic.LoadUint64(&s.Fail)
}

func (s *Stats) IncTotal() {
	atomic.AddUint64(&s.Total, 1)
}

func (s *Stats) IncSuccess() {
	atomic.AddUint64(&s.Success, 1)
}

func (s *Stats) IncFail() {
	atomic.AddUint64(&s.Fail, 1)

}

func main() {
	// Use GOMAXPROCS * 20 for I/O heavy workload (HTTP requests)
	numWorkers := runtime.GOMAXPROCS(0) * 20
	fmt.Printf("Running with %d workers (GOMAXPROCS * 20)...\n", numWorkers)

	s := &Stats{}

	//custom transport, as making a new client in the function everytime would
	//Default max idle conns is 100, but per host is 2, we need it to be more to reuse conns and avoid the handshake overhead each time
	transport := http.Transport{
		MaxIdleConnsPerHost: numWorkers, //Default is 2, should be >=numWorkers
	}
	client := http.Client{
		Transport: &transport,
		Timeout:   5 * time.Second,
	}
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			worker(s, id, &client)
			wg.Done()
		}(i)
	}

	go reporter(s, os.Args[1])

	wg.Wait()

}

func worker(s *Stats, workerId int, client *http.Client) {
	//give each worker its own random generator
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerId)))
	for {
		//generate a random number
		url := generateTarget(r)
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		if resp.StatusCode == 200 {
			s.IncSuccess()
		}
		if resp.StatusCode == 404 {
			s.IncFail()
		}
		s.IncTotal()

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		// time.Sleep(time.Millisecond * 10)
	}
}

func generateTarget(r *rand.Rand) string {
	num := r.Int()
	validUrl := fmt.Sprintf("http://localhost:8080/check?id=user_%d", r.Intn(1000000))
	maliciousAttacker := fmt.Sprintf("http://localhost:8080/check?id=random_%d", r.Intn(1000000))
	if num%2 == 0 {
		return validUrl
	} else {
		return maliciousAttacker
	}
}

func reporter(s *Stats, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer f.Close()

	// Write header
	f.WriteString("Time,RPS\n")

	ticker := time.NewTicker(1 * time.Second)
	var prevTotal uint64
	startTime := time.Now()

	//TICKER.C -> channel on which tickes are delivered
	for range ticker.C {
		total, success, fail := s.Get()

		//req per secon
		reqPerSecond := total - prevTotal
		prevTotal = total

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("\rRps: %d | Total: %d | Success: %d | Fail(or Blocked): %d", reqPerSecond, total, success, fail)

		// Write to CSV
		f.WriteString(fmt.Sprintf("%.1f,%d\n", elapsed, reqPerSecond))
	}
	time.Sleep(time.Second)
}
