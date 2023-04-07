package Core

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"sync"
)

type PanicErr struct {
}

var ReturnCode = 0

func (p PanicErr) Error() string {
	return `Panic happened`
}

var Errchan = make(chan error, 1)

var MainPanicHandler = NewPanicHandler()

func CatchPanic(output io.Writer) {
	if err := recover(); err != nil {
		MainPanicHandler.Receive(err, debug.Stack())
		PrintPanic(output, Errchan)
	}
}

func PrintPanic(output io.Writer, e chan error) {
	_, _ = fmt.Fprintln(output, "version: ", NowVersion)
	_, _ = fmt.Fprintln(output, "panic: ", MainPanicHandler)
	_, _ = fmt.Fprintln(output, "please report this issue to", IssueURL(), "or send email to", FeedbackEmail())
	ReturnCode = 1
	e <- PanicErr{}
}

type PanicHandler struct {
	panic  []any
	m      sync.RWMutex
	stacks [][]byte
}

func (p *PanicHandler) Bytes() []byte {
	p.m.RLock()
	defer p.m.RUnlock()
	builder := bytes.Buffer{}
	for i, v := range p.panic {
		builder.WriteString(fmt.Sprint(v))
		builder.WriteString("\nStack:\n")
		builder.Write(p.stacks[i])
		builder.WriteByte('\n')
	}

	return builder.Bytes()
}

func (p *PanicHandler) Receive(P any, stack []byte) {
	p.m.Lock()
	defer p.m.Unlock()
	p.panic = append(p.panic, P)
	p.stacks = append(p.stacks, stack)
}

func (p *PanicHandler) Panics() []any {
	p.m.RLock()
	defer p.m.RUnlock()
	return p.panic
}

func (p *PanicHandler) String() string {
	p.m.RLock()
	defer p.m.RUnlock()
	builder := strings.Builder{}
	for i, v := range p.panic {
		builder.WriteString(fmt.Sprint(v))
		builder.WriteString("\nStack:\n")
		builder.Write(p.stacks[i])
		builder.WriteByte('\n')
	}

	return builder.String()
}

func NewPanicHandler() *PanicHandler {
	return &PanicHandler{stacks: make([][]byte, 0, 20), panic: make([]any, 0, 20), m: sync.RWMutex{}}
}
