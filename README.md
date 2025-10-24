# microservice-apc-pdu

[OpenAV](https://github.com/Dartmouth-OpenAV) compatible microservice for controlling APC Switched Power Distribution Units (PDU's)

## Overview

This microservice should work with all current APC switched PDU's.

Developed with and tested against "NetShelter" models 7900B's and APDU9941's.

## Device Configuration

APC devices come with SSH and DHCP enabled out of the box.  For now, this microservice only supports Telnet.

*Login over SSH to the device then:*

If this is a new device, you'll be asked to change the password on first login.  Default user is `apc`, default password is `apc`.  

Follow the steps to change the password if prompted, then run these commands:

- `console -t enable`
- `reboot`
- `YES`

### Microservice setup and use

By default, this microservice presumes the protocol is defined in the URL.  This is because these devices can support both Telnet and SSH. This setup allows the flexibility to use either protocol within the same container.

ex: `http://<microserviceAddr>/telnet|<user>:<pw>@<deviceAddr>:<devicePort>`

To enable 'legacy' telnet-only mode: add `framework.UseTelnet = true` to to 'setFrameworkGlobals' in microservice.go.
