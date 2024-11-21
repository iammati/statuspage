package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iammati/statuspage/config"
	"iammati/statuspage/handlers"
	"iammati/statuspage/websocket"
)

func main() {
	// Bootstrapping the application
	config.Bootstrap()

	// Connect to the database
	config.DbConn = config.Database()
	defer config.DbConn.Close()

	port := "8080"
	srv := &http.Server{
		Addr: ":" + port,
	}

	// HTTP server with CORS middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/up", handlers.HandleUp)
	mux.HandleFunc("/certinfo", handlers.HandleCertInfo)

	// WebSocket server
	mux.HandleFunc("/ws", websocket.Handle)

	// Wrap the mux with the CORS middleware
	corsHandler := corsMiddleware(mux)

	// Start WebSocket broadcast routine
	go websocket.BroadcastMessages()

	for range 5 {
		PrintLog("", false)
	}

	PrintLog("======================================= STATUSPAGE =======================================", false)
	PrintLog("", false)

	srv.Handler = corsHandler

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			PrintLog("listen: ", true)
			log.Fatalf("%s\n", err)
		}
	}()
	PrintLog(fmt.Sprintf("HTTP server listening on port *:%s...", port), false)

	go handlers.MonitorHostChanges(5 * time.Second)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	PrintLog("Shutting down server...", false)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		PrintLog("Server forced to shutdown: ", true)
		log.Fatalf("%s\n", err)
	}

	PrintLog("Server exiting", false)
}

func PrintLog(reason string, doNotLogIt bool) {
	if !doNotLogIt {
		handlers.LogUpdatetimeEvent(handlers.UpdatetimeEvent{
			Time:   time.Now(),
			Reason: reason,
		}, !doNotLogIt)

		log.Println(reason)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://app:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
