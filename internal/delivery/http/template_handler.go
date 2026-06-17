package http

import (
	"encoding/json"
	"net/http"
	"refer-mate/internal/delivery/http/middlewear"
	"refer-mate/internal/usecase"
	"strconv"

	"github.com/gorilla/mux"
)

type TemplateHandler struct {
	tmplUC *usecase.TemplateUseCase
}

func NewTemplateHandler(tmplUC *usecase.TemplateUseCase) *TemplateHandler {
	return &TemplateHandler{tmplUC: tmplUC}
}

func registerTemplateRoutes(parent *mux.Router, tmplUC *usecase.TemplateUseCase) {
	h := NewTemplateHandler(tmplUC)
	r := parent.PathPrefix("/templates").Subrouter()
	r.HandleFunc("", h.Create).Methods("POST")
	r.HandleFunc("", h.List).Methods("GET")
	r.HandleFunc("/{id}", h.GetByID).Methods("GET")
	r.HandleFunc("/{id}", h.Update).Methods("PUT")
	r.HandleFunc("/{id}", h.Delete).Methods("DELETE")
}

type templateRequest struct {
	Title   string `json:"title"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	var req templateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	t, err := h.tmplUC.Create(r.Context(), userID, req.Title, req.Subject, req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	templates, err := h.tmplUC.List(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *TemplateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	t, err := h.tmplUC.GetByID(r.Context(), id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req templateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	t, err := h.tmplUC.Update(r.Context(), id, userID, req.Title, req.Subject, req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.tmplUC.Delete(r.Context(), id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
}
