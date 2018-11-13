package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/crash", func(w http.ResponseWriter, r *http.Request) {
		myVal := fib(5000, true)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strconv.Itoa(myVal)))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		myVal := fib(5000, false)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strconv.Itoa(myVal)))
	})

	log.Println("server starting")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func fib(n int, crash bool) int {
	result := n
	if n > 0 {
		result += fib(n-1, false)
	} else {
		result = 1
	}

	if result > 90 {
		time.Sleep(50 * time.Millisecond)
	}

	if crash && result > 9000 {
		panic("it's over 9000")
	}

	return result
}
