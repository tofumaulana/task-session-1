package handlers

import (
	"encoding/json"
	"net/http"
	"task-sesion-1/repositories"
)

type ReportHandler struct {
    repo *repositories.TransactionRepository
}

func NewReportHandler(repo *repositories.TransactionRepository) *ReportHandler {
    return &ReportHandler{repo: repo}
}

func (h *ReportHandler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
    // Panggil repository
    report, err := h.repo.GetDailyReport()
    if err != nil {
        http.Error(w, "Failed to generate report: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Return JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}