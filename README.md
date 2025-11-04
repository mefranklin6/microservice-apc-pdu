# microservice-apc-pdu

[OpenAV](https://github.com/Dartmouth-OpenAV) compatible microservice for controlling APC Switched Power Distribution Units (PDU's)

Originally written by [Matthew Franklin](https://github.com/mefranklin6) (mefranklin6) at Chico State in October of 2025.

## Device Overview

[APC NetShelter Switched Rack PDU's](https://www.se.com/us/en/product-range/61799-apc-netshelter-switched-rack-pdus/#products) are rack mounted power distribution units that allow remote control of their AC electrical outlets.  They are useful for power cycling crashed devices, and can be used to reduce energy consumption and heat in racks by powering off devices that are not needed.

## Microservice Overview

This microservice should work with all current APC switched PDU's.

Developed with and tested with "NetShelter" models 7900B's and APDU9941's.

## Device Configuration

APC devices come with SSH and DHCP enabled out of the box.  For now, this microservice only supports Telnet.

1. **Use a workstation to connect to the device over SSH or serial port.**  If this is a new device, you'll be asked to change the password on first login.  Default user is `apc`, default password is `apc`.  

2. Follow the steps to change the password if prompted, then **run these commands:**

    - `console -t enable`
    - `reboot`
    - `YES`

## Microservice setup and use

By default, this microservice presumes the protocol is defined in the URL.  This is because these devices can support both Telnet and SSH. This setup allows the flexibility to use either protocol within the same container.

`http://<microserviceAddr>/telnet|<user>:<pw>@<deviceAddr>:<devicePort>`

If user and port are omitted, then default user `apc` and telnet port 23 will be used.

ex: `http://192.168.1.5/telnet|:secretpw@192.168.1.10/`

To enable 'legacy' Telnet-only or SSH-only mode: add `framework.UseTelnet = true` or ``framework.UseSSH = true`` to 'setFrameworkGlobals' in microservice.go.

*Important: This microservice only supports Telnet for now*

## Endpoints

### Get state

GET `/state/<outletNum>`

Returns "on", "off", or "unknown".  See container logs for any errors.

### Set state

PUT `/state/<outletNum>` with "on" or "off" or "reboot" in the body of the request.

`outletNum` can be a single outlet number ex:"1", a range of outlets (ex: "1-6"), or "all".

Returns "ok".  Check logs for errors.

### Reboot Outlet

PUT `/state/<outletNum>`

### Get all outlets

GET `/alloutlets`

Returns all outlets delimited by pipes `|`
Note: this is not an official OpenAV endpoint.

## Note to maintainers

This microservice requires built-in echo handling, at least for Telnet.  At the time of development, the devices tested did not seem to care that we sent "DON'T ECHO" at the IAC stage; Even though they reply "WON'T ECHO", they still send echo.  It was deemed more reliable to just ignore all IAC negotiation, process reads with IAC default values, and handle the echo.
