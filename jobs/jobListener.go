package jobs

import (
	"gogs.mikescher.com/BlackForestBytes/goext/rfctime"
	"locbunny/logic"
	"locbunny/models"
	"time"
)

type JobListener struct {
	execID  models.JobExecutionID
	start   time.Time
	jobName string
	logs    []models.JobLog
	app     *logic.Application
}

func NewJobListener(app *logic.Application, id models.JobExecutionID, jobName string) *JobListener {
	return &JobListener{
		execID:  id,
		jobName: jobName,
		start:   time.Now(),
		logs:    make([]models.JobLog, 0),
		app:     app,
	}
}

func (lstr *JobListener) Log(lvl models.JobLogLevel, logtype string, msg string, extra any) {
	logentry := models.JobLog{
		JobLogID:       models.NewJobLogID(),
		JobExecutionID: lstr.execID,
		JobName:        lstr.jobName,
		Type:           logtype,
		Time:           rfctime.NowRFC3339Nano(),
		Message:        msg,
		Level:          lvl,
		Extra:          extra,
	}

	lstr.logs = append(lstr.logs, logentry)
}

func (lstr *JobListener) LogDebug(logtype string, msg string, extra any) {
	lstr.Log(models.JobLogLevelDebug, logtype, msg, extra)
}

func (lstr *JobListener) LogInfo(logtype string, msg string, extra any) {
	lstr.Log(models.JobLogLevelInfo, logtype, msg, extra)
}

func (lstr *JobListener) LogWarn(logtype string, msg string, extra any) {
	lstr.Log(models.JobLogLevelWarn, logtype, msg, extra)
}

func (lstr *JobListener) LogError(logtype string, msg string, extra any) {
	lstr.Log(models.JobLogLevelError, logtype, msg, extra)
}

func (lstr *JobListener) LogFatal(logtype string, msg string, extra any) {
	lstr.Log(models.JobLogLevelFatal, logtype, msg, extra)
}
