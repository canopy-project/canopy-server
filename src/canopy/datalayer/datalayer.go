/*
 * Copyright 2014-2015 Canopy Services, Inc.
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
package datalayer

import (
    "canopy/cloudvar"
    "canopy/sddl"
    "github.com/gocql/gocql"
    "time"
    "errors"
)

var InvalidPasswordError = errors.New("Incorrect password")

// AccessLevel is the access permissions an account has for a device.
type AccessLevel int
const (
    NoAccess = iota
    ReadOnlyAccess
    ReadWriteAccess
)

// ShareLevel is the sharing permissions an account has for a device.
type ShareLevel int
const (
    NoSharing = iota
    SharingAllowed
    ShareRevokeAllowed
)

type NotificationType int
const (
    NotificationType_LowPriority = iota
    NotificationType_MedPriority
    NotificationType_HighPriority
    NotificationType_SMS
    NotificationType_Email
    NotificationType_InApp
)

// Datalayer provides an abstracted interface for interacting with Canopy's
// backend perstistant datastore.
type Datalayer interface {
    // Connect to the database named <keyspace>.
    Connect(keyspace string) (Connection, error)

    // Completely erase the database named <keyspace>.  Handle with care!
    EraseDb(keyspace string) error

    // Prepare (i.e., create) a new database named <keyspace>.
    PrepDb(keyspace string) error

    // Migrate database from one version to another
    MigrateDB(keyspace, startVersion, endVersion string) error
}

// Connection is a connection to the database.
type Connection interface {
    // Truncate all sensor data from the database.  Use with care!
    ClearSensorData()

    // Close this database connection.  Any subsequent calls using this
    // interface will return an error.
    Close()

    // Create a new user account in the database.
    CreateAccount(username, email, password string) (Account, error)

    // Create a new device in the database.  If <uuid> is nil, then the
    // implementation will assign a newly created UUID.  If <secretKey> is nil,
    // then the implementation will assign a newly created Secret Key.
    CreateDevice(name string, uuid *gocql.UUID, secretKey string, publicAccessLevel AccessLevel) (Device, error)

    // Remove a user account from the database.
    DeleteAccount(username string) error

    // Lookup a user account from the database (without password verification).
    LookupAccount(usernameOrEmail string) (Account, error)

    // Lookup a user account from the database (with password verification).
    // Returns an error if the account is not found, or if the password is
    // incorrect.
    LookupAccountVerifyPassword(usernameOrEmail, password string) (Account, error)

    // Lookup a device from the database (without secret key verification).
    LookupDevice(deviceId gocql.UUID) (Device, error)

    // Lookup a device from the database and verify the secret key
    LookupDeviceVerifySecretKey(deviceId gocql.UUID, secret string) (Device, error)

    // Lookup a device from the database, using string representation of its
    // UUID (without secret key verification).
    LookupDeviceByStringID(id string) (Device, error)

    // Lookup a device from the database, using string representation of its
    // UUID, and verify the secret key.
    LookupDeviceByStringIDVerifySecretKey(id, secret string) (Device, error)

    // Get the datalayer interface for the Pigeon system
    PigeonSystem() PigeonSystem
}

// Account is a user account
type Account interface {
    // Get the account's activation code.
    ActivationCode() string

    // Mark this account as activated, using an activation code.
    Activate(username, code string) error

    // Get all devices that user has access to.
    Devices() ([]Device, error)

    // Get device by ID, but only if this account has access to it.
    Device(id gocql.UUID) (Device, error)

    // Get user's email address.
    Email() string

    // Generate a new Reset Password Code that expires in 24 hours, replacing
    // any existing Reset Password Code.  Saves it to the database.
    GenResetPasswordCode() (code string, err error)

    // Has this account been activated?
    IsActivated() bool

    // Reset password.  Like SetPassword but requires a valid Password Reset
    // Code, and invalidates <code> on success.
    ResetPassword(code, newPassword string) error

    // Set email.  This also causes the account to go back to un-activated
    // status and a new activation code is generated.  Saves changes to the
    // database.
    SetEmail(newEmail string) error

    // Set password
    SetPassword(string) error

    // Get user's username.
    Username() string

    // Verify user's password.  Returns true if password is correct.
    VerifyPassword(password string) bool
}

// Device is a Canopy-enabled device
type Device interface {
    // Extend the SDDL by adding Cloud Variables
    ExtendSDDL(jsn map[string]interface{}) error

    // Get historic sample data for a Cloud Variable.
    HistoricData(varDef sddl.VarDef, curTime, startTime, endTime time.Time) ([]cloudvar.CloudVarSample, error)

    // Get historic sample data for a Cloud Variable, by name.
    HistoricDataByName(cloudVarName string, curTime, startTime, endTime time.Time) ([]cloudvar.CloudVarSample, error)

    // Get historic notifications originating from this device
    HistoricNotifications() ([]Notification, error)

    // Get the UUID of this device.
    ID() gocql.UUID

    // Get the UUID of this device.
    IDString() string

    // Store a Cloud Variable data sample.
    // <value> must have an appropriate dynamic type.  See documentation in
    // cloudvar/cloudvar.go for more details.
    InsertSample(varDef sddl.VarDef, t time.Time, value interface{}) error

    // Store a record of a notification.
    InsertNotification(notifyType int, t time.Time, msg string) error

    // Get last time communication occurred with the server
    // Return nil if device has never interacted with the server.
    LastActivityTime() *time.Time

    // Get latest sample data for a Cloud Variable.
    LatestData(varDef sddl.VarDef) (*cloudvar.CloudVarSample, error)

    // Get latest sample data for a Cloud Variable, by name.
    LatestDataByName(cloudVarName string) (*cloudvar.CloudVarSample, error)

    // Get the user-assigned note about device's location
    LocationNote() string

    // Lookup a Cloud Variable by name.  Essentially, shorthand for:
    //      device.SDDLDocument().LookupVarDef(cloudVarName)
    LookupVarDef(cloudVarName string) (sddl.VarDef, error)

    // Get the user-assigned name for this device.
    Name() string

    // Get the public access level
    PublicAccessLevel() AccessLevel

    // Get the SDDL document for this device.  Returns nil if document is
    // unknown (which may happen for newly provisioned devices that haven't
    // sent any reports yet).
    SDDLDocument() sddl.Document

    // Get the SDDL document for this device, as a marshalled JSON string.
    // Returns "" if document is unknown (which may happen for newly
    // provisioned devices that haven't sent any reports yet).
    SDDLDocumentString() string

    // Get the Secret Key for this device.  The Secret Key is used to
    // authenticate messages coming from the device.
    SecretKey() string

    // Set the access and sharing permissions that an account has for this
    // device.
    SetAccountAccess(account Account, access AccessLevel, sharing ShareLevel) error

    // Set the user-assigned location note for this device.
    SetLocationNote(locationNote string) error

    // Set the user-assigned name for this device.
    SetName(name string) error

    // Set the SDDL class associated with this device.
    SetSDDLDocument(doc sddl.Document) error

    // Update the last activity timestamp.
    // If <t> is nil, the current server time is used.  Otherwise, the last
    // activity timestamp is set to *t.
    // Saves the data to the database.
    UpdateLastActivityTime(t *time.Time) error

    // Update websocket connectivity status
    // Saves the data to the database
    UpdateWSConnected(connected bool) error

    // Gets whether or not this device is connected to the database.
    // (Does not fetch from DB)
    WSConnected() bool
}

// Notification is a record of a message sent to the device owner originiating
// from the device.
type Notification interface {
    // Get the date & time that this notification was sent.
    Datetime() time.Time

    // Mark this notification as dismissed.
    Dismiss() error

    // Has this notification been dismissed?
    IsDismissed() bool
 
    // Get the notification message
    Msg() string

    // Get the requested notification type.
    NotifyType() int
}

type PigeonSystem interface {
    // List all workers that are listening for <key>.
    // Returns list of hostnames
    GetListeners(key string) ([]string, error)
    
    // Register that a worker is listening for <key>.
    RegisterListener(hostname, key string) error

    // Register that a worker exists
    RegisterWorker(hostname string) error
}

type ValidationError struct {
    msg string
}

func (err ValidationError) Error() string {
    return err.msg
}

func NewValidationError(msg string) *ValidationError {
    return &ValidationError{msg}
}
