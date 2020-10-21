#!/usr/bin/env sh

# Issue a NGINX reload signal (SIGUSR2 indicates to the wrapper to ONLY reload
# NGINX and not other configuration) to the nginx-wrapper.
WRAPPER_PID="$(cat /opt/nginx-wrapper/run/nginx-wrapper.pid)"
kill -s SIGUSR2 "${WRAPPER_PID}"
