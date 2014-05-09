package main

import (
    "fmt"
    "io"
    "net/http"
    
    "code.google.com/p/go.net/websocket"
    "github.com/gocql/gocql"
)
// Echo the data received on the WebSocket
func EchoServer(ws *websocket.Conn) {
    io.Copy(ws, ws)
}


func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    // setup cassandra
    SayHi()

    // setup web & ws servers
    http.Handle("/echo", websocket.Handler(EchoServer))
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
