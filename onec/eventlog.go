package onec

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/golang-lru"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/encoding/charmap"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

type Eventlog struct {
	LastId                uint64
	Name                  string
	Path                  string
	sessionDataSplitCache *lru.Cache
}

type Event struct {
	id uint64
	severity, connectID, session, transactionStatus, transactionId, userCode,
	computerCode, appCode, eventCode, sessionDataSplitCode, dataType, workServerCode,
	primaryPortCode, secondaryPortCode, metadataCode int
	userName, userUuid, computerName, appName, eventName, comment,
	metadataCodes, sessionDataSplitPresentation, data, dataPresentation, workServerName, primaryPortName,
	secondaryPortName, metadataName, metadataUuid string
	date, transactionDate time.Time
}

func NewEventlog(name string, path string) *Eventlog {
	e := &Eventlog{
		Name: name,
		Path: path,
	}
	e.sessionDataSplitCache, _ = lru.New(64) // <- is enough for anyone
	return e
}

func (eventlog *Eventlog) ReadEvents() ([]common.MapStr, uint64, time.Time, error) {
	var events []common.MapStr
	db, err := sql.Open("sqlite3", eventlog.Path)
	if err != nil {
		logp.WTF("%s", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logp.WTF("%s", err)
	}
	rows, err := db.Query(getEventlogQuery(), eventlog.LastId)
	if err != nil {
		logp.WTF("%s", err)
	}
	defer rows.Close()
	var rowNumber, lastId uint64 = 1, 0
	var event Event
	for rows.Next() {
		if rowNumber == 1 {
			events = make([]common.MapStr, 0, 50) // TODO POLLING LENGTH!
		}
		var iDate, iTransactionDate int64
		err = rows.Scan(&event.id, &event.severity, &iDate, &event.connectID,
			&event.session, &event.transactionStatus, &iTransactionDate, &event.transactionId,
			&event.userCode, &event.userName, &event.userUuid, &event.computerCode,
			&event.computerName, &event.appCode, &event.appName, &event.eventCode,
			&event.eventName, &event.comment, &event.metadataCodes, &event.sessionDataSplitCode,
			&event.dataType, &event.data, &event.dataPresentation, &event.workServerCode,
			&event.workServerName, &event.primaryPortCode, &event.primaryPortName, &event.secondaryPortCode,
			&event.secondaryPortName, &event.metadataCode, &event.metadataName, &event.metadataUuid)
		if err != nil {
			logp.WTF("%s", err)
		}
		if event.id == lastId {
			logp.Info("EventLog[%s] multiple occurrences of rowid %d omitted", eventlog.Name, event.id)
			continue
		}
		event.data = encodeWindows1251(event.data)
		event.date = decodeOnecDate(iDate)
		event.transactionDate = decodeOnecDate(iTransactionDate)
		event.sessionDataSplitPresentation = eventlog.getSessionDataSplitPresentation(db, event.sessionDataSplitCode)
		events = append(events, common.MapStr{
			"id":                           &event.id,
			"severity":                     &event.severity,
			"date":                         &event.date,
			"connectId":                    &event.connectID,
			"session":                      &event.session,
			"transactionStatus":            &event.transactionStatus,
			"transactionDate":              &event.transactionDate,
			"transactionId":                &event.transactionId,
			"userName":                     &event.userName,
			"userUuid":                     &event.userUuid,
			"computerName":                 &event.computerName,
			"appName":                      &event.appName,
			"eventName":                    &event.eventName,
			"comment":                      &event.comment,
			"sessionDataSplitPresentation": &event.sessionDataSplitPresentation,
			"dataType":                     &event.dataType,
			"data":                         &event.data,
			"dataPresentation":             &event.dataPresentation,
			"workServerName":               &event.workServerName,
			"primaryPortName":              &event.primaryPortName,
			"secondaryPortName":            &event.secondaryPortName,
			"metadataName":                 &event.metadataName,
			"metadataUuid":                 &event.metadataUuid,
		})
		rowNumber++
		lastId = event.id
	}
	err = rows.Err()
	if err != nil {
		logp.WTF("%s", err)
	}
	return events, event.id, event.date, nil
}

func (eventlog *Eventlog) getSessionDataSplitPresentation(db *sql.DB, sessionDataSplitCode int) string {
	var presentation string
	value, found := eventlog.sessionDataSplitCache.Get(sessionDataSplitCode)
	if found {
		presentation, _ = value.(string)
	} else {
		logp.Info("EventLog[%s] Session data split cache miss", eventlog.Name)
		rows, err := db.Query(getDataSplitQuery(), sessionDataSplitCode)
		if err != nil {
			logp.WTF("%v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var name, data string
			var dataType int
			err = rows.Scan(&name, &dataType, &data)
			if err != nil {
				logp.WTF("%v", err)
			}
			data = encodeWindows1251(data)
			if presentation != "" {
				presentation += ", "
			}
			presentation += fmt.Sprintf("%v: [%d] %v", name, dataType, data)
		}
		err = rows.Err()
		if err != nil {
			logp.WTF("%v", err)
		}
		eventlog.sessionDataSplitCache.Add(sessionDataSplitCode, presentation)
	}
	return presentation
}

func decodeOnecDate(date int64) time.Time {
	// Magic numbers to convert Nuraliev Epoch to Unix Epoch
	var magicNumber1, magicNumber2 int64 = 10000, 62135596800
	epoch := date/magicNumber1 - magicNumber2
	return time.Unix(epoch, 0)
}

func encodeWindows1251(s string) string {
	enc := charmap.Windows1251.NewEncoder()
	out, _ := enc.String(s)
	return out
}
