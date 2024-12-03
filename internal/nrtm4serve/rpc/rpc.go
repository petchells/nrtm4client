/*
Package rpc is a JSONRPC 2.0 server implementation.

See the JSONRPC 2.0 specification at https://www.jsonrpc.org/specification

This is not a 100% compliant implementation. It differs in these ways:

	Requests
	- Parameter 'by-name' structures are not implemented
	- Notifications are not implemented
	- Arguments which are slices are not implemented

The handler will forward RPC commands to matching functions on the API object. A matching API
implementation has the following form:

	type MyAPI struct {
		rpc.API
	}

	func (api MyAPI) GetAuth(w http.ResponseWriter, r *http.Request, rpcreq rpc.JSONRPCRequest)
		(*sessions.Session, error) {
		...
	}

	func (api MyAPI) ExampleRPCFunction(
		w http.ResponseWriter,
		r *http.Request,
		rpcParam1: type,
		...
		rpcParamN: type,
	) (response[, error]) {
		...
	}

A Javascript number parameter will be cast to any Golang number type which matches the method
signature. Note that you will loose precision if the JS value has fractional digits and the
receiving Go function expects an integer type.

GetAuth is called before every function invocation. If no authentication is required to execute
a function, then GetAuth should return nil, nil.
*/
package rpc

import "gitlab.com/etchells/nrtm4client/internal/nrtm4/util"

var logger = util.Logger
