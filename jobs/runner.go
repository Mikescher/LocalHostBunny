package jobs

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gogs.mikescher.com/BlackForestBytes/goext/exerr"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"gogs.mikescher.com/BlackForestBytes/goext/mathext"
	"gogs.mikescher.com/BlackForestBytes/goext/rfctime"
	"gogs.mikescher.com/BlackForestBytes/goext/syncext"
	bunny "locbunny"
	"locbunny/logic"
	"locbunny/models"
	"math/rand"
	"time"
)

type JobFunction[TData any] func(ctx context.Context, app *logic.Application, lstr *JobListener, data *TData) (int, error)

type JobRunner[TData any] struct {
	name       string
	app        *logic.Application
	isRunning  *syncext.AtomicBool
	isStarted  bool
	interval   time.Duration
	sigChannel chan string
	runTimeout time.Duration
	jobFunc    JobFunction[TData]
	data       *TData
}

func NewJobRunner[TData any](app *logic.Application, name string, interval time.Duration, timeout time.Duration, fn JobFunction[TData], data *TData) *JobRunner[TData] {
	return &JobRunner[TData]{
		app:        app,
		isRunning:  syncext.NewAtomicBool(false),
		isStarted:  false,
		sigChannel: make(chan string),
		interval:   interval,
		runTimeout: timeout,
		name:       name,
		jobFunc:    fn,
		data:       data,
	}
}

func (j *JobRunner[TData]) Start() error {
	if j.isRunning.Get() {
		return exerr.New(bunny.ErrJob, "job already running").Build()
	}
	if j.isStarted {
		return exerr.New(bunny.ErrJob, "job was already started").Build() // re-start after stop is not allowed
	}

	j.isStarted = true

	go j.mainLoop()

	return nil
}

func (j *JobRunner[TData]) Stop() {
	log.Info().Msg(fmt.Sprintf("Stopping Job [%s]", j.name))
	syncext.WriteNonBlocking(j.sigChannel, "stop")
	j.isRunning.Wait(false)
	log.Info().Msg(fmt.Sprintf("Stopped Job [%s]", j.name))
}

func (j *JobRunner[TData]) Running() bool {
	return j.isRunning.Get()
}

func (j *JobRunner[TData]) mainLoop() {
	j.isRunning.Set(true)

	firstRun := true
	for {

		interval := j.interval
		if firstRun {
			// randomize first interval to spread jobs around
			perc := mathext.Clamp(rand.Float64(), 0.1, 0.5)
			interval = time.Duration(int64(float64(interval) * perc))
		}
		firstRun = false

		signal, okay := syncext.ReadChannelWithTimeout(j.sigChannel, interval)
		if okay {
			if signal == "stop" {
				log.Info().Msg(fmt.Sprintf("Job [%s] received <stop> signal", j.name))
				break
			} else if signal == "run" {
				log.Info().Msg(fmt.Sprintf("Job [%s] received <run> signal", j.name))
				continue
			} else {
				log.Error().Msg(fmt.Sprintf("Received unknown job signal: <%s> in job [%s]", signal, j.name))
			}
		}

		log.Debug().Msg(fmt.Sprintf("Run job [%s]", j.name))

		err := j.execute()
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("Failed to execute job [%s]: %s", j.name, err.Error()))
		}

	}

	log.Info().Msg(fmt.Sprintf("Job [%s] exiting main-loop", j.name))

	j.isRunning.Set(false)
}

func (j *JobRunner[TData]) execute() (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = exerr.New(bunny.ErrJob, "Recovered panic in JobRunner::execute").Any("recover", rec).Build()
		}
	}()

	runCtx, cancelRunCtx := context.WithTimeout(context.Background(), j.runTimeout)
	defer cancelRunCtx()

	jobExec := models.JobExecution{
		JobExecutionID: models.NewJobExecutionID(),
		JobName:        j.name,
		StartTime:      rfctime.NowRFC3339Nano(),
		EndTime:        nil,
		Changes:        0,
		Status:         models.JobStatusRunning,
	}

	lstr := NewJobListener(j.app, jobExec.JobExecutionID, j.name)

	lstr.LogInfo("JOB_START", "Job started", nil)

	changes, err := langext.RunPanicSafeR2(func() (int, error) { return j.jobFunc(runCtx, j.app, lstr, j.data) })

	//goland:noinspection GoTypeAssertionOnErrors
	if panicerr, ok := err.(langext.PanicWrappedErr); ok {

		jobExec.EndTime = langext.Ptr(rfctime.NowRFC3339Nano())
		jobExec.Error = langext.Ptr(panicerr.Error())
		jobExec.Status = models.JobStatusFailed
		jobExec.Changes = changes
		lstr.LogFatal("JOB_PANIC", "Job finished with a panic", langext.H{"msg": panicerr.Error(), "obj": panicerr.ReoveredObj()})
		log.Error().Str("panic", panicerr.Error()).Msg(fmt.Sprintf("Job '%s' <%s> finished with panic and %d changes after %f minutes", j.name, jobExec.JobExecutionID, changes, jobExec.Delta().Minutes()))

	} else if err != nil {

		jobExec.EndTime = langext.Ptr(rfctime.NowRFC3339Nano())
		jobExec.Error = langext.Ptr(err.Error())
		jobExec.Status = models.JobStatusFailed
		jobExec.Changes = changes
		lstr.LogFatal("JOB_ERR", "Job finished with an error", langext.H{"msg": err.Error(), "err_obj": exerr.FromError(err).ToAPIJson(false, true, true)})
		log.Error().Str("err", err.Error()).Msg(fmt.Sprintf("Job '%s' <%s> finished with an error and %d changes after %f minutes", j.name, jobExec.JobExecutionID, changes, jobExec.Delta().Minutes()))

	} else {

		jobExec.EndTime = langext.Ptr(rfctime.NowRFC3339Nano())
		jobExec.Error = nil
		jobExec.Status = models.JobStatusSuccess
		jobExec.Changes = changes
		lstr.LogInfo("JOB_FINISH", "Job finished", nil)
		log.Info().Msg(fmt.Sprintf("Job '%s' <%s> finished successfully with %d changes after %f minutes", j.name, jobExec.JobExecutionID, changes, jobExec.Delta().Minutes()))

	}

	return nil
}
