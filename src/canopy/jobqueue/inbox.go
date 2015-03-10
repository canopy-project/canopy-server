// Copyright 2015 Canopy Services, Inc.
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
    "fmt"
)

type PigeonInbox struct {
    server *PigeonServer
    msgKey string
    handler Handler
    userCtx interface{}
}

type funcHandler struct {
    fn HandlerFunc
}

func (inbox PigeonInbox) Close() error {
    return fmt.Errorf("Close not implemented")
}

func (inbox PigeonInbox) MsgKey() string {
    return inbox.msgKey
}

func (inbox PigeonInbox) Resume() error {
    return fmt.Errorf("Resume not implemented")
}

func (inbox PigeonInbox) Server() Server {
    return inbox.server
}

func (inbox PigeonInbox) SetHandler(handler Handler) error {
    inbox.handler = handler
    return nil
}

func (inbox PigeonInbox) SetHandlerFunc(fn HandlerFunc) error {
    inbox.handler = &funcHandler{fn}
    return nil
}

func (inbox PigeonInbox) Suspend() error {
    return fmt.Errorf("Suspend not implemented")
}

func (inbox PigeonInbox) SetUserCtx(userCtx interface{}) {
    inbox.userCtx = userCtx
}

func (handler *funcHandler)Handle(jobkey string, userCtx map[string]interface{}, req Request, resp Response) {
    handler.fn(jobkey, userCtx, req, resp)
}
