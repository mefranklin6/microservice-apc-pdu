package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mefranklin6/microservice-framework/framework"
)

///////////////////////////////////////////////////////////////////////////////
// Main Functions //
///////////////////////////////////////////////////////////////////////////////

// Get Functions //

// Gets the state of an outlet
func getState(socketKey string, outlet string) (string, error) {
	function := "getState"

	cmdStr := "olStatus " + outlet + "\r\n"
	response, err := sendCommand(socketKey, cmdStr)
	if err != nil {
		errMsg := function + " - error getting outlet status: " + err.Error()
		framework.AddToErrors(socketKey, errMsg)
		return "", err
	}

	if strings.Contains(response, "On") {
		return `"on"`, nil
	} else if strings.Contains(response, "Off") {
		return `"off"`, nil
	} else {
		errMsg := function + " - Could not determine state from data: " + response
		framework.AddToErrors(socketKey, errMsg)
		return "", errors.New(errMsg)
	}
}

// Gets the status of all outlets.  Not an official endpoint.
func getAllOutlets(socketKey string) (string, error) {
	function := "getAllOutlets"

	cmdStr := "olStatus all\r\n"
	response, err := sendCommand(socketKey, cmdStr)
	if err != nil {
		errMsg := function + " - error getting all outlets status: " + err.Error()
		framework.AddToErrors(socketKey, errMsg)
		return "", err
	}
	return `"` + response + `"`, nil
}

// Set Functions //

// Sets the state of one or more outlets
// num can be a single outlet number, a range of outlets (ex: "1-6"), or "all"
// state can be "on", "off", or "reboot"
func setState(socketKey string, num string, state string) (string, error) {
	function := "setState"

	framework.Log(function + " Setting outlet(s): " + num + " to: " + state)
	cmd := "ol" + state + " " + num + "\r\n"

	resp, err := sendCommand(socketKey, cmd)
	if err != nil {
		errMsg := function + " - error setting outlet state: " + err.Error()
		framework.AddToErrors(socketKey, errMsg)
		return "", err
	}
	return resp, nil
}

///////////////////////////////////////////////////////////////////////////////
// Helper Functions //
///////////////////////////////////////////////////////////////////////////////

func telnetLoginNegotiation(socketKey string) bool {
	function := "telnetLoginNegotiation"
	framework.Log("Starting Telnet login negotiation for: " + socketKey)

	username := "apc" // default if not specified
	password := ""

	if strings.Count(socketKey, "@") == 1 {
		sanitizedKey := framework.StripProtocolPrefix(socketKey)
		credentials := strings.Split(sanitizedKey, "@")[0]
		if strings.Count(credentials, ":") == 1 {
			username = strings.Split(credentials, ":")[0]
			password = strings.Split(credentials, ":")[1]
		}
	}

	if password == "" {
		noPwMsg := function + " - Password is required"
		framework.AddToErrors(socketKey, noPwMsg)
		return false
	}

	userSent := false
	passSent := false

	const maxRounds = 50 // Usually ~30 rounds due to a large welcome banner
	for round := 0; round < maxRounds; round++ {
		raw := framework.ReadLineFromSocket(socketKey)
		if raw == "" {
			continue
		}

		// Clean up the text for matching
		plain := strings.TrimSpace(raw)
		lower := strings.ToLower(plain)

		// framework.Log(fmt.Sprintf("Round %d - Received: %q", round, plain))

		// Skip echo of username or password we sent

		// Check for username prompt
		if strings.Contains(lower, "user name") && !userSent {
			framework.Log("Sending username: " + username)
			framework.WriteLineToSocket(socketKey, username+"\r\n")
			userSent = true
			continue
		}

		// Check for password prompt
		if strings.Contains(lower, "password") && !passSent {
			framework.Log("Sending password")
			framework.WriteLineToSocket(socketKey, password+"\r\n")
			passSent = true
			continue
		}

		// Check for command prompt (apc>)
		if strings.Contains(plain, "apc>") || strings.HasSuffix(plain, ">") {
			framework.Log("APC PDU login successful - prompt detected: " + plain)
			return true
		}
	}

	errMsg := function + " - Stopped negotiation after " + fmt.Sprintf("%d", maxRounds) + " rounds; No prompt detected"
	framework.AddToErrors(socketKey, errMsg)
	return false
}

func ensureConnected(socketKey string) bool {
	function := "ensureConnected"

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		framework.Log(function + " - No existing connection found. Creating new connection.")
		negotiation := telnetLoginNegotiation(socketKey)
		if !negotiation {
			framework.Log(function + " - Telnet login negotiation failed.")
			return false
		}
	}
	return true
}

func sendCommand(socketKey string, command string) (string, error) {
	function := "sendCommand"
	if !ensureConnected(socketKey) {
		errMsg := function + " - Unable to connect to device: " + socketKey
		framework.AddToErrors(socketKey, errMsg)
		err := errors.New(errMsg)
		return "", err
	}
	framework.Log("Sending command to device: " + command)
	framework.WriteLineToSocket(socketKey, command)

	resultCache := []string{}
	seenPrompt := false

	// Make a read loop. APC's are really chatty and can have multi-line responses
	const maxReads = 20
	for i := 0; i < maxReads; i++ {
		// Remove any null bytes, carriage returns, or line feeds (there's a lot)
		line := framework.ReadLineFromSocket(socketKey)
		line = strings.Map(func(r rune) rune {
			if r == 0 || r == '\n' || r == '\r' {
				return -1
			}
			return r
		}, line)

		switch {
		case line == "":
			continue
		case strings.Contains(line+"\r\n", command): // command echo
			seenPrompt = true
			continue
		case strings.Contains(line, "E000"): // Success code.  Should be the last line.
			switch len(resultCache) {
			case 0:
				return "ok", nil // Nothing in the cache if it was a successfull SET command
			case 1:
				return resultCache[0], nil
			default:
				return strings.Join(resultCache, "|"), nil
			}
		case strings.Contains(line, "E"): // device error code
			return "", errors.New("Device returned error code: " + line)
		default:
			if !seenPrompt { // ignore chatter before the echo
				continue
			}
			resultCache = append(resultCache, line)
			// framework.Log("Appending: " + line + " to result cache")
		}

	}
	return "", errors.New(function + " - Unable to parse result") // should not happen
}
