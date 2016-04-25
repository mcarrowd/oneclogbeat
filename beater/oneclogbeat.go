package beater

import (
	"fmt"
	"sync"
	"time"

	"github.com/mcarrowd/oneclogbeat/config"
	"github.com/mcarrowd/oneclogbeat/onec"

	"github.com/elastic/beats/winlogbeat/checkpoint"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/paths"
	"github.com/elastic/beats/libbeat/publisher"
)

var (
	debugf = logp.MakeDebug("winlogbeat")
)

type Oneclogbeat struct {
	config     *config.Settings
	client     publisher.Client
	eventlogs  []*onec.Eventlog
	checkpoint *checkpoint.Checkpoint
	done       chan struct{}
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
	debugf("Configuration validated. config=%v", ob.config)
	// Registry file grooming
	if ob.config.Oneclogbeat.RegistryFile == "" {
		ob.config.Oneclogbeat.RegistryFile = config.DefaultRegistryFile
	}
	ob.config.Oneclogbeat.RegistryFile = paths.Resolve(paths.Data, ob.config.Oneclogbeat.RegistryFile)
	logp.Info("State will be read from and persisted to %s", ob.config.Oneclogbeat.RegistryFile)
	return nil
}

func (ob *Oneclogbeat) Setup(b *beat.Beat) error {
	ob.client = b.Publisher.Connect()
	ob.done = make(chan struct{})
	// Registry file setup
	var err error
	ob.checkpoint, err = checkpoint.NewCheckpoint(ob.config.Oneclogbeat.RegistryFile, 10, 5*time.Second)
	if err != nil {
		return err
	}
	// Populate []eventlogs
	ob.eventlogs = make([]*onec.Eventlog, 0, len(ob.config.Oneclogbeat.Eventlogs))
	for _, config := range ob.config.Oneclogbeat.Eventlogs {
		eventlog := onec.NewEventlog(config.Name, config.Path)
		debugf("Initialized Eventlog[%s]", eventlog.Name)
		ob.eventlogs = append(ob.eventlogs, eventlog)
	}
	return nil
}

func (ob *Oneclogbeat) Run(b *beat.Beat) error {
	persistedState := ob.checkpoint.States()
	var wg sync.WaitGroup
	for _, eventlog := range ob.eventlogs {
		state, _ := persistedState[eventlog.Name]
		// Run goroutine for each eventlog
		wg.Add(1)
		go ob.processEventLog(&wg, eventlog, state)
	}
	wg.Wait()
	ob.checkpoint.Shutdown()
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
		ob.client.Close()
	}
}

func (ob *Oneclogbeat) processEventLog(wg *sync.WaitGroup, eventlog *onec.Eventlog, state checkpoint.EventLogState) {
	defer wg.Done()
	eventlog.LastId = state.RecordNumber
	debugf("EventLog[%s] goroutine started", eventlog.Name)
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
			logp.Warn("EventLog[%s] ReadEvents() error: %v", eventlog.Name, err)
			break
		}
		debugf("EventLog[%s] ReadEvents() returned %d events", eventlog.Name, len(events))

		// Polling
		if len(events) == 0 {
			time.Sleep(time.Second)
			continue
		}

		// Format
		lastId, timestamp := events[len(events)-1].Id, events[len(events)-1].Date
		eventMaps := make([]common.MapStr, 0, len(events))
		for _, event := range events {
			eventMaps = append(eventMaps, event.ToMapStr())
		}

		// Publish
		numEvents := len(eventMaps)
		ok := ob.client.PublishEvents(eventMaps, publisher.Sync, publisher.Guaranteed)
		if !ok {
			// due to using Sync and Guaranteed the ok will only be false on shutdown.
			// Do not update the internal state and return in this case
			return
		}
		eventlog.LastId = lastId
		logp.Info("EventLog[%s] Successfully published %d events", eventlog.Name, numEvents)

		// Persist achievements!
		ob.checkpoint.Persist(eventlog.Name, lastId, timestamp)
	}
}
