package main

import (
	"sync"

	DDNS "GodDns/core"
	log "GodDns/log"
	"GodDns/netinterface"
)

type ServiceCronJob struct {
	ps           []*DDNS.Parameters
	GlobalDevice *netinterface.Device
	wg           *sync.WaitGroup
	times        int // times to run
}

func (r *ServiceCronJob) SetTimes(times int) {
	r.times = times
	r.wg.Add(times)
}

func (r *ServiceCronJob) SetWg(wg *sync.WaitGroup) {
	r.wg = wg
}

func NewServiceCronJob(g *netinterface.Device, ps ...*DDNS.Parameters) *ServiceCronJob {
	return &ServiceCronJob{ps: ps, GlobalDevice: g}
}

func (r *ServiceCronJob) Run() {
	if r.times == 0 {
		return
	}
	defer r.wg.Done()
	defer func() { r.times-- }()
	ps := r.ps
	gd := r.GlobalDevice
	err := ModeController(ps, gd)
	if err != nil {
		log.Error("error running ddns: ", log.String("error", err.Error()))
	}
}
