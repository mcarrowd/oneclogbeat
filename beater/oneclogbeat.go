package beater

import (
	"fmt"
	"sync"
	"time"

	"github.com/mcarrowd/oneclogbeat/config"
	"github.com/mcarrowd/oneclogbeat/onec"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
)

type Oneclogbeat struct {
	config    *config.Settings
	client    publisher.Client
	eventlogs []onec.Eventlog
	done      chan struct{}
}

func New() *Oneclogbeat {
	return &Oneclogbeat{}
}

func (ob *Oneclogbeat) Config(b *beat.Beat) error {
	// Read
	err := b.RawConfig.Unpack(&ob.config)
	if err != nil {
		return fmt.Errorf("Error reading configuration file. %v", err)
	}
	// Validate
	err = ob.config.Validate()
	if err != nil {
		return fmt.Errorf("Error validating configuration file. %v", err)
	}
	logp.Info("Configuration validated. config=%v", ob.config)
	return nil
}

func (ob *Oneclogbeat) Setup(b *beat.Beat) error {
	ob.client = b.Events
	ob.done = make(chan struct{})
	// Populate []eventlogs
	ob.eventlogs = make([]onec.Eventlog, 0, len(ob.config.Oneclogbeat.Eventlogs))
	for _, config := range ob.config.Oneclogbeat.Eventlogs {
		eventlog := onec.Eventlog{
			Name: config.Name,
			Path: config.Path,
		}
		logp.Info("Initialized Eventlog[%s]", eventlog.Name)
		ob.eventlogs = append(ob.eventlogs, eventlog)
	}
	return nil
}

func (ob *Oneclogbeat) Run(b *beat.Beat) error {
	var wg sync.WaitGroup
	for _, eventlog := range ob.eventlogs {
		wg.Add(1)
		go ob.processEventLog(&wg, eventlog)
	}
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

func (ob *Oneclogbeat) processEventLog(wg *sync.WaitGroup, eventlog onec.Eventlog) {
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
		events, err := eventlog.ReadEvents()
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
