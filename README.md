# HTTPQ
Httpq is a simple message broker.
This project is a demo application for an interview with [Schneider Electric](https://www.se.com/ww/en/), see the [tasks](./task.md) for more details about the assignemnt.

## Usage
httpq takes the following environment variables:
```
HTTPQ_PORT			// Optional set the port on which to run the webserver, defaults to 23411
HTTPQ_TLS_KEY_PATH		// Optional set the path to the TLS key file, if not set the server will run without TLS
HTTPQ_TLS_CERT_PATH		// Optional set the path to the TLS certificate file, if not set the server will run without TLS
``

for example setting up the server with a self signed certificate on port 8080:
```sh
$ # generating the certificate & key
$ mkdir certs
$ openssl genrsa -out certs/server.key 2048
$ openssl ecparam -genkey -name secp384r1 -out certs/server.key
$ openssl req -new -x509 -sha256 -key certs/server.key -out certs/server.crt -days 3650
$
$ # setting the environment variables
$ export HTTPQ_PORT=8080
$ export HTTPQ_TLS_KEY_PATH=certs/server.key
$ export HTTPQ_TLS_CERT_PATH=certs/server.crt
$
$ # run the application
$ httpq
```

Or run it using our certificates with docker on port 80:
```sh
$ docker run -e HTTPQ_TLS_KEY_PATH=/certs/server.key -e HTTPQ_TLS_CERT_PATH=/certs/server.crt -v ./certs:/certs -p 80:23411 ghcr.io/jobstoit/schneider-httpq:latest
```

