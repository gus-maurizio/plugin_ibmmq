#!/bin/bash
# -*- mode: sh -*-
# © Copyright IBM Corporation 2015, 2018
#
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Fail on any non-zero return code
set -ex

test -f /usr/bin/yum && RHEL=true || RHEL=false
test -f /usr/bin/apt-get && UBUNTU=true || UBUNTU=false

# If MQ_PACKAGES isn't specifically set, then choose a valid set of defaults
if [ -z "$MQ_PACKAGES" ]; then
  $UBUNTU && MQ_PACKAGES="ibmmq-sdk ibmmq-client"
  $RHEL && MQ_PACKAGES="MQSeriesSDK-*.rpm MQSeriesClient-*.rpm"
fi

if ($UBUNTU); then
  export DEBIAN_FRONTEND=noninteractive
  # Use a reduced set of apt repositories.
  # This ensures no unsupported code gets installed, and makes the build faster
  source /etc/os-release
  # Figure out the correct apt URL based on the CPU architecture
  CPU_ARCH=$(uname -p)
  if [ ${CPU_ARCH} == "x86_64" ]; then
     APT_URL="http://archive.ubuntu.com/ubuntu/"
  else
     APT_URL="http://ports.ubuntu.com/ubuntu-ports/"
  fi
fi

# Install additional packages required by MQ, this install process and the runtime scripts
$RHEL && yum -y install \
  bash \
  bc \
  ca-certificates \
  coreutils \
  curl \
  file \
  findutils \
  gawk \
  glibc-common \
  grep \
  passwd \
  procps-ng \
  sed \
  tar \
  util-linux

# Download and extract the MQ installation files
DIR_EXTRACT=/tmp/mq
mkdir -p ${DIR_EXTRACT}
cd ${DIR_EXTRACT}
curl -LO $MQ_URL
tar -zxvf ./*.tar.gz

# Remove packages only needed by this script
$UBUNTU && apt-get purge -y \
  ca-certificates \
  curl

# Note: ca-certificates and curl are installed by default in RHEL

# Remove any orphaned packages
$UBUNTU && apt-get autoremove -y

# Find directory containing .deb files
$UBUNTU && DIR_DEB=$(find ${DIR_EXTRACT} -name "*.deb" -printf "%h\n" | sort -u | head -1)
$RHEL && DIR_RPM=$(find ${DIR_EXTRACT} -name "*.rpm" -printf "%h\n" | sort -u | head -1)
# Find location of mqlicense.sh
MQLICENSE=$(find ${DIR_EXTRACT} -name "mqlicense.sh")

# Accept the MQ license
${MQLICENSE} -text_only -accept

$UBUNTU && echo "deb [trusted=yes] file:${DIR_DEB} ./" > /etc/apt/sources.list.d/IBM_MQ.list

# Install MQ using the DEB packages
$UBUNTU && apt-get update
$UBUNTU && apt-get install -y $MQ_PACKAGES

$RHEL && cd $DIR_RPM && rpm -ivh $MQ_PACKAGES


# Remove tar.gz files unpacked by RPM postinst scripts
find /opt/mqm -name '*.tar.gz' -delete

# Clean up all the downloaded files
$UBUNTU && rm -f /etc/apt/sources.list.d/IBM_MQ.list
rm -rf ${DIR_EXTRACT}

# Clean up cached files
$UBUNTU && rm -rf /var/lib/apt/lists/*

