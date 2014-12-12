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
// If the files above cannot be written to, then falls back to logging to
// STDOUT.
package canolog

import (
    "log"
    "io"
    "os"
    "fmt"
)

type CanopyLogger struct {
    logger *log.Logger
    logFile *os.File
    errorLogger *log.Logger
    errorLogFile *os.File
    warnLogger *log.Logger
    logRequests bool
    logTraces bool
}

var std = CanopyLogger{}

//var noopLogger = log.New(io.MultiWriter(), "", log.LstdFlags | log.Lshortfile)

// If /var/log/canopy files cannot be opened, then fallback to just logging to STDOUT
func InitFallback() error {
    std.logger = log.New(os.Stdout, "", log.LstdFlags | log.Lshortfile)
    std.errorLogger = log.New(os.Stdout, "ERROR ", log.LstdFlags | log.Lshortfile)
    std.warnLogger = log.New(os.Stdout, "WARN ", log.LstdFlags | log.Lshortfile)

    return nil
}

// Initialize Canopy logger
func Init(logFilename string) error {
    var err error
    std.logFile, err = os.OpenFile(logFilename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    if err != nil {
        fmt.Println("Error opening file " + logFilename + ": ", err)
        fmt.Println("Falling back to STDOUT for logging")
        return InitFallback()
    }
    std.logger = log.New(std.logFile, "", log.LstdFlags | log.Lshortfile)

    std.errorLogFile, err = os.OpenFile("/var/log/canopy/ccs-errors.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    if err != nil {
        fmt.Println("Error opening file /var/log/canopy/ccs-errors.log: ", err)
        fmt.Println("Falling back to STDOUT for logging")
        return InitFallback()
    }
    /*std.errorLogger = log.New(io.MultiWriter(std.errorLogFile, std.logFile), "ERROR ", log.LstdFlags | log.Lshortfile)
    std.warnLogger = log.New(io.MultiWriter(std.errorLogFile, std.logFile), "WARN ", log.LstdFlags | log.Lshortfile)

    return nil
}

// Close Canopy log file
func Shutdown() {
    std.logger.Output(2, fmt.Sprintln("Goodbye"));
    if (std.logFile != nil) {
        std.logFile.Close()
    }
    if (std.errorLogFile != nil) {
        std.errorLogFile.Close()
    }
}

// Log a request or response body
func Request(v ...interface{}) {
    if (std.logRequests) {
        std.logger.Output(2, fmt.Sprintln(v...))
    }
}

// Log an error
func Error(v ...interface{}) {
    std.errorLogger.Output(2, fmt.Sprintln(v...))
}

// Log a warning
func Warn(v ...interface{}) {
    std.warnLogger.Output(2, fmt.Sprintln(v...))
}

// Log an information statement
func Info(v ...interface{}) {
    std.logger.Output(2, fmt.Sprintln(v...))
}

// Log a debug trace message
func Trace(v ...interface{}) {
    std.logger.Output(2, fmt.Sprintln(v...))
}

// Log a debug trace message
func Websocket(v ...interface{}) {
    std.logger.Output(2, fmt.Sprintln(v...))
}
