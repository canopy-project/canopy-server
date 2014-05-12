package main

import (
    "fmt"
    "net/http"
    "code.google.com/p/go.net/websocket"
    "github.com/gorilla/sessions"
    "github.com/gorilla/context"
    "canopy/datalayer"
)

var store = sessions.NewCookieStore([]byte("my_production_secret"))

func loginHandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "canopy-login-session")
    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    if dl.VerifyAccountPassword("greg", "mypass") {
        session.Values["logged_in_username"] = "greg"
        err := session.Save(r, w)
        if err != nil {
            fmt.Fprintf(w, "", err);
        }
        fmt.Fprintf(w, "logged in!")
    } else {
        fmt.Fprintf(w, "incorrect password")
    }
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "canopy-login-session")
    session.Values["logged_in_username"] = ""
    err := session.Save(r, w)
    if err != nil {
        fmt.Fprintf(w, "", err);
    }
    fmt.Fprintf(w, "logged out!")
}
func privHandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "canopy-login-session")
    if session.Values["logged_in_username"].(string) == "sam" {
        fmt.Fprintf(w, "access granted");
    } else {
        fmt.Fprintf(w, "ACCESS DENIED");
    }
}

func main() {
    fmt.Println("starting server");
    http.Handle("/echo", websocket.Handler(CanopyWebsocketServer))
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/logout", logoutHandler)
    http.HandleFunc("/private", privHandler)
    http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}
