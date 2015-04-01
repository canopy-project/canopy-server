Upgrade Process
-------------------------------------------------------------------------------

0.9.1 to 15.04.03
-------------------------------------------------------------------------------
*** Backup Database ***

    nodetool -h localhost -p 7199 snapshot canopy

*** Upgrade source ***

    git fetch
    git checkout v15.04.03
    make
    sudo make install

*** Migrate the database ***

    canodevtool migrate-db "0.9.1" "15.04.03"

*** Stop the old version and start the new version ***

    sudo /etc/init.d/canopy-cloud-service stop
    sudo /etc/init.d/canopy-server start

*** Cleanup obsolete files ***

    sudo rm /etc/init.d/canopy-cloud-service

0.9.0 to 0.9.1
-------------------------------------------------------------------------------

*** Backup Database ***

    nodetool -h localhost -p 7199 snapshot canopy

*** Upgrade source ***

    git fetch
    git checkout v0.9.1
    make

*** Upgrade config file ***

Add the following fields to `/etc/canopy/server.conf`:

    "enable-http": false,
    "enable-https": true,
    "https-cert-file": "/etc/canopy/cert.pem",
    "https-priv-key-file": "/etc/canopy/key.pem",
    "password-hash-cost" : 10,
    "password-secret-salt" : "HCZIloQgIzAq5USZ17dvg",

*** Install and run the new version ***

    sudo make update

*** Migrate the database ***

    canodevtool migrate-db "0.9.0" "0.9.1"

*** Reset password ***

    Version 0.9.1 requires everyone to reset their passwords (we switched from
    hardcoded password salt to configurable password salt).  Go to
    https://sandbox.canopy.link/mgr/ and click "Forgot password?" and follow
    instructions.
