# Onepaq
### A cross platform 1password read-only client

## Security

This was created out of convenience. Use at your own risk ;) (where's the linux
love 1password). It effectively exposes your 1password secrets through a HTTP
server (when the vault is unlocked). Mutual TLS is supported to restrict who can
communiate to the server.

## Installing

- Manual:
```sh
go get github.com/hspak/onepaq
cd $GOPATH/src/github.com/hspak/onepaq
go install
```
- Arch Linux: https://aur.archlinux.org/packages/onepaq/

**TLS Setup**
Example config (require `openssl`):
```sh
$ openssl genrsa -out ca.key 4096
$ openssl req -new -key ca.key -out ca.csr   # Details are up to you
$ openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 -out ca.pem
$ openssl req -new -key ca.key -out cert.csr # Make sure the common name lines up with the server name
                                             # We should also be creating a seperate key for the certificate, but I don't think that buys any more security here
$ openssl x509 -req -in cert.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out cert.pem -days 365 -sha256
```
By the end, you should have a CA file and a cert/key pair you can specify for onepaq.
We're cheating by using the same cert/key pair for both the server and client (by default).

## Usage
```sh
$ onepaq -h
Usage of server:
  -config-path string
    	path to the config file (default "/etc/onepaq.d/onepaq.conf")

Usage of client:
  -act string
    	action to perform
  -addr string
    	server to query (default "localhost:8080")
  -config-path string
    	path to the config file (default "/etc/onepaq.d/onepaq.conf")
  -item string
    	item to take action on
  -pass string
    	password to unlock
```

**Client**
```sh
To unlock vault
$ onepaq client -act unlock

To read passwords
$ onepaq client -act read <entry>
```
