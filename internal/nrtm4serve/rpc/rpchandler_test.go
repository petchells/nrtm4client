package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testAPI struct {
	API
}

type ComplexTextObject struct {
	Name string   `json:"Name"`
	List []string `json:"List"`
}

// GetAuth returns return auth (if it's required, otherwise nil), or an error (no session)
func (ta testAPI) GetAuth(w http.ResponseWriter, r *http.Request, rpcreq JSONRPCRequest) (WebSession, bool) {
	name := "test-user"
	return WebSession{Session: name}, true
}

func (ta testAPI) ShowSessionUser(s WebSession) string {
	if ws, ok := s.Session.(string); ok {
		return ws
	}
	return ""
}

func (ta testAPI) EchoNumber(i int64) int64 {
	return i
}

func (ta testAPI) Echo(name string) string {
	return name
}

func (ta testAPI) EchoI(n int) int {
	return n
}

func (ta testAPI) EchoI64(n int64) int64 {
	return n
}

func (ta testAPI) EchoOnlyError() error {
	return errors.New("EchoOnlyError")
}

func (ta testAPI) EchoNilWithError() (int64, error) {
	err := JSONRPCError{Code: -32000, Message: "Testing 1, 2, 4"}
	return 0, err
}

func (ta testAPI) ComplexObjectAsParamAndResultTest(cto ComplexTextObject) (ComplexTextObject, error) {
	return cto, nil
}

func (ta testAPI) MapAsParamAndResultTest(cto map[string]interface{}) (map[string]interface{}, error) {
	return cto, nil
}

func (ta testAPI) StringSliceResultTest(p []string) ([]string, error) {
	return p, nil
}

func (ta testAPI) IntSliceResultTest(p []int64) ([]int64, error) {
	return p, nil
}

func (ta testAPI) ComplexSliceResultTest(w http.ResponseWriter, r *http.Request, p []ComplexTextObject) ([]ComplexTextObject, error) {
	return p, nil
}

/*
	rpc call with an empty Array:

--> []
<-- {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}

rpc call with an invalid Batch (but not empty):

--> [1]
<-- [

	{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}

]

rpc call with invalid Batch:

--> [1,2,3]
<-- [

	{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
	{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
	{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}

]
*/
func TestRpcErrorHandling(t *testing.T) {
	var jsonStr string
	var expected string
	var doErrorTest = func(jsonStr string, expected string) {
		res := doRequest(t, jsonStr)
		got := res.Error.Error()
		if !strings.EqualFold(expected, got) {
			t.Errorf("RPC error handling failed. Expected '%v' got '%v'", expected, got)
		}
	}

	jsonStr = `{
		"jsonRpc": "2.0",
		"id": "1",
		"method": "Heartbeat"
	}`
	expected = "-32099: "
	doErrorTest(jsonStr, expected)

	jsonStr = `{
		"jsonrpc": "2.0",
		"id": "1",
		"method": "Echo",
		"params": [
			"wrong",
			"params"
		]
	}`
	expected = invalidParamsResponse.Error.Error()
	doErrorTest(jsonStr, expected)

	jsonStr = `[{
		"jsonrpc": "2.0",
		"id": "1"`
	expected = parseErrorResponse.Error.Error()
	doErrorTest(jsonStr, expected)

	jsonStr = `{
		"jsonrpc": "2.0",
		"id": "1"`
	expected = parseErrorResponse.Error.Error()
	doErrorTest(jsonStr, expected)

	jsonStr = `{
		"jsonrpc": "2.0",
		"method": "Echo",
		"params": [""]}`
	expected = emptyResponse.Error.Error()
	doErrorTest(jsonStr, expected)

	jsonStr = `{
		"jsonrpc": "2",
		"id": 77777,
		"method": "Echo",
		"params": [""]}`
	expected = emptyResponse.Error.Error()
	doErrorTest(jsonStr, expected)

	jsonStr = `{
		"jsonrpc": "2.0",
		"id": 77777,
		"method": "EchoOnlyError"}`
	expected = "-32099: EchoOnlyError"
	doErrorTest(jsonStr, expected)

	jsonStr = `{
		"jsonrpc": "2.0",
		"id": 77787,
		"method": "EchoNilWithError"}`
	expected = "-32000: Testing 1, 2, 4"
	doErrorTest(jsonStr, expected)
}

func TestOptionsRequest(t *testing.T) {
	req, err := http.NewRequest("OPTIONS", "/rpc", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := setupResponseRecorder(req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestArrayOfCommandsRequest(t *testing.T) {
	var jsonStr = `[{
		"jsonrpc": "2.0",
		"id": "1",
		"method": "Echo",
		"params": [
			"Hello World!"
		]
	},{
		"jsonrpc": "2.0",
		"id": "2",
		"method": "Heartbeat"
	}]`
	req, err := http.NewRequest("POST", "/rpc", bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rr := setupResponseRecorder(req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	res := []JSONRPCResponse{}
	json.Unmarshal([]byte(rr.Body.String()), &res)

	expectNumResponses := 2
	if len(res) != expectNumResponses {
		t.Fatalf("should have %d responses: got %d", expectNumResponses, len(res))
	}

	var expectString string
	var gotString string

	expectString = "Hello World!"
	gotString, ok := res[0].Result.(string)
	if !ok {
		t.Fatal("Echo should return a string")
	}
	if expectString != gotString {
		t.Fatalf("Echo should say '%s' but said '%v'", expectString, gotString)
	}
	if res[0].Error != nil {
		t.Fatal("Error should not be nil if Result was returned")
	}
	if res[1].Result != nil {
		t.Fatal("When called with no such method, result should be nil")
	}
	var errorCode int64
	errorCode = res[1].Error.Code
	if errorCode != emptyResponse.Error.Code {
		t.Fatalf("No such method error code expected %d but was %d", emptyResponse.Error.Code, errorCode)
	}
}

func TestSingleSimpleParams(t *testing.T) {
	{
		var expected = "Hello World!"
		var jsonStr = `{
		"jsonrpc": "2.0",
		"id": "1",
		"method": "Echo",
		"params": ["` + expected + `"]
	}`
		res := doRequest(t, jsonStr)
		if res.Result != expected {
			t.Errorf("Expected '%v' but got '%v'", expected, res.Result)
		}
	}
	{
		var expected = "test-user"
		var jsonStr = `{
		"jsonrpc": "2.0",
		"id": "1",
		"method": "ShowSessionUser",
		"params": []
	}`
		res := doRequest(t, jsonStr)
		if res.Result != expected {
			t.Errorf("Expected '%v' but got '%v'", expected, res.Result)
		}
	}
	{
		var expected int64 = 7
		var jsonStr = `{
		"jsonrpc": "2.0",
		"id": "1",
		"method": "EchoI",
		"params": [
			7
		]
	}`
		res := doRequest(t, jsonStr)
		if int64(res.Result.(float64)) != expected {
			t.Errorf("Expected '%d' but got '%d'", expected, res.Result)
		}
	}
}

func TestSingleMapParam(t *testing.T) {

	var expected string = "Middlestown"
	var jsonStr = `{
			"jsonrpc": "2.0",
			"id": "1",
			"method": "MapAsParamAndResultTest",
			"params": [
				{
					"Name": "` + expected + `",
					"List": ["White Swan", "Little Bull", "WMC"],
					"ZFlist": [ "contact" ]
				}
			]
		}`
	jpcres := doRequest(t, jsonStr)
	if jpcres.Error != nil {
		t.Errorf("Expected successful call to rpcHandler, got '%v'", jpcres.Error)
	}
	bs, err := json.Marshal(jpcres.Result)
	if err != nil {
		t.Errorf("Should be serializable")
	}
	var cc ComplexTextObject
	//res := jpcres.Result.(map[string]interface{})
	json.Unmarshal(bs, &cc)
	if cc.Name != expected {
		t.Errorf("Expected '%v' but got '%v'", expected, cc.Name)
	}
	if len(cc.List) != 3 {
		t.Errorf("Expected 3 items in list but got '%v'", len(cc.List))
	}
}

func TestSingleComplexParam(t *testing.T) {

	var expected string = "Headingley"
	var jsonStr = `{
			"jsonrpc": "2.0",
			"id": "1",
			"method": "ComplexObjectAsParamAndResultTest",
			"params": [
				{
					"Name": "` + expected + `",
					"List": ["Rugby", "Cricket"]
				}
			]
		}`
	jpcres := doRequest(t, jsonStr)
	if jpcres.Error != nil {
		t.Fatalf("Expected successful call to rpcHandler, got '%v'", jpcres.Error)
	}

	bs, err := json.Marshal(jpcres.Result)
	if err != nil {
		t.Errorf("Should be serializable")
	}
	var cc ComplexTextObject
	json.Unmarshal(bs, &cc)
	if cc.Name != expected {
		t.Errorf("Expected '%v' but got '%v'", expected, cc.Name)
	}
	if len(cc.List) != 2 {
		t.Errorf("Expected 2 items in list but got '%v'", len(cc.List))
	} else {
		if cc.List[0] != "Rugby" {
			t.Errorf("Expected first item to be Rugby but was '%v'", cc.List[0])
		}
		if cc.List[1] != "Cricket" {
			t.Errorf("Expected first item to be Cricket but was '%v'", cc.List[1])
		}
	}
}

//	func TestInt64ConversionBounds(t *testing.T) {
//		var jsonStr = `{
//		"jsonrpc": "2.0",
//		"id": "1",
//		"method": "EchoNumber",
//		"params": [%v]
//	}`
//
//		var expected int64 = 211585055384929410
//		jpcres := doRequest(t, fmt.Sprintf(jsonStr, expected))
//		if jpcres.Error != nil {
//			t.Fatalf("Expected successful call to rpcHandler, got '%v'", jpcres.Error)
//		}
//		res := jpcres.Result.(int64)
//		if res != expected {
//			t.Fatalf("Expected '%v' but got '%v'", expected, res)
//		}
//	}
func TestSingleBadCommand(t *testing.T) {
	var jsonStr = `{
		"jsonrpc": "2.0",
		"id": "1",
		"method": "Echo",
		"params": [
			7
		]
	}`
	res := doRequest(t, jsonStr)
	if res.Error == nil {
		t.Fatal("Expected an error but it was nil")
	}
	var expectedCode int64 = invalidParamsResponse.Error.Code
	if res.Error.Code != expectedCode {
		t.Fatalf("Expected '%v' but got '%v'", expectedCode, res.Error.Code)
	}
}

func TestParseErrorHandling(t *testing.T) {

	var doTestWith = func(jsonStr string, expectedErr error) {
		req, err := http.NewRequest("POST", "/rpc", bytes.NewBuffer([]byte(jsonStr)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		rr := setupResponseRecorder(req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		res := JSONRPCResponse{}
		if err = json.Unmarshal([]byte(rr.Body.String()), &res); err != nil {
			t.Errorf("Should not encounter unmarshal error")
		}
		if res.Result != nil {
			t.Errorf("Result should be nil")
		}
		if res.Error == nil {
			t.Errorf("Error should not be nil")
		}
		if res.Error.Error() != expectedErr.Error() {
			t.Errorf("Should get Parse Error but not %v", res.Error.Error())
		}
	}
	var jsonStr = `[{
		"jsonrpc": "2.0",
		"id": "1"`
	doTestWith(jsonStr, parseErrorResponse.Error)
	jsonStr = `{
		"jsonrpc": "2.0",
		"id": "1"`
	doTestWith(jsonStr, parseErrorResponse.Error)
	jsonStr = `{
		"jsonrpc": "2.0",
		"method": "Echo",
		"params": [""]}`
	doTestWith(jsonStr, emptyResponse.Error)
}
func TestSliceTypeHandling(t *testing.T) {

	var doTestWith = func(jsonStr string) JSONRPCResponse {
		req, err := http.NewRequest("POST", "/rpc", bytes.NewBuffer([]byte(jsonStr)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		rr := setupResponseRecorder(req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		res := JSONRPCResponse{}
		if err = json.Unmarshal([]byte(rr.Body.String()), &res); err != nil {
			t.Errorf("Should not encounter unmarshal error")
		}
		return res
	}
	var jsonStr string
	{
		jsonStr = `{
		"jsonrpc": "2.0",
		"id": "id",
		"method": "StringSliceResultTest",
		"params": [[]]}`
		o := doTestWith(jsonStr)
		if o.Error != nil {
			t.Errorf("StringSliceResultTest empty failed with error")
		}
	}
	{
		jsonStr = `{
		"jsonrpc": "2.0",
		"id": "id",
		"method": "StringSliceResultTest",
		"params": [["how", "now", "brown", "cow"]]}`
		o := doTestWith(jsonStr)
		if o.Error != nil {
			t.Errorf("StringSliceResultTest failed with error")
		}
		if fmt.Sprint(o.Result) != "[how now brown cow]" {
			t.Errorf("StringSliceResultTest failed with bad result %v", o.Result)
		}
	}
	// TODO: implement RPC so this and ComplexSliceResultTest work
	// {
	// 	jsonStr = `{
	// 	"jsonrpc": "2.0",
	// 	"id": "id",
	// 	"method": "IntSliceResultTest",
	// 	"params": [[3, 4, 5]]}`
	// 	o := doTestWith(jsonStr)
	// 	if o.Error != nil {
	// 		t.Errorf("IntSliceResultTest failed %v", o.Error)
	// 	}
	// 	if fmt.Sprint(o.Result) != "[3 4 5]" {
	// 		t.Errorf("IntSliceResultTest failed %v", o.Result)
	// 	}
	// }
}

var doRequest = func(t *testing.T, jsonStr string) JSONRPCResponse {
	jsonBytes := []byte(jsonStr)
	req, err := http.NewRequest("POST", "/rpc", bytes.NewBuffer(jsonBytes))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	rr := setupResponseRecorder(req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	res := JSONRPCResponse{}
	json.Unmarshal([]byte(rr.Body.String()), &res)
	return res
}

func setupResponseRecorder(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	rpcHandler := Handler{API: testAPI{}}
	server := http.HandlerFunc(rpcHandler.RPCServiceWrapper)
	server.ServeHTTP(rr, req)
	return rr
}
