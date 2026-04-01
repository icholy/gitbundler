package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sync/semaphore"
)

type Bundler struct {
	Name       string
	URL        string
	Interval   time.Duration
	RepoPath   string
	BundlePath string
	Sem        *semaphore.Weighted
}

// Run clones the repo (if needed), then loops fetching and re-bundling.
func (b *Bundler) Run(ctx context.Context) error {
	for {
		if err := b.sync(ctx); err != nil {
			slog.Error("sync failed", "name", b.Name, "err", err)
		}
		select {
		case <-time.After(b.Interval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// sync pulls changes from the repo and updates the bundle
func (b *Bundler) sync(ctx context.Context) error {
	if b.Sem != nil {
		if err := b.Sem.Acquire(ctx, 1); err != nil {
			return err
		}
		defer b.Sem.Release(1)
	}
	if _, err := os.Stat(b.RepoPath); os.IsNotExist(err) {
		slog.Info("cloning", "name", b.Name, "url", b.URL)
		if err := b.clone(ctx); err != nil {
			return fmt.Errorf("clone: %w", err)
		}
		slog.Info("clone complete", "name", b.Name)
	} else {
		slog.Info("fetching", "name", b.Name)
		if err := b.fetch(ctx); err != nil {
			return fmt.Errorf("fetch: %w", err)
		}
	}
	slog.Info("bundling", "name", b.Name)
	if err := b.bundle(ctx); err != nil {
		return fmt.Errorf("bundle: %w", err)
	}
	slog.Info("bundle complete", "name", b.Name)
	return nil
}

// clone a bare repo
func (b *Bundler) clone(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(b.RepoPath), 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "clone", "--bare", b.URL, b.RepoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// fetch updates for a repo
func (b *Bundler) fetch(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "-C", b.RepoPath, "fetch", "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// bundle updates the git bundle in the data directory
func (b *Bundler) bundle(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(b.BundlePath), 0o755); err != nil {
		return err
	}
	tmpPath, err := filepath.Abs(b.BundlePath + ".tmp")
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "-C", b.RepoPath, "bundle", "create", tmpPath, "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, b.BundlePath)
}
