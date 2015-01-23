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

type Config interface {
    ToString() string
    ToJsonObject() map[string]interface{}

    LoadConfig() error
    LoadConfigFile(filename string) error
    LoadConfigJson(jsonObj map[string]interface{}) error
    LoadConfigCLI() error
    LoadConfigEnv() error

    OptAllowAnonDevices() bool
    OptAllowOrigin() string
    OptEmailService() string
    OptEnableHTTP() bool
    OptEnableHTTPS() bool
    OptForwardOtherHosts() string
    OptHostname() string
    OptHTTPPort() int16
    OptHTTPSPort() int16
    OptJavascriptClientPath() string
    OptLogFile() string
    OptProductionSecret() string
    OptSendgridUsername() string
    OptSendgridSecretKey() string
    OptWebManagerPath() string
}

func NewDefaultConfig() Config {
    return &CanopyConfig{
        enableHTTPS: true,
        httpPort: 80,
        httpsPort: 443,
        logFile: "/var/log/canopy/canopy-server.log",
    }
}

// Reads the "log-file" option without reading the whole config.  This allows
// us to log errors relating to configuration to the right place.
func JustGetOptLogFile() string {
    return justGetOptLogFile()
}
