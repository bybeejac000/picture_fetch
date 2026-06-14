package routes

import (
	"fmt"
	"net/http"
)

func refreshPictures() {
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if err := RefreshSlideshowList(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to refresh slideshow list: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Slideshow list refreshed successfully!")
		fmt.Printf("Slideshow list refreshed successfully!\n")
	})

}

func SetRoutes() {
	refreshPictures()
	fmt.Printf("Routes set.\n")
}
