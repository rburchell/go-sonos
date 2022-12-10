# Credits

Forked from https://github.com/szatmary/sonos so all credit goes to them. There are few changes have been made mostly around the structure of the generated code.

# Services

The service implimentations are automatically generated from the service definition XML files obtained from the Sonos devices via `makeservice.go.`

`cmd/makeservices/downloadallservices.sh` fetches them from the device and `cmd/makeservices/makeallservices.sh` generates the code.

# More

Please see https://svrooij.io/sonos-api-docs/sonos-communication.html and https://svrooij.io/sonos-api-docs/services/

