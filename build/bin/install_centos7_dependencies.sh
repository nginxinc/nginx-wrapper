#!/usr/bin/env bash

#
#  Copyright 2020 F5 Networks
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

# Exit the script and an error is encountered
set -o errexit
# Exit the script when a pipe operation fails
set -o pipefail
# Exit the script when there are undeclared variables
set -o nounset

curl --retry 6 -Ls -o /tmp/gh.tar.gz https://github.com/cli/cli/releases/download/v1.2.0/gh_1.2.0_linux_amd64.tar.gz
echo "3e4e3c49497b18f5da46989b188091a25171a8090bfcf42bc75befa115f49322  /tmp/gh.tar.gz" | sha256sum -c
tar -xz --strip-components=1 -C /usr/local/ -f /tmp/gh.tar.gz
rm /usr/local/LICENSE /var/gh.tar.gz

yum --assumeyes --quiet install gcc git