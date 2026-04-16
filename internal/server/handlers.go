package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/IhorXsh/Thread-Safe-Inventory-Service/internal/inventory"
)

func (s *Server) handleGetStock(w http.ResponseWriter, r *http.Request) {
	productID := pathVar(r, "productID")
	stock := s.svc.GetStock(productID)

	writeJSON(w, http.StatusOK, map[string]any{
		"product_id": productID,
		"stock":      stock,
	})
}

func (s *Server) handleReserve(w http.ResponseWriter, r *http.Request) {
	var item ReserveRequest
	if err := decodeJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.svc.Reserve(inventory.ReserveItem{
		ProductID: inventory.ProductID(item.ProductID),
		Quantity:  item.Quantity,
	}); err != nil {
		writeInventoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleReserveMultiple(w http.ResponseWriter, r *http.Request) {
	var req ReserveMultipleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "items must not be empty")
		return
	}

	reserveItems := make([]inventory.ReserveItem, len(req.Items))
	for i, item := range req.Items {
		reserveItems[i] = inventory.ReserveItem{
			ProductID: inventory.ProductID(item.ProductID),
			Quantity:  item.Quantity,
		}
	}

	if err := s.svc.ReserveMultiple(reserveItems); err != nil {
		writeInventoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func decodeJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

func writeInventoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, inventory.ErrProductNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, inventory.ErrInsufficientStock):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, inventory.ErrInvalidQuantity):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}
