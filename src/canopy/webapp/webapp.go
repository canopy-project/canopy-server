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
package webapp

import (
    "fmt"
    "github.com/gorilla/mux"
    "net/http"
)

func AddRoutes(r *mux.Router) {
    r.HandleFunc("/device/{id}", GET_device__id).Methods("GET")
}

func GET_device__id(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    uuidString := vars["id"]
    fmt.Fprint(w, `<!DOCTYPE html>
<meta charset="utf-8">
<!--meta name="viewport" content="width=1024px, initial-scale=1024px, user-scalable=no"-->
<html>
<head>
    <script>
        var DEVICE_UUID = "`, uuidString, `";
    </script>
    <script src="../mgr/canoweb.conf.js"></script>

    <script src="http://ajax.googleapis.com/ajax/libs/jquery/1.11.0/jquery.min.js"></script>

     <link rel="stylesheet" href="//code.jquery.com/ui/1.11.0/themes/smoothness/jquery-ui.css">
     <script src="//code.jquery.com/ui/1.11.0/jquery-ui.js"></script>
     <script '3rdparty/jquery.ui.touch-punch.min.js'></script>

    <script type="text/javascript" src="http://www.google.com/jsapi"></script>
    <script>
        document.write('<script src="../mgr/' + gCanopyWebAppConfiguration.javascriptClientURL + '" type="text/javascript"><\/script>');
    </script>
    <script src="../mgr/nodes/cano.js"></script>
    <script src="../mgr/nodes/cano_account_dropdown.js"></script>
    <script src="../mgr/nodes/cano_button.js"></script>
    <script src="../mgr/nodes/cano_control_small.js"></script>
    <script src="../mgr/nodes/cano_device_control_widget.js"></script>
    <script src="../mgr/nodes/cano_device_settings_popup.js"></script>
    <script src="../mgr/nodes/cano_devices_list_item.js"></script>
    <script src="../mgr/nodes/cano_devices_list.js"></script>
    <script src="../mgr/nodes/cano_devices_menu.js"></script>
    <script src="../mgr/nodes/cano_devices_screen.js"></script>
    <script src="../mgr/nodes/cano_event_panel.js"></script>
    <script src="../mgr/nodes/cano_hover_plot.js"></script>
    <script src="../mgr/nodes/cano_login_dialog.js"></script>
    <script src="../mgr/nodes/cano_login_page.js"></script>
    <script src="../mgr/nodes/cano_main_page.js"></script>
    <script src="../mgr/nodes/cano_option.js"></script>
    <script src="../mgr/nodes/cano_popup.js"></script>
    <script src="../mgr/nodes/cano_plot.js"></script>
    <script src="../mgr/nodes/cano_sharing_popup.js"></script>
    <script src="../mgr/nodes/cano_sensor_small.js"></script>
    <script src="../mgr/nodes/cano_signup_dialog.js"></script>
    <script src="../mgr/nodes/cano_topbar.js"></script>
    <script src="../mgr/nodes/demo/dizon.js"></script>
    <script src="../mgr/canoweb_dispatcher.js"></script>
    <script src="../mgr/canopy_app_util.js"></script>

    <link href='http://fonts.googleapis.com/css?family=Source+Sans+Pro:200,300,400,700|ABeeZee|Titillium+Web:200,300,400,700' rel='stylesheet' type='text/css'>
    <link href='../mgr/canoweb.css' rel='stylesheet' type='text/css'>
</head>

<body>
    <div id="main"></div>
</body>

<script>

var gCanopy = new CanopyClient(gCanopyWebAppConfiguration);
$(function() {
    dispatcher = new CanowebDispatcher(gCanopy);
    gCanopy.onReady(function() {
        if (gCanopy.account) {
            dispatcher.showPage("main")
        }
        else {
            dispatcher.showPage("login");
        }

        gCanopy.onLogout(function() {
            dispatcher.showPage("login");
        });
    });
});

</script>
</html>`)
}

