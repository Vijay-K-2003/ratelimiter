package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Vijay-K-2003/ratelimiter/internal/ratelimiter_custom"
)

func TestServer(t *testing.T) {
	maxRequestsPerMinute := 10 // Test limit

	rl := ratelimiter_custom.RateLimiter{M: make(map[string]*ratelimiter_custom.RequestCounter)}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.HandleRequest(w, r, maxRequestsPerMinute)
	}))
	defer srv.Close()

	/*
		Send multiple requests
	*/
	for i := 0; i < maxRequestsPerMinute+1; i++ {
		resp, err := http.Get(srv.URL)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		if i < maxRequestsPerMinute {
			/*
				First N successful requests
			*/
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Unexpected status code: %d", resp.StatusCode)
			}
			var msg map[string]string
			err = json.NewDecoder(resp.Body).Decode(&msg)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			if msg["message"] != "Hello, Gopher!" {
				t.Errorf("Unexpected message: %s", msg["message"])
			}
		} else {
			/*
				Error on next request 429
			*/
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Errorf("Unexpected status code: %d", resp.StatusCode)
			}
			var errResp map[string]string
			err = json.NewDecoder(resp.Body).Decode(&errResp)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			if errResp["error"] != "Too Many Requests" {
				t.Errorf("Unexpected error message: %s", errResp["error"])
			}
		}
		resp.Body.Close()
	}

	/*
		Wait for backoff time
	*/
	time.Sleep(ratelimiter_custom.ExponentialBackoff(maxRequestsPerMinute))

	/*
		Try to check if rate limit is removed
	*/
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
	var msg map[string]string
	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if msg["message"] != "Hello, Gopher!" {
		t.Errorf("Unexpected message: %s", msg["message"])
	}
	resp.Body.Close()
}
