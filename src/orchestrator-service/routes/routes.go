package routes

import (
	"net/http"
	"orchestrator-service/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func Routes(app *handlers.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Logger)

	// Routes
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test endpoint working"))
	})
	r.Get("/generate-meeting-id", app.GenerateMeetingId)
	r.Post("/capture-screenshots", app.CaptureScreenshots)
	r.Post("/capture-audio", app.CaptureAudio)
	r.Post("/end-meeting", app.EndMeeting)
	r.Post("/start-transcription", app.StartTranscription)

	return r
}
