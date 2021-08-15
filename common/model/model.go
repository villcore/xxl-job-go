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
	Id          int64     `xorm:"id"`
	AppName     string    `xorm:"app_name"`
	Title       string    `xorm:"title"`
	AddressType int32     `xorm:"address_type"`
	AddressList string    `xorm:"address_list"`
	UpdateTime  time.Time `xorm:"update_time"`
}

type JobLog struct {
	Id                     int64     `xorm:"id"`
	JobGroup               int32     `xorm:"job_group"`
	JobId                  int32     `xorm:"job_id"`
	ExecutorAddress        string    `xorm:"executor_address"`
	ExecutorHandler        string    `xorm:"executor_handler"`
	ExecutorParam          string    `xorm:"executor_param"`
	ExecutorShardingParam  string    `xorm:"executor_sharding_param"`
	ExecutorFailRetryCount int32     `xorm:"executor_fail_retry_count"`
	TriggerTime            time.Time `xorm:"trigger_time"`
	TriggerCode            int32     `xorm:"trigger_code"`
	TriggerMsg             string    `xorm:"trigger_msg"`
	HandleTime             time.Time `xorm:"handle_time"`
	HandleCode             int32     `xorm:"handle_code"`
	HandleMsg              string    `xorm:"handle_msg"`
	AlarmStatus            int32     `xorm:"alarm_status"`
}

type JobUser struct {
	Id         int64  `xorm:"id" json:"id"`
	Username   string `xorm:"username" json:"username"`
	Password   string `xorm:"password" json:"password"`
	Role       int32  `xorm:"role" json:"role"`
	Permission string `xorm:"permission" json:"permission"`
}
