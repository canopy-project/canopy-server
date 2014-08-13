/*
 * Copyright 2014 Gregory Prisament
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Logging library for Canopy.
//
// Warnings and errors are always logged to /var/log/canopy/ccs-error.log
// Additional trace and debug info is logged to /var/log/canopy/ccs.log
//
// ENV VAR                  cmd-line arg                  Result                    Default
// ----------------------------------------------------------------------------------------
// CANOPY_LOG_HTTP_ACCESS   --canopy-(no)-log-access      HTTP access is logged     ENABLED
// CANOPY_LOG_REQUESTS      --canopy-(no)-log-requests    HTTP access is logged     ENABLED
package canolog

import (
    "log"
    "os"
    "fmt"
)

type CanopyLogger struct {
    logger *log.Logger
    logFile *os.File
    errorLogFile *os.File
    logRequests bool
    logTraces bool
}

var std = CanopyLogger{}

// Initialize Canopy logger
func Init() error {
    var err error
    std.logFile, err = os.OpenFile("/var/log/canopy/ccs.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    if err != nil {
        return fmt.Errorf("Error opening file /var/log/canopy/ccs.log: ", err)
    }
    std.logger = log.New(std.logFile, "", log.LstdFlags | log.Lshortfile)
    return nil
}

// Close Canopy log file
func Shutdown() {
    std.logger.Println("Terminating Canopy Cloud Service");
    std.logFile.Close()
}

// Log a request or response body
func Request(v ...interface{}) {
    if (std.logRequests) {
        std.logger.Output(2, fmt.Sprintln(v...))
    }
}

// Log an error
func Error(v ...interface{}) {
    std.logger.Output(2, fmt.Sprintln(v...))
}

// Log a warning
func Warn(v ...interface{}) {
    std.logger.Output(2, fmt.Sprintln(v...))
}

// Log an information statement
func Info(v ...interface{}) {
    std.logger.Output(2, fmt.Sprintln(v...))
}

// Log a debug trace message
func Trace(v ...interface{}) {
    std.logger.Output(2, fmt.Sprintln(v...))
}
