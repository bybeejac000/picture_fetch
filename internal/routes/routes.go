package routes

import (
	"fmt"
	"log"
	"net/http"

	"photo_fetch/internal/slideshow"
)

func Register(mux *http.ServeMux, svc *slideshow.Service) {
	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Refresh(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to refresh slideshow list: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Slideshow list refreshed successfully!")
		log.Println("slideshow list refreshed successfully")
	})

	mux.HandleFunc("/photo", func(w http.ResponseWriter, r *http.Request) {
		imageName := r.URL.Query().Get("file")
		http.ServeFile(w, r, imageName)
	})

	log.Println("routes set")
}
