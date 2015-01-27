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
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
)

type CanopyConfig struct {
    allowAnonDevices bool
    allowOrigin string
    emailService string
    enableHTTP bool
    enableHTTPS bool
    forwardOtherHosts string
    hostname string
    httpPort int16
    httpsCertFile string
    httpsPrivKeyFile string
    httpsPort int16
    logFile string
    webManagerPath string
    passwordHashCost int16
    passwordSecretSalt string
    productionSecret string
    sendgridSecretKey string
    sendgridUsername string
    javascriptClientPath string
}

func (config *CanopyConfig) ToString() string {
    return fmt.Sprint(`SERVER CONFIG SETTINGS:
allow-anon-devices:  `, config.allowAnonDevices, `
allow-origin:        `, config.allowOrigin, `
email-service:       `, config.emailService, `
enable-http:         `, config.enableHTTP, `
enable-https:        `, config.enableHTTPS, `
forward-other-hosts: `, config.forwardOtherHosts, `
hostname:            `, config.hostname, `
http-port:           `, config.httpPort, `
https-cert-file:     `, config.httpsCertFile, `
https-port:          `, config.httpsPort, `
https-priv-key-file: `, config.httpsPrivKeyFile, `
js-client-path:      `, config.javascriptClientPath, `
log-file:            `, config.logFile, `
sendgrid-username:   `, config.sendgridUsername, `
web-manager-path:    `, config.webManagerPath)
}

func (config *CanopyConfig) ToJsonObject() map[string]interface{}{
    return map[string]interface{} {
        "allow-anon-devices" : config.allowAnonDevices,
        "allow-origin" : config.allowOrigin,
        "email-service" : config.emailService,
        "enable-http" : config.enableHTTP,
        "enable-https" : config.enableHTTPS,
        "forward-other-hosts" : config.forwardOtherHosts,
        "hostname" : config.hostname,
        "http-port" : config.httpPort,
        "https-cert-file" : config.httpsCertFile,
        "https-port" : config.httpsPort,
        "https-priv-key-file" : config.httpsPrivKeyFile,
        "js-client-path" : config.javascriptClientPath,
        "log-file" : config.logFile,
        "sendgrid-username" : config.sendgridUsername,
        "web-manager-path" : config.webManagerPath,
    }
}

func (config *CanopyConfig) LoadConfig() error {
    err := config.LoadConfigFile("/etc/canopy/server.conf")
    if os.IsNotExist(err) {
        // If file doesn't exist, just move on to the next one.
    } else if err != nil {
        return err
    }

    homeDir := os.Getenv("HOME")
    if homeDir != "" {
        err = config.LoadConfigFile(homeDir + "/.canopy/server.conf")
        if os.IsNotExist(err) {
            // If file doesn't exist, just move on to the next one.
        } else if err != nil {
            return err
        }
    }

    confFile := os.Getenv("CANOPY_SERVER_CONFIG_FILE")
    if confFile != "" {
        err = config.LoadConfigFile(confFile)
        if err != nil {
            // If config file is specified explicitely, it must be readable
            return err
        }
    }

    err = config.LoadConfigEnv()
    if err != nil {
        return err
    }

    err = config.LoadConfigCLI()
    if err != nil {
        return err
    }

    return nil
}

func (config *CanopyConfig) LoadConfigEnv() error {
    allowAnonDevices := os.Getenv("CCS_ALLOW_ANON_DEVICES")
    if allowAnonDevices == "1" || allowAnonDevices == "true" {
        config.allowAnonDevices = true
    } else if allowAnonDevices == "0" || allowAnonDevices == "false" {
        config.allowAnonDevices = false
    } else if allowAnonDevices != "" {
        return fmt.Errorf("Invalid value for CCS_ALLOW_ANON_DEVICES: %s",  allowAnonDevices)
    }

    allowOrigin := os.Getenv("CCS_ALLOW_ORIGIN")
    if allowOrigin != "" {
        config.allowOrigin = allowOrigin
    }

    emailService := os.Getenv("CCS_EMAIL_SERVICE")
    if emailService != "" {
        if !(emailService == "none" || emailService == "sendgrid") {
            return fmt.Errorf("Unknown email service: %s",  emailService)
        }
        config.emailService = emailService
    }

    enableHTTP := os.Getenv("CCS_ENABLE_HTTP")
    if enableHTTP == "1" || enableHTTP == "true" {
        config.enableHTTP = true
    } else if enableHTTP == "0" || enableHTTP == "false" {
        config.enableHTTP = false
    } else if enableHTTP != "" {
        return fmt.Errorf("Invalid value for CCS_ENABLE_HTTP: %s",  enableHTTP)
    }

    enableHTTPS := os.Getenv("CCS_ENABLE_HTTPS")
    if enableHTTPS == "1" || enableHTTPS == "true" {
        config.enableHTTPS = true
    } else if enableHTTPS == "0" || enableHTTPS == "false" {
        config.enableHTTPS = false
    } else if enableHTTPS != "" {
        return fmt.Errorf("Invalid value for CCS_ENABLE_HTTPS: %s",  enableHTTPS)
    }

    forwardOtherHosts := os.Getenv("CCS_FORWARD_OTHER_HOSTS")
    if forwardOtherHosts != "" {
        config.forwardOtherHosts = forwardOtherHosts
    }

    hostname := os.Getenv("CCS_HOSTNAME")
    if hostname != "" {
        config.hostname = hostname
    }

    httpPort := os.Getenv("CCS_HTTP_PORT")
    if httpPort != "" {
        port, err := strconv.ParseInt(httpPort, 0, 16)
        if err != nil {
            return fmt.Errorf("Invalid value for CCS_HTTP_PORT: %s",  httpPort)
        }
        config.httpPort = int16(port)
    }

    httpsCertFile := os.Getenv("CCS_HTTPS_CERT_FILE")
    if httpsCertFile != "" {
        config.httpsCertFile = httpsCertFile
    }

    httpsPort := os.Getenv("CCS_HTTPS_PORT")
    if httpPort != "" {
        port, err := strconv.ParseInt(httpsPort, 0, 16)
        if err != nil {
            return fmt.Errorf("Invalid value for CCS_HTTPS_PORT: %s",  httpsPort)
        }
        config.httpsPort = int16(port)
    }

    httpsPrivKeyFile := os.Getenv("CCS_HTTPS_PRIV_KEY_FILE")
    if httpsPrivKeyFile != "" {
        config.httpsPrivKeyFile = httpsPrivKeyFile
    }

    jsClientPath := os.Getenv("CCS_JS_CLIENT_PATH")
    if jsClientPath != "" {
        config.javascriptClientPath = jsClientPath
    }

    logFile := os.Getenv("CCS_LOG_FILE")
    if logFile != "" {
        config.logFile = logFile
    }

    passwordHashCost := os.Getenv("CCS_PASSWORD_HASH_COST")
    if passwordHashCost != "" {
        hashCost, err := strconv.ParseInt(passwordHashCost, 0, 16)
        if err != nil {
            return fmt.Errorf("Invalid value for CCS_PASSWORD_HASH_COST: %s", passwordHashCost)
        }
        config.passwordHashCost = int16(hashCost)
    }

    passwordSecretSalt := os.Getenv("CCS_PASSWORD_SECRET_SALT")
    if passwordSecretSalt != "" {
        config.passwordSecretSalt = passwordSecretSalt
    }

    productionSecret := os.Getenv("CCS_PRODUCTION_SECRET")
    if productionSecret != "" {
        config.productionSecret = productionSecret
    }

    sendgridSecretKey := os.Getenv("CCS_SENDGRID_SECRET_KEY")
    if sendgridSecretKey != "" {
        config.sendgridSecretKey = sendgridSecretKey
    }

    sendgridUsername := os.Getenv("CCS_SENDGRID_USERNAME")
    if sendgridUsername != "" {
        config.sendgridUsername = sendgridUsername
    }

    webMgrPath := os.Getenv("CCS_WEB_MANAGER_PATH")
    if webMgrPath != "" {
        config.webManagerPath = webMgrPath
    }

    return nil
}

func (config *CanopyConfig) LoadConfigCLI() error {
    allowAnonDevices := flag.String("allow-anon-devices", "", "")
    allowOrigin := flag.String("allow-origin", "", "")
    emailService := flag.String("email-service", "", "")
    enableHTTP := flag.String("enable-http", "", "")
    enableHTTPS := flag.String("enable-https", "", "")
    forwardOtherHosts := flag.String("forward-other-hosts", "", "")
    hostname := flag.String("hostname", "", "")
    httpPort := flag.String("http-port", "", "")
    httpsCertFile := flag.String("https-cert-file", "", "")
    httpsPort := flag.String("https-port", "", "")
    httpsPrivKeyFile := flag.String("https-priv-key-file", "", "")
    jsClientPath := flag.String("js-client-path", "", "")
    logFile := flag.String("log-file", "", "")
    passwordHashCost := flag.String("password-hash-cost", "", "")
    passwordSecretSalt := flag.String("password-secret-salt", "", "")
    productionSecret := flag.String("production-secret", "", "")
    sendgridSecretKey := flag.String("sendgrid-secret-key", "", "")
    sendgridUsername := flag.String("sendgrid-username", "", "")
    webMgrPath := flag.String("web-manager-path", "", "")

    flag.Parse()

    if *allowAnonDevices != "" {
        if *allowAnonDevices == "1" || *allowAnonDevices == "true" {
            config.allowAnonDevices = true
        } else if *allowAnonDevices == "0" || *allowAnonDevices == "false" {
            config.allowAnonDevices = false
        } else if *allowAnonDevices != "" {
            return fmt.Errorf("Invalid value for --allow-anon-devices: %s",  *allowAnonDevices)
        }
    }

    if *allowOrigin != "" {
        config.allowOrigin = *allowOrigin
    }

    if *emailService != "" {
        if !(*emailService == "none" || *emailService == "sendgrid") {
            return fmt.Errorf("Unknown email service: %s",  *emailService)
        }
        config.emailService = *emailService
    }

    if *enableHTTP != "" {
        if *enableHTTP == "1" || *enableHTTP == "true" {
            config.enableHTTP = true
        } else if *enableHTTP == "0" || *enableHTTP == "false" {
            config.enableHTTP = false
        } else if *enableHTTP != "" {
            return fmt.Errorf("Invalid value for --enable-http: %s",  *enableHTTP)
        }
    }

    if *enableHTTPS != "" {
        if *enableHTTPS == "1" || *enableHTTPS == "true" {
            config.enableHTTPS = true
        } else if *enableHTTPS == "0" || *enableHTTPS == "false" {
            config.enableHTTPS = false
        } else if *enableHTTPS != "" {
            return fmt.Errorf("Invalid value for --enable-http: %s",  *enableHTTPS)
        }
    }

    if *forwardOtherHosts != "" {
        config.forwardOtherHosts = *forwardOtherHosts
    }

    if *hostname != "" {
        config.hostname = *hostname
    }

    if *httpPort != "" {
        port, err := strconv.ParseInt(*httpPort, 0, 16)
        if err != nil {
            return fmt.Errorf("Invalid value for --http-port: %s",  *httpPort)
        }
        config.httpPort = int16(port)
    }

    if *httpsCertFile != "" {
        config.httpsCertFile = *httpsCertFile
    }

    if *httpsPort != "" {
        port, err := strconv.ParseInt(*httpsPort, 0, 16)
        if err != nil {
            return fmt.Errorf("Invalid value for --http-ports: %s",  *httpsPort)
        }
        config.httpsPort = int16(port)
    }

    if *httpsPrivKeyFile != "" {
        config.httpsPrivKeyFile = *httpsPrivKeyFile
    }

    if *jsClientPath != "" {
        config.javascriptClientPath = *jsClientPath
    }

    if *logFile != "" {
        config.logFile = *logFile
    }

    if *passwordHashCost != "" {
        hashCost, err := strconv.ParseInt(*passwordHashCost, 0, 16)
        if err != nil {
            return fmt.Errorf("Invalid value for --password-hash-cost: %s",  *passwordHashCost)
        }
        config.passwordHashCost = int16(hashCost)
    }

    if *passwordSecretSalt != "" {
        config.passwordSecretSalt = *passwordSecretSalt
    }

    if *productionSecret != "" {
        config.productionSecret = *productionSecret
    }

    if *sendgridSecretKey != "" {
        config.sendgridSecretKey = *sendgridSecretKey
    }

    if *sendgridUsername != "" {
        config.sendgridUsername = *sendgridUsername
    }

    if *webMgrPath != "" {
        config.webManagerPath = *webMgrPath
    }

    return nil
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
        case "allow-anon-devices":
            config.allowAnonDevices, ok = v.(bool)
        case "allow-origin":
            config.allowOrigin, ok = v.(string)
        case "email-service":
            var emailService string
            emailService, ok = v.(string)
            if !(emailService == "none" || emailService == "sendgrid") {
                return fmt.Errorf("Unknown email service: %s", emailService)
            }
            config.emailService = emailService
        case "enable-http":
            config.enableHTTP, ok = v.(bool)
        case "enable-https":
            config.enableHTTPS, ok = v.(bool)
        case "forward-other-hosts": 
            config.forwardOtherHosts, ok = v.(string)
        case "hostname": 
            config.hostname, ok = v.(string)
        case "http-port": 
            port, ok := v.(int)
            if ok {
                config.httpPort = int16(port)
            }
        case "https-cert-file": 
            config.httpsCertFile, ok = v.(string)
        case "https-port": 
            port, ok := v.(int)
            if ok {
                config.httpsPort = int16(port)
            }
        case "https-priv-key-file": 
            config.httpsPrivKeyFile, ok = v.(string)
        case "js-client-path": 
            config.javascriptClientPath, ok = v.(string)
        case "log-file": 
            config.logFile, ok = v.(string)
        case "password-hash-cost": 
            passwordHashCost, ok := v.(int)
            if ok {
                config.passwordHashCost = int16(passwordHashCost)
            }
        case "password-secret-salt": 
            config.passwordSecretSalt, ok = v.(string)
        case "production-secret": 
            config.productionSecret, ok = v.(string)
        case "sendgrid-secret-key": 
            config.sendgridSecretKey, ok = v.(string)
        case "sendgrid-username": 
            config.sendgridUsername, ok = v.(string)
        case "web-manager-path": 
            config.webManagerPath, ok = v.(string)
        default:
            return fmt.Errorf("Unknown configuration option: %s", k)
        }

        if !ok {
            return fmt.Errorf("Incorrect JSON type for %s", k)
        }
    }
    return nil
}
func (config *CanopyConfig) OptAllowAnonDevices() bool {
    return config.allowAnonDevices
}

func (config *CanopyConfig) OptAllowOrigin() string {
    return config.allowOrigin
}

func (config *CanopyConfig) OptEmailService() string {
    return config.emailService
}

func (config *CanopyConfig) OptEnableHTTP() bool {
    return config.enableHTTP
}

func (config *CanopyConfig) OptEnableHTTPS() bool {
    return config.enableHTTPS
}

func (config *CanopyConfig) OptForwardOtherHosts() string {
    return config.forwardOtherHosts
}

func (config *CanopyConfig) OptHostname() string {
    return config.hostname
}

func (config *CanopyConfig) OptHTTPPort() int16 {
    return config.httpPort
}

func (config *CanopyConfig) OptHTTPSCertFile() string {
    return config.httpsCertFile
}

func (config *CanopyConfig) OptHTTPSPort() int16 {
    return config.httpsPort
}

func (config *CanopyConfig) OptHTTPSPrivKeyFile() string {
    return config.httpsPrivKeyFile
}

func (config *CanopyConfig) OptJavascriptClientPath() string {
    return config.javascriptClientPath
}

func (config *CanopyConfig) OptLogFile() string {
    return config.logFile
}

func (config *CanopyConfig) OptPasswordHashCost() int16 {
    return config.passwordHashCost
}

func (config *CanopyConfig) OptPasswordSecretSalt() string {
    return config.passwordSecretSalt
}

func (config *CanopyConfig) OptProductionSecret() string {
    return config.productionSecret
}

func (config *CanopyConfig) OptSendgridUsername() string {
    return config.sendgridUsername
}

func (config *CanopyConfig) OptSendgridSecretKey() string {
    return config.sendgridSecretKey
}

func (config *CanopyConfig) OptWebManagerPath() string {
    return config.webManagerPath
}

func justGetOptLogFile() string {
    out := "/var/log/canopy/canopy-server.log"

    logFile := os.Getenv("CCS_LOG_FILE")
    if logFile != "" {
        out = logFile
    }

    // TODO: also read config files and command-line
    return out;
}
