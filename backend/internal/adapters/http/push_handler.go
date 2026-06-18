package httpadapter

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"

	"micha/backend/internal/domain/pushsubscription"
	"micha/backend/internal/ports/outbound"
)

// PushHandlerDeps contains the dependencies needed by the push handler.
type PushHandlerDeps struct {
	PushRepo        outbound.PushSubscriptionRepository
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDContact    string
}

// pushHandler handles Web Push subscription management.
type pushHandler struct {
	deps PushHandlerDeps
}

// newPushHandler constructs a pushHandler.
func newPushHandler(deps PushHandlerDeps) pushHandler {
	return pushHandler{deps: deps}
}

// subscribeRequest is the JSON body for POST /v1/push/subscribe.
type subscribeRequest struct {
	Endpoint  string `json:"endpoint"`
	P256dhKey string `json:"p256dh"`
	AuthKey   string `json:"auth"`
}

// handleSubscribe saves a new push subscription for the authenticated user.
func (h pushHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	if req.Endpoint == "" || req.P256dhKey == "" || req.AuthKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("endpoint, p256dh, and auth are required"))
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	sub, err := pushsubscription.New(
		pushsubscription.ID(uuid.NewString()),
		userID,
		req.Endpoint,
		req.P256dhKey,
		req.AuthKey,
		time.Now(),
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create subscription"))
		return
	}

	if err := h.deps.PushRepo.Save(r.Context(), sub); err != nil {
		slog.Error("failed to save push subscription", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save subscription"))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "subscribed"})
}

// handleTest sends a test push notification to all subscriptions of the authenticated user.
func (h pushHandler) handleTest(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	subs, err := h.deps.PushRepo.ListByUserID(r.Context(), userID)
	if err != nil {
		slog.Error("failed to list push subscriptions", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list subscriptions"))
		return
	}

	if len(subs) == 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("no push subscriptions found — allow notifications first"))
		return
	}

	sent := 0
	var lastErr error
	for _, sub := range subs {
		attrs := sub.Attributes()
		s := &webpush.Subscription{
			Endpoint: attrs.Endpoint,
			Keys: webpush.Keys{
				P256dh: attrs.P256dhKey,
				Auth:   attrs.AuthKey,
			},
		}

		resp, err := webpush.SendNotification([]byte(`{"title":"🔔 Micha","body":"¡Notificaciones activadas! Recibirás alertas de pagos y recordatorios aquí.","icon":"/icon-192.png","badge":"/badge.png","tag":"micha-test"}`), s, &webpush.Options{
			Subscriber:      h.deps.VAPIDContact,
			VAPIDPublicKey:  h.deps.VAPIDPublicKey,
			VAPIDPrivateKey: h.deps.VAPIDPrivateKey,
			TTL:             60,
		})
		if err != nil {
			lastErr = err
			slog.Warn("failed to send push notification, removing subscription",
				"endpoint", attrs.Endpoint, "error", err)
			// Subscription likely expired — remove it.
			if delErr := h.deps.PushRepo.DeleteByEndpoint(r.Context(), userID, attrs.Endpoint); delErr != nil {
				slog.Error("failed to delete stale subscription", "error", delErr)
			}
			continue
		}
		resp.Body.Close()
		sent++
	}

	if sent == 0 && lastErr != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("all subscriptions expired — re-allow notifications"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "sent",
		"sent":   sent,
	})
}

// VAPIDPublicKey returns the VAPID public key so the frontend can subscribe.
func (h pushHandler) handleVapidPublicKey(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"publicKey": h.deps.VAPIDPublicKey,
	})
}

func errorResponse(msg string) map[string]any {
	return map[string]any{
		"error": map[string]string{"message": msg},
	}
}
