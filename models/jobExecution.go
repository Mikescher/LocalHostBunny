package models

import (
	"gogs.mikescher.com/BlackForestBytes/goext/rfctime"
	"time"
)

type JobStatus string //@enum:type

const (
	JobStatusRunning JobStatus = "RUNNING"
	JobStatusSuccess JobStatus = "SUCCESS"
	JobStatusFailed  JobStatus = "FAILED"
)

type JobExecution struct {
	JobExecutionID JobExecutionID           `bson:"_id,omitempty" json:"id"`
	JobName        string                   `bson:"jobName"       json:"jobName"`
	StartTime      rfctime.RFC3339NanoTime  `bson:"startTime"     json:"startTime"`
	EndTime        *rfctime.RFC3339NanoTime `bson:"endTime"       json:"endTime"`
	Changes        int                      `bson:"changes"       json:"changes"`
	Status         JobStatus                `bson:"status"        json:"status"`
	Error          *string                  `bson:"error"         json:"error"`
}

func (u JobExecution) GetID() AnyID {
	return u.JobExecutionID.AsAny()
}

func (u JobExecution) GetCreationTime() time.Time {
	return u.StartTime.Time()
}

func (u JobExecution) Delta() time.Duration {
	if u.EndTime != nil {
		return u.EndTime.Sub(u.StartTime)
	} else {
		return time.Now().Sub(u.StartTime.Time())
	}
}
