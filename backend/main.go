package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Task struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var InMemoryTask Task = Task{
	Title:     "Learn Dagger",
	Completed: false,
}

const ListenAddr string = ":8080"

func main() {
	taskMux := http.NewServeMux()
	taskMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("Access-Control-Allow-Origin", "*")
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
		headers.Add("Access-Control-Allow-Methods", "GET,PUT")

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			slog.Info("request", "path", "/", "method", http.MethodGet)

			data, err := json.Marshal(InMemoryTask)
			if err != nil {
				slog.Error("error in marshal", "error", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			headers.Set("Content-Type", "application/json")
			if _, err = w.Write(data); err != nil {
				slog.Error("error in writing response", "error", err.Error())
			}
		case http.MethodPut:
			slog.Info("request", "path", "/", "method", http.MethodPut)

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("error in reading body", "error", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			contentTypeHeader := r.Header.Get("Content-Type")
			var contentType string
			if contentTypeHeader != "" {
				contentType = contentTypeHeader
			} else {
				contentType = http.DetectContentType(bodyBytes)
			}
			if contentType != "application/json" {
				slog.Warn("unsupported content type", "contentType", contentType)
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}

			if len(bodyBytes) <= 0 {
				slog.Warn("empty body")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var task Task
			if err := json.Unmarshal(bodyBytes, &task); err != nil {
				slog.Warn("error in unmarshal", "error", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if task.Title == "" {
				slog.Warn("empty title in put")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if len(task.Title) > 100 {
				slog.Warn("got title that is too long")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}

			slog.Info("overwriting task in memory", "newTitle", task.Title, "newCompleted", task.Completed)
			InMemoryTask = task
			w.WriteHeader(http.StatusAccepted)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	rootMux := http.NewServeMux()
	rootMux.Handle("/api/task", taskMux)

	s := &http.Server{
		Addr:              ListenAddr,
		Handler:           rootMux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	slog.Info("starting server", "addr", ListenAddr)
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("error while serving", "error", err.Error())
			os.Exit(1)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		slog.Error("error while shutting down server", "error", err.Error())
	}
}
