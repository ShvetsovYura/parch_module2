package webapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"oyevents/internal/types"
	"sync"
)

type EventsWebapi struct {
	eventsCh chan types.EventMessage
	config   types.WebapiConfig
}

func NewEventsWebapi(cfg types.WebapiConfig, eventsCh chan types.EventMessage) *EventsWebapi {
	return &EventsWebapi{
		eventsCh: eventsCh,
		config:   cfg,
	}
}
func (a *EventsWebapi) userCreateHandler(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.eventsCh <- types.EventMessage{
		Topic: "user-events",
		Msg:   content,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (a *EventsWebapi) paymentCreateHandler(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.eventsCh <- types.EventMessage{
		Topic: "payment-events",
		Msg:   content,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (a *EventsWebapi) movieCreateHandler(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.eventsCh <- types.EventMessage{
		Topic: "movie-events",
		Msg:   content,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (a *EventsWebapi) Run(ctx context.Context, wg *sync.WaitGroup) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/events/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"status": true})
	})
	mux.HandleFunc("POST /api/events/user", a.userCreateHandler)
	mux.HandleFunc("POST /api/events/payment", a.paymentCreateHandler)
	mux.HandleFunc("POST /api/events/movie", a.movieCreateHandler)
	http.ListenAndServe(":"+a.config.Listen, mux)
}
