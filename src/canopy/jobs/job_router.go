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
    "canopy/config"
    "canopy/mail"
    "canopy/datalayer/cassandra_datalayer"
    "canopy/jobqueue"
    "canopy/jobs/rest"
)

func InitJobServer(cfg config.Config, pigeonServer jobqueue.Server) error {
    mailer, err := mail.NewMailClient(cfg)
    if err != nil {
        return err
    }

    userCtx := map[string]interface{}{
        "cfg" : cfg,
        "mailer" : mailer,
    }

    dl := cassandra_datalayer.NewDatalayer(cfg)
    conn, err := dl.Connect("canopy")
    if err != nil {
        return err
    }
    userCtx["db-conn"] = conn

    routes := map[string]jobqueue.HandlerFunc{
        "api/activate": rest.RestJobWrapper(rest.ApiActivateHandler),
        "api/create_account": rest.RestJobWrapper(rest.ApiCreateAccountHandler),
        "api/create_devices": rest.RestJobWrapper(rest.ApiCreateDevicesHandler),
        "GET:api/device/id": rest.RestJobWrapper(rest.GET__api__device__id),
        "POST:api/device/id": rest.RestJobWrapper(rest.POST__api__device__id),
        "api/device/id/var": rest.RestJobWrapper(rest.GET__api__device__id__var),
        "api/devices": rest.RestJobWrapper(rest.GET__api__devices),
        "api/finish_share_transaction": rest.RestJobWrapper(rest.POST__api__finish_share_transaction),
        "api/info": rest.RestJobWrapper(rest.GET__api__info),
        "api/login": rest.RestJobWrapper(rest.POST__api__login),
        "api/logout": rest.RestJobWrapper(rest.GET_POST__api__logout),
        "GET:api/user/self": rest.RestJobWrapper(rest.GET__api__user__self),
        "POST:api/user/self": rest.RestJobWrapper(rest.POST__api__user__self),
        "DELETE:api/user/self": rest.RestJobWrapper(rest.DELETE__api__user__self),
        "api/reset_password": rest.RestJobWrapper(rest.POST__api__reset_password),
        "api/share": rest.RestJobWrapper(rest.POST__api__share),
    }

    // Register handlers
    for msgKey, handler := range routes {
        inbox, err := pigeonServer.CreateInbox(msgKey)
        if err != nil {
            return err
        }
        inbox.SetUserCtx(userCtx)
        inbox.SetHandlerFunc(handler)
    }

    return nil
}
