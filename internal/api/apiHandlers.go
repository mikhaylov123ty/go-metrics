package api

import (
	"log"
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

	query := &storage.Data{}

	if err := query.NewData(
		req.PathValue("type"),
		req.PathValue("name"),
		req.PathValue("value"),
	); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := query.UniqueID()

	if query.Type == "counter" {
		prev, err := h.repo.Read(id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if prev != nil {
			query.Value = prev.Value.(int64) + query.Value.(int64)
		}
	}

	if err := h.repo.Create(query); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
