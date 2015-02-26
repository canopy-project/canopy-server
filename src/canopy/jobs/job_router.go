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

package jobs

import (
    "canopy/canolog"
    "canopy/config"
    "canopy/jobqueue"
    "canopy/jobs/rest"
)

func InitJobServer(cfg config.Config) error {
    pigeon, err := jobqueue.NewPigeonSystem(cfg)
    if err != nil {
        return err
    }

    userCtx := map[string]interface{}{
        "cfg" : cfg,
    }

    server, err := pigeon.StartServer("localhost") // TODO use configured hostname
    if err != nil {
        return err
    }

    err = server.Handle("api/info", rest.ApiInfoHandler, userCtx)
    if err != nil {
        return err
    }

    err = server.Handle("api/activate", rest.RestJobWrapper(rest.ApiActivateHandler), userCtx)
    if err != nil {
        return err
    }
    return nil
}

func InitJobClient(cfg config.Config) error {
    pigeon, err := jobqueue.NewPigeonSystem(cfg)
    if err != nil {
        return err
    }

    client := pigeon.NewClient()

    respChan, err := client.Launch("api/info", map[string]interface{}{"hi" : "there"})

    resp := <-respChan
    canolog.Info(resp)

    return nil
}
    
