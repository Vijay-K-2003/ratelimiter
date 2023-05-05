package ratelimiter_custom

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

/*
RateLimiter struct for a request counter and Mutex lock for thread safety and concurrency
*/
type RateLimiter struct {
	M map[string]*RequestCounter
	sync.Mutex
}

type RequestCounter struct {
	count     int
	timestamp time.Time
}

func (rl *RateLimiter) addRequest(ip string, maxRequestsPerMinute int) bool {
	rl.Lock()
	defer rl.Unlock()

	if counter, ok := rl.M[ip]; ok {
		if time.Since(counter.timestamp) > time.Minute {
			counter.count = 1
			counter.timestamp = time.Now()
			return true
		} else if counter.count >= maxRequestsPerMinute {
			return false
		} else {
			counter.count++
			return true
		}
	} else {
		rl.M[ip] = &RequestCounter{1, time.Now()}
		return true
	}
}

func (rl *RateLimiter) HandleRequest(w http.ResponseWriter, r *http.Request, maxRequestsPerMinute int) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var (
		success bool
		n       int
	)

	backoffChan := make(chan bool)

	/*
		Goroutine to wait for backoff time to expire
	*/
	go func() {
		for {
			select {
			case <-time.After(ExponentialBackoff(n)):
				backoffChan <- true
				return
			case <-r.Context().Done():
				/*
					Request cancelled so stop waiting
				*/
				return
			}
		}
	}()

	for !success {
		success = rl.addRequest(ip, maxRequestsPerMinute)
		if !success {
			if n > 10 {
				/*
					Retry > 10 : Return error
				*/
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{"error": "Too Many Requests"})
				return
			}

			/*
				Wait for backoff time to expire or request to be cancelled
			*/
			select {
			case <-backoffChan:
				n++
				if ExponentialBackoff(n) >= 10*time.Second {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					json.NewEncoder(w).Encode(map[string]string{"error": "Too Many Requests"})
					return
				}
				backoffChan = make(chan bool)
				go func() {
					for {
						select {
						case <-time.After(ExponentialBackoff(n)):
							backoffChan <- true
							return
						case <-r.Context().Done():
							return
						}
					}
				}()
			case <-r.Context().Done():
				/*
					request cancelled so stop waiting
				*/
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, Gopher!"})
}
