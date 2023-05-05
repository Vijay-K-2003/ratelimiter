package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/Vijay-K-2003/ratelimiter/internal/ratelimiter_custom"
)

func main() {
	maxRequestsPerMinute := flag.Int("max-requests-per-minute", 10, "maximum number of requests per minute per IP address")
	listenAddr := flag.String("listen", ":8080", "listening address for the server")

	flag.Parse()
	mp := make(map[string]*ratelimiter_custom.RequestCounter)
	rl := ratelimiter_custom.RateLimiter{M: mp}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rl.HandleRequest(w, r, *maxRequestsPerMinute)
	})

	fmt.Printf("Listening on %s, Rate Limit of %d requests per minute \n", *listenAddr, *maxRequestsPerMinute)
	err := http.ListenAndServe(*listenAddr, nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
