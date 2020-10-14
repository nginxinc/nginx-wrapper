# Events

## Event List

| Name            | Description                                                |
|-----------------|------------------------------------------------------------|
| `pre-start`     | Event emitted when before NGINX starts for the first time  |
| `start`         | Event emitted when NGINX starts for the first time         |
| `start-worker`  | Event emitted every time a NGINX worker process is started |
| `exit`          | Event emitted when the main NGINX process exits            |
| `exit-worker`   | Event emitted every time a NGINX worker process exits      |
| `pre-reload`    | Event emitted before NGINX reloads                         |
| `reload`        | Event emitted when NGINX reloads                           |