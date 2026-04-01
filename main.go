package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"golang.org/x/sync/semaphore"
)

func main() {
	configPath := flag.String("config", "gitbundler.yaml", "path to config file")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var sem *semaphore.Weighted
	if cfg.MaxConcurrent > 0 {
		sem = semaphore.NewWeighted(int64(cfg.MaxConcurrent))
	}

	var wg sync.WaitGroup
	for _, repo := range cfg.Repos {
		b := &Bundler{
			Name:        repo.Name,
			URL:         repo.URL,
			Interval:    repo.Interval,
			RepoPath:    filepath.Join(cfg.DataDir, repo.Name+".git"),
			BundlePath:  filepath.Join(cfg.DataDir, repo.Name+".bundle"),
			Env:         repo.Env,
			Repack:      repo.Repack,
			CloneFlags:  repo.CloneFlags,
			FetchFlags:  repo.FetchFlags,
			BundleFlags: repo.BundleFlags,
			Sem:         sem,
		}
		wg.Go(func() {
			if err := b.Run(ctx); err != nil {
				slog.Error("bundler stopped", "name", b.Name, "err", err)
			}
		})
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: &Server{DataDir: cfg.DataDir},
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	slog.Info("listening", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("http server error", "err", err)
		return
	}

	stop()
	wg.Wait()
}
