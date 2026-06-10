package main

import (
	"fmt"
	"net/http"
	"os"
)

func SetRoutes() {
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

	fmt.Printf("Server is listening on port %s\n", os.Getenv("GO_LISTEN_PORT"))
	http.ListenAndServe(":"+os.Getenv("GO_LISTEN_PORT"), nil)
}
