package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// handleWebSocket upgrades to WebSocket and manages session delivery.
func (rt *Routes) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	rid := r.PathValue("rid")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Missing rid: raw error format (not typed envelope)
	if rid == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Missing request ID"}`))
		return
	}

	// Check if request exists
	stored := rt.storage.GetRequest(rid)
	if stored == nil {
		// Raw error format matching Routes.swift:215
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Request not found"}`))
		return
	}

	// Send initial status (typed envelope)
	if err := writeWSMessage(conn, WebSocketMessage{
		Type:    "status",
		Payload: StatusPayload{Status: stored.Status, Timestamp: nowISO8601()},
	}); err != nil {
		return
	}

	// If already ready, send session and close
	if stored.Status == StatusReady && stored.EncryptedSession != nil {
		if err := writeWSMessage(conn, WebSocketMessage{
			Type: "session",
			Payload: SessionPayload{
				DeliveredAt:      nowISO8601(),
				EncryptedSession: *stored.EncryptedSession,
			},
		}); err != nil {
			return // Don't mark delivered if write failed
		}
		rt.storage.MarkDelivered(rid)
		return
	}

	// If expired, send error and close
	if stored.ExpiresAt.Before(time.Now()) {
		rt.storage.UpdateStatus(rid, StatusExpired)
		writeWSMessage(conn, WebSocketMessage{
			Type:    "error",
			Payload: ErrorPayload{Code: "expired", Message: "Request has expired"},
		})
		return
	}

	// Wait for session: pre-register waiter, then start reader
	waiterID := generateID()
	waiterCh, registered := rt.storage.RegisterWaiter(rid, waiterID)
	done := make(chan struct{})

	// Writer mutex for WebSocket connection (gorilla requires serialized writes)
	var writeMu sync.Mutex

	// Only start reader if we're actually waiting (not immediate result)
	if registered {
		// Reader goroutine: handles ping/pong and detects disconnection
		go func() {
			defer close(done)
			for {
				msgType, data, err := conn.ReadMessage()
				if err != nil {
					// Client disconnected — unblock waiter channel
					rt.storage.RemoveWaiter(rid, waiterID)
					return
				}
				if msgType == websocket.TextMessage && strings.TrimSpace(string(data)) == "ping" {
					writeMu.Lock()
					conn.WriteMessage(websocket.TextMessage, []byte("pong"))
					writeMu.Unlock()
				}
			}
		}()
	}

	// Block until session, error, or cancellation
	msg := <-waiterCh

	// If we got a cancelled message from RemoveWaiter due to disconnect, just return
	if msg.Type == "error" {
		if p, ok := msg.Payload.(ErrorPayload); ok && p.Code == "cancelled" {
			return
		}
	}

	writeMu.Lock()
	writeErr := writeWSMessage(conn, msg)
	writeMu.Unlock()

	// Only mark delivered if the session was successfully written to the client
	if msg.Type == "session" && writeErr == nil {
		rt.storage.MarkDelivered(rid)
	}

	// Close connection and wait for reader to finish (if it was started)
	conn.Close()
	if registered {
		<-done
	}
}

// writeWSMessage writes a typed WebSocket message as JSON.
func writeWSMessage(conn *websocket.Conn, msg WebSocketMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

// generateID returns a unique waiter identifier.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
