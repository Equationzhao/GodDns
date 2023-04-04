// Package DDNS
// basic interfaces and tools for DDNS service
package DDNS

import (
	"fmt"
	"strings"
)

const (
	Success = iota
	NotExecute
	Failed
	Timeout
)

type MsgLevel int

const (
	Info MsgLevel = iota * 4
	Warn
	Error
)

type Msg interface {
	String() string
	Level() MsgLevel
}

type StringMsg struct {
	Msg   []string
	level MsgLevel
	Sep   string
}

func (s *StringMsg) SetSep(sep string) *StringMsg {
	s.Sep = sep
	return s
}

func (s *StringMsg) String() string {
	return strings.Join(s.Msg, s.Sep)
}

func (s *StringMsg) Level() MsgLevel {
	return s.level
}

func (s *StringMsg) AppendAssign(msg ...string) *StringMsg {
	s.Msg = append(s.Msg, msg...)
	return s
}

func NewStringMsg(level MsgLevel) *StringMsg {
	return &StringMsg{level: level, Sep: " "}
}

func (s *StringMsg) SetLevel(level MsgLevel) (ok bool) {
	if level == Info || level == Warn || level == Error {
		s.level = level
		return true
	}
	return false
}

type AnyMsg struct {
	Msg   []any
	level MsgLevel
	Sep   string
}

func (a *AnyMsg) AppendAssign(msg *AnyMsg) *AnyMsg {
	a.Msg = append(a.Msg, msg.Msg...)
	return a
}

func (a *AnyMsg) SetLevel(level MsgLevel) (ok bool) {
	if level == Info || level == Warn || level == Error {
		a.level = level
		return true
	}
	return false
}

func NewAnyMsg(level MsgLevel) *AnyMsg {
	return &AnyMsg{level: level, Sep: " "}
}

func (a *AnyMsg) String() string {
	// join by space
	MsgStr := make([]string, 0, len(a.Msg))
	for _, msg := range a.Msg {
		MsgStr = append(MsgStr, fmt.Sprintf("%v", msg))
	}
	return strings.Join(MsgStr, " ")
}

func (a *AnyMsg) Level() MsgLevel {
	return a.level
}

func (a *AnyMsg) SetSep(sep string) *AnyMsg {
	a.Sep = sep
	return a
}

type MsgGroup interface {
	GetMsgOf(level MsgLevel) []string
	GetInfo() []Msg
	GetWarn() []Msg
	GetError() []Msg
	Add(msg Msg)
	AddInfo(...string)
	AddWarn(...string)
	AddError(...string)
}

type DefaultMsgGroup struct {
	Info  []string
	Warn  []string
	Error []string
}

func (d *DefaultMsgGroup) GetMsgOf(level MsgLevel) []string {
	switch level {
	case Info:
		return d.Info
	case Warn:
		return d.Warn
	case Error:
		return d.Error
	default:
		panic("unknown msg level")
	}
}

func NewDefaultMsgGroup(l ...uint8) *DefaultMsgGroup {
	defaultLen := 10
	switch len(l) {
	case 0:
		return &DefaultMsgGroup{
			Info:  make([]string, 0, defaultLen),
			Warn:  make([]string, 0, defaultLen),
			Error: make([]string, 0, defaultLen),
		}
	case 1:
		return &DefaultMsgGroup{
			Info:  make([]string, 0, l[0]),
			Warn:  make([]string, 0, defaultLen),
			Error: make([]string, 0, defaultLen),
		}
	case 2:
		return &DefaultMsgGroup{
			Info:  make([]string, 0, l[0]),
			Warn:  make([]string, 0, l[1]),
			Error: make([]string, 0, defaultLen),
		}
	case 3:
		return &DefaultMsgGroup{
			Info:  make([]string, 0, l[0]),
			Warn:  make([]string, 0, l[1]),
			Error: make([]string, 0, l[2]),
		}
	default:
		panic("too many arguments")
	}
}

func (d *DefaultMsgGroup) GetInfo() []Msg {
	msg := make([]Msg, 0, len(d.Info))
	for _, s := range d.Info {
		msg = append(msg, NewStringMsg(Info).AppendAssign(s))
	}
	return msg
}

func (d *DefaultMsgGroup) GetWarn() []Msg {
	msg := make([]Msg, 0, len(d.Info))
	for _, s := range d.Warn {
		msg = append(msg, NewStringMsg(Info).AppendAssign(s))
	}
	return msg
}

func (d *DefaultMsgGroup) GetError() []Msg {
	msg := make([]Msg, 0, len(d.Info))
	for _, s := range d.Error {
		msg = append(msg, NewStringMsg(Info).AppendAssign(s))
	}
	return msg
}

func (d *DefaultMsgGroup) Add(msg Msg) {
	switch msg.Level() {
	case Info:
		d.AddInfo(msg.String())
	case Warn:
		d.AddWarn(msg.String())
	case Error:
		d.AddError(msg.String())
	default:
		panic("unknown msg level")
	}
}

func (d *DefaultMsgGroup) AddInfo(s ...string) {
	d.Info = append(d.Info, s...)
}

func (d *DefaultMsgGroup) AddWarn(s ...string) {
	d.Warn = append(d.Warn, s...)
}

func (d *DefaultMsgGroup) AddError(s ...string) {
	d.Error = append(d.Error, s...)
}

type Request interface {
	ToParameters() Service
	GetName() string    // return like "dnspod"
	MakeRequest() error // MakeRequest will return error if exist
	Status() Status
	Target() string
}

// ThroughProxy is an interface for service that can make request through a proxy
type ThroughProxy interface {
	Request
	RequestThroughProxy() error
}

type Status struct {
	Name   string
	MG     MsgGroup
	Status int
}

// const Success = "success"
// const Failed = "failed"

// AppendMsg
// append msg to Status.MG, using fmt.Sprint
func (s *Status) AppendMsg(msg ...Msg) *Status {
	for _, m := range msg {
		s.MG.Add(m)
	}
	return s
}
