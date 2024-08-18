package web

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rwlist/food/internal/blobs"
	"github.com/rwlist/food/internal/conf"
	"github.com/rwlist/food/internal/db"
)

type Server struct {
	cfg    *conf.App
	db     *db.Queries
	photos *blobs.Photos
}

func NewServer(cfg *conf.App, db *db.Queries, photos *blobs.Photos) *Server {
	return &Server{
		cfg:    cfg,
		db:     db,
		photos: photos,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHtml)
	})
	mux.HandleFunc("/week/{date}", s.handleWeek)
	mux.HandleFunc("/photo/{id}", s.handlePhoto)

	server := &http.Server{Addr: s.cfg.HttpBind, Handler: mux}
	slog.Info("starting server", "addr", fmt.Sprintf("http://localhost%s", s.cfg.HttpBind))

	go func() {
		<-ctx.Done()
		slog.Info("shutting down server")
		server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}

// handleWeek displays all photos taken in a range [date, date + 7 days]
func (s *Server) handleWeek(w http.ResponseWriter, r *http.Request) {
	dateStr := r.PathValue("date")

	// parse a date in a format "2024-12-31"
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}
	slog.Debug("fetching photos for week", "date", date)

	nextDate := date.AddDate(0, 0, 7)
	// prevDate := date.AddDate(0, 0, -7)

	data := PhotosData{
		Title: fmt.Sprintf("Week %s", date.Format("2006-01-02")),
		Rows:  make([]PhotoRow, 0),
	}

	dbPhotos, err := s.db.FindFoodPostsByDateRange(context.Background(), db.FindFoodPostsByDateRangeParams{
		FromTs:  pgtype.Timestamp{Time: date, Valid: true},
		UntilTs: pgtype.Timestamp{Time: nextDate, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to fetch photos", http.StatusInternalServerError)
		return
	}

	for _, p := range dbPhotos {
		data.Rows = append(data.Rows, PhotoRow{
			Timestamp: p.PostTs.Time,
			URL:       fmt.Sprintf("/photo/%s", p.PhotoKey),
		})
	}

	if err := photosTmpl.Execute(w, data); err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}

func (s *Server) handlePhoto(w http.ResponseWriter, r *http.Request) {
	slog.Debug("fetching photo", "id", r.PathValue("id"))

	id := r.PathValue("id")
	photo, err := s.photos.GetPhoto(context.Background(), id)
	if err != nil {
		http.Error(w, "failed to fetch photo", http.StatusInternalServerError)
		return
	}
	defer photo.Body.Close()

	w.Header().Set("Content-Type", *photo.ContentType)
	io.Copy(w, photo.Body)
}
