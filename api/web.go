package api

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ents-source/door-control/doors"
)

type HealthOptions struct {
	ExpectedDoors    int
	DoorOfflineAfter time.Duration
}

var srv *http.Server

func Init(addr string, static string, healthOpts HealthOptions) *sync.WaitGroup {
	wg := new(sync.WaitGroup)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if doors.Count(healthOpts.DoorOfflineAfter) != healthOpts.ExpectedDoors {
			w.WriteHeader(http.StatusExpectationFailed)
			_, _ = w.Write([]byte(http.StatusText(http.StatusExpectationFailed)))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(http.StatusText(http.StatusOK)))
		}
	})

	http.Handle("/", http.FileServer(http.Dir(static)))

	go func() {
		srv = &http.Server{Addr: addr, Handler: http.DefaultServeMux}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}

		wg.Done()
	}()

	return wg
}

func Stop() {
	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}
}
