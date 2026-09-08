package profiling

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"time"

	"github.com/railzwaylabs/github-actions-samples/internal/platform/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("platform.profiling", fx.Invoke(Start))

func Start(lifecycle fx.Lifecycle, cfg config.Config, log *zap.Logger) {
	if !cfg.ProfilingEnabled {
		return
	}

	server := &http.Server{
		Addr:              cfg.ProfilingAddress,
		Handler:           handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				log.Info("profiling server started", zap.String("address", server.Addr))
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("profiling server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", index)
	mux.HandleFunc("/debug/pprof/profile", cpuProfile)
	mux.HandleFunc("/debug/pprof/trace", executionTrace)
	for _, name := range []string{"allocs", "block", "goroutine", "heap", "mutex", "threadcreate"} {
		mux.HandleFunc("/debug/pprof/"+name, func(w http.ResponseWriter, r *http.Request) {
			profile := pprof.Lookup(name)
			if profile == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			debug, _ := strconv.Atoi(r.URL.Query().Get("debug"))
			if err := profile.WriteTo(w, debug); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}
	return mux
}

func index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<html><body><h1>Profiles</h1><a href="profile">profile</a><br><a href="heap">heap</a><br><a href="goroutine?debug=1">goroutine</a><br><a href="trace">trace</a></body></html>`)
}

func cpuProfile(w http.ResponseWriter, r *http.Request) {
	seconds := profileSeconds(r, 30)
	w.Header().Set("Content-Type", "application/octet-stream")
	if err := pprof.StartCPUProfile(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	timer := time.NewTimer(time.Duration(seconds) * time.Second)
	select {
	case <-timer.C:
	case <-r.Context().Done():
		if !timer.Stop() {
			<-timer.C
		}
	}
	pprof.StopCPUProfile()
}

func executionTrace(w http.ResponseWriter, r *http.Request) {
	seconds := profileSeconds(r, 1)
	w.Header().Set("Content-Type", "application/octet-stream")
	if err := trace.Start(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	timer := time.NewTimer(time.Duration(seconds) * time.Second)
	select {
	case <-timer.C:
	case <-r.Context().Done():
		if !timer.Stop() {
			<-timer.C
		}
	}
	trace.Stop()
}

func profileSeconds(r *http.Request, fallback int) int {
	seconds, err := strconv.Atoi(r.URL.Query().Get("seconds"))
	if err != nil || seconds < 1 {
		return fallback
	}
	if seconds > 300 {
		return 300
	}
	return seconds
}
