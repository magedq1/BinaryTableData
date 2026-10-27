package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Version is the GUI version.
const Version = "0.2.0"

// FileFilter describes a native file dialog filter.
type FileFilter struct {
	DisplayName string `json:"displayName"`
	Pattern     string `json:"pattern"`
}

// App is the Wails binding object. It owns the open table and binary-list
// sessions and exposes editor operations to the frontend.
type App struct {
	ctx context.Context

	mu     sync.Mutex
	tables map[string]*tableSession
	lists  map[string]*listSession
}

// NewApp creates the application state.
func NewApp() *App {
	return &App{
		tables: make(map[string]*tableSession),
		lists:  make(map[string]*listSession),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(context.Context) {
	a.CloseAll()
}

// CloseAll closes every open table and list. The frontend calls it when the
// window is closed; Wails also invokes it through shutdown.
func (a *App) CloseAll() {
	a.mu.Lock()
	tables := make([]*tableSession, 0, len(a.tables))
	for _, session := range a.tables {
		tables = append(tables, session)
	}
	a.tables = make(map[string]*tableSession)
	lists := make([]*listSession, 0, len(a.lists))
	for _, session := range a.lists {
		lists = append(lists, session)
	}
	a.lists = make(map[string]*listSession)
	a.mu.Unlock()

	for _, session := range tables {
		session.mu.Lock()
		_ = session.close()
		session.mu.Unlock()
	}
	for _, session := range lists {
		session.mu.Lock()
		_ = session.close()
		session.mu.Unlock()
	}
}

// Version returns the GUI version.
func (a *App) Version() string { return Version }

// PickFile opens a native single-file dialog.
func (a *App) PickFile(title string, filters []FileFilter) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application is not ready")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: runtimeFilters(filters),
	})
}

// PickFiles opens a native multi-file dialog.
func (a *App) PickFiles(title string, filters []FileFilter) ([]string, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("application is not ready")
	}
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: runtimeFilters(filters),
	})
}

// PickSaveFile opens a native save dialog.
func (a *App) PickSaveFile(title, defaultFilename string, filters []FileFilter) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application is not ready")
	}
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                title,
		DefaultFilename:      defaultFilename,
		Filters:              runtimeFilters(filters),
		CanCreateDirectories: true,
	})
}

func runtimeFilters(filters []FileFilter) []runtime.FileFilter {
	out := make([]runtime.FileFilter, 0, len(filters))
	for _, filter := range filters {
		if strings.TrimSpace(filter.Pattern) == "" {
			continue
		}
		out = append(out, runtime.FileFilter{
			DisplayName: filter.DisplayName,
			Pattern:     filter.Pattern,
		})
	}
	return out
}

func normalizePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}
