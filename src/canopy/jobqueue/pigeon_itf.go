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
//  Pigeon is Canopy's message passing system.  Its flexible design allows
//  it to be used for a variety of distributed computing tasks.
//
//  A "Worker" is a server that can listen for Requests.
//
//  A "Launcher" is an object used for sending Requests to workers.
//
//  A "Request" consists of a string name (key) and a JSON payload.
//
//  A "Response" consists of a JSON payload and an optional error object.
//
// WORKERS
//
//  A Worker is identified by IP Address or Hostname.  Information about each
//  Worker is stored in the database.
//
//  To register a new Worker, use:
//
//      worker, err := pigeonSys.StartWorker(hostname)
//
//  You can then start listening for requests that match a desired key:
//
//      err = worker.ListenHandler("generic", myHandlerFunc)
//
//  Before your program quits, we advise that you Stop the worker.  Otherwise,
//  requests will continue to be sent to it and will time out.
//
//      err = worker.Stop()
//
//  All data about workers including their status and what they are listening
//  for are stored in the DB.
//
// LAUNCHERS
//
//  To send a message you must first create a Launcher object.  A Launcher
//  contains the settings that will be used to send the request.
//
//      launcher := pigeonSys.NewLauncher()
//
//  You can then set options:
//
//      launcher.SetProtocol(pigeon.GO_CHANNEL, pigeon.HTTP)
//      launcher.SetTimeoutms(1000)
//
//  To send a request that should be consumed by exactly one Worker:
//
//      responseChan := launcher.Launch("generic", myPayload)
//
//  To block waiting for the response:
//
//      response := <-responseChan
//


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

type System interface {
    // Adds worker to the DB, if not already present, and sets its
    // status to "active".
    StartWorker(hostname string) (Worker, error)

    // Lookup a specific worker by hostname.
    Worker(hostname string) (Worker, error)

    // Obtain list of workers from the DB.
    Workers() ([]Worker, error)

}

type Launcher interface {
    // Broadcast a payload to every listener interested in these messages
    Broadcast(name string, payload map[string]inteface{})

    // Launches a work item that will be consumed by exactly one listener
    Launch(key string, payload map[string]inteface{}) <-chan PigeonResponse
    
    // Launches a work item that is idemponent and can be consumed by multiple
    // listeners without ill effect.  This allows the job to be sent to
    // multiple consumers simultaneously, for low latency response (whoever
    // responds first wins).
    LaunchIdempotent(key string, int numParallel, payload map[string]inteface{}) <-chan PigeonResponse

    // Set the protocols to use.  Use pigeon.AUTO to let the implementation
    // decide.
    // <localProtocol> is the protocol to use when the hostname matches the
    // local system's.
    // <globalProtocol> is the protocol to use when the hostname is for a
    // remote server.
    SetProtocol(localProtocol ProtocolEnum, remoteProtocol ProtocolEnum)

    // Set the timeout for non-broadcast requests.
    SetTimeoutMilliseconds(timeout uint32)
}

type Worker interface {
    // Listen for requests that match <key>.
    // This is the low-level interface for listening for requests.
    Listen(key string, request <-chan Request, response ->chan Response) error

    // Listen for requests that match <key>, triggering a handle function each
    // time a request is recieved.
    ListenHandler(key string, func handler(req Request) resp Response) error

    // Set the worker's status to "active".  Does nothing if worker is already
    // "active".
    Start() error

    // Get the worker's status
    Status() error

    // Set the worker's status to "stopped".  It will no longer recieve
    // requests until started again.  Does nothing if worker is already
    // "stopped".
    Stop() error

    // Stop listening for a specific <key>.
    StopListening(key string) error
}

type Request interface {
    Body map[string]interface{}
}

type Response interface {
    Error() err
    Body() map[string]interface{}
}
