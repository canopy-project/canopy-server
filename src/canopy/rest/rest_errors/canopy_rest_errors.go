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
package rest_errors

import(
    "net/http"
    "fmt"
)

type CanopyRestError interface {
    // Set error headers and payload depending on the error that occured
    WriteTo(http.ResponseWriter)
}

type DatabaseConnectionError struct {}
func (DatabaseConnectionError) WriteTo(w http.ResponseWriter) {
    w.WriteHeader(http.StatusInternalServerError);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "could_not_connect_to_database"}`)
}

func NewDatabaseConnectionError() CanopyRestError {
    return &DatabaseConnectionError{}
}

type NotLoggedInError struct {}
func (NotLoggedInError) WriteTo(w http.ResponseWriter) {
    w.WriteHeader(http.StatusUnauthorized);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "not_logged_in"}`)
}

func NewNotLoggedInError() CanopyRestError {
    return &NotLoggedInError{}
}

type BadInputError struct {
    msg string
}
func (err BadInputError) WriteTo(w http.ResponseWriter) {
    w.WriteHeader(http.StatusBadRequest);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "bad_input", "error_msg" : "%s"}`, err.msg)
}

func NewBadInputError(msg string) CanopyRestError {
    return &BadInputError{msg}
}

type InternalServerError struct {
    msg string
}
func (err InternalServerError) WriteTo(w http.ResponseWriter) {
    w.WriteHeader(http.StatusInternalServerError);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "internal_error", "error_msg" : "%s"}`, err.msg)
}

func NewInternalServerError(msg string) CanopyRestError {
    return &BadInputError{msg}
}
