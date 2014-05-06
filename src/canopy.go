package main

import (
    "fmt"
    "io"
    "net/http"
    
    "code.google.com/p/go.net/websocket"
)

// Echo the data received on the WebSocket
func EchoServer(ws *websocket.Conn) {
    io.Copy(ws, ws)
}


func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    http.Handle("/echo", websocket.Handler(EchoServer))
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
