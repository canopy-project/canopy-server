# Use with CAUTION!
#
# This script is meant to run on a brand new Ubuntu 14.04 machine.  Running
# multiple times or on an already-setup machine may cause serious problems.
#
# Configuring this script prior to running:
# 
# Required:
#   export INSTALL_CANOPY_HOSTNAME="dev05.canopy.link"
#
# Optional:
#   export INSTALL_CANOPY_SERVER_BRANCH="v15.05.01"
#   export INSTALL_CANOPY_JS_CLIENT_BRANCH="v15.05.01"
#   export INSTALL_CANOPY_DEVICE_MGR_BRANCH="v15.05.01"
#   export INSTALL_CANOPY_EMAIL_SERVICE="sendgrid"
#   export INSTALL_CANOPY_SENDGRID_USERNAME="myusername"
#   export INSTALL_CANOPY_SENDGRID_SECRET_KEY="mysecret"

sudo bash -c 'echo 127.0.0.1 $HOSTNAME >> /etc/hosts'

# Upgrade
DEBIAN_FRONTEND=noninteractive sudo -E apt-get update -y
DEBIAN_FRONTEND=noninteractive sudo -E apt-get upgrade -y

# Install Cassandra
sudo bash -c 'echo deb http://debian.datastax.com/community stable main > /etc/apt/sources.list.d/cassandra.sources.list'
curl -L http://debian.datastax.com/debian/repo_key | sudo apt-key add -
DEBIAN_FRONTEND=noninteractive sudo -E apt-get update -y
DEBIAN_FRONTEND=noninteractive sudo -E apt-get install -y cassandra=2.0.7

# Install Oracle Java 1.7
DEBIAN_FRONTEND=noninteractive sudo -E apt-get install -y python-software-properties
echo debconf shared/accepted-oracle-license-v1-1 select true | sudo debconf-set-selections
echo debconf shared/accepted-oracle-license-v1-1 seen true | sudo debconf-set-selections
DEBIAN_FRONTEND=noninteractive sudo -E add-apt-repository -y ppa:webupd8team/java
DEBIAN_FRONTEND=noninteractive sudo -E apt-get update -y
DEBIAN_FRONTEND=noninteractive sudo -E apt-get install -y oracle-java7-installer


# Install GO
DEBIAN_FRONTED=noninteractive sudo -E apt-get install -y git make mercurial
wget https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
echo PATH=/usr/local/go/bin:\$PATH >> ~/.profile
sudo bash -c 'echo PATH=/usr/local/go/bin:\$PATH >> ~/.profile'
source ~/.profile
rm go1.4.2.linux-amd64.tar.gz

# Install Canopy
git clone https://github.com/canopy-project/canopy-server
cd canopy-server
if [[ $INSTALL_CANOPY_SERVER_BRANCH ]]; then
    git checkout $INSTALL_CANOPY_SERVER_BRANCH
fi 
make
sudo make install
cd ..

git clone https://github.com/canopy-project/canopy-js-client
if [[ $INSTALL_CANOPY_JS_CLIENT_BRANCH ]]; then
    cd canopy-js-client
    git checkout $INSTALL_CANOPY_JS_CLIENT_BRANCH
    cd ..
fi 

git clone https://github.com/canopy-project/canopy-device-mgr
if [[ $INSTALL_CANOPY_DEVICE_MGR_BRANCH ]]; then
    cd canopy-device-mgr
    git checkout $INSTALL_CANOPY_DEVICE_MGR_BRANCH
    cd ..
fi 

# Start Cassandra.  Need to wait for it to be ready.
sudo cassandra
sleep 4
echo "DONE"

# Configure Canopy
export CONF_CANOPY_JS_CLIENT_PATH="/home/ubuntu/canopy-js-client"
export CONF_CANOPY_WEB_MANAGER_PATH="/home/ubuntu/canopy-device-mgr"
export CONF_CANOPY_SECRET_SALT=`< /dev/urandom tr -dc _A-Z-a-z-0-9 | head -c20`
export CONF_CANOPY_EMAIL_SERVICE="none"
export CONF_CANOPY_CONF_HOSTNAME="none"
if [[ $INSTALL_CANOPY_EMAIL_SERVICE ]]; then
    export CONF_CANOPY_EMAIL_SERVICE=$INSTALL_CANOPY_EMAIL_SERVICE
fi 
if [[ $INSTALL_CANOPY_SENDGRID_USERNAME ]]; then
    export CONF_CANOPY_SENDGRID_USERNAME=$INSTALL_CANOPY_SENDGRID_USERNAME
fi 
if [[ $INSTALL_CANOPY_SENDGRID_SECRET_KEY ]]; then
    export CONF_CANOPY_SENDGRID_SECRET_KEY=$INSTALL_CANOPY_SENDGRID_SECRET_KEY
fi 
eval "echo `cat canopy-server/scripts/server.conf.template`" > server.conf
# Create database
canopy-ops create-db

# Start Canopy
sudo /etc/init.d/canopy-server start
sleep 3
tail -n 80 /var/log/canopy/server.log
