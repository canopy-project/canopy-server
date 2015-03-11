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
//  Pigeon is Canopy's distributed message passing system.  Pigeon can
//  efficient pass messages locally or to remote servers and integrates
//  naturally with golang's native channels.  Pigeon uses Canopy's database to
//  persistently store routing info, server status and load.
//
//  A "Message" consists of:
//      - A string label called the "MsgKey" that controls where the message
//      is sent.
//      - A gob-able payload of type map[string]interface{}
//
//  The "Outbox" interface allows you to send messages.  When a message is
//  sent, it is sent to one or more inboxes listening for the message's MsgKey.
//  You can control how many inboxes recieve the message.  For example, using
//  outbox.Launch sends the message to exactly 1 inbox, simulating a job queue.
//  Using oubox.Broadcast sends the message to all inboxes that are listening
//  for MsgKey.
//
//  The "Inbox" interface allows you to recieve messages with a particular
//  MsgKey.
//
//  To create an Inbox, you must first be running a Pigeon RPC Server.
//
package jobqueue

import (
    "canopy/config"
    "canopy/datalayer/cassandra_datalayer"
    "time"
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

type HandlerFunc func(msgKey string, userCtx map[string]interface{}, req Request, resp Response)

type Handler interface {
    Handle(jobkey string, userCtx map[string]interface{}, req Request, resp Response)
}

type System interface {
    // Create a new outbox object.
    NewOutbox() Outbox

    // Create a new empty response object.
    NewResponse() Response

    // Starts RPC server, adds worker to the DB, if not already present, and
    // sets its status to "active".
    StartServer(hostname string) (Server, error)

    // Lookup a specific Server by hostname.
    //Server(hostname string) (Server, error)

    // Obtain list of Servers from the DB.
    //Servers() ([]ServerInfo, error)
}

type Server interface {
    // Create (and register) a new Inbox that recieves messages labelled
    // msgKey.
    CreateInbox(msgKey string) (Inbox, error)

    // Set the Server's status to "active".  Does nothing if server is already
    // "active".
    Start() error

    // Get the Server's status
    Status() (StatusEnum, error)

    // Set the Server's status to "stopped".  It will no longer recieve
    // requests until started again.  Does nothing if worker is already
    // "stopped".
    Stop() error
}

type Inbox interface {
    // Close (cleanup & shutdown) this inbox.
    // After this is called Handler will no longer be triggered and this
    // object's methods will all return "Inbox closed" errors.
    Close() error

    // Get the MsgKey that this inbox is listening for.
    MsgKey() string

    // Resume listening for MsgKey after a call to .Suspend().  Returns an
    // error if inbox is not suspended.
    Resume() error

    // Obtain the Server object that this Inbox is using for RPC.
    Server() Server

    // Set the handler that should be triggered when messages with MsgKey are
    // recieved.
    SetHandler(handler Handler) error

    // Set the handler that should be triggered when messages with MsgKey are
    // retrieved.  This sets the handler using a function instead of a Handler
    // object.
    SetHandlerFunc(fn HandlerFunc) error

    // Temporarily stop listening for MsgKey.  Call .Resume() to resume.
    // Returns an error if inbox is already suspended.
    Suspend() error

    // Set additional data that should be passed to handler.
    SetUserCtx(userCtx interface{})
}

type Outbox interface {
    // Broadcast a request to every interested Inbox
    Broadcast(msgKey string, payload map[string]interface{}) error

    // Launches a request that will be handled by exactly one Server
    Launch(msgKey string, payload map[string]interface{}) (<-chan Response, error)
    
    // Launches a request that is idemponent and can be consumed by multiple
    // Servers without ill effect.  This allows the job to be sent to
    // multiple consumers simultaneously, for low latency response (whoever
    // responds first wins).
    LaunchIdempotent(msgKey string, numParallel uint32, payload map[string]interface{}) (<-chan Response, error)

    // Set the timeout for non-broadcast requests.
    // Use a negative value for no timeout.
    SetTimeoutms(timeout int32)
}

type RecieveHandler interface {
    Recieve(timeout time.Duration) (map[string]interface{}, error)
    Handle(jobkey string, userCtx map[string]interface{}, req Request, resp Response)
}

func NewRecieveHandler() RecieveHandler{
    return NewPigeonRecieveHandler()
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
