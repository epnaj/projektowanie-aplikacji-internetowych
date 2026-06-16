package logging

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// RotatingWriter is a minimal, dependency-free size-based rotating file writer.
// When a write would push the active file past maxBytes it rotates:
// app.log -> app.log.1 -> app.log.2 ... up to maxBackups, oldest discarded.
// It is safe for concurrent use, so it can back an slog handler directly.
type RotatingWriter struct {
	mu         sync.Mutex
	path       string
	maxBytes   int64
	maxBackups int
	file       *os.File
	size       int64
}

func NewRotatingWriter(path string, maxBytes int64, maxBackups int) (*RotatingWriter, error) {
	if maxBytes <= 0 {
		maxBytes = 10 << 20
	}
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	w := &RotatingWriter{path: path, maxBytes: maxBytes, maxBackups: maxBackups}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *RotatingWriter) open() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	w.file = f
	w.size = info.Size()
	return nil
}

func (w *RotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Rotate before writing, but never on an empty file: a single line larger
	// than maxBytes still goes through rather than looping forever.
	if w.size > 0 && w.size+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// rotate closes the active file, shifts the numbered backups up by one
// (dropping any beyond maxBackups), moves the active file to .1, and reopens a
// fresh active file.
func (w *RotatingWriter) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	if w.maxBackups <= 0 {
		_ = os.Remove(w.path)
		return w.open()
	}
	_ = os.Remove(w.backupName(w.maxBackups))
	for i := w.maxBackups - 1; i >= 1; i-- {
		_ = os.Rename(w.backupName(i), w.backupName(i+1))
	}
	_ = os.Rename(w.path, w.backupName(1))
	return w.open()
}

func (w *RotatingWriter) backupName(i int) string {
	return w.path + "." + strconv.Itoa(i)
}

func (w *RotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}
