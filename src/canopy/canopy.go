package main

import (
    "fmt"
    "net/http"
    "code.google.com/p/go.net/websocket"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    fmt.Println("starting server");
    http.Handle("/echo", websocket.Handler(CanopyWebsocketServer))
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
