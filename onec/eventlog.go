package onec

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/encoding/charmap"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

type Eventlog struct {
	LastId uint64
	Name   string
	Path   string
}

type Event struct {
	id uint64
	severity, connectID, session, transactionStatus, transactionId, userCode,
	computerCode, appCode, eventCode, sessionDataSplitCode, dataType, workServerCode,
	primaryPortCode, secondaryPortCode int
	userName, userUuid, computerName, appName, eventName, comment,
	metadataCodes, data, dataPresentation, workServerName, primaryPortName, secondaryPortName string
	date, transactionDate time.Time
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
	rowNumber := 1
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
			&event.secondaryPortName)
		if err != nil {
			logp.WTF("%s", err)
		}
		event.data = encodeWindows1251(event.data)
		event.date = decodeOnecDate(iDate)
		event.transactionDate = decodeOnecDate(iTransactionDate)
		events = append(events, common.MapStr{
			"id":                &event.id,
			"severity":          &event.severity,
			"date":              &event.date,
			"connectId":         &event.connectID,
			"session":           &event.session,
			"transactionStatus": &event.transactionStatus,
			"transactionDate":   &event.transactionDate,
			"transactionId":     &event.transactionId,
			"userName":          &event.userName,
			"userUuid":          &event.userUuid,
			"computerName":      &event.computerName,
			"appName":           &event.appName,
			"eventName":         &event.eventName,
			"comment":           &event.comment,
			"dataType":          &event.dataType,
			"data":              &event.data,
			"dataPresentation":  &event.dataPresentation,
			"workServerName":    &event.workServerName,
			"primaryPortName":   &event.primaryPortName,
			"secondaryPortName": &event.secondaryPortName,
		})
		rowNumber++
	}
	err = rows.Err()
	if err != nil {
		logp.WTF("%s", err)
	}
	return events, event.id, event.date, nil
}

func getEventlogQuery() string {
	sql := `
	SELECT
	T1.rowID as id,
	T1.severity,
	T1.date,
	T1.connectID,
	T1.session,
	T1.transactionStatus,
	T1.transactionDate,
	T1.transactionID,
	T1.userCode,
	ifnull(T2.name, "") as userName,
	ifnull(T2.uuid, "") as userUuid,
	T1.computerCode,
	ifnull(T3.name, "") as computerName,
	T1.appCode,
	ifnull(T4.name, "") as appName,
	T1.eventCode,
	ifnull(T5.name, "") as eventName,
	T1.comment,
	T1.metadataCodes,
	T1.sessionDataSplitCode,
	T1.dataType,
	T1.data,
	T1.dataPresentation,
	T1.workServerCode,
	ifnull(T6.name, "") as workServerName,
	T1.primaryPortCode,
	ifnull(T7.name, "") as primaryPortName,
	T1.secondaryPortCode,
	ifnull(T8.name, "") as secondaryPortName
	FROM  EventLog T1  
	LEFT OUTER JOIN AppCodes T4 ON T1.appCode = T4.code 
	LEFT OUTER JOIN ComputerCodes T3 ON T1.computerCode = T3.code 
	LEFT OUTER JOIN EventCodes T5 ON T1.eventCode = T5.code 
	LEFT OUTER JOIN UserCodes T2 ON T1.userCode = T2.code 
	LEFT OUTER JOIN WorkServerCodes T6 ON T1.workServerCode = T6.code 
	LEFT OUTER JOIN PrimaryPortCodes T7 ON T1.primaryPortCode = T7.code 
	LEFT OUTER JOIN SecondaryPortCodes T8 ON T1.secondaryPortCode = T8.code 
	WHERE (id > ?) 
	ORDER BY id limit 3
	`
	return sql
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
