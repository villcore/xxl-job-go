package executor

import (
	"log"
	"villcore.com/common/api"
)

type CallbackApi interface {
	Beat() api.ReturnT
	IdleBeat(param *api.IdleBeatParam) api.ReturnT
	Run(param *api.TriggerParam) api.ReturnT
	Kill(param *api.KillParam) api.ReturnT
	Log(param *api.LogParam) api.ReturnT
}

type ClientExecutorBizImpl struct {
	executor *XxlJobSimpleExecutor
}

func NewClientExecutorBiz(executor *XxlJobSimpleExecutor) *ClientExecutorBizImpl {
	return &ClientExecutorBizImpl{
		executor: executor,
	}
}

func (receiver *ClientExecutorBizImpl) Beat() api.ReturnT {
	log.Println("Receive admin beat request ")
	return api.NewSuccessReturnT(nil)
}

func (receiver *ClientExecutorBizImpl) IdleBeat(idleBeatParam *api.IdleBeatParam) api.ReturnT {
	if job := receiver.executor.getJob(idleBeatParam.JobId); job != nil {
		return api.NewFailReturnT("job goroutine is running or has trigger queue.")
	} else {
		return api.NewSuccessReturnT(nil)
	}
}

func (receiver *ClientExecutorBizImpl) Run(triggerParam *api.TriggerParam) api.ReturnT {
	if _, err := receiver.executor.runJob(triggerParam); err != nil {
		return api.NewSuccessReturnT(nil)
	} else {
		return api.NewFailReturnT("job goroutine invoke failed.")
	}
}

func (receiver *ClientExecutorBizImpl) Kill(killParam *api.KillParam) api.ReturnT {
	receiver.executor.removeJob(killParam.JobId)
	return api.NewSuccessReturnT(nil)
}

func (receiver *ClientExecutorBizImpl) Log(logParam *api.LogParam) api.ReturnT {
	// TODO: just ignore
	return api.NewSuccessReturnT(nil)
}
