package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"villcore.com/client/executor"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	log.Println("Start demo.")
	jobConfig := executor.JobConfig{
		AminAddresses:    "http://127.0.0.1:8081/xxl-job-admin",
		AccessToken:      "",
		Appname:          "xxl-job-executor-sample",
		Address:          "http://127.0.0.1:9998/",
		Ip:               "",
		Port:             9998,
		LogPath:          "",
		LogRetentionDays: "",
	}

	simpleExecutor := executor.New(&jobConfig)
	_ = simpleExecutor.AddJobHandler("demoJobHandler",
		func(ctx context.Context, param string) {
			log.Println("Trigger job run ")
		},
		func() {
			log.Println("demoJobHandler init func ")
		},
		func() {
			log.Println("demoJobHandler destroy func ")
		})

	_ = simpleExecutor.AddJobHandler("demoJobHandler2",
		func(ctx context.Context, param string) {
			log.Println("Trigger job run 2 ")
		},
		func() {
			log.Println("demoJobHandler init func 2 ")
		},
		func() {
			log.Println("demoJobHandler destroy func 2 ")
			log.Println("==============================")
		})

	log.Println("Start simpleExecutor.")
	err := simpleExecutor.Start()
	if err != nil {
		log.Fatalln("Start simpleExecutor %V failed ", jobConfig.Appname)
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	<-sigChan
_:
	simpleExecutor.Destroy()
	log.Println("Stop simpleExecutor ", jobConfig.Appname)
}
