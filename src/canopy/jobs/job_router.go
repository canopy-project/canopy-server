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

func ApiInfoHandler(req pigeon.Request, resp pigeon.Response) {
    resp.SetBody(map[string]interface{}{
        "result" : "ok",
        "service-name" : "Canopy Cloud Service",
        "version" : "0.9.2-beta",
        "config" : info.Config.ToJsonObject(),
    })
}

func InitJobServer(cfg config.Config) error {
    pigeon, err = jobqueue.NewPigeonSystem(cfg)
    if err != nil {
        return err
    }

    worker, err = pigeon.StartWorker("localhost") // TODO use configured hostname

    worker.ListenHandlerFunc("api/info", ApiInfoHandler)
}
