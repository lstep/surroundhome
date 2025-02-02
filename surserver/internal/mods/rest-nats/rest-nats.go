package rest

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/lstep/surroundhome/surserver/internal/app"
	"github.com/nats-io/nats.go"
)

type RestModule struct {
	// internal dependencies, e.g., connection config for an identity server
}

func (m *RestModule) Name() string {
	return "bridge"
}

func (m *RestModule) Init(config map[string]any) error {
	// Set up your identity server connection, initialize services...
	return nil
}
func (m *RestModule) HTTPHandlers(pub app.Publisher) []app.HTTPHandler {
	return []app.HTTPHandler{
		{
			Method:  "POST",
			Path:    "/{topic}",
			Handler: withPub(handleNatsProxy, pub),
		},
	}
}

func (m *RestModule) MsgHandlers(pub app.Publisher) []app.MsgHandler {
	return []app.MsgHandler{
		{
			Subject: "user.created",
			Handler: func(msg *nats.Msg) {
				// Handle user created event
			},
		},
	}
}

func withPub(h func(http.ResponseWriter, *http.Request, app.Publisher), pub app.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r, pub)
	}
}

func handleNatsProxy(w http.ResponseWriter, r *http.Request, pub app.Publisher) {
	start := time.Now()

	// Get topic from URL pattern
	topic := r.PathValue("topic")
	slog.Info("received request",
		"method", r.Method,
		"path", r.URL.Path,
		"topic", topic,
		"remote_addr", r.RemoteAddr,
	)

	if topic == "" {
		http.Error(w, "Invalid URL path: missing topic", http.StatusBadRequest)
		slog.Error("missing topic in URL path")
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		slog.Error("failed to read request body",
			"error", err,
		)
		return
	}
	defer r.Body.Close()

	// Verify the body is valid JSON
	var jsonBody interface{}
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		slog.Error("invalid JSON in request body",
			"error", err,
		)
		return
	}

	slog.Info("publishing to NATS",
		"topic", topic,
		"payload_size", len(body),
	)

	// Send request to NATS and wait for response
	msg, err := pub.Request(topic, body)
	if err != nil {
		if err == nats.ErrTimeout {
			http.Error(w, "Request to NATS timed out", http.StatusGatewayTimeout)
			slog.Error("NATS request timed out",
				"topic", topic,
				"timeout", time.Duration(15)*time.Second,
			)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			slog.Error("failed to publish to NATS",
				"topic", topic,
				"error", err,
			)
		}
		return
	}

	// Set content type to JSON if the response is JSON
	w.Header().Set("Content-Type", "application/json")

	// Write the NATS response directly to the HTTP response
	w.Write(msg.Data)

	// Log completion time and response size
	elapsed := time.Since(start)
	slog.Info("request completed",
		"topic", topic,
		"response_size", len(msg.Data),
		"duration", elapsed,
	)
}
