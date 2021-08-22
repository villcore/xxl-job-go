package core

import (
	"context"
	"log"
	"os"
	"strings"
	"time"
	"villcore.com/admin/misc"
	"villcore.com/admin/service"
)

func init() {
	log.SetOutput(os.Stdout)
}

type ExecutorRegistryMonitor interface {
	Start() error
	Stop() error
}

type SimpleJobRegistryMonitor struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRegistryMonitor() *SimpleJobRegistryMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &SimpleJobRegistryMonitor{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (receiver *SimpleJobRegistryMonitor) Start() error {
	log.Println("SimpleRegistryMonitor start ")
	go func() {
		timer := time.NewTimer(time.Second * 10)
		defer timer.Stop()
		for {
			select {
			case <-receiver.ctx.Done():
				break
			case <-timer.C:
				updateJobGroupOnLineRegistry()
				timer.Reset(time.Second * 10)
				break
			}
		}
	}()
	return nil
}

func updateJobGroupOnLineRegistry() {
	// 1. get job group by address type 0
	// 2. remove all dead registry
	// 3. get all online registry
	// 4. update job group with online registry
	jobGroups, err := service.GetJobGroupByAddressType(int32(misc.ADDRESS_AUTO_REGISTER))
	if err != nil {
		log.Println("Get auto register job group error ", err)
	}

	if len(jobGroups) == 0 {
		log.Println("Get empty job group ")
		return
	}

	jobRegistries, err := service.GetDeadJobRegistry(int64(30 * 1000))
	if err != nil {
		log.Println("Get dead job register error ", err)
		return
	}

	for _, jobRegistry := range jobRegistries {
		if err = service.RemoveRegistry(jobRegistry.Id); err != nil {
			log.Println("Remove dead job registry error ", err)
		}
	}

	aliveJobRegistries, err := service.GetAliveJobRegistry(int64(30 * 1000))
	if err != nil {
		log.Println("Get dead job register error ", err)
	}

	groupAliveRegistry := make(map[string][]string)
	for _, jobRegistry := range aliveJobRegistries {
		aliveRegistries := groupAliveRegistry[jobRegistry.RegistryKey]
		if aliveRegistries == nil {
			groupAliveRegistry[jobRegistry.RegistryKey] = aliveRegistries
		}
		aliveRegistries = append(aliveRegistries, jobRegistry.RegistryValue)
		groupAliveRegistry[jobRegistry.RegistryKey] = aliveRegistries
	}

	for _, group := range jobGroups {
		var addressList = ""
		aliveRegistries := groupAliveRegistry[group.AppName]
		if aliveRegistries != nil {
			addressList = strings.Join(aliveRegistries, ",")
		}
		_ = service.UpdateJobGroupAddress(group.Id, addressList)
	}
}

func (receiver *SimpleJobRegistryMonitor) Stop() error {
	log.Println("SimpleRegistryMonitor stop ")
	if receiver.ctx != nil {
		receiver.cancel()
	}
	return nil
}
