package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotatingWriterRotatesAndCapsBackups(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	// maxBytes small so each ~10-byte line forces a rotation; keep 2 backups.
	w, err := NewRotatingWriter(path, 8, 2)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer w.Close()

	for i := 0; i < 6; i++ {
		if _, err := w.Write([]byte("0123456789\n")); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	// Active file plus at most maxBackups (.1, .2) must exist; .3 must not.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("active log missing: %v", err)
	}
	for _, suffix := range []string{".1", ".2"} {
		if _, err := os.Stat(path + suffix); err != nil {
			t.Fatalf("backup %s missing: %v", suffix, err)
		}
	}
	if _, err := os.Stat(path + ".3"); !os.IsNotExist(err) {
		t.Fatalf("backup .3 should not exist (maxBackups=2)")
	}
}

func TestRotatingWriterReopenAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	w, err := NewRotatingWriter(path, 1<<20, 1)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if _, err := w.Write([]byte("first\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	w.Close()

	// Reopening the same path must append, not truncate.
	w2, err := NewRotatingWriter(path, 1<<20, 1)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer w2.Close()
	if _, err := w2.Write([]byte("second\n")); err != nil {
		t.Fatalf("write 2: %v", err)
	}

	data, _ := os.ReadFile(path)
	if got := string(data); !strings.Contains(got, "first") || !strings.Contains(got, "second") {
		t.Fatalf("expected both lines, got %q", got)
	}
}
