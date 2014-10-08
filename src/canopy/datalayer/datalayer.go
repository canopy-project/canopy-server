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
package datalayer

import (
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
    // implementation will assign a newly created UUID.
    CreateDevice(name string, uuid *gocql.UUID, publicAccessLevel AccessLevel) (Device, error)

    // Remove a user account from the database.
    DeleteAccount(username string)

    // Lookup a user account from the database (without password verification).
    LookupAccount(usernameOrEmail string) (Account, error)

    // Lookup a user account from the database (with password verification).
    // Returns an error if the account is not found, or if the password is
    // incorrect.
    LookupAccountVerifyPassword(usernameOrEmail, password string) (Account, error)

    // Lookup a device from the database.
    LookupDevice(deviceId gocql.UUID) (Device, error)

    // Lookup a device from the database, using string representation of its
    // UUID.
    LookupDeviceByStringID(id string) (Device, error)

    // Lookup a device from the database by UUID, creating a new device with
    // that UUID if none already exists.
    LookupOrCreateDevice(deviceId gocql.UUID, publicAccessLevel AccessLevel) (Device, error)
}

// Account is a user account
type Account interface {
    // Get all devices that user has access to.
    Devices() ([]Device, error)

    // Get device by ID, but only if this account has access to it.
    Device(id gocql.UUID) (Device, error)

    // Get user's email address.
    Email() string

    // Get user's username.
    Username() string

    // Verify user's password.  Returns true if password is correct.
    VerifyPassword(password string) bool
}

// Device is a Canopy-enabled device
type Device interface {
    // Get historic sample data for a property.
    // <property> must be an sddl.Control or an sddl.Sensor.
    HistoricData(property sddl.Property, startTime, endTime time.Time) ([]sddl.PropertySample, error)

    // Get historic sample data for a property, by property name.
    HistoricDataByPropertyName(propertyName string, startTime, endTime time.Time) ([]sddl.PropertySample, error)

    // Get historic notifications originating from this device
    HistoricNotifications() ([]Notification, error)

    // Get the UUID of this device.
    ID() gocql.UUID

    // Store a data sample from a control or sensor.
    // <property> must be an sddl.Control (with ControlType() == "parameter") or
    // an sddl.Sensor.
    // <value> must have an appropriate dynamic type.  See documentation in
    // sddl/sddl_sample.go for more details.
    InsertSample(property sddl.Property, t time.Time, value interface{}) error

    // Store a record of a notification.
    InsertNotification(notifyType int, t time.Time, msg string) error

    // Get latest sample data for a property.
    //
    // property must be an sddl.Control or an sddl.Sensor.
    LatestData(property sddl.Property) (*sddl.PropertySample, error)

    // Get latest sample data for a property, by property name.
    LatestDataByPropertyName(propertyName string) (*sddl.PropertySample, error)

    // Lookup a property by name.  Essentially, shorthand for:
    //      device.SDDLClass().LookupProperty(propertyName)
    LookupProperty(propertyName string) (sddl.Property, error)

    // Get the user-assigned name for this device.
    Name() string

    // Get the public access level
    PublicAccessLevel() AccessLevel

    // Get the SDDL class for this device.  Returns nil if class is unknown
    // (which may happen for newly provisioned devices that haven't sent any
    // reports yet).
    SDDLClass() *sddl.Class

    // Get the SDDL class for this device, as a marshalled JSON string.
    // Returns "" if class is unknown (which may happen for newly provisioned
    // devices that haven't sent any reports yet).
    SDDLClassString() string

    // Set the access and sharing permissions that an account has for this
    // device.
    SetAccountAccess(account Account, access AccessLevel, sharing ShareLevel) error

    // Set the user-assigned location note for this device.
    SetLocationNote(locationNote string) error

    // Set the user-assigned name for this device.
    SetName(name string) error

    // Set the SDDL class associated with this device.
    SetSDDLClass(class *sddl.Class) error
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

