type CassConnection struct {
    dl *CasssandraDatalayer
    session *gocql.Session
}

func (conn *CassConnection) Close() {
    dl.session.Close()
}

func (conn *CassConnection) CreateAccount(username, email, password string) (*CassAccount, error) {
    password_hash, _ := bcrypt.GenerateFromPassword([]byte(password + salt), hashCost)

    // TODO: transactionize
    if err := conn.session.Query(`
            INSERT INTO accounts (username, email, password_hash)
            VALUES (?, ?, ?)
    `, username, email, password_hash).Exec(); err != nil {
        return nil, err
    }

    if err := conn.session.Query(`
            INSERT INTO account_emails (email, username)
            VALUES (?, ?)
    `, email, username).Exec(); err != nil {
        return nil, err
    }

    return &CassAccount{conn, username, email, password_hash}, nil
}

func (conn *CassConnection) CreateDevice(name string) (*CassDevice, error) {
    id := gocql.TimeUUID()
    
    err := conn.session.Query(`
            INSERT INTO devices (device_id, friendly_name)
            VALUE (?, ?)
    `, id, name).Exec()
    if err != nil {
        return nil, err
    }
    return &CassDevice{
        conn: conn,
        deviceId: id,
        friendlyName: name,
        class: nil,         // class gets initialized during first report
        classString: ""
    }
}

func (conn *CassConnection) DeleteAccount(username string) {
    account, _ := conn.LookupAccount(username)
    email := account.Email()

    if err := conn.session.Query(`
            DELETE FROM accounts
            WHERE username = ?
    `, username).Exec(); err != nil {
        log.Print(err)
    }

    if err := conn.session.Query(`
            DELETE FROM account_emails
            WHERE email = ?
    `, email).Exec(); err != nil {
        log.Print(err)
    }
}

func (conn *CassDatalayer)LookupAccount(usernameOrEmail string) (*CassAccount, error) {
    var account CassAccount

    if err := conn.session.Query(`
            SELECT username, email, password_hash FROM accounts 
            WHERE username = ?
            LIMIT 1
    `, usernameOrEmail).Consistency(gocql.One).Scan(
         &account.username, &account.email, &account.password_hash); err != nil {
            return nil, err
    }
    /* TODO: try email if username not found */
    account.conn = conn
    return &account, nil
}

func (dl *CassDatalayer)LookupAccountVerifyPassword(usernameOrEmail string, password string) (*CassAccount, error) {
    account, err := dl.LookupAccount(usernameOrEmail)
    if err != nil {
        return nil, err
    }

    verified := account.VerifyPassword(password)
    if (!verified) {
        return nil, InvalidPasswordError
    }

    return account, nil
}

func (dl *CassDatalayer) LookupDevice(deviceId gocql.UUID) (*CassDevice, error) {
    var device CassDevice

    device.deviceId = deviceId
    device.dl = dl

    err := dl.session.Query(`
        SELECT friendly_name, sddl
        FROM devices
        WHERE device_id = ?
        LIMIT 1`, deviceId).Consistency(gocql.One).Scan(
            &device.friendlyName,
            &device.classString)
    if err != nil {
        return nil, err
    }

    if device.classString != "" {
        device.class, err = sddl.ParseClassString("anonymous", device.classString)
        if err != nil {
            return nil, err
        }
    }

    return &device, nil
}

func (dl *CassDatalayer) LookupDeviceByStringId(id string) (*CassDevice, error) {
    deviceId, err := gocql.ParseUUID(id)
    if err != nil {
        return nil, err
    }
    return dl.LookupDevice(deviceId)
}

