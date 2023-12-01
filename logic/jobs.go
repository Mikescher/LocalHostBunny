package logic

type Job interface {
	Start() error
	Stop()
}
