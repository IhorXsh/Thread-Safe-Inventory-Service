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
	logger := s.requestLogger(r)
	productID := pathVar(r, "productID")
	logger.Info("request received", "method", r.Method, "path", routeTemplate(r), "product_id", productID)

	stock := s.svc.GetStock(productID)
	logger.Info("get stock computed", "product_id", productID, "stock", stock)

	writeJSON(w, logger, http.StatusOK, Response{
		Status: StatusOK,
		Data: map[string]any{
			"product_id": productID,
			"stock":      stock,
		},
	})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	logger := s.requestLogger(r)
	logger.Info("request received", "method", r.Method, "path", routeTemplate(r))
	writeJSON(w, logger, http.StatusOK, Response{Status: StatusOK})
}

func (s *Server) handleReserve(w http.ResponseWriter, r *http.Request) {
	logger := s.requestLogger(r)
	logger.Info("request received", "method", r.Method, "path", routeTemplate(r))

	var item ReserveRequest
	if err := decodeJSON(r, &item); err != nil {
		logger.Warn("reserve decode failed", "error", err)
		writeError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	logger.Info("reserve request parsed", "product_id", item.ProductID, "quantity", item.Quantity)

	if err := s.svc.Reserve(inventory.ReserveItem{
		ProductID: inventory.ProductID(item.ProductID),
		Quantity:  item.Quantity,
	}); err != nil {
		logger.Warn("reserve failed", "product_id", item.ProductID, "quantity", item.Quantity, "error", err)
		writeInventoryError(w, logger, err)
		return
	}

	logger.Info("reserve success", "product_id", item.ProductID, "quantity", item.Quantity)
	writeJSON(w, logger, http.StatusOK, Response{Status: StatusOK})
}

func (s *Server) handleReserveMultiple(w http.ResponseWriter, r *http.Request) {
	logger := s.requestLogger(r)
	logger.Info("request received", "method", r.Method, "path", routeTemplate(r))

	var req ReserveMultipleRequest
	if err := decodeJSON(r, &req); err != nil {
		logger.Warn("reserve-multiple decode failed", "error", err)
		writeError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	logger.Info("reserve-multiple request parsed", "items_count", len(req.Items))

	if len(req.Items) == 0 {
		logger.Warn("reserve-multiple failed: empty items")
		writeError(w, logger, http.StatusBadRequest, "items must not be empty")
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
		logger.Warn("reserve-multiple failed", "items_count", len(req.Items), "error", err)
		writeInventoryError(w, logger, err)
		return
	}

	logger.Info("reserve-multiple success", "items_count", len(req.Items))
	writeJSON(w, logger, http.StatusOK, Response{Status: StatusOK})
}

func decodeJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

func writeInventoryError(w http.ResponseWriter, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, inventory.ErrProductNotFound):
		writeError(w, logger, http.StatusNotFound, err.Error())
	case errors.Is(err, inventory.ErrInsufficientStock):
		writeError(w, logger, http.StatusConflict, err.Error())
	case errors.Is(err, inventory.ErrInvalidQuantity):
		writeError(w, logger, http.StatusBadRequest, err.Error())
	default:
		writeError(w, logger, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, logger *slog.Logger, status int, message string) {
	writeJSON(w, logger, status, Response{Status: StatusError, Error: message})
}

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	logger.Info("response sent", "status_code", status, "response", payload)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("failed to write json response", "error", err)
	}
}
