package beater

import (
	"sync"
	"time"

	"github.com/mcarrowd/oneclogbeat/onec"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
)

type Oneclogbeat struct {
	client   publisher.Client
	eventlog *onec.Eventlog
	done     chan struct{}
}

func New() *Oneclogbeat {
	return &Oneclogbeat{
		eventlog: &onec.Eventlog{
			DbPath: ".\\src\\github.com\\mcarrowd\\oneclogbeat\\testing\\infobase\\1Cv8Log\\1Cv81.lgd",
		},
	}
}

func (ob *Oneclogbeat) Config(b *beat.Beat) error {
	return nil
}

func (ob *Oneclogbeat) Setup(b *beat.Beat) error {
	ob.client = b.Events
	ob.done = make(chan struct{})
	return nil
}

func (ob *Oneclogbeat) Run(b *beat.Beat) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go ob.processEventLog(&wg)
	wg.Wait()
	return nil
}

func (ob *Oneclogbeat) Cleanup(b *beat.Beat) error {
	logp.Info("Cleaning up Oneclogbeat")
	return nil
}

func (ob *Oneclogbeat) Stop() {
	logp.Info("Stopping Oneclogbeat")
	if ob.done != nil {
		close(ob.done)
	}
}

func (ob *Oneclogbeat) processEventLog(wg *sync.WaitGroup) {
	defer wg.Done()
	logp.Info("Goroutine started")
loop:
	for {
		select {
		case <-ob.done:
			break loop
		default:
		}

		// Read
		events, err := ob.eventlog.ReadEvents()
		if err != nil {
			logp.Warn("ReadEvents() error: %v", err)
			break
		}
		logp.Info("ReadEvents() returned %d records", len(events))

		// Polling
		if len(events) == 0 {
			time.Sleep(time.Second)
			continue
		}

		// Publish
		numEvents := len(events)
		ok := ob.client.PublishEvents(events, publisher.Sync, publisher.Guaranteed)
		if ok {
			logp.Info("Successfully published %d events",
				numEvents)
		} else {
			logp.Warn("Failed to publish %d events",
				numEvents)
		}
	}
}
