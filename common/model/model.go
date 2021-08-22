package model

import "time"

type JobRegistry struct {
	Id            int64     `xorm:"id"`
	RegistryGroup string    `xorm:"registry_group"`
	RegistryKey   string    `xorm:"registry_key"`
	RegistryValue string    `xorm:"registry_value"`
	UpdateTime    time.Time `xorm:"update_time"`
}

type JobInfo struct {
	Id                     int64     `xorm:"id" json:"id"`
	JobGroup               int64     `xorm:"job_group" json:"jobGroup"`
	JobDesc                string    `xorm:"job_desc" json:"jobDesc"`
	AddTime                time.Time `xorm:"add_time" json:"addTime"`
	Author                 string    `xorm:"author" json:"author"`
	AlarmEmail             string    `xorm:"alarm_email" json:"alarmEmail"`
	ScheduleType           string    `xorm:"schedule_type" json:"scheduleType"`
	ScheduleConf           string    `xorm:"schedule_conf" json:"scheduleConf"`
	MisfireStrategy        string    `xorm:"misfire_strategy" json:"misfireStrategy"`
	ExecutorRouteStrategy  string    `xorm:"executor_route_strategy" json:"executorRouteStrategy"`
	ExecutorHandler        string    `xorm:"executor_handler" json:"executorHandler"`
	ExecutorParam          string    `xorm:"executor_param" json:"executorParam"`
	ExecutorBlockStrategy  string    `xorm:"executor_block_strategy" json:"executorBlockStrategy"`
	ExecutorTimeout        int32     `xorm:"executor_timeout" json:"executorTimeout"`
	ExecutorFailRetryCount int32     `xorm:"executor_fail_retry_count" json:"executorFailRetryCount"`
	GlueType               string    `xorm:"glue_type" json:"glueType"`
	GlueSource             string    `xorm:"glue_source" json:"glueSource"`
	GlueRemark             string    `xorm:"glue_remark" json:"glueRemark"`
	GlueUpdatetime         time.Time `xorm:"glue_updatetime" json:"glueUpdatetime"`
	ChildJobid             string    `xorm:"child_jobid" json:"childJobid"`
	TriggerStatus          int32     `xorm:"trigger_status" json:"triggerStatus"`
	TriggerLastTime        int64     `xorm:"trigger_last_time" json:"triggerLastTime"`
	TriggerNextTime        int64     `xorm:"trigger_next_time" json:"triggerNextTime"`
	UpdateTime             time.Time `xorm:"update_time" json:"updateTime"`
}

type JobGroup struct {
	Id          int64     `xorm:"id" json:"id"`
	AppName     string    `xorm:"app_name" json:"appname"`
	Title       string    `xorm:"title" json:"title"`
	AddressType int32     `xorm:"address_type" json:"addressType"`
	AddressList string    `xorm:"address_list" json:"addressList"`
	UpdateTime  time.Time `xorm:"update_time" json:"updateTime"`
}

type JobLog struct {
	Id                     int64     `xorm:"id" json:"id"`
	JobGroup               int32     `xorm:"job_group" json:"jobGroup"`
	JobId                  int32     `xorm:"job_id" json:"jobId"`
	ExecutorAddress        string    `xorm:"executor_address" json:"executorAddress"`
	ExecutorHandler        string    `xorm:"executor_handler" json:"executorHandler"`
	ExecutorParam          string    `xorm:"executor_param" json:"executorParam"`
	ExecutorShardingParam  string    `xorm:"executor_sharding_param" json:"executorShardingParam"`
	ExecutorFailRetryCount int32     `xorm:"executor_fail_retry_count" json:"executorFailRetryCount"`
	TriggerTime            time.Time `xorm:"trigger_time" json:"triggerTime"`
	TriggerCode            int32     `xorm:"trigger_code" json:"triggerCode"`
	TriggerMsg             string    `xorm:"trigger_msg" json:"triggerMsg"`
	HandleTime             time.Time `xorm:"handle_time"`
	HandleCode             int32     `xorm:"handle_code" json:"handleCode"`
	HandleMsg              string    `xorm:"handle_msg" json:"handleMsg"`
	AlarmStatus            int32     `xorm:"alarm_status" json:"alarmStatus"`
}

type JobUser struct {
	Id         int64  `xorm:"id" json:"id"`
	Username   string `xorm:"username" json:"username"`
	Password   string `xorm:"password" json:"password"`
	Role       int32  `xorm:"role" json:"role"`
	Permission string `xorm:"permission" json:"permission"`
}

type JobLogReport struct {
	Id           int64     `xorm:"id" json:"id"`
	TriggerDay   time.Time `xorm:"trigger_day" json:"triggerDay"`
	RunningCount int32     `xorm:"running_count" json:"runningCount"`
	SucCount     int32     `xorm:"suc_count" json:"sucCount"`
	FailCount    int32     `xorm:"fail_count" json:"failCount"`
	UpdateTime   time.Time `xorm:"update_time" json:"updateTime"`
}
