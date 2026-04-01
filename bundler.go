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
	Name     string
	URL      string
	Interval time.Duration
	DataDir  string
	Sem      *semaphore.Weighted
}

// RepoDir returns the bare repo directory path.
func (b *Bundler) RepoDir() string {
	return filepath.Join(b.DataDir, b.Name+".git")
}

// BundlePath returns the path where the bundle file is served from.
func (b *Bundler) BundlePath() string {
	return filepath.Join(b.DataDir, b.Name+".bundle")
}

// Run clones the repo (if needed), then loops fetching and re-bundling.
func (b *Bundler) Run(ctx context.Context) error {
	repoDir := b.RepoDir()
	bundlePath := b.BundlePath()

	for {
		if err := b.sync(ctx, repoDir, bundlePath); err != nil {
			slog.Error("sync failed", "name", b.Name, "err", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(b.Interval):
		}
	}
}

// sync pulls changes from the repo and updates the bundle
func (b *Bundler) sync(ctx context.Context, repoDir, bundlePath string) error {
	if b.Sem != nil {
		if err := b.Sem.Acquire(ctx, 1); err != nil {
			return err
		}
		defer b.Sem.Release(1)
	}
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		slog.Info("cloning", "name", b.Name, "url", b.URL)
		if err := b.clone(ctx, repoDir); err != nil {
			return fmt.Errorf("clone: %w", err)
		}
		slog.Info("clone complete", "name", b.Name)
	} else {
		slog.Info("fetching", "name", b.Name)
		if err := b.fetch(ctx, repoDir); err != nil {
			return fmt.Errorf("fetch: %w", err)
		}
	}
	slog.Info("bundling", "name", b.Name)
	if err := b.bundle(ctx, repoDir, bundlePath); err != nil {
		return fmt.Errorf("bundle: %w", err)
	}
	slog.Info("bundle complete", "name", b.Name)
	return nil
}

// clone a bare repo
func (b *Bundler) clone(ctx context.Context, repoDir string) error {
	if err := os.MkdirAll(filepath.Dir(repoDir), 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "clone", "--bare", b.URL, repoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// fetch updates for a repo
func (b *Bundler) fetch(ctx context.Context, repoDir string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "fetch", "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// bundle updates the git bundle in the data directory
func (b *Bundler) bundle(ctx context.Context, repoDir, bundlePath string) error {
	if err := os.MkdirAll(filepath.Dir(bundlePath), 0o755); err != nil {
		return err
	}
	tmpPath, err := filepath.Abs(bundlePath + ".tmp")
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "bundle", "create", tmpPath, "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, bundlePath)
}
