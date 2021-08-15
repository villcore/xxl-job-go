package executor

type JobHandler struct {
	name           string
	jobHandler     func(param string) JobHandlerReturnResult
	intHandler     func()
	destroyHandler func()
}

type JobHandlerReturnResult struct {
	// TODO
}
