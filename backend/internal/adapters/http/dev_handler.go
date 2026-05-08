package httpadapter

import (
	"net/http"
	"strconv"
	"time"

	appshared "micha/backend/internal/application/shared"
)

type devHandler struct{}

func newDevHandler() devHandler {
	return devHandler{}
}

func (h devHandler) handleTimeOffset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is allowed")
		return
	}

	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_DAYS", "days must be an integer")
		return
	}

	if days == 0 {
		appshared.ResetTimeOffset()
	} else {
		appshared.AddTimeOffset(time.Duration(days) * 24 * time.Hour)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Time offset updated",
		"now":     appshared.Now(),
	})
}
