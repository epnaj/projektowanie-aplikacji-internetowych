package web

import (
	"net/http"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := h.users.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := h.users.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if err := h.sessions.Save(w, r, user.Id); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessions.Clear(w, r); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	user, err := h.users.Get(r.Context(), uid)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}
