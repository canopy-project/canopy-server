// Copyright 2014 SimpleThings, Inc.
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

package config

import (
    "fmt"
    "io/ioutil"
    "encoding/json"
)

type CanopyConfig struct {
    allowOrigin string
    hostname string
    defaultProxyTarget string
    webManagerPath string
    javascriptClientPath string
}

func (config *CanopyConfig) LoadConfig() error {
    return config.LoadConfigFile("/etc/canopy/canopy-server.conf")
}


func (config *CanopyConfig) LoadConfigFile(filename string) error {
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }

    s := string(bytes)

    return config.LoadConfigJsonString(s)
}

func (config *CanopyConfig) LoadConfigJsonString(jsonString string) error {
    var jsonObj map[string]interface{}

    err := json.Unmarshal([]byte(jsonString), &jsonObj)
    if err != nil {
        return err
    }

    return config.LoadConfigJson(jsonObj)

}

func (config *CanopyConfig) LoadConfigJson(jsonObj map[string]interface{}) error {
    for k, v := range jsonObj {
        ok := false
        switch k {
        case "allow-origin":
            config.allowOrigin, ok = v.(string)
        case "hostname": 
            config.hostname, ok = v.(string)
        case "default-proxy-target": 
            config.defaultProxyTarget, ok = v.(string)
        case "web-manager-path": 
            config.webManagerPath, ok = v.(string)
        case "js-client-path": 
            config.javascriptClientPath, ok = v.(string)
        default:
            return fmt.Errorf("Unknown configuration option: %s", k)
        }

        if !ok {
            return fmt.Errorf("Expected string value for %s", k)
        }
    }
    return nil
}

func (config *CanopyConfig) OptAllowOrigin() string {
    return config.allowOrigin
}

func (config *CanopyConfig) OptHostname() string {
    return config.hostname
}

func (config *CanopyConfig) OptDefaultProxyTarget() string {
    return config.defaultProxyTarget
}

func (config *CanopyConfig) OptJavascriptClientPath() string {
    return config.javascriptClientPath
}

func (config *CanopyConfig) OptWebManagerPath() string {
    return config.webManagerPath
}
