package main

import (
    "fmt"
    "canopy/datalayer"
    "time"
    "net/http"
    "encoding/json"
    "code.google.com/p/go.net/websocket"
)

func EchoServer(ws *websocket.Conn) {
    message := "hello"
    i := 0

    // receive binary frame
    for {
        var inmessage string
        var f interface{}
        ws.SetReadDeadline(time.Now().Add(50*time.Millisecond))
        err := websocket.Message.Receive(ws, &inmessage)
        if (err == nil) {
            fmt.Print("Recieved: ", inmessage);
            err = json.Unmarshal([]byte(inmessage), &f)
            if err == nil {
                m := f.(map[string]interface{})
                for k, v := range m {
                    switch vv := v.(type) {
                        case string:
                            fmt.Println(k, "is string", vv)
                        case int:
                            fmt.Println(k, "is int", vv)
                        case []interface{}:
                            fmt.Println(k, "is an array:")
                        default:
                            fmt.Println(k, "is of a type I don't know how to handle")
                    }
                }
            } else {
                fmt.Println(err)
            }
        }

        if (i % 20 == 0) {
            websocket.Message.Send(ws, message)
        }
        i++;
    }
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    dl := datalayer.NewCassandraDatalayer()
    /*dl.PrepDb() */
    dl.Connect()
    /*dl.StorePropertyValue("abcdef", "cpu", 0.87);*/
    print("starting server\n");

    http.Handle("/echo", websocket.Handler(EchoServer))
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
