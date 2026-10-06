package errs

import "errors"

var (
	ErrPlanNotFound   = errors.New("plan not found")
	ErrPlanClosed     = errors.New("plan is closed")
	ErrOpenPlanExists = errors.New("channel already has an open plan")
	ErrNoOpenPlan     = errors.New("no open plan in channel")
	ErrInvalidHora    = errors.New("invalid hora")
)
