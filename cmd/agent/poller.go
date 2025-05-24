package main

import (
	"log"
	"time"
)

type Poller struct {
	stats         *Stats
	iteration     uint
	sleepDuration time.Duration
}

func (p *Poller) poll() {
	p.iteration++
	log.Println("Poll - start iteration", p.iteration)
	p.stats.collect()
}

func (p *Poller) runRoutine() {
	for {
		p.poll()
		time.Sleep(p.sleepDuration)
	}
}

func newPoller(stats *Stats, sleepDuration time.Duration) *Poller {
	return &Poller{
		stats:         stats,
		sleepDuration: sleepDuration,
	}
}
