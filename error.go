package mediator

type IError interface {
	error
	Code() uint
}

//go:generate stringer -type=Error -trimprefix=Error

type Error uint

const (
	ErrorRequestHandlerAlreadyExists Error = iota + 1
	ErrorRequestPipelineBehaviorAlreadyExists
	ErrorRequestHandlerNotFound
	ErrorRequestHandlerNotValid
	ErrorNotificationHandlerNotValid
)

func (r Error) Error() string {
	return r.String()
}

func (r Error) Code() uint {
	return uint(r)
}
