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

package rest

import(
    "canopy/canolog"
    "encoding/json"
    "fmt"
    "net/http"
)

type RestError interface {
    StatusCode() int
    ResponseBody() string
    
    // Log the error and return self
    Log() RestError
}

func WriteErrorResponse(w http.ResponseWriter, err RestError) {
    w.WriteHeader(err.StatusCode())
    fmt.Fprintf(w, err.ResponseBody())
}

type GenericRestError struct {
    statusCode int
    responseBody string
}

func (err *GenericRestError) StatusCode() int {
    return err.statusCode
}

func (err *GenericRestError) ResponseBody() string {
    return err.responseBody
}

func (err *GenericRestError) Log() RestError {
    canolog.ErrorCalldepth(3, fmt.Sprintf("HTTP %d: %s", err.statusCode, err.responseBody))
    return err
}

func NewGenericRestError(statusCode int, errorType string, msg string) *GenericRestError {
    body := map[string]interface{}{
        "result" : "error",
        "error_type" : errorType,
    }
    if msg != "" {
        body["error_msg"] = msg
    }

    jsonBytes, err := json.MarshalIndent(body, "", "    ")
    if err != nil {
        canolog.Error("Error marshalling error response.  That's ironic.", err)
        return &GenericRestError{
            statusCode: statusCode,
            responseBody: `{"result" : "error", "error_type" : "internal_error", "error_msg" : "Error encoding error response"}`,
        }
    }
    responseBody := string(jsonBytes)

    return &GenericRestError{
        statusCode: statusCode,
        responseBody: responseBody,
    }
}

func BadInputError(msg string) *GenericRestError {
    return NewGenericRestError(http.StatusBadRequest, "bad_input", msg)
}

func DatabaseConnectionError() *GenericRestError {
    return NewGenericRestError(http.StatusInternalServerError, "internal_error", "Database connection error")
}

func EmailTakenError() *GenericRestError {
    return NewGenericRestError(http.StatusBadRequest, "email_taken", "")
}

func IncorrectUsernameOrPasswordError() *GenericRestError {
    return NewGenericRestError(http.StatusUnauthorized, "incorrect_username_or_password", "")
}

func InternalServerError(msg string) *GenericRestError {
    return NewGenericRestError(http.StatusInternalServerError, "internal_error", msg)
}

func NotLoggedInError() *GenericRestError {
    return NewGenericRestError(http.StatusUnauthorized, "not_logged_in", "")
}

func URLNotFoundError() *GenericRestError {
    return NewGenericRestError(http.StatusNotFound, "url_not_found", "")
}

func UnauthorizedError(msg string) *GenericRestError {
    return NewGenericRestError(http.StatusUnauthorized, "unauthorized", msg)
}

func UsernameNotAvailableError() *GenericRestError {
    return NewGenericRestError(http.StatusBadRequest, "username_not_available", "")
}
