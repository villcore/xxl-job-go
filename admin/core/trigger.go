package core

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"villcore.com/admin/misc"
	"villcore.com/admin/service"
	"villcore.com/common/api"
	"villcore.com/common/model"
)

var serverExecutorApi *ServerExecutorApiImpl

func init() {
	log.SetOutput(os.Stdout)
	serverExecutorApi = &ServerExecutorApiImpl{
		accessToken: "XXL-JOB-ACCESS-TOKEN",
	}
}

type TriggerJobParam struct {
	JobId                 int32
	TriggerType           misc.TriggerType
	FailRetryCount        int
	ExecutorShardingParam string
	ExecutorParam         string
	AddressList           string
}

func TriggerJob(param *TriggerJobParam) error {
	jobInfo, err := service.GetJobInfo(param.JobId)
	if err != nil {
		return err
	}

	if jobInfo == nil {
		return nil
	}

	if param.ExecutorParam != "" {
		jobInfo.ExecutorParam = param.ExecutorParam
	}

	jobGroup, err := service.GetJobGroup(jobInfo.JobGroup)
	if err != nil {
		return err
	}

	addressListStr := strings.TrimSpace(param.AddressList)
	if utf8.RuneCountInString(addressListStr) > 0 {
		jobGroup.AddressType = int32(misc.ADDRESS_MANUAL_REGISTER)
		jobGroup.AddressList = addressListStr
	}

	shardingIndex, shardingTotal, shadingValid := parseShardingParam(param.ExecutorShardingParam)
	executorList := parseRegistryExecutorList(jobGroup.AddressList)
	if jobInfo.ExecutorRouteStrategy == string(ROUTER_SHARDING_BROADCAST) && len(executorList) > 0 && !shadingValid {
		for seq := range executorList {
			go processTrigger(jobGroup, jobInfo, param.FailRetryCount, param.TriggerType, seq, len(executorList))
		}
	} else {
		go processTrigger(jobGroup, jobInfo, param.FailRetryCount, param.TriggerType, shardingIndex, shardingTotal)
	}
	return nil
}

func parseShardingParam(shardingParamStr string) (int, int, bool) {
	shardingParam := strings.TrimSpace(shardingParamStr)
	if utf8.RuneCountInString(shardingParam) <= 0 {
		return 0, 1, false
	}

	shardingAddrSplit := strings.Split(shardingParam, "/")
	if len(shardingAddrSplit) == 2 {
		shardingIndex, err := strconv.Atoi(shardingAddrSplit[0])
		shardingTotal, err2 := strconv.Atoi(shardingAddrSplit[1])
		if err == nil && err2 == nil {
			return shardingIndex, shardingTotal, true
		}
	}
	return 0, 1, false
}

func parseRegistryExecutorList(registryListStr string) []string {
	registryList := strings.TrimSpace(registryListStr)
	if utf8.RuneCountInString(registryList) > 0 {
		return strings.Split(registryList, ",")
	}
	return make([]string, 0)
}

func processTrigger(group *model.JobGroup, jobInfo *model.JobInfo,
	failRetryCount int, triggerType misc.TriggerType,
	shardingIndex int, shardingTotal int) {

	// save log
	jobLog := model.JobLog{
		JobId:       int32(jobInfo.Id),
		JobGroup:    int32(group.Id),
		TriggerTime: time.Now(),
	}

	err := service.SaveJobLog(&jobLog)
	if err != nil {
		log.Println("Save job log err ", err)
	}

	// prepare trigger param
	executorTriggerParam := &api.TriggerParam{
		JobId:                 int32(jobInfo.Id),
		ExecutorHandler:       jobInfo.ExecutorHandler,
		ExecutorParams:        jobInfo.ExecutorParam,
		ExecutorBlockStrategy: jobInfo.ExecutorBlockStrategy,
		ExecutorTimeout:       int64(jobInfo.ExecutorTimeout),
		LogId:                 jobLog.Id,
		LogDateTime:           jobLog.TriggerTime.Unix(),
		GlueType:              jobInfo.GlueType,
		GlueSource:            jobInfo.GlueSource,
		GlueUpdateTime:        jobInfo.GlueUpdatetime.Unix(),
		BroadcastIndex:        int32(shardingIndex),
		BroadcastTotal:        int32(shardingTotal),
	}

	result := api.NewFailReturnT("jobconf_trigger_address_empty")
	routeStrategy := jobInfo.ExecutorRouteStrategy
	executorAddressList := parseRegistryExecutorList(group.AddressList)
	var targetExecutorAddress string
	if len(executorAddressList) > 0 {
		if routeStrategy == string(ROUTER_SHARDING_BROADCAST) {
			if shardingIndex < len(executorAddressList) {
				targetExecutorAddress = executorAddressList[shardingIndex]
			} else {
				targetExecutorAddress = executorAddressList[0]
			}
		} else {
			// router
			targetExecutorAddress, err = Route(RouterType(jobInfo.ExecutorRouteStrategy), executorTriggerParam, executorAddressList)
			if err != nil {
				result = api.NewFailReturnT(err.Error())
			}
		}
	}

	if utf8.RuneCountInString(targetExecutorAddress) > 0 {
		runExecutor(executorTriggerParam, targetExecutorAddress)
	}

	log.Println(result)
	// TODO: update log
	//blockStrategy := jobInfo.ExecutorBlockStrategy
	//var shardingParam = ""
	//if routeStrategy == string(ROUTER_SHARDING_BROADCAST) {
	//	shardingParam = strings.Join([]string{strconv.Itoa(shardingIndex), strconv.Itoa(shardingTotal)}, "/")
	//}
}

func runExecutor(param *api.TriggerParam, targetExecutorAddress string) api.ReturnT {
	result := serverExecutorApi.run(targetExecutorAddress, param)
	log.Println("Server run trigger param ", param, " result ", result)
	return result
}
