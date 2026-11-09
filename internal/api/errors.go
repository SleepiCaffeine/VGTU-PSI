package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/api/gen"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/game"
)

func writeRequestError(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, http.StatusBadRequest, err.Error())
}

func writeServiceError(w http.ResponseWriter, _ *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, game.ErrUnknownGame), errors.Is(err, game.ErrInvalidSelection):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("request failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(gen.Error{Error: message})
	if err != nil {
		log.Printf("write error response: %v", err)
	}
}
