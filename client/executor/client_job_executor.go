package executor

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"villcore.com/common/api"
)

func init() {
	log.SetOutput(os.Stdout)
}

type JobExecutor interface {
	AddJobHandler(name string, handlerFunc func(param string), initFunc func(), destroyFunc func()) error
	Start() error
	Destroy() error

	getJob(jobId int32) *Job
	runJob(param api.TriggerParam) (*Job, error)
	removeJob(jobId int32) (*Job, error)
}

// job state
const (
	NOT_FOUND = iota
	NOT_START
	RUNNING
	RUNNING_END
)

type Job struct {
	jobId            int32
	state            int32
	ctx              context.Context
	cancel           context.CancelFunc
	triggerParamChan chan *api.TriggerParam

	handlerRegistryEntry *HandlerRegistryEntry
}

type HandlerRegistryEntry struct {
	handlerName string
	handlerFunc func(ctx context.Context, param string)
	initFunc    func()
	destroyFunc func()
}

type XxlJobSimpleExecutor struct {
	jobHandlerRegistryMapping map[string]HandlerRegistryEntry
	runtimeJobHandlerMapping  map[int32]Job
	rwLock                    *sync.RWMutex
	embedHttpServer           EmbedServer
	ctx                       context.Context
	cancel                    context.CancelFunc
	jobConfig                 *JobConfig
	registryTicker            *time.Ticker
}

func New(config *JobConfig) *XxlJobSimpleExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &XxlJobSimpleExecutor{
		jobHandlerRegistryMapping: make(map[string]HandlerRegistryEntry),
		runtimeJobHandlerMapping:  make(map[int32]Job),
		rwLock:                    &sync.RWMutex{},
		ctx:                       ctx,
		cancel:                    cancel,
		jobConfig:                 config,
	}
}

func (e *XxlJobSimpleExecutor) AddJobHandler(name string, handlerFunc func(ctx context.Context, param string), initFunc func(), destroyFunc func()) error {
	if _, ok := e.jobHandlerRegistryMapping[name]; ok {
		return errors.New("Add duplicate job handler '" + name + "'")
	}

	e.jobHandlerRegistryMapping[name] = HandlerRegistryEntry{
		handlerName: name,
		handlerFunc: handlerFunc,
		initFunc:    initFunc,
		destroyFunc: destroyFunc,
	}
	log.Println("Add job handler '" + name + "'")
	return nil
}

func (e *XxlJobSimpleExecutor) Start() error {
	// prepare job handlers

	// init embed server, start listen
	clientExecutorBiz := NewClientExecutorBiz(e)

	e.embedHttpServer = NewHttpServer(e.jobConfig, clientExecutorBiz)
	if err := e.embedHttpServer.Start(); err != nil {
		log.Print("Start embed http server failed ")
		return nil
	}
	log.Print("Start embed http server listen ", e.jobConfig.Port)

	// init admin biz
	adminBiz := ClientAdminApiImpl{
		hostUrl:       e.jobConfig.AminAddresses,
		accessToken:   e.jobConfig.AccessToken,
		timeoutSecond: 3,
	}

	// register this executor
	ticker := time.NewTicker(5 * time.Second)
	e.registryTicker = ticker
	registryCtx := e.ctx
	// defer cancel()
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					registryParam := api.RegistryParam{
						RegistryGroup: "EXECUTOR",
						RegistryKey:   e.jobConfig.Appname,
						RegistryValue: e.jobConfig.Address,
					}
					returnT := adminBiz.Registry(registryParam)
					log.Println("Registry executor client ", returnT)
				}

			case <-registryCtx.Done():
				{
					log.Println("Admin biz client cancel ")
					return
				}
			}
		}
	}()
	return nil
}

func (e *XxlJobSimpleExecutor) Destroy() error {
	log.Println("Executor destroy.")
	// stop accept running request
	//
	// stop all running job; max wait
	if err := e.embedHttpServer.Stop(); err != nil {
		log.Println("Stop embed http server error ", err)
	}
	e.registryTicker.Stop()
	e.cancel()
	return nil
}

func (e *XxlJobSimpleExecutor) getJob(jobId int32) *Job {
	e.rwLock.RLock()
	defer e.rwLock.RUnlock()
	if job, ok := e.runtimeJobHandlerMapping[jobId]; ok {
		return &job
	}
	return nil
}

func (e *XxlJobSimpleExecutor) runJob(param *api.TriggerParam) (*Job, error) {
	log.Println("Run job ", param)
	e.rwLock.Lock()
	defer e.rwLock.Unlock()
	runtimeJob, ok := e.runtimeJobHandlerMapping[param.JobId]
	log.Println("Run job ", runtimeJob)
	if !ok {
		// not exist
		switch param.GlueType {
		case "BEAN":
			// 1. get runtimeJob
			if handlerRegistry, ok := e.jobHandlerRegistryMapping[param.ExecutorHandler]; ok {
				log.Println("Find handler ", handlerRegistry.handlerName, runtimeJob)
				runtimeJob.jobId = param.JobId
				runtimeJob.state = RUNNING
				runtimeJob.ctx, runtimeJob.cancel = context.WithCancel(context.Background())
				runtimeJob.handlerRegistryEntry = &handlerRegistry
				runtimeJob.triggerParamChan = make(chan *api.TriggerParam)
				e.runtimeJobHandlerMapping[param.JobId] = runtimeJob
				go func() {
					for {
						select {
						case <-runtimeJob.ctx.Done():
							log.Printf("Go routine handler %v cancel \n", runtimeJob.handlerRegistryEntry.handlerName)
							return

						case triggerParam, ok := <-runtimeJob.triggerParamChan:
							if !ok {
								log.Printf("Job %v dispatchar cancel \n", runtimeJob.handlerRegistryEntry.handlerName)
								return
							}

							log.Printf("Dispatch %v job trigger param %v \n", triggerParam.ExecutorHandler, triggerParam)
							var ctx context.Context
							var cancel context.CancelFunc
							if triggerParam.ExecutorTimeout > 0 {
								ctx, cancel = context.WithTimeout(runtimeJob.ctx, time.Duration(triggerParam.ExecutorTimeout*1000))
							} else {
								ctx, cancel = context.WithCancel(runtimeJob.ctx)
							}

							var waitGroup sync.WaitGroup
							waitGroup.Add(1)
							go func() {
								defer func() {
									waitGroup.Done()
									log.Println("Job run complete with param ")
								}()

								runtimeJob.handlerRegistryEntry.initFunc()
								runtimeJob.handlerRegistryEntry.handlerFunc(ctx, triggerParam.ExecutorParams)
								runtimeJob.handlerRegistryEntry.destroyFunc()
							}()
							waitGroup.Wait()
							cancel()
						}
					}
				}()
				runtimeJob.triggerParamChan <- param
				return &runtimeJob, nil
			} else {
				return nil, errors.New("Not found job named '" + param.ExecutorHandler + "'")
			}
		}
	} else {
		switch param.GlueType {
		case "BEAN":
			if strings.Compare(runtimeJob.handlerRegistryEntry.handlerName, param.ExecutorHandler) == 0 {
				log.Println("Already run job add param ", param)
				runtimeJob.triggerParamChan <- param
				return &runtimeJob, nil
			} else {
				close(runtimeJob.triggerParamChan)
				if runtimeJob.ctx != nil {
					runtimeJob.ctx.Done()
					runtimeJob.cancel()
				}
				runtimeJob.jobId = param.JobId
				runtimeJob.state = RUNNING
				runtimeJob.ctx, runtimeJob.cancel = context.WithCancel(context.Background())
				runtimeJob.triggerParamChan = make(chan *api.TriggerParam)
			}

			// 1. get runtimeJob
			if handlerRegistry, ok := e.jobHandlerRegistryMapping[param.ExecutorHandler]; ok {
				log.Println("Find handler ", handlerRegistry.handlerName, runtimeJob)
				runtimeJob.handlerRegistryEntry = &handlerRegistry
				e.runtimeJobHandlerMapping[param.JobId] = runtimeJob
				go func() {
					for {
						select {
						case <-runtimeJob.ctx.Done():
							log.Printf("Go routine handler %v cancel \n", runtimeJob.handlerRegistryEntry.handlerName)
							return

						case triggerParam, ok := <-runtimeJob.triggerParamChan:
							if !ok {
								log.Printf("Job %v dispatchar cancel \n", runtimeJob.handlerRegistryEntry.handlerName)
								return
							}

							log.Printf("Dispatch %v job trigger param %v \n", triggerParam.ExecutorHandler, triggerParam)
							var ctx context.Context
							var cancel context.CancelFunc
							if triggerParam.ExecutorTimeout > 0 {
								ctx, cancel = context.WithTimeout(runtimeJob.ctx, time.Duration(triggerParam.ExecutorTimeout*1000))
							} else {
								ctx, cancel = context.WithCancel(runtimeJob.ctx)
							}

							var waitGroup sync.WaitGroup
							waitGroup.Add(1)
							go func() {
								defer func() {
									waitGroup.Done()
									log.Println("Job run complete with param ")
								}()

								runtimeJob.handlerRegistryEntry.initFunc()
								runtimeJob.handlerRegistryEntry.handlerFunc(ctx, triggerParam.ExecutorParams)
								runtimeJob.handlerRegistryEntry.destroyFunc()
							}()
							waitGroup.Wait()
							cancel()
						}
					}
				}()
				runtimeJob.triggerParamChan <- param
				return &runtimeJob, nil
			} else {
				return nil, errors.New("Not found job named '" + param.ExecutorHandler + "'")
			}
		}
	}
	return &runtimeJob, nil
}

func (e *XxlJobSimpleExecutor) removeJob(jobId int32) (*Job, error) {
	e.rwLock.RLock()
	defer e.rwLock.RUnlock()
	if runtimeJob, ok := e.runtimeJobHandlerMapping[jobId]; ok {
		close(runtimeJob.triggerParamChan)
		runtimeJob.ctx.Done()
		delete(e.runtimeJobHandlerMapping, jobId)
		return &runtimeJob, nil
	}
	return nil, errors.New("Not found runtime job '" + strconv.Itoa(int(jobId)) + "'")
}
