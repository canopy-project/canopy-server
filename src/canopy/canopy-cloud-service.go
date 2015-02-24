// Copyright 2014-2015 SimpleThings, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
    "canopy/pigeon"
    "canopy/jobqueue"
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

func handleHelloWorld(sys jobqueue.System, worker jobqueue.Worker) {
    canolog.Info("Handler Start")
    reqChan := make(chan jobqueue.Request)
    respChan := make(chan jobqueue.Response)
    canolog.Info("Handler Listening")
    err := worker.Listen("generic", reqChan, respChan)
    if err != nil {
        canolog.Error(err.Error())
        return
    }

    canolog.Info("Handler done listening")
    req := <-reqChan
    canolog.Info("hello4")
    if (req.Body()["msg"] == "hello") {
        resp := sys.NewResponse()
        resp.SetBody(map[string]interface{}{"msg" : "world"})
        canolog.Info("hello5")
        respChan <- resp
        canolog.Info("hello6")
    }
    canolog.Info("hello7")
}

func main() {
    r := mux.NewRouter()

    cfg := config.NewDefaultConfig()
    err := cfg.LoadConfig()
    if err != nil {
        logFilename := config.JustGetOptLogFile()

        err2 := canolog.Init(logFilename)
        if err2 != nil {
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

    pigeonSys, err := pigeon.InitPigeonSystem()
    if (err != nil) {
        canolog.Error("Error starting pigeon system")
        return
    }

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

    if (cfg.OptPasswordSecretSalt() == "") {
        canolog.Error("You must set the configuration option \"password-secret-salt\"")
        return
    }
    canolog.Info(cfg.ToString())

    // Register worker
    canolog.Info("Init Pigeon Sys: ")
    pigeonSys2, err := jobqueue.NewPigeonSystem(cfg)
    if err != nil {
        canolog.Error(err.Error())
        return
    }
    canolog.Info("Create worker")
    worker, _ := pigeonSys2.StartWorker("localhost")
    canolog.Info("Handle hello msg")
    go handleHelloWorld(pigeonSys2, worker)

    launcher := pigeonSys2.NewLauncher()
    canolog.Info("Launch hello message")
    respChan, err := launcher.Launch("generic", map[string]interface{}{"msg" : "hello"})
    canolog.Info("Waiting for response")
    resp := <-respChan
   
    canolog.Info("Response: ", resp)
    // ---

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
    http.Handle(hostname + "/echo", websocket.Handler(ws.NewCanopyWebsocketServer(cfg, pigeonSys)))

    webapp.AddRoutes(r)
    rest.AddRoutes(r, cfg, pigeonSys)

    http.Handle(hostname + "/", r)

    if (webManagerPath != "") {
        http.Handle(hostname + "/mgr/", http.StripPrefix("/mgr/", http.FileServer(http.Dir(webManagerPath))))
    }

    if (jsClientPath != "") {
        http.Handle(hostname + "/canopy-js-client/", http.StripPrefix("/canopy-js-client/", http.FileServer(http.Dir(jsClientPath))))
    }

    // Run HTTP and HTTPS servers simultaneously (if both are enabled)
    httpResultChan := make(chan error)
    httpsResultChan := make(chan error)
    if cfg.OptEnableHTTP() {
        go func() {
            httpPort := cfg.OptHTTPPort()
            srv := &http.Server{
                Addr: fmt.Sprintf(":%d", httpPort),
                Handler: context.ClearHandler(http.DefaultServeMux),
            }
            err = srv.ListenAndServe()
            httpResultChan <- err
        }()
    }
    if cfg.OptEnableHTTPS() {
        go func() {
            httpsPort := cfg.OptHTTPSPort()
            httpsCertFile := cfg.OptHTTPSCertFile()
            httpsPrivKeyFile := cfg.OptHTTPSPrivKeyFile()
            srv := &http.Server{
                Addr: fmt.Sprintf(":%d", httpsPort),
                Handler: context.ClearHandler(http.DefaultServeMux),
            }
            err := srv.ListenAndServeTLS(httpsCertFile, httpsPrivKeyFile)
            httpsResultChan <- err
        }()
    }

    // Exit if either server has error
    select {
        case err := <- httpResultChan:
            canolog.Error(err)
        case err := <- httpsResultChan:
            canolog.Error(err)
    }

}

/*
 * NOTES: Check out https://leanpub.com/gocrypto/read for good intro to crypto.
 */
