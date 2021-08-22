package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"villcore.com/admin/config"
	"villcore.com/admin/controller"
	"villcore.com/admin/core"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	log.Println("AdminServer start.")

	// start scheduler
	scheduler := core.NewSimpleScheduler("scheduler")
	err := scheduler.Start()
	if err != nil {
		log.Fatalln("Start scheduler error ", err)
	}

	// start gin
	router := gin.Default()
	router.LoadHTMLFiles(
		"./asset/templates/index.html",
		"./asset/templates/login.html",
		"./asset/templates/jobinfo/jobinfo.html",
		"./asset/templates/jobgroup/jobgroup.index.html",
		"./asset/templates/user/user.index.html",
		"./asset/templates/help.html",
		"./asset/templates/joblog/joblog.index.html",
		"./asset/templates/joblog/joblog.detail.html",
		"./asset/templates/common/common.exception.html",
	)

	routerGroup := router.Group(config.ServerConfig.ServerContextPath)
	routerGroup.Static("/static", "./asset/static")
	routerGroup.POST("/api/registry", controller.Registry)
	routerGroup.POST("/api/registryRemove", controller.RegistryRemove)
	routerGroup.GET("/", controller.Index)
	routerGroup.GET("/index", controller.Index)
	routerGroup.POST("/index", controller.Index)
	routerGroup.POST("/chartInfo", controller.ChartInfo)
	routerGroup.GET("/help", controller.HelpPage)
	// routerGroup.GET("/error", controller.ErrorPage)

	// job group
	routerGroup.GET("/jobgroup", controller.JobGroup)
	routerGroup.POST("/jobgroup/pageList", controller.JobGroupPageList)
	routerGroup.POST("/jobgroup/loadById", controller.LoadJobGroupById)
	//routerGroup.POST("/jobgroup/remove", controller.RemoveLoadJobGroup)
	//routerGroup.POST("/jobgroup/save", controller.SaveLoadJobGroup)
	//routerGroup.POST("/jobgroup/update", controller.UpdateLoadJobGroup)

	// job info
	routerGroup.GET("/jobinfo", controller.JobInfo)
	//routerGroup.GET("/jobinfo/add", controller.AddJobInfo)
	//routerGroup.POST("/jobinfo/nextTriggerTime", controller.JobInfoNextTriggerTime)
	routerGroup.POST("/jobinfo/pageList", controller.JobInfoPageList)
	//routerGroup.POST("/jobinfo/remove", controller.RemoveJobInfo)
	routerGroup.POST("/jobinfo/start", controller.StartJobInfo)
	routerGroup.POST("/jobinfo/stop", controller.StopJobInfo)
	routerGroup.POST("/jobinfo/trigger", controller.TriggerJobInfo)
	//routerGroup.POST("/jobinfo/update", controller.UpdateJobInfo)

	// login & logout
	routerGroup.GET("/login", controller.LoginPage)
	routerGroup.POST("/login", controller.Login)
	routerGroup.POST("/logout", controller.Logout)

	// user
	routerGroup.GET("/user", controller.User)
	routerGroup.POST("/user/pageList", controller.UserPageList)
	routerGroup.POST("/user/updatePwd", controller.UpdateUserPwd)
	routerGroup.POST("/user/add", controller.AddUser)
	routerGroup.POST("/user/remove", controller.RemoveUser)
	routerGroup.POST("/user/update", controller.UpdateUser)

	// job log
	routerGroup.GET("/joblog", controller.JobLogPage)
	routerGroup.POST("/joblog/pageList", controller.JobLogPageList)
	routerGroup.POST("/joblog/getJobsByGroup", controller.GetJobsByGroup)
	routerGroup.GET("/joblog/logDetailPage", controller.LogDetailPage)
	routerGroup.POST("/joblog/logDetailCat", controller.LogDetailCat)
	routerGroup.POST("/joblog/logKill", controller.LogKill) // TODO
	routerGroup.POST("/joblog/clearLog", controller.ClearLog)

	// TODO
	// 路由类型
	// 其他监控
	_ = router.Run(":" + strconv.Itoa(config.ServerConfig.ServerPort))

	// listen sign
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	<-sigChan

	_ = scheduler.Stop()
	log.Println("AdminServer stop. ")
}
