# Use with CAUTION!
#
# This script is meant to run on a brand new Ubuntu 14.04 machine.  Running
# multiple times or on an already-setup machine may cause serious problems.

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

