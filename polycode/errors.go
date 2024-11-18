package polycode

import (
	"fmt"
)

var ErrBadRequest = DefineError("polycode.client", 2, "bad request")
var ErrTaskExecError = DefineError("polycode.client", 3, "task execution error")
var ErrPanic = DefineError("polycode.client", 5, "task in progress")
var ErrTaskStopped = &ErrPanic
var ErrContextNotFound = DefineError("polycode.client", 6, "context not found")
var ErrInternal = DefineError("polycode.client", 7, "internal error")
var ErrTaskNotFound = DefineError("polycode.client", 8, "task not found")
var ErrServiceExecError = DefineError("polycode.client", 9, "service error")
var ErrApiExecError = DefineError("polycode.client", 10, "api error")

type Error struct {
	Module   string
	ErrorNo  int
	Format   string
	Args     []any
	CauseBy  string
	CanRetry bool
}

func (t Error) Wrap(err error) Error {
	return Error{
		Module:   t.Module,
		ErrorNo:  t.ErrorNo,
		Format:   t.Format,
		Args:     t.Args,
		CauseBy:  err.Error(),
		CanRetry: t.CanRetry,
	}
}

func (t Error) Retry(b bool) Error {
	return Error{
		Module:   t.Module,
		ErrorNo:  t.ErrorNo,
		Format:   t.Format,
		Args:     t.Args,
		CauseBy:  t.CauseBy,
		CanRetry: b,
	}
}

func (t Error) With(args ...any) Error {
	return Error{
		Module:   t.Module,
		ErrorNo:  t.ErrorNo,
		Format:   t.Format,
		Args:     args,
		CauseBy:  t.CauseBy,
		CanRetry: t.CanRetry,
	}
}

func (t Error) Error() string {
	if t.CauseBy == "" {
		return fmt.Sprintf("module: [%s], errorNo : [%d], reason: [%s]", t.Module, t.ErrorNo, fmt.Sprintf(t.Format, t.Args...))
	} else {
		return fmt.Sprintf("module: [%s], errorNo : [%d], reason: [%s], causeBy: [%s]", t.Module, t.ErrorNo, fmt.Sprintf(t.Format, t.Args...), t.CauseBy)
	}
}

func (t Error) ToJson() string {
	if t.CauseBy == "" {
		return fmt.Sprintf(`{"module":"%s","errorNo":%d,"reason":"%s"}`, t.Module, t.ErrorNo, fmt.Sprintf(t.Format, t.Args...))
	} else {
		return fmt.Sprintf(`{"module":"%s","errorNo":%d,"reason":"%s","causeBy":"%s"}`, t.Module, t.ErrorNo, fmt.Sprintf(t.Format, t.Args...), t.CauseBy)
	}
}

func DefineError(module string, errorNo int, format string) Error {
	return Error{
		Module:   module,
		ErrorNo:  errorNo,
		Format:   format,
		CanRetry: false,
	}

}

func IsError(err error, dst Error) bool {
	ret, ok := err.(Error)
	if !ok {
		ret2, ok := err.(*Error)
		if !ok {
			return false
		}
		ret = *ret2
	}
	return ret.Module == dst.Module && ret.ErrorNo == dst.ErrorNo
}

func IsPolycodeError(err error) bool {
	_, ok := err.(Error)
	return ok
}
