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
package endpoints

import (
    "encoding/json"
    "fmt"
    "canopy/datalayer/cassandra_datalayer"
    "canopy/canolog"
    //"canopy/mail"
    "net/http"
)

func POST_create_account(w http.ResponseWriter, r *http.Request) {
    canolog.Trace("POST_create_account")
    writeStandardHeaders(w);

    canolog.Trace("Decoding payload")
    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    canolog.Trace("Reading username")
    username, ok := data["username"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"string_username_expected\"}")
        return
    }

    canolog.Trace("Reading email")
    email, ok := data["email"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"string_email_expected\"}")
        return
    }

    canolog.Trace("Reading password")
    password, ok := data["password"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"string_password_expected\"}")
        return
    }

    canolog.Trace("Getting session")
    session, _ := store.Get(r, "canopy-login-session")

    canolog.Trace("Connecting to DB")
    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()

    canolog.Trace("Creating acct")
    account, err := conn.CreateAccount(username, email, password)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"creating_account\"}")
        return
    }

    canolog.Trace("Updating Session")
    session.Values["logged_in_username"] = username
    err = session.Save(r, w)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"saving_session\"}")
        return
    }


    /*canolog.Trace("Sending email")
    mailer, err := mail.NewDefaultMailClient()
    if (err != nil) {
        canolog.Error(err)
        fmt.Fprintf(w, "{\"error\" : \"initializing_mail_client\"}")
        return
    }

    msg := mailer.NewMail();
    msg.AddTo(account.Email(), account.Username())
    msg.SetFrom("no-reply@canopy.link", "Canopy Cloud Service")
    msg.SetReplyTo("no-reply@canopy.link")
    msg.SetSubject("Welcome to Canopy")
    msg.SetHTML("Thank you for creating a Canopy account!")
    err = mailer.Send(msg)
    if (err != nil) {
        fmt.Fprintf(w, "{\"error\" : \"sending_email\"}")
        return
    }*/

    canolog.Trace("All done!")
    fmt.Fprintf(w, "{\"result\" : \"ok\", \"username\" : \"%s\", \"email\" : \"%s\"}",
        account.Username(),
        account.Email())
    return
}

