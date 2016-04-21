package onec

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

func getDataSplitQuery() string {
	sql := `
	SELECT
	T2.name as sessionParamName,
	T3.dataType as sessionValDataType,
	T3.data as sessionValData
	FROM SessionDataSplits as T1
	INNER JOIN SessionParamCodes as T2 ON T1.sessionParamCode = T2.code
	INNER JOIN SessionDataCodes as T3 ON T1.sessionParamCode = T3.sessionParamCode and T1.sessionValCode = T3.sessionValCode
	WHERE T1.code = ?
	ORDER BY T1.sessionParamCode
	`
	return sql
}
