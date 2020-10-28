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

M="$(printf "\033[34;1m▶\033[0m")"

function check_path() {
  echo -n "${M} Checking to see if ${1} is installed... "
  if ! command -v "${1}" &> /dev/null; then
      if [ "${2:0:0}" == "" ]; then
        echo "${1} is not in PATH"
        exit 1
      else
        echo "no - checking for alternative"
        check_path "${2}"
      fi
  else
      echo "yes"
  fi
}

echo -n "${M} Checking to see if bash is version 4+... "
if [ ${BASH_VERSINFO} -lt 4 ]; then
  echo "bash is version ${BASH_VERSINFO}"
  exit 1
else
  echo "yes - [version ${BASH_VERSINFO}]"
fi

check_path uname

if [[ "$(uname -s)" == "Darwin" ]]; then
  echo;
  echo '  If you are running on MacOS with homebrew, be sure to have your path'
  echo '  setup up so that the gnubin directory takes precedence:'
  echo '  export PATH="/usr/local/opt/make/libexec/gnubin:$PATH"'
  echo;
fi

check_path awk
check_path cat
check_path column
check_path date
check_path egrep
check_path find
check_path gh
check_path git
check_path go
check_path grep
check_path gzip
check_path head
check_path make
check_path mkdir
check_path mv
check_path sed gsed
check_path sort
check_path sha256sum
check_path tail
check_path tr
check_path xargs

# Check to make sure we aren't running in docker
if [ ! -f /.dockerenv ]; then
  echo -n "${M} Checking to see if docker is installed... "
  if ! command -v "docker" &> /dev/null; then
      echo "docker is not in PATH (docker is optional)"
  else
      echo "yes"
  fi
fi

echo -n "${M} Checking to see if make is GNU Make... "
MAKE_VERSION="$(make --version)"
if [[ "${MAKE_VERSION}" =~ ^GNU\ Make.*$ ]]; then
  echo "yes"
else
  echo "make installed in PATH is not GNU make"
  exit 1
fi

echo -n "${M} Checking to see if make is GNU Make 4.1+... "
if [[ ${MAKE_VERSION:9:1} -ge 4 && ${MAKE_VERSION:11:1} -ge 1 ]]; then
  echo "yes - [${MAKE_VERSION:9:3}]"
else
  echo "make version is too old: ${MAKE_VERSION:9:3}"
  exit 1
fi

echo -n "${M} Checking to see if make is go is 1.15+... "
GO_VERSION="$(go version)"
if [[ ${GO_VERSION:13:1} -eq 1 && ${GO_VERSION:15:2} -ge 15 ]]; then
  echo "yes - [${GO_VERSION:13:4}]"
else
  echo "go version is too old: ${GO_VERSION:13:4}"
  exit 1
fi