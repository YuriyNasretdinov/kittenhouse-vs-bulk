package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var (
	delay = flag.Duration("delay-per-byte", 100*time.Nanosecond, "how long to wait after each byte read")
)

func main() {
	flag.Parse()

	log.Printf("Processing delay: %s per byte", *delay)

	start := time.Now()
	totalInsertedBytes := int64(0)
	insertPrefix := `INSERT INTO test(a,b,c) VALUES`
	insertBodyPrefix := []byte(`INSERT INTO test (a,b,c) VALUES`)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var buf [1024]byte

		recognised := false
		first := true
		sz := 0
		for {
			n, err := r.Body.Read(buf[:])
			if first {
				first = false
				if bytes.HasPrefix(buf[0:n], insertBodyPrefix) {
					n -= len(insertBodyPrefix)
					recognised = true
				}
			}

			sz += n
			if err != nil {
				break
			}
		}

		// wait to simulate query processing time by ClickHouse
		time.Sleep((*delay) * time.Duration(sz))

		if !recognised {
			r.ParseForm()
			if q := r.Form.Get("query"); !strings.HasPrefix(q, insertPrefix) {
				log.Printf("Unknown insert query: %s", q)
				return
			}
		}

		atomic.AddInt64(&totalInsertedBytes, int64(sz))
		totalMB := float64(atomic.LoadInt64(&totalInsertedBytes)) / 1e6
		elapsedSec := float64(time.Since(start)) / float64(time.Second)
		log.Printf("INSERT: %db, total: %.2f MB, avg %.2f MB/sec", sz, totalMB, totalMB/elapsedSec)
	})

	log.Fatal(http.ListenAndServe(":8123", nil))
}
