package main

import (
	"fmt"
	"net/http"
	"time"
)

type bucket struct {
	// IpAddr is the ID to identify the user to whom the bucket belongs
	IpAddr string
	// capacity is the maximum number of tokens in the bucket
	capacity int
	// remaining is the number of tokens remaining in the bucket
	remaining int
	// tokenIncrementRate is the rate at which tokens are added to the bucket per second
	tokenIncrementRate float64
	// lastRequestTime is the last time a request was made
	lastRequestTime time.Time
}

var IPAddrToBucketMapping map[string]*bucket

func validateRateLimit(bkt *bucket) bool {
	// first fill in all the necessary values for the bucket before validating whether this request should go in or not.
	// remaining should be remaining + (whatever tokens that were incremented because of rate since the last request). Note: It should not overflow the bucket's capacity.
	bkt.remaining = min(bkt.capacity, bkt.remaining+int(time.Since(bkt.lastRequestTime).Seconds()*bkt.tokenIncrementRate))
	bkt.lastRequestTime = time.Now()

	// Now validate whether this request should really go in or not
	if bkt.remaining < 1 {
		return false
	}
	bkt.remaining--
	return true
}

func main() {

	IPAddrToBucketMapping = make(map[string]*bucket)
	http.HandleFunc("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Unlimited access"))
	})
	http.HandleFunc("/limited", func(w http.ResponseWriter, r *http.Request) {
		userIPAddr := r.RemoteAddr
		fmt.Println("User IP address is : ", userIPAddr)
		if bkt, exists := IPAddrToBucketMapping[userIPAddr]; exists {
			fmt.Println("User Rate-Limit Bucket is : \n", bkt)
			allow := validateRateLimit(bkt)
			if allow {
				w.Write([]byte("Request Allowed!\n"))
			} else {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
		} else {
			// Create a new bucket for the user
			bkt := &bucket{
				IpAddr:             r.RemoteAddr,
				capacity:           10,
				remaining:          9,
				tokenIncrementRate: 1,
				lastRequestTime:    time.Now(),
			}
			IPAddrToBucketMapping[userIPAddr] = bkt
		}
		w.Write([]byte("limited access"))
	})
	http.ListenAndServe(":8080", nil)
}
