package main


import (
    "context"
    "cyberditto-backend/internal/api"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)


func main() {
    // Initialize logger
    log.SetFlags(log.LstdFlags | log.Lshortfile)
   
    router := api.SetupRouter()
   
    // Server configuration
    server := &http.Server{
        Addr:         ":8080",
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }


    // Server startup
    go func() {
        log.Printf("Starting server on %s", server.Addr)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()


    // Graceful shutdown handling
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit


    log.Println("Shutting down server...")


    // Shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()


    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }


    log.Println("Server stopped gracefully")
}