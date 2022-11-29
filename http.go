package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
)

const (
	URLKeyParam = "key"
)

// HTTPQ is a http message broker
type HTTPQ struct {
	RxBytes  int                    // number of bytes (message body) consumed
	TxBytes  int                    // number of bytes (message body) published
	PubFails int                    // number of publish failures
	SubFails int                    // number of subscribe failures
	channels map[string]chan []byte // content of saved messages
	mutex    *sync.Mutex
}

func (h HTTPQ) getChannel(key string) chan []byte {
	_, ok := h.channels[key]
	if !ok {
		h.mutex.Lock()
		defer h.mutex.Unlock()

		h.channels[key] = make(chan []byte)
	}

	return h.channels[key]
}

// NewHTTPQ retursn a new instance of a httpq
func NewHTTPQ() *HTTPQ {
	h := &HTTPQ{}
	h.mutex = &sync.Mutex{}
	h.channels = map[string]chan []byte{}

	return h
}

// ServeHTTP is an implementation of http.Handler
func (h *HTTPQ) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rtr := chi.NewRouter()

	rtr.Get("/{"+URLKeyParam+"}", h.Consume)
	rtr.Post("/{"+URLKeyParam+"}", h.Publish)

	rtr.ServeHTTP(w, r)
}

// Pubish is the http.HandlerFunc for the POST /{key} request
// publish will store the
func (h *HTTPQ) Publish(w http.ResponseWriter, r *http.Request) {
	ch := h.getChannel(chi.URLParam(r, URLKeyParam))

	buff := &bytes.Buffer{}
	_, err := buff.ReadFrom(r.Body)
	if err != nil {
		h.PubFails += 1
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error reading from body in POST /{key} request: %v", err)
		return
	}

	for {
		select {
		case ch <- buff.Bytes():
			h.RxBytes += buff.Len()
			w.WriteHeader(http.StatusOK)
			return
		case <-r.Context().Done():
			h.PubFails += 1
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}
	}
}

// Consume is the http.HandlerFunc for the GET /{key} request
// consume will retreive the first in line message published by a publish request
func (h *HTTPQ) Consume(w http.ResponseWriter, r *http.Request) {
	ch := h.getChannel(chi.URLParam(r, URLKeyParam))

	for {
		select {
		case <-r.Context().Done():
			h.SubFails += 1
			w.WriteHeader(http.StatusRequestTimeout)
			return
		case val := <-ch:
			c, err := w.Write(val)
			if err != nil {
				h.SubFails += 1
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("error writing to body in GET /{key} request: %v", err)
				return
			}
			h.TxBytes += c
			return
		}
	}
}

// StatsResponse is the returned object for the GET /stats request
type StatsResponse struct {
	PublishedBytes int `json:"published_bytes"`
	ConsumedBytes  int `json:"consumed_bytes"`
	PublishedFails int `json:"published_fails"`
	ConsumedFails  int `json:"consumed_fails"`
}

// Stats is the http.HandlerFunc for the GET /stats request
// stats will return the statistics of the server
func (h *HTTPQ) Stats(w http.ResponseWriter, r *http.Request) {
	resp := StatsResponse{
		PublishedBytes: h.RxBytes,
		ConsumedBytes:  h.TxBytes,
		PublishedFails: h.PubFails,
		ConsumedFails:  h.SubFails,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding /stats request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
