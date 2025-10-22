package main

import "github.com/mefranklin6/microservice-framework/framework"

///////////////////////////////////////////////////////////////////////////////
// Main functions //
///////////////////////////////////////////////////////////////////////////////

// Get functions //

func getOutlet(socketKey string, outlet string) (string, error) {
	//function := "getOutlet"
	framework.Log(socketKey, "HelloWorld!")
	return "", nil
}
