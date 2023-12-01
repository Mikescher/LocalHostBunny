package models

import (
	"gogs.mikescher.com/BlackForestBytes/goext/rfctime"
	"time"
)

type JobLogLevel string //@enum:type

const (
	JobLogLevelDebug JobLogLevel = "DEBUG"
	JobLogLevelInfo  JobLogLevel = "INFO"
	JobLogLevelWarn  JobLogLevel = "WARN"
	JobLogLevelError JobLogLevel = "ERROR"
	JobLogLevelFatal JobLogLevel = "FATAL"
)

type JobLog struct {
	JobLogID       JobLogID                 `bson:"_id,omitempty" json:"id"`
	JobExecutionID JobExecutionID           `bson:"executionId"   json:"executionId"`
	JobName        string                   `bson:"jobName"       json:"jobName"`
	Type           string                   `bson:"type"          json:"type"`
	Time           rfctime.RFC3339NanoTime  `bson:"time"          json:"time"`
	Message        string                   `bson:"message"       json:"message"`
	Level          JobLogLevel              `bson:"level"         json:"level"`
	Extra          any                      `bson:"extra"         json:"extra"`
	Deleted        *rfctime.RFC3339NanoTime `bson:"deleted"       json:"deleted"`
}

func (u JobLog) GetID() AnyID {
	return u.JobLogID.AsAny()
}

func (u JobLog) GetCreationTime() time.Time {
	return u.Time.Time()
}
