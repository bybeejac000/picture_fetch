package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"photo_fetch/internal/cleanup"
	"photo_fetch/internal/config"
	"photo_fetch/internal/database"
	"photo_fetch/internal/faces"
	"photo_fetch/internal/realtime"
	"photo_fetch/internal/routes"
	"photo_fetch/internal/slideshow"
	"photo_fetch/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cache := store.New(cfg)
	defer cache.Close()

	slideshowSvc := slideshow.NewService(db, cache, cfg)
	faceProcessor := faces.NewProcessor(db, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := slideshowSvc.Refresh(ctx); err != nil {
		log.Fatal(err)
	}
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	go cleanup.StartCleanupLoop(cleanupCtx, cfg.CleanupDir, 10*time.Second)

	mux := http.NewServeMux()
	routes.Register(mux, slideshowSvc)

	manager := realtime.NewManager(cfg, faceProcessor)
	manager.Register(mux)
	manager.Start()

	log.Printf("server is listening on port %s", cfg.GoListenPort)
	log.Fatal(http.ListenAndServe(":"+cfg.GoListenPort, mux))
}
