package engine

import "context"

func (e *Executor) ctx() context.Context {
	if e.Session != nil && e.Session.Ctx != nil {
		return e.Session.Ctx
	}
	return context.Background()
}
