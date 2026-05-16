package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"cloud-agent-runtime/shared"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var backendURL string

var runCmd = &cobra.Command{
	Use:   "run <repo>",
	Args:  cobra.ExactArgs(1),
	Short: "Run a repo in a remote sandbox",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := args[0]
		body, _ := json.Marshal(shared.CreateSessionRequest{RepoURL: repo})
		resp, err := http.Post(backendURL+"/sessions", "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var out shared.CreateSessionResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return err
		}
		fmt.Printf("session: %s\n", out.Session.ID)

		u, _ := url.Parse(backendURL)
		scheme := "ws"
		if u.Scheme == "https" {
			scheme = "wss"
		}
		wsURL := fmt.Sprintf("%s://%s/sessions/%s/stream", scheme, u.Host, out.Session.ID)
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		go func() {
			<-ctx.Done()
			_, _ = http.Post(fmt.Sprintf("%s/sessions/%s/stop", backendURL, out.Session.ID), "application/json", bytes.NewReader(nil))
			_ = conn.Close()
		}()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return nil
			}
			fmt.Print(string(msg))
		}
	},
}

func init() {
	runCmd.Flags().StringVar(&backendURL, "backend", "http://localhost:8080", "Backend API URL")
}
