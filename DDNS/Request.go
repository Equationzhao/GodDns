// Package DDNS
// basic interfaces and tools for DDNS service
package DDNS

import (
	"fmt"
)

const (
	Success = iota
	NotExecute
	Failed
	Timeout
)

type Request interface {
	ToParameters() Parameters
	GetName() string    // return like "dnspod"
	MakeRequest() error // MakeRequest will return error if exist
	Status() Status
	Target() string
}

// ThroughProxy is an interface for service that can make request through a proxy
type ThroughProxy interface {
	RequestThroughProxy() error
}

type Status struct {
	Name   string
	Msg    string
	Status int
}

// const Success = "success"
// const Failed = "failed"

// AppendMsg
// append msg to Status.Msg, using fmt.Sprint
func (s *Status) AppendMsg(msg ...any) *Status {
	s.Msg += fmt.Sprint(msg...)
	return s
}

// AppendMsgF
// append msg to Status.Msg, using fmt.Sprintf
func (s *Status) AppendMsgF(format string, msg ...any) *Status {
	s.Msg += fmt.Sprintf(format, msg...)
	return s
}

func ExecuteRequest(request Request) error {
	return request.MakeRequest()
}
