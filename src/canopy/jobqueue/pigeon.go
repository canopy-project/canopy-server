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

package jobqueue

import (
    "canopy/datalayer"
    "errors"
    "fmt"
    "time"
)

type PigeonSystem struct {
    dl datalayer.PigeonSystem
}

type PigeonRequest struct {
    ReqJobKey string
    ReqBody map[string]interface{}
}

type PigeonResponse struct {
    RespBody map[string]interface{}
}

type PigeonRecieveHandler struct {
    ch chan map[string]interface{}
}

func NewPigeonRecieveHandler() *PigeonRecieveHandler{
    return &PigeonRecieveHandler{
        ch: make(chan map[string]interface{}),
    }
}

func (recvHandler *PigeonRecieveHandler) Handle(jobkey string, 
        userCtx map[string]interface{}, 
        req Request, 
        resp Response) {

    // Send it into channel
    recvHandler.ch <- req.Body()
}

func (recvHandler *PigeonRecieveHandler) Recieve(timeout time.Duration) (map[string]interface{}, error) {

    select {
        case msg := <- recvHandler.ch:
            return msg, nil
        case <- time.After(timeout):
            return nil, errors.New("Recieve timed out")
    }
    
}

func (pigeon *PigeonSystem) NewOutbox() Outbox {
    return &PigeonOutbox{
        sys: pigeon,
        timeoutms: -1,
    }
}

func (pigeon *PigeonSystem) NewResponse() Response {
    return &PigeonResponse{}
}

func (pigeon *PigeonSystem) StartServer(hostname string) (Server, error) {
    server := &PigeonServer{
        sys : pigeon,
        hostname: hostname,
        handlers : map[string]*pigeonHandler{},
    }

    err := server.Start()
    if err != nil {
        return nil, err
    }

    return server, nil
}

func (pigeon *PigeonSystem) Server(hostname string) (Server, error) {
    return nil, fmt.Errorf("Not Implemented")
}

func (pigeon *PigeonSystem) Servers() ([]Server, error) {
    return nil, fmt.Errorf("Not Implemented")
}

func (resp *PigeonResponse) Body() map[string]interface{} {
    return resp.RespBody
}

func (resp *PigeonResponse) SetBody(body map[string]interface{}) {
    resp.RespBody = body
}

func (resp *PigeonResponse) AppendToBody(key string, value interface{}) {
    resp.RespBody[key] = value
}

func (req *PigeonRequest) Body() map[string]interface{} {
    return req.ReqBody
}


