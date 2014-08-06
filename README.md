Canopy Cloud Service
------------------------------------------------------------------------------

Canopy Cloud Service is the server-side component of Canopy.  The main
executable is `canopy-cloud-service`, which is written in golang.  Some
of its responsibilities include:

 - Talking over websockets to each device.
 - Stores data in a Cassandra database.
 - Serving the Canopy REST API.


Building and Installing (Quick-and-easy method, Ubuntu 14.04)
------------------------------------------------------------------------------
Install Cassandra:

    sudo apt-get install apache

Build Canopy Cloud Service:

    make

Install Canopy Cloud Service:

    sudo make install

Start it running:

    sudo /etc/init.d/canopy-cloud-service start

Stop it:

    sudo /etc/init.d/canopy-cloud-service stop 

Notes for older systems (specifically: Ubuntu 12.04 LTS):
------------------------------------------------------------------------------

Installing cassandra 2.07 on Ubuntu 12.04LTS:

    $ sudo vim /etc/apt/sources.list.d/cassandra.sources.list
    deb http://debian.datastax.com/community stable main

    $ curl -L http://debian.datastax.com/debian/repo_key | sudo apt-key add -
    $ sudo apt-get update
    $ sudo apt-get install cassandra=2.0.7

For cassandra to run, it needs oracle java 1.7:

    $ sudo apt-get install python-software-properties
    $ sudo add-apt-repository ppa:webupd8team/java
    $ sudo apt-get update
    $ sudo apt-get install oracle-java7-installer

The gocql package requires golang 1.1 or later.  Ubuntu 12.04 installs 1.0 by
default.  To updgrade:

    $ sudo add-apt-repository ppa:duh/golang
    $ sudo apt-get update
    $ sudo apt-get install golang
