package ipc

func newInvalidRequestResponse(id interface{}) response {
	return response{"2.0", id, nil, &errorObj{Code: -32600, Message: "Invalid Request"}}
}

func newMethodNotFoundResponse(id interface{}) response {
	return response{"2.0", id, nil, &errorObj{Code: -32601, Message: "Method Not Found"}}
}

func newParseErrorResponse(details string) response {
	return response{"2.0", nil, nil, &errorObj{Code: -32700, Message: "Parse error", Data: details}}
}
