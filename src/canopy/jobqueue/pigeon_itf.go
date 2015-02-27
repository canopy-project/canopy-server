// Copyright 2015 Gregory Prisament
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

// OVERVIEW
//
//  Pigeon is Canopy's message passing system.  It follows a client/server
//  model, although most nodes will act as both a server and a client.  The
//  process goes something like this:
//
//  1) The Server registers which jobs it can handle.  These are identified by
//  "job key" strings.
//  2) The Client launches (or broadcasts) a request.
//  2) The Pigeon system determines which server(s) to send the request to.
//  3) The Server receives, processes the request (perhaps by launching further
//  requests), and generates a response.
//  4) The Pigeon system reports the Server's response to the Client.
//
//  Pigeon's flexible design allows it to be used for a variety of distributed
//  computing tasks.
//
//  A "Server" listens for Requests.
//
//  A "Client" issues Requests.
//
//  A "Request" consists of a string name ("job key") and a JSON payload.
//
//  A "Response" consists of a JSON payload and an optional error object.
//
// SERVERS
//
//  A Server is identified by IP Address or Hostname.  Information about each
//  Server is stored in the database.
//
//  To register a new Server, use:
//
//      server, err := pigeonSys.StartServer(hostname)
//
//  You can then start listening for requests that match a desired key:
//
//      err = server.Handle("myJobKey", myHandlerFunc)
//
//  Before your program quits, we advise that you Stop the server.  Otherwise,
//  requests will continue to be sent to it and will time out.
//
//      err = server.Stop()
//
//  All data about Servers including their status and what they are listening
//  for are stored in the DB.
//
// LAUNCHING REQUESTS
//
//  To send a message you must first create a Client object.  A Client contains
//  the settings that will be used to send the request.
//
//      client := pigeonSys.NewClient()
//
//  You can then set options:
//
//      client.SetTimeoutms(1000)
//
//  To send a request that should be consumed by exactly one Server:
//
//      responseChan := client.Launch("generic", myPayload)
//
//  To block waiting for the response:
//
//      response := <-responseChan
//
package jobqueue

import (
    "canopy/config"
    "canopy/datalayer/cassandra_datalayer"
)

// StatusEnum is the status of a Worker
type StatusEnum int
const (
    // DOES_NOT_EXIST means no worker exists for the provided hostname
    DOES_NOT_EXIST StatusEnum = iota

    // STOPPED means the worker has been stopped and will not be sent requests.
    STOPPED

    // RUNNING means the worker is currently running.
    RUNNING

    // UNRESPONSIVE means the worker has not been stopped, but is not
    // responding to requests. 
    UNRESPONSIVE
)

type HandlerFunc func(jobKey string, userCtx map[string]interface{}, req Request, resp Response)

type System interface {
    // Create a new empty response object.
    NewClient() Client

    // Create a new empty response object.
    NewResponse() Response

    // Starts RPC server, adds worker to the DB, if not already present, and
    // sets its status to "active".
    StartServer(hostname string) (Server, error)

    // Lookup a specific Server by hostname.
    Server(hostname string) (Server, error)

    // Obtain list of Servers from the DB.
    Servers() ([]Server, error)
}

type Client interface {
    // Broadcast a request to every Server interested
    Broadcast(jobKey string, payload map[string]interface{}) error

    // Launches a request that will be handled by exactly one Server
    Launch(jobKey string, payload map[string]interface{}) (<-chan Response, error)
    
    // Launches a request that is idemponent and can be consumed by multiple
    // Servers without ill effect.  This allows the job to be sent to
    // multiple consumers simultaneously, for low latency response (whoever
    // responds first wins).
    LaunchIdempotent(jobKey string, numParallel uint32, payload map[string]interface{}) (<-chan Response, error)

    // Set the timeout for non-broadcast requests.
    // Use a negative value for no timeout.
    SetTimeoutms(timeout int32)
}

type Server interface {
    // Listen for requests that match <key>, triggering a handler function each
    // time such a request is recieved.
    // Registers that this Server can handle jobs named <jobKey> in the
    // database.
    // <userCtx> is optional user-provided data that will be passed to the
    //      handler as req.UserContext().
    Handle(jobKey string, fn HandlerFunc, userCtx map[string]interface{}) error

    // Set the Server's status to "active".  Does nothing if server is already
    // "active".
    Start() error

    // Get the Server's status
    Status() error

    // Set the Server's status to "stopped".  It will no longer recieve
    // requests until started again.  Does nothing if worker is already
    // "stopped".
    Stop() error

    // Stop listening for a specific <jobKey>.
    StopHandling(jobKey string) error
}

type Request interface {
    Body() map[string]interface{}
}

type Response interface {
    Body() map[string]interface{}

    // Must be a gob-able value
    SetBody(body map[string]interface{})

    // <value> must be a gob-able value
    AppendToBody(key string, value interface{})
}

func NewPigeonSystem(cfg config.Config) (System, error) {
    dl := cassandra_datalayer.NewDatalayer(cfg)
    // TODO: share DB connection
    conn, err := dl.Connect("canopy")
    if err != nil {
        return nil, err
    }

    dlpigeon := conn.PigeonSystem()
    
    return &PigeonSystem{dlpigeon}, nil
}
