package backend

import (
	"context"
	"net/http"
	"time"

	"cloud-agent-runtime/shared"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	manager  *SessionManager
	upgrader websocket.Upgrader
}

func NewServer(m *SessionManager) *Server {
	return &Server{manager: m, upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}}
}

func (s *Server) Router() *gin.Engine {
	r := gin.Default()
	r.POST("/sessions", s.createSession)
	r.GET("/sessions/:id", s.getSession)
	r.POST("/sessions/:id/stop", s.stopSession)
	r.GET("/sessions/:id/stream", s.streamSession)
	return r
}

func (s *Server) createSession(c *gin.Context) {
	var req shared.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session, err := s.manager.Create(c.Request.Context(), req.RepoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, shared.CreateSessionResponse{Session: session.Session})
}

func (s *Server) getSession(c *gin.Context) {
	session, ok := s.manager.Get(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"session": session.Session})
}

func (s *Server) stopSession(c *gin.Context) {
	if err := s.manager.Stop(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

func (s *Server) streamSession(c *gin.Context) {
	session, ok := s.manager.Get(c.Param("id"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(c.Request.Context())
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
