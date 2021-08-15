package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"syscall"
	"villcore.com/admin/config"
	"villcore.com/admin/controller"
	"villcore.com/admin/core"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	log.Println(config.I18n)

	log.Println("AdminServer start.")

	// start scheduler
	scheduler := core.NewSimpleScheduler("scheduler")
	err := scheduler.Start()
	if err != nil {
		log.Fatalln("Start scheduler error ", err)
	}

	router := gin.Default()
	router.LoadHTMLFiles(
		"./asset/templates/index.html",
		"./asset/templates/login.html",
		"./asset/templates/jobinfo/jobinfo.html",
	)
	routerGroup := router.Group(config.ServerConfig.ServerContextPath)
	routerGroup.Static("/static", "./asset/static")
	routerGroup.POST("/api/registry", controller.Registry)
	routerGroup.POST("/api/registryRemove", controller.RegistryRemove)

	// template
	routerGroup.GET("/", controller.Index)
	routerGroup.GET("/index", controller.Index)
	routerGroup.GET("/login", controller.LoginPage)
	routerGroup.POST("/login", controller.Login)
	routerGroup.POST("/logout", controller.Logout)
	routerGroup.GET("/jobinfo", controller.JobInfo)
	routerGroup.POST("/jobinfo/pageList", controller.JobInfoPageList)
	_ = router.Run(":8081")

	// listen sign
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	<-sigChan

	_ = scheduler.Stop()
	log.Println("AdminServer stop. ")
}
