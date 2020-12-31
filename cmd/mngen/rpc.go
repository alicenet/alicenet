package main

var rpcDef = `{{ $Services := .Services }}// Code generated. DO NOT EDIT.

package {{.Package}}

import (
	"context"
	"errors"
	"sync"
)
{{range $service := $Services}}{{range $rpc := $service.RPC}}
// {{$rpc.Service}}{{$rpc.Name}}Handler is an interface class that only contains
// the method Handle{{$rpc.Service}}{{$rpc.Name}}
// The class that implements this method MUST handle the RPC call for
// the method {{$rpc.Name}} of the RPC service {{$rpc.Service}}
type {{$rpc.Service}}{{$rpc.Name}}Handler interface {
	Handle{{$rpc.Service}}{{$rpc.Name}}(context.Context, *{{$rpc.RequestType}}) (*{{$rpc.ReturnsType}}, error)
}
{{end}}{{end}}
// inboundRPCDispatch allows handlers to be registered for all RPC methods
// using the Register<Service><Name> methods.
// After registration, the inboundRPCDispatch struct will dispatch calls
// to an rpc method via the methods named as <Service><Name>(...)
type inboundRPCDispatch struct {
	sync.Mutex{{range $service := $Services}}{{range $rpc := $service.RPC}}
  //	handler{{$rpc.Service}}{{$rpc.Name}} is the registered handler for the
	//  {{$rpc.Name}} RPC method of service {{$rpc.Service}}
	handler{{$rpc.Service}}{{$rpc.Name}} {{$rpc.Service}}{{$rpc.Name}}Handler
	// waitChan{{$rpc.Service}}{{$rpc.Name}} will cause a caller of the RPC
	// method {{$rpc.Name}} on service {{$rpc.Service}} to block until the
	// method has been registered.
	waitChan{{$rpc.Service}}{{$rpc.Name}} chan struct{}{{end}}{{end}}
}
{{range $service := $Services}}{{range $rpc := $service.RPC}}
// Register{{$rpc.Service}}{{$rpc.Name}} will register the object 't' as the service
// handler for the RPC method {{$rpc.Name}} from service {{$rpc.Service}}
func (d *inboundRPCDispatch) Register{{$rpc.Service}}{{$rpc.Name}}(t {{$rpc.Service}}{{$rpc.Name}}Handler) {
	d.Lock()
	defer d.Unlock()
	// double registration is not allowed
	if d.handler{{$rpc.Service}}{{$rpc.Name}} != nil {
		panic("double registration of {{$rpc.Service}}{{$rpc.Name}}")
	}
	// register the service handler
	d.handler{{$rpc.Service}}{{$rpc.Name}} = t
	// close the wait channel to signal that the method is ready to use
	close(d.waitChan{{$rpc.Service}}{{$rpc.Name}})
}

// {{$rpc.Service}}{{$rpc.Name}} will invoke the handler for the RPC method
// {{$rpc.Name}} from service {{$rpc.Service}}
func (d *inboundRPCDispatch) {{$rpc.Service}}{{$rpc.Name}}(ctx context.Context, r *{{$rpc.RequestType}}) (*{{$rpc.ReturnsType}}, error) {
	// wait for registration to complete or context to be canceled
	select {
	case <-ctx.Done():
		return nil, errors.New("context canceled")
	case <-d.waitChan{{$rpc.Service}}{{$rpc.Name}}:
		// return the invoked methods response
		return d.handler{{$rpc.Service}}{{$rpc.Name}}.Handle{{$rpc.Service}}{{$rpc.Name}}(ctx, r)
	}
}
{{end}}{{end}}
// NewInboundRPCDispatch will construct a new inboundRPCDispatcher with all fields properly
// initialized.
func NewInboundRPCDispatch() *inboundRPCDispatch {
	return &inboundRPCDispatch{ {{range $service := $Services}}{{range $rpc := $service.RPC}}
		// initialize the wait channel for method {{$rpc.Name}} on service {{$rpc.Service}}
		waitChan{{$rpc.Service}}{{$rpc.Name}}: make(chan struct{}),{{end}}{{end}}
	}
}
{{range $service := $Services}}
// Generated{{$service.Service}}Server implements the {{$service.Service}} service as a gRPC
// server. Generated{{$service.Service}}Server invokes methods on the services
// through the inboundRPCDispatch handlers.
type Generated{{$service.Service}}Server struct {
	dispatch *inboundRPCDispatch
}
{{range $rpc := $service.RPC}}
// {{$rpc.Name}} will invoke the method {{$rpc.Name}} on the RPC service {{$rpc.Service}}
// using the inboundRPCDispatch handler.
func (s *Generated{{$rpc.Service}}Server) {{$rpc.Name}}(ctx context.Context, r *{{$rpc.RequestType}}) (*{{$rpc.ReturnsType}}, error) {
	return s.dispatch.{{$rpc.Service}}{{$rpc.Name}}(ctx, r)
}

{{end}}

// NewGenerated{{$service.Service}}Server constructs a new server for the service.
func NewGenerated{{$service.Service}}Server(dispatch *inboundRPCDispatch) *Generated{{$service.Service}}Server {
  return &Generated{{$service.Service}}Server{
    dispatch: dispatch,
  }
}

{{end}}`

var rpcTest = `{{ $Services := .Services }}// Code generated. DO NOT EDIT.
package {{.Package}}

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)
{{range $service := $Services}}{{range $rpc := $service.RPC}}
type test{{$rpc.Service}}{{$rpc.Name}}Handler struct{}

func (th *test{{$rpc.Service}}{{$rpc.Name}}Handler) Handle{{$rpc.Service}}{{$rpc.Name}}(context.Context, *{{$rpc.RequestType}}) (*{{$rpc.ReturnsType}}, error) {
	return &{{.ReturnsType}}{}, nil
}

func Test{{$rpc.Service}}{{$rpc.Name}}(t *testing.T) {
	// Setup the dispatch handler
	d := NewInboundRPCDispatch()

	// Setup the handler for the TestService
	h := &test{{$rpc.Service}}{{$rpc.Name}}Handler{}

	// Register the handler with the dispatch class
	d.Register{{$rpc.Service}}{{$rpc.Name}}(h)

	// Create the server and pass in the dispatch class
	srvr := Generated{{$rpc.Service}}Server{
		dispatch: d,
	}

	// Test calling the method TestCall
	_, err := srvr.{{$rpc.Name}}(context.Background(), &{{$rpc.RequestType}}{})
	if err != nil {
		t.Error(err)
	}
}

func TestDoubleregistration{{$rpc.Service}}{{$rpc.Name}}(t *testing.T) {
	// Setup the dispatch handler
	d := NewInboundRPCDispatch()

	// Setup the handler for the TestService
	h := &test{{$rpc.Service}}{{$rpc.Name}}Handler{}

	// Register the handler with the dispatch class
	d.Register{{$rpc.Service}}{{$rpc.Name}}(h)

	fn := func() {
		d.Register{{$rpc.Service}}{{$rpc.Name}}(h)
	}
	assert.Panics(t, fn, "double registration must panic")
}

func Test{{$rpc.Service}}{{$rpc.Name}}Cancel(t *testing.T) {
	// Setup the dispatch handler
	d := NewInboundRPCDispatch()

	// Create the server and pass in the dispatch class
	srvr := Generated{{$rpc.Service}}Server{
		dispatch: d,
	}

	// Test calling the method TestCall
	errChan := make(chan error)
	defer close(errChan)
	ctx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	fn := func() {
		_, err := srvr.{{$rpc.Name}}(cancelCtx, &{{$rpc.RequestType}}{})
		errChan <- err
	}
	go fn()
	cancelFunc()
	cancelErr := <-errChan
	assert.EqualError(t, cancelErr, "context canceled", "the error returned must be a context canceled error")
}
{{end}}{{end}}
`
