/*
 * Copyright 2014 Gregory Prisament
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
    "fmt"
    "net/http"
    "net/http/httputil"
    "net/url"
    "code.google.com/p/go.net/websocket"
    "github.com/gorilla/context"
    "github.com/gorilla/mux"
    "canopy/canolog"
    "canopy/config"
    "canopy/rest"
    "canopy/webapp"
    "canopy/ws"
    "os"
    "os/signal"
    "syscall"
)

var gConfAllowOrigin = ""

func shutdown() {
    canolog.Shutdown()
}

func main() {
    r := mux.NewRouter()

    cfg := config.NewDefaultConfig()
    err := cfg.LoadConfig()
    if err != nil {
        logFilename := config.JustGetOptLogFile()

        err := canolog.Init(logFilename)
        if err != nil {
            fmt.Println(err)
            return
        }
        canolog.Info("Starting Canopy Cloud Service")
        canolog.Error("Configuration error: %s", err)
        canolog.Info("Exiting")
        return
    }

    err = canolog.Init(cfg.OptLogFile())
    if err != nil {
        fmt.Println(err)
        return
    }

    canolog.Info("Starting Canopy Cloud Service")

    // handle SIGINT & SIGTERM
    defer shutdown()
    c := make (chan os.Signal, 1)
    c2 := make (chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    signal.Notify(c2, syscall.SIGTERM)
    go func() {
        <-c
        canolog.Info("SIGINT recieved")
        shutdown()
        os.Exit(1)
    }()
    go func() {
        <-c2
        canolog.Info("SIGTERM recieved")
        shutdown()
        os.Exit(1)
    }()

    if (cfg.OptHostname() == "") {
        canolog.Error("You must set the configuration option \"hostname\"")
        return
    }

    canolog.Info(cfg.DumpToString())

    if (cfg.OptForwardOtherHosts() != "") {
        canolog.Info("Requests to hosts other than ", cfg.OptHostname(), " will be forwarded to ", cfg.OptForwardOtherHosts())
        targetUrl, _ := url.Parse(cfg.OptForwardOtherHosts())
        reverseProxy := httputil.NewSingleHostReverseProxy(targetUrl)
        http.Handle("/", reverseProxy)
    } else {
        canolog.Info("No reverse proxy for other hosts consfigured.")
    }

    hostname := cfg.OptHostname()
    webManagerPath := cfg.OptWebManagerPath()
    jsClientPath := cfg.OptJavascriptClientPath()
    httpPort := cfg.OptHTTPPort()
    http.Handle(hostname + "/echo", websocket.Handler(ws.CanopyWebsocketServer))

    webapp.AddRoutes(r)
    rest.AddRoutes(r)

    http.Handle(hostname + "/", r)

    if (webManagerPath != "") {
        http.Handle(hostname + "/mgr/", http.StripPrefix("/mgr/", http.FileServer(http.Dir(webManagerPath))))
    }

    if (jsClientPath != "") {
        http.Handle(hostname + "/canopy-js-client/", http.StripPrefix("/canopy-js-client/", http.FileServer(http.Dir(jsClientPath))))
    }

    //err := http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", context.ClearHandler(http.DefaultServeMux))
    srv := &http.Server{
        Addr: fmt.Sprintf(":%d", httpPort),
        Handler: context.ClearHandler(http.DefaultServeMux),
        //ReadTimeout: 10*time.Second,
        //WriteTimeout: 10*time.Second,
    }
    err = srv.ListenAndServe()
    if err != nil {
        canolog.Error(err);
    }
}

/*
 * NOTES: Check out https://leanpub.com/gocrypto/read for good intro to crypto.
 */
