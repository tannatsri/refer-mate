package http

import (
	"net/http"
	"net/url"
	"refer-mate/internal/usecase"
	"strconv"

	"github.com/gorilla/mux"
)

// transparentGIF is a 1x1 transparent GIF pixel.
var transparentGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
	0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21,
	0xf9, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

type TrackingHandler struct {
	trackingUC *usecase.TrackingUseCase
}

func NewTrackingHandler(trackingUC *usecase.TrackingUseCase) *TrackingHandler {
	return &TrackingHandler{trackingUC: trackingUC}
}

func registerTrackingRoutes(r *mux.Router, trackingUC *usecase.TrackingUseCase) {
	h := NewTrackingHandler(trackingUC)
	r.HandleFunc("/track/open/{emailID}", h.Open).Methods("GET")
	r.HandleFunc("/track/click/{emailID}", h.Click).Methods("GET")
}

func (h *TrackingHandler) Open(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["emailID"], 10, 64)
	if err == nil {
		h.trackingUC.TrackOpen(r.Context(), id)
	}
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write(transparentGIF)
}

func (h *TrackingHandler) Click(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["emailID"], 10, 64)
	if err == nil {
		h.trackingUC.TrackClick(r.Context(), id)
	}

	redirectURL := r.URL.Query().Get("url")
	if redirectURL == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}

	// Validate URL to prevent open redirect to non-http schemes
	parsed, err := url.Parse(redirectURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}
