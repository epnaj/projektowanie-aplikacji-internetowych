package core

import "errors"

// Errors are to be collapsed in dto in order to keep the
// information away form malicious users
var (
	ErrNotFound = errors.New("not found") // 404
	// ErrNotOwner: the resource exists but belongs to another user
	ErrNotOwner           = errors.New("not owner")           // 404
	ErrConflict           = errors.New("already exists")      // 409
	ErrLinkInactive       = errors.New("link inactive")       // 404 or 410 (gone)
	ErrInvalidCredentials = errors.New("invalid credentials") // 401
)
