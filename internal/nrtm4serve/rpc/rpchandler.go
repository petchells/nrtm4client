package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
)

var (
	// ErrResponseUnauthorized returned when user has no session or is not authorized to execute method
	ErrResponseUnauthorized = JSONRPCError{Code: -32000, Message: "Unauthorized"}
	// ErrNoPrivilege user does not have the privilege to view resource
	ErrNoPrivilege = JSONRPCError{Code: -32098, Message: "No privilege"}

	parseErrorResponse     = JSONRPCResponse{JSONRPC: "2.0", Error: &JSONRPCError{Code: -32700, Message: "Parse error"}}
	invalidRequestResponse = JSONRPCResponse{JSONRPC: "2.0", Error: &JSONRPCError{Code: -32600, Message: "Invalid Request"}}
	methodNotFoundResponse = JSONRPCResponse{JSONRPC: "2.0", Error: &JSONRPCError{Code: -32601, Message: "Method not found"}}
	invalidParamsResponse  = JSONRPCResponse{JSONRPC: "2.0", Error: &JSONRPCError{Code: -32602, Message: "Invalid params"}}
	unauthorizedResponse   = JSONRPCResponse{JSONRPC: "2.0", Error: &ErrResponseUnauthorized}
	// User-defined codes from -32000 to -32099
	emptyResponse = JSONRPCResponse{JSONRPC: "2.0", Error: &JSONRPCError{Code: -32099, Message: ""}}
)

// WebSession is the user's session
type WebSession struct {
	Session any
}

// API Implementation of API for customer web client
type API interface {
	GetAuth(w http.ResponseWriter, r *http.Request, req JSONRPCRequest) (WebSession, bool)
}

// JSONRPCRequest A request
type JSONRPCRequest struct {
	JSONRPC string  `json:"jsonrpc"`
	ID      any     `json:"id"`
	Method  *string `json:"method"`
	Params  []any   `json:"params"`
}

// JSONRPCError An error
type JSONRPCError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func (e JSONRPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// JSONRPCResponse A response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Error   *JSONRPCError `json:"error,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
}

// Handler which implements the JSONRPC 2.0 specification at https://www.jsonrpc.org/specification
type Handler struct {
	API API
}

// RPCServiceWrapper Convenience function
func (handler Handler) RPCServiceWrapper(w http.ResponseWriter, r *http.Request) {
	if len(os.Getenv("APP_MODE")) > 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		w.WriteHeader(http.StatusOK)
		return
	}
	handler.ProcessRPC(w, r)
}

// ProcessRPC Calls a method on your RPCAPI implementation and writes the response
func (handler Handler) ProcessRPC(w http.ResponseWriter, r *http.Request) {
	bodyBuffer := new(bytes.Buffer)
	bodyBuffer.ReadFrom(r.Body)
	cleanjson := strings.TrimSpace(bodyBuffer.String())
	var response JSONRPCResponse
	if len(cleanjson) < 37 {
		response = parseErrorResponse
	} else if cleanjson[0] == '[' {
		responses := []JSONRPCResponse{}
		rpcreqs := []JSONRPCRequest{}
		err := json.Unmarshal([]byte(cleanjson), &rpcreqs)
		if err == nil {
			var session WebSession
			for _, rpcreq := range rpcreqs {
				var ok bool
				session, ok = handler.API.GetAuth(w, r, rpcreq)
				if !ok {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			for _, rpcreq := range rpcreqs {
				response := handler.execRPCRequest(w, r, session, rpcreq)
				responses = append(responses, response)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(responses)
			return
		}
		response = parseErrorResponse
	} else {
		rpcreq := JSONRPCRequest{}
		err := json.Unmarshal([]byte(cleanjson), &rpcreq)
		if err == nil {
			sess, ok := handler.API.GetAuth(w, r, rpcreq)
			if ok {
				response = handler.execRPCRequest(w, r, sess, rpcreq)
			} else {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		} else {
			response = parseErrorResponse
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (handler Handler) execRPCRequest(w http.ResponseWriter, r *http.Request, session WebSession, req JSONRPCRequest) JSONRPCResponse {
	if req.JSONRPC != "2.0" || req.Method == nil || len(*req.Method) == 0 || req.ID == nil {
		return emptyResponse
	}
	targetMethod := reflect.ValueOf(handler.API).MethodByName(*req.Method)
	if reflect.Invalid == targetMethod.Kind() {
		return emptyResponse
	}
	in := make([]reflect.Value, targetMethod.Type().NumIn())
	if targetMethod.Type().NumIn() < len(req.Params) {
		return invalidParamsResponse
	}

	filled := 0
	for paramIdx := 0; paramIdx < targetMethod.Type().NumIn(); paramIdx++ {
		if reflect.TypeOf(w).AssignableTo(targetMethod.Type().In(paramIdx)) {
			in[paramIdx] = reflect.ValueOf(w)
			filled++
			continue
		}
		if reflect.TypeOf(r).AssignableTo(targetMethod.Type().In(paramIdx)) {
			in[paramIdx] = reflect.ValueOf(r)
			filled++
			continue
		}
		if reflect.TypeOf(session).AssignableTo(targetMethod.Type().In(paramIdx)) {
			in[paramIdx] = reflect.ValueOf(session)
			filled++
			continue
		}
		if len(req.Params) != targetMethod.Type().NumIn()-filled {
			logger.Debug("RPCHandler error: Method found, but number of args is incorrect", "method", req.Method)
			rpcErr := invalidParamsResponse
			rpcErr.ID = req.ID
			return rpcErr
		}
		paramType := reflect.TypeOf(req.Params[paramIdx-filled])
		if paramType == nil {
			rpcErr := invalidParamsResponse
			rpcErr.ID = req.ID
			return rpcErr
		}
		if paramType == targetMethod.Type().In(paramIdx) {
			in[paramIdx] = reflect.ValueOf(req.Params[paramIdx-filled])
		} else if paramType.Kind() == reflect.Float64 {
			flt := req.Params[paramIdx-filled].(float64)
			switch targetMethod.Type().In(paramIdx).Kind() {
			case reflect.Int:
				in[paramIdx] = reflect.ValueOf(int(flt))
			case reflect.Int8:
				in[paramIdx] = reflect.ValueOf(int8(flt))
			case reflect.Int16:
				in[paramIdx] = reflect.ValueOf(int16(flt))
			case reflect.Int32:
				in[paramIdx] = reflect.ValueOf(int32(flt))
			// No werky... big numbers fail
			case reflect.Int64:
				in[paramIdx] = reflect.ValueOf(int64(flt))
			default:
				rpcErr := invalidParamsResponse
				rpcErr.ID = req.ID
				return rpcErr
			}
		} else if paramType.Kind() == reflect.Map {
			bs, err := json.Marshal(req.Params[paramIdx-filled])
			if err != nil {
				rpcErr := invalidParamsResponse
				rpcErr.ID = req.ID
				return rpcErr
			}
			val := reflect.New(targetMethod.Type().In(paramIdx)).Elem()
			err = json.Unmarshal(bs, val.Addr().Interface())
			if err != nil {
				logger.Debug("rpchandler Could not unmarshall value", "error", err)
				rpcErr := invalidParamsResponse
				rpcErr.ID = req.ID
				return rpcErr
			}
			in[paramIdx] = val
		} else if paramType.Kind() == reflect.Slice {
			bs, err := json.Marshal(req.Params[paramIdx-filled])
			if err != nil {
				rpcErr := invalidParamsResponse
				rpcErr.ID = req.ID
				return rpcErr
			}
			val := []string{}
			err = json.Unmarshal(bs, &val)
			if err != nil {
				rpcErr := invalidParamsResponse
				rpcErr.ID = req.ID
				return rpcErr
			}
			in[paramIdx] = reflect.ValueOf(val)
		} else {
			logger.Info("Unknown parameter type in rpchandler.go")
			rpcErr := invalidParamsResponse
			rpcErr.ID = req.ID
			return rpcErr
		}
	}
	res := targetMethod.Call(in)

	var rpcResponse JSONRPCResponse
	rpcResponse.JSONRPC = "2.0"
	rpcResponse.ID = req.ID

	switch targetMethod.Type().NumOut() {
	case 0:
	case 1:
		rs1, err := collectReturns(res[0].Interface())
		if err != nil {
			rpcResponse.Error = err
		} else {
			rpcResponse.Result = rs1
		}
	case 2:
		rs1 := res[0].Interface()
		_, err := collectReturns(res[1].Interface())
		if err != nil {
			rpcResponse.Error = err
		} else {
			rpcResponse.Result = rs1
		}
	}
	return rpcResponse
}

func collectReturns(apiResp interface{}) (interface{}, *JSONRPCError) {
	switch resp := apiResp.(type) {
	case JSONRPCError:
		return nil, &resp
	case error:
		var rpcErr JSONRPCError
		rpcErr.Code = emptyResponse.Error.Code
		rpcErr.Message = resp.Error()
		return nil, &rpcErr
	default:
		if apiResp == nil {
			return "OK", nil
		}
		return apiResp, nil
	}
}
