package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cloud-agent-runtime/shared"
	"github.com/gorilla/websocket"
)

type Server struct {
	manager  *SessionManager
	upgrader websocket.Upgrader
}

func NewServer(m *SessionManager) *Server {
	return &Server{manager: m, upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sessions", s.handleSessions)
	mux.HandleFunc("/sessions/", s.handleSessionByID)
	return mux
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/sessions" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	s.createSession(w, r)
}

func (s *Server) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/sessions/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]

	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		s.getSession(w, r, id)
	case len(parts) == 2 && parts[1] == "stop" && r.Method == http.MethodPost:
		s.stopSession(w, r, id)
	case len(parts) == 2 && parts[1] == "stream" && r.Method == http.MethodGet:
		s.streamSession(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	var req shared.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	session, err := s.manager.Create(r.Context(), req.RepoURL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, shared.CreateSessionResponse{Session: session.Session})
}

func (s *Server) getSession(w http.ResponseWriter, _ *http.Request, id string) {
	session, ok := s.manager.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]shared.Session{"session": session.Session})
}

func (s *Server) stopSession(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.manager.Stop(r.Context(), id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) streamSession(w http.ResponseWriter, r *http.Request, id string) {
	session, ok := s.manager.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	stream, err := session.Sandbox.Stream(ctx)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}
	for {
		select {
		case chunk, ok := <-stream:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, chunk); err != nil {
				return
			}
		case <-time.After(30 * time.Second):
			_ = conn.WriteMessage(websocket.PingMessage, []byte("ping"))
		case <-ctx.Done():
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
