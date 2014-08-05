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

// Datalayer provides an abstracted interface for interacting with Canopy's
// backend perstistant datastore.
type Datalayer interface {
    // Connect to the database named <keyspace>.
    func Connect(keyspace string) Connection, error

    // Completely erase the database named <keyspace>.  Handle with care!
    func EraseDb(keyspace string) error

    // Prepare (i.e., create) a new database named <keyspace>.
    func PrepDb(keyspace string) error
}

// Connection is a connection to the database.
type Connection interface {
    // Close this database connection.  Any subsequent calls using this
    // interface will return an error.
    func Close()

    // Create a new user account in the database.
    func CreateAccount(username, email, password string) (*Account, error)

    // Create a new device in the database.
    func CreateDevice(name string) (*Device, error)

    // Remove a user account from the database.
    func DeleteAccount(username string)

    // Lookup a user account from the database (without password verification).
    func LookupAccount(usernameOrEmail string) (*Account, error)

    // Lookup a user account from the database (with password verification).
    // Returns an error if the account is not found, or if the password is
    // correct.
    func LookupAccountVerifyPassword(usernameOrEmail, password string) (*Account, error)

    // Lookup a device from the database.
    func LookupDevice(deviceId gocql.UUID) (*Device, error)

    // Lookup a device from the database, using string representation of its
    // UUID.
    func LookupDeviceByStringID(id string) (*Device, error)
}

// Account is a user account
type Account interface {

    // Get all devices that user has access to.
    func Devices() ([]*Device, error)

    // Get device by ID, but only if this account has access to it.
    func Device(id gocql.UUID) (*Device, error)

    // Get user's email address.
    func Email() string

    // Get user's username.
    func Username() string

    // Verify user's password.  Returns true if password is correct.
    func VerifyPassword(password string) bool
}

// Device is a Canopy-enabled device
type Device interface {

    // Get historic sample data for a property.
    // <property> must be an sddl.Control or an sddl.Sensor.
    func HistoricData(property sddl.Property, startTime, endTime time.Time) ([]sddl.PropertySample, error)

    // Get historic sample data for a property, by property name.
    func HistoricDataByPropertyName(propertyName string, startTime, endTime time.Time) ([]sddl.PropertySample, error)

    // Get the UUID of this device.
    func ID() gocql.UUID

    // Store a data sample from a control or sensor.
    // <property> must be an sddl.Control (with ControlType() == "parameter") or
    // an sddl.Sensor.
    // <value> must have an appropriate dynamic type.  See documentation in
    // sddl/sddl_sample.go for more details.
    func InsertSample(property sddl.Property, t time.Time, value interface{}) error

    // Get latest sample data for a property.
    //
    // property must be an sddl.Control or an sddl.Sensor.
    func LatestData(property sddl.Property) ([]sddl.PropertySample, error)

    // Get latest sample data for a property, by property name.
    func LatestDataByPropertyName(propertyName string) ([]sddl.PropertySample, error)

    // Lookup a property by name.  Essentially, shorthand for:
    //      device.SDDLClass().LookupProperty(propertyName)
    func LookupProperty(propertyName string) (sddl.Property, error)

    // Get the user-assigned name for this device.
    func Name() string

    // Get the SDDL class for this device.  Returns nil if class is unknown
    // (which may happen for newly provisioned devices that haven't sent any
    // reports yet).
    func SDDLClass() *sddl.Class

    // Get the SDDL class for this device, as a marshalled JSON string.
    // Returns "" if class is unknown (which may happen for newly provisioned
    // devices that haven't sent any reports yet).
    func SDDLClassString() string

    // Set the access and sharing permissions that an account has for this
    // device.
    func SetAccountAccess(account *Account, access AccessLevel, sharing ShareLevel) error

    // Set the user-assigned location note for this device.
    func SetLocationNote(locationNote string) error

    // Set the user-assigned name for this device.
    func SetName(name string) error

    // Set the SDDL class associated with this device.
    func SetSDDLClass(class *sddl.Class) error
}

// Obtain default Datalayer interface. (Currently, there is only one backend:
// Cassandra).
func NewDatalayer() Datalayer {
    return cassandra_datalayer.NewCassDatalayer()
}
