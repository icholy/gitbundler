package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Bundler struct {
	Name     string
	URL      string
	Interval time.Duration
	DataDir  string
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

func (b *Bundler) sync(ctx context.Context, repoDir, bundlePath string) error {
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		slog.Info("cloning", "name", b.Name, "url", b.URL)
		if err := b.gitClone(ctx, repoDir); err != nil {
			return fmt.Errorf("clone: %w", err)
		}
		slog.Info("clone complete", "name", b.Name)
	} else {
		slog.Info("fetching", "name", b.Name)
		if err := b.gitFetch(ctx, repoDir); err != nil {
			return fmt.Errorf("fetch: %w", err)
		}
	}
	slog.Info("bundling", "name", b.Name)
	if err := b.createBundle(ctx, repoDir, bundlePath); err != nil {
		return fmt.Errorf("bundle: %w", err)
	}
	slog.Info("bundle complete", "name", b.Name)
	return nil
}

func (b *Bundler) gitClone(ctx context.Context, repoDir string) error {
	if err := os.MkdirAll(filepath.Dir(repoDir), 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "clone", "--bare", b.URL, repoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (b *Bundler) gitFetch(ctx context.Context, repoDir string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "fetch", "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (b *Bundler) createBundle(ctx context.Context, repoDir, bundlePath string) error {
	if err := os.MkdirAll(filepath.Dir(bundlePath), 0o755); err != nil {
		return err
	}
	tmpFile, err := filepath.Abs(bundlePath + ".tmp")
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "bundle", "create", tmpFile, "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Remove(tmpFile)
		return err
	}
	return os.Rename(tmpFile, bundlePath)
}
