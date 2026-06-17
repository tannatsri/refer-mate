package http

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"refer-mate/internal/delivery/http/middlewear"
	"refer-mate/internal/domain"
	"refer-mate/internal/usecase"

	"github.com/gorilla/mux"
)

type CampaignHandler struct {
	campaignUC *usecase.CampaignUseCase
	authUC     *usecase.AuthUseCase
}

func NewCampaignHandler(campaignUC *usecase.CampaignUseCase, authUC *usecase.AuthUseCase) *CampaignHandler {
	return &CampaignHandler{campaignUC: campaignUC, authUC: authUC}
}

func registerCampaignRoutes(parent *mux.Router, campaignUC *usecase.CampaignUseCase, authUC *usecase.AuthUseCase) {
	h := NewCampaignHandler(campaignUC, authUC)
	r := parent.PathPrefix("/campaigns").Subrouter()
	r.HandleFunc("", h.Create).Methods("POST")
	r.HandleFunc("", h.List).Methods("GET")
	r.HandleFunc("/{id}", h.GetByID).Methods("GET")
	r.HandleFunc("/{id}", h.Update).Methods("PUT")
	r.HandleFunc("/{id}/recipients", h.UploadRecipients).Methods("POST")
	r.HandleFunc("/{id}/launch", h.Launch).Methods("POST")
	r.HandleFunc("/{id}/pause", h.Pause).Methods("POST")
	r.HandleFunc("/{id}/analytics", h.Analytics).Methods("GET")
}

type campaignRequest struct {
	TemplateID   int64  `json:"template_id"`
	CampaignName string `json:"campaign_name"`
	Description  string `json:"description"`
}

func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	var req campaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	c, err := h.campaignUC.Create(r.Context(), userID, req.TemplateID, req.CampaignName, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	campaigns, err := h.campaignUC.List(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaigns)
}

func (h *CampaignHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	c, err := h.campaignUC.GetByID(r.Context(), id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req campaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	c, err := h.campaignUC.Update(r.Context(), id, userID, req.TemplateID, req.CampaignName, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// UploadRecipients accepts multipart/form-data with a "file" field (CSV).
// CSV columns: email, name, company, role [, custom_key...]
func (h *CampaignHandler) UploadRecipients(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	recipients, err := parseCSV(file)
	if err != nil {
		http.Error(w, "invalid CSV: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.campaignUC.AddRecipients(r.Context(), id, userID, recipients); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"added": len(recipients)})
}

func (h *CampaignHandler) Launch(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	oauthToken, err := h.authUC.GetOAuthToken(r.Context(), userID)
	if err != nil {
		http.Error(w, "no Gmail access: "+err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := h.authUC.GetMe(r.Context(), userID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}

	if err := h.campaignUC.Launch(r.Context(), id, userID, oauthToken, user.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "running"})
}

func (h *CampaignHandler) Pause(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.campaignUC.Pause(r.Context(), id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "paused"})
}

func (h *CampaignHandler) Analytics(w http.ResponseWriter, r *http.Request) {
	userID := middlewear.GetUserID(r)
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	campaign, emails, err := h.campaignUC.GetAnalytics(r.Context(), id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"campaign": campaign,
		"emails":   emails,
	})
}

func parseCSV(r io.Reader) ([]domain.CampaignRecipient, error) {
	reader := csv.NewReader(r)
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	stdCols := map[string]int{"email": -1, "name": -1, "company": -1, "role": -1}
	customCols := map[string]int{}

	for i, h := range headers {
		if _, ok := stdCols[h]; ok {
			stdCols[h] = i
		} else {
			customCols[h] = i
		}
	}

	var recipients []domain.CampaignRecipient
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		rec := domain.CampaignRecipient{
			CustomVariables: map[string]string{},
		}
		if i := stdCols["email"]; i >= 0 && i < len(row) {
			rec.RecipientEmail = row[i]
		}
		if i := stdCols["name"]; i >= 0 && i < len(row) {
			rec.RecipientName = row[i]
		}
		if i := stdCols["company"]; i >= 0 && i < len(row) {
			rec.Company = row[i]
		}
		if i := stdCols["role"]; i >= 0 && i < len(row) {
			rec.Role = row[i]
		}
		for col, i := range customCols {
			if i < len(row) {
				rec.CustomVariables[col] = row[i]
			}
		}

		if rec.RecipientEmail == "" {
			continue
		}
		recipients = append(recipients, rec)
	}
	return recipients, nil
}
