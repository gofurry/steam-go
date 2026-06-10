package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"time"

	steam "github.com/gofurry/steam-go"
)

type aggregate struct {
	mu             sync.Mutex
	byClass        map[steam.TrafficClass]int
	byStatus       map[int]int
	byErrorKind    map[string]int
	cacheHits      int
	conditionalHit int
	blockDetected  int
}

func newAggregate() *aggregate {
	return &aggregate{
		byClass:     make(map[steam.TrafficClass]int),
		byStatus:    make(map[int]int),
		byErrorKind: make(map[string]int),
	}
}

func (a *aggregate) record(event steam.RequestEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.byClass[event.TrafficClass]++
	a.byStatus[event.StatusCode]++
	if event.ErrorKind != "" {
		a.byErrorKind[event.ErrorKind]++
	}
	if event.CacheHit {
		a.cacheHits++
	}
	if event.ConditionalHit {
		a.conditionalHit++
	}
	if event.BlockDetected {
		a.blockDetected++
	}
}

func (a *aggregate) print() {
	a.mu.Lock()
	defer a.mu.Unlock()

	fmt.Println("by traffic class:")
	printClassMap(a.byClass)
	fmt.Println("by status:")
	printIntMap(a.byStatus)
	fmt.Println("by error kind:")
	printStringMap(a.byErrorKind)
	fmt.Printf("cache_hit=%d conditional_hit=%d block_detected=%d\n", a.cacheHits, a.conditionalHit, a.blockDetected)
}

func main() {
	events := make(chan steam.RequestEvent, 64)
	aggregate := newAggregate()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for event := range events {
			aggregate.record(event)
		}
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"observer-demo"`)
		if r.Header.Get("If-None-Match") == `"observer-demo"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithRequestObserver(steam.RequestObserverFunc(func(event steam.RequestEvent) {
			select {
			case events <- event:
			default:
			}
		})),
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			RateLimiter: &steam.TrafficRateLimiterPolicy{Limit: 1, Burst: 1},
			Cache:       &steam.TrafficCachePolicy{TTL: 25 * time.Millisecond},
		}),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for i := 0; i < 2; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/status", nil)
		if err != nil {
			panic(err)
		}
		if _, err := client.DoRawHTTPRequest(ctx, req, &steam.RawHTTPRequestOptions{
			TrafficClass: steam.TrafficClassPublicStorePage,
		}); err != nil {
			panic(err)
		}
		time.Sleep(40 * time.Millisecond)
	}

	close(events)
	<-done
	aggregate.print()
}

func printClassMap(values map[steam.TrafficClass]int) {
	keys := make([]string, 0, len(values))
	lookup := make(map[string]int, len(values))
	for key, value := range values {
		text := string(key)
		keys = append(keys, text)
		lookup[text] = value
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("  %s=%d\n", key, lookup[key])
	}
}

func printIntMap(values map[int]int) {
	keys := make([]int, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	for _, key := range keys {
		fmt.Printf("  %d=%d\n", key, values[key])
	}
}

func printStringMap(values map[string]int) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("  %s=%d\n", key, values[key])
	}
}
