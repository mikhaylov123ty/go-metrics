package api

import (
	"net/http"

	"metrics/internal/storage"
)

type Handlers struct {
	repo storage.Storage
}

func NewHandlers(repo storage.Storage) *Handlers {
	return &Handlers{repo: repo}
}

func (h Handlers) Update(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	query := &storage.Data{
		Type:  req.PathValue("type"),
		Name:  req.PathValue("name"),
		Value: req.PathValue("value"),
	}

	if err := h.repo.Create(query); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
