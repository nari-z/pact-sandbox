package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type (
	todo struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	todos struct {
		m       sync.Mutex
		records map[string]todo
	}
)

const (
	port = 50000
)

func main() {
	if err := Run(); err != nil {
		log.Fatalf(`Failed to run: %v`, err)
	}
}

func Run() error {
	db := todos{
		m:       sync.Mutex{},
		records: map[string]todo{},
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/todo", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Println(`Failed to read request body`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var req struct {
				Title string `json:"title"`
			}
			if err := json.Unmarshal(body, &req); err != nil {
				log.Println(`Failed to convert request from body`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			db.m.Lock()
			defer db.m.Unlock()
			todo := todo{
				ID:    strconv.Itoa(len(db.records) + 1),
				Title: req.Title,
				Done:  false,
			}
			db.records[todo.ID] = todo

			res, err := json.Marshal(todo)
			if err != nil {
				log.Println(`Failed to convert response to body`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if _, err := w.Write(res); err != nil {
				log.Println(`Failed to write response body`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			return
		})
		r.Patch("/{todoID}", func(w http.ResponseWriter, r *http.Request) {
			todoID := chi.URLParam(r, "todoID")

			db.m.Lock()
			defer db.m.Unlock()
			todo, ok := db.records[todoID]
			if !ok {
				log.Printf(`Not found todo: %s\n`, todoID)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			todo.Done = true
			db.records[todo.ID] = todo

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return
		})
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		return err
	}

	return nil
}
