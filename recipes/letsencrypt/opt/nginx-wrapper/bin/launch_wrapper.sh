#!/usr/bin/env bash

# Exit the script and an error is encountered
set -o errexit
# Exit the script when a pipe operation fails
set -o pipefail

if [ -z "${TLS_HOSTNAME}" ]; then
  >&2 echo "TLS_HOSTNAME must be set to the domain you wish to use with Let's Encrypt"
  exit 1
fi

if [ -z "${LETS_ENCRYPT_EMAIL}" ]; then
  >&2 echo "LETS_ENCRYPT_EMAIL must be set to the email that you wish to use with Let's Encrypt"
  exit 1
fi

if [ -z "${DNS_RESOLVER}" ]; then
  >&2 echo "DNS_RESOLVER was not set - using default of 1.1.1.1"
  export DNS_RESOLVER="1.1.1.1"
fi

# Path to the Let's Encrypt TLS certificates
export CERT_DIR="/etc/letsencrypt/live/${TLS_HOSTNAME}"
export PATH="${PATH}:/opt/nginx-wrapper/bin"

if [ "" = "${LETS_ENCRYPT_STAGING:-}" ] || [ "0" = "${LETS_ENCRYPT_STAGING}" ]; then
  CERTBOT_STAGING_FLAG=""
else
  CERTBOT_STAGING_FLAG="--staging"
fi

# Exit the script when there are undeclared variables
set -o nounset

if [ ! -f "${CERT_DIR}/fullchain.pem" ]; then
  echo "Generating certificates with Let's Encrypt"
  certbot certonly --standalone \
         -m "${LETS_ENCRYPT_EMAIL}" \
         ${CERTBOT_STAGING_FLAG} \
         --agree-tos --force-renewal --non-interactive \
         -d "${TLS_HOSTNAME}"
fi

# Assigns a unique machine ID that can be read by the wrapper
if [ ! -f /etc/machine-id ]; then
  uuid -F STR | tr -d '-' > /etc/machine-id
fi

# Start up the wrapper
exec /opt/nginx-wrapper/bin/nginx-wrapper \
  --config /opt/nginx-wrapper/nginx-wrapper.toml \
  run
