package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	kittenhouse = flag.Bool("kittenhouse", false, "Whether or not to send INSERTs in kittenhouse format")
	persistent  = flag.Bool("persistent", false, "(for kittenhouse) Whether or not use persistent mode")
	addr        = flag.String("addr", "127.0.0.1:8124", "Address of bulk inserter")
)

func main() {
	flag.Parse()

	table := `test (a,b,c)`
	if *kittenhouse {
		table = `test(a,b,c)`
	}

	const N = 200000
	// const N = 1
	const REPEAT = 5000

	postURL := fmt.Sprintf("http://%s/", *addr)
	if *kittenhouse {
		postURL += "?table=" + table
		if *persistent {
			postURL += "&persistent=1"
		}
	}

	bodyStr := `('` + strings.Repeat("a", REPEAT) +
		`','` + strings.Repeat("b", REPEAT) +
		`','` + strings.Repeat("c", REPEAT) + `')`
	dataLen := len(bodyStr)

	if !*kittenhouse {
		bodyStr = `INSERT INTO ` + table + ` VALUES ` + bodyStr
	}

	body := []byte(bodyStr)
	inserted := 0
	start := time.Now()

	for i := 0; i < N; i++ {
		resp, err := http.Post(postURL, "application/x-www-form-urlencoded", bytes.NewReader(body))
		if err != nil {
			log.Fatalf("Could not POST: %v", err)
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()

		// log.Printf("Inserting %s", bodyStr)

		inserted += dataLen

		if i%1000 == 0 {
			total := float64(inserted) / 1e6
			timePassed := float64(time.Since(start)) / float64(time.Second)
			log.Printf("Inserted %d/%d (%.2f MB, avg %.2f MB/sec)", i, N, total, total/timePassed)
		}
	}

	log.Printf("Inserter inserted %.2f MB total", float64(inserted)/1e6)
}
