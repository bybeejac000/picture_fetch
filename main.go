package main

import (
	"fmt"
	"net/http"
	"os"
	"photo_fetch/routes"
	"photo_fetch/startup"
	"photo_fetch/websocket"
)

func main() {
	startup.StartupScript()
	websocket.InitializeSockets()
	routes.SetRoutes()

	fmt.Printf("Server is listening on port %s\n", os.Getenv("GO_LISTEN_PORT"))
	http.ListenAndServe(":"+os.Getenv("GO_LISTEN_PORT"), nil)
}
