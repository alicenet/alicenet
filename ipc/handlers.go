package ipc

type errorObj struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type response struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *errorObj   `json:"error,omitempty"`
}

type request struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Params  interface{} `json:"params"`
	Method  string      `json:"method"`
}

var serverMethods = map[string]func(*Server, request) (interface{}, *errorObj){

	"subscribe": func(s *Server, req request) (interface{}, *errorObj) {
		s.pushId = req.Id
		return nil, nil
	},
	"healthz": func(s *Server, req request) (interface{}, *errorObj) {
		return struct{ Success bool }{Success: true}, nil
	},
}
