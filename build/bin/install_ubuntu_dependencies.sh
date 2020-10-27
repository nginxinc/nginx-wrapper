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

apt-get update -qq
apt-get install -y -qq build-essential software-properties-common git curl gnupg2 bsdmainutils coreutils

apt-key adv --keyserver keyserver.ubuntu.com --recv-key C99B11DEB97541F0
apt-add-repository https://cli.github.com/packages

apt-get update -qq
apt-get install -y -qq  gh