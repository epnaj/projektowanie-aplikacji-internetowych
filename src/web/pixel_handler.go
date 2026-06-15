package web

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// transparentPixel is a 1x1 transparent GIF89a, the smallest valid image we can
// return to the embedding page.
var transparentPixel = mustDecodeBase64("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")

func mustDecodeBase64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// trackPixel resolves the link hash, records a hit, and always returns the
// pixel. It never reports whether the hash exists or is active: a tracking
// pixel must not break the host page, and returning a uniform 200 stops a
// caller from probing which hashes are valid. Recording is best-effort; only
// genuinely unexpected errors are logged.
func (h *Handler) trackPixel(w http.ResponseWriter, r *http.Request) {
	hash := strings.TrimSuffix(r.PathValue("hash"), ".gif")

	if err := h.stats.RecordHit(r.Context(), hash, time.Now()); err != nil {
		if !errors.Is(err, core.ErrNotFound) && !errors.Is(err, core.ErrLinkInactive) {
			slog.Error("record hit", "err", err)
		}
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(transparentPixel); err != nil {
		slog.Error("write pixel", "err", err)
	}
}
