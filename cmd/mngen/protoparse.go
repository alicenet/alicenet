package main

import (
	"os"

	"github.com/emicklei/proto"
)

type service struct {
	Service string
	RPC     map[*proto.RPC]*rPC
}

type rPC struct {
	Service     string
	Name        string
	RequestType string
	ReturnsType string
}

type registrar struct {
	Services    map[*proto.Service]*service
	thisService *proto.Service
	Package     string
}

func newregistrar() *registrar {
	return &registrar{
		Services: make(map[*proto.Service]*service),
	}
}

func (reg *registrar) HandleService(s *proto.Service) {
	reg.Services[s] = &service{
		Service: s.Name,
		RPC:     make(map[*proto.RPC]*rPC),
	}
	reg.thisService = s
}

func (reg *registrar) HandleRPC(r *proto.RPC) {
	reg.Services[reg.thisService].RPC[r] = &rPC{
		reg.thisService.Name,
		r.Name,
		r.RequestType,
		r.ReturnsType,
	}
}

func parseProto(filepath string) *registrar {
	r := newregistrar()
	reader, _ := os.Open(filepath)
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, _ := parser.Parse()

	proto.Walk(definition,
		proto.WithService(r.HandleService),
		proto.WithRPC(r.HandleRPC))
	return r
}
