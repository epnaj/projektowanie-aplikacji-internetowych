package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

// Client is a black-box HTTP client for the pixel-tracker API. It knows nothing
// about the server's internals: the request and response shapes below are
// hand-declared from the public wire contract, not imported from the app. A
// cookie jar carries the session cookie returned by /api/login, exactly as a
// browser would.
type Client struct {
	base string
	http *http.Client
}

func NewClient(base string, timeout time.Duration) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		base: strings.TrimRight(base, "/"),
		http: &http.Client{Jar: jar, Timeout: timeout},
	}, nil
}

// Wire types: the subset of the public API this tool uses. Field names mirror
// the JSON the server speaks; they are intentionally duplicated here so the
// tool stays decoupled from the server's own DTOs.

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type nameRequest struct {
	Name string `json:"name"`
}

type activeRequest struct {
	Active bool `json:"active"`
}

type idResponse struct {
	Id uint64 `json:"id"`
}

type linkResponse struct {
	Id       uint64 `json:"id"`
	LinkHash string `json:"linkHash"`
}

// ErrExists is returned by Register when the account already exists (HTTP 409),
// so re-running the seeder tops up data instead of failing.
var ErrExists = errors.New("account already exists")

// httpError reports an unexpected status from the API.
type httpError struct {
	Method string
	Path   string
	Status int
	Want   int
	Body   string
}

func (e *httpError) Error() string {
	msg := fmt.Sprintf("%s %s: got %d, want %d", e.Method, e.Path, e.Status, e.Want)
	if e.Body != "" {
		msg += ": " + e.Body
	}
	return msg
}

// Register creates an account. A 409 is reported as ErrExists rather than a
// hard error.
func (c *Client) Register(ctx context.Context, email, password string) error {
	err := c.do(ctx, http.MethodPost, "/api/register", credentials{email, password}, nil, http.StatusCreated)
	var he *httpError
	if errors.As(err, &he) && he.Status == http.StatusConflict {
		return ErrExists
	}
	return err
}

// Login authenticates and stores the session cookie in the jar.
func (c *Client) Login(ctx context.Context, email, password string) error {
	return c.do(ctx, http.MethodPost, "/api/login", credentials{email, password}, nil, http.StatusOK)
}

// CreateProject creates a project and returns its id.
func (c *Client) CreateProject(ctx context.Context, name string) (uint64, error) {
	var out idResponse
	if err := c.do(ctx, http.MethodPost, "/api/projects", nameRequest{name}, &out, http.StatusCreated); err != nil {
		return 0, err
	}
	return out.Id, nil
}

// CreateLink creates a tracking link under a project and returns its id and the
// public hash used to build the pixel URL.
func (c *Client) CreateLink(ctx context.Context, projectID uint64, name string) (uint64, string, error) {
	var out linkResponse
	path := "/api/projects/" + strconv.FormatUint(projectID, 10) + "/links"
	if err := c.do(ctx, http.MethodPost, path, nameRequest{name}, &out, http.StatusCreated); err != nil {
		return 0, "", err
	}
	return out.Id, out.LinkHash, nil
}

// SetLinkActive toggles a link's active flag through the public API.
func (c *Client) SetLinkActive(ctx context.Context, linkID uint64, active bool) error {
	path := "/api/links/" + strconv.FormatUint(linkID, 10)
	return c.do(ctx, http.MethodPatch, path, activeRequest{active}, nil, http.StatusOK)
}

// Hit fetches the tracking pixel, the same request a third-party page makes when
// it loads the image. The endpoint always returns 200.
func (c *Client) Hit(ctx context.Context, hash string) error {
	return c.do(ctx, http.MethodGet, "/pixel/"+hash, nil, nil, http.StatusOK)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any, want int) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != want {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return &httpError{
			Method: method,
			Path:   path,
			Status: resp.StatusCode,
			Want:   want,
			Body:   strings.TrimSpace(string(snippet)),
		}
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
