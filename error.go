package mediator

type IError interface {
	error
	Code() uint
}

//go:generate stringer -type=tError -trimprefix=Error -output=error_string.go

type tError uint

const (
	ErrorRequestHandlerAlreadyExists tError = iota + 1
	ErrorRequestPipelineBehaviorAlreadyExists
	ErrorRequestHandlerNotFound
	ErrorRequestHandlerNotValid
	ErrorNotificationHandlerNotValid
)

func (r tError) Error() string {
	return r.String()
}

func (r tError) Code() uint {
	return uint(r)
}
