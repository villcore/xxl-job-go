package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"villcore.com/admin/config"
	"villcore.com/admin/db"
	"villcore.com/admin/misc"
	"villcore.com/common/model"
)

func init() {
	log.SetOutput(os.Stdout)
}

func GetJobInfo(jobId int32) (*model.JobInfo, error) {
	records := make([]model.JobInfo, 0)
	err := db.DbEngine.Table("xxl_job_info").Where("id = ?", jobId).Find(&records)
	if err != nil {
		err := "Get job info query db err " + err.Error()
		log.Println(err)
		return nil, errors.New(err)
	}

	if len(records) < 1 {
		return nil, nil
	}
	return &records[0], nil
}

func GetAllJobGroup() ([]model.JobGroup, error) {
	records := make([]model.JobGroup, 0)
	err := db.DbEngine.Table("xxl_job_group").Find(&records)
	if err != nil {
		err := "Get job group query db err " + err.Error()
		log.Println(err)
		return nil, errors.New(err)
	}

	if len(records) < 1 {
		return nil, nil
	}
	return records, nil
}

func GetJobGroup(jobGroupId int64) (*model.JobGroup, error) {
	records := make([]model.JobGroup, 0)
	err := db.DbEngine.Table("xxl_job_group").Where("id = ?", jobGroupId).Find(&records)
	if err != nil {
		err := "Get job group query db err " + err.Error()
		log.Println(err)
		return nil, errors.New(err)
	}

	if len(records) < 1 {
		return nil, nil
	}
	return &records[0], nil
}

func UpdateJobGroupAddress(id int64, addressList string) error {
	_, err := db.DbEngine.Exec("UPDATE xxl_job_group SET address_list = ?, update_time = ? WHERE id = ? ", addressList, time.Now(), id)
	if err != nil {
		err := "Update job group address err " + err.Error()
		log.Println(err)
		return errors.New(err)
	}
	return nil
}

func SaveJobLog(jobLog *model.JobLog) error {
	_, err := db.DbEngine.Table("xxl_job_log").InsertOne(jobLog)
	return err
}

func GetJobGroupByAddressType(addressType int32) ([]model.JobGroup, error) {
	records := make([]model.JobGroup, 0)
	err := db.DbEngine.Table("xxl_job_group").Where("address_type = ?", addressType).Find(&records)
	return records, err
}

func GetDashboardInfo() (map[string]interface{}, error) {
	jobInfoCount, err := db.DbEngine.Table("xxl_job_info").Count()
	if err != nil {
		log.Println("Get job info count error ", err)
	}

	// TODO: log

	executorCount, err := db.DbEngine.Table("xxl_job_group").Count()
	if err != nil {
		log.Println("Get executor count error ", err)
	}

	dashboardMap := map[string]interface{}{
		"jobInfoCount":       jobInfoCount,
		"jobLogCount":        100,
		"jobLogSuccessCount": 100,
		"executorCount":      executorCount,
	}
	return dashboardMap, nil
}

func GetJobInfoList(start, length, jobGroup, triggerStatus int32, jobDesc, executorHandler, author string) ([]model.JobInfo, int64, error) {
	session := db.DbEngine.Table("xxl_job_info")
	if jobGroup > 0 {
		session = session.Where("job_group = ?", jobGroup)
	}

	if triggerStatus >= 0 {
		session = session.Where("trigger_status = ?", triggerStatus)
	}

	if utf8.RuneCountInString(jobDesc) > 0 {
		session = session.Where("job_desc like ? ", "%"+jobDesc+"%")
	}

	if utf8.RuneCountInString(executorHandler) > 0 {
		session = session.Where("executor_handler like ? ", "%"+executorHandler+"%")
	}

	if utf8.RuneCountInString(author) > 0 {
		session = session.Where("author like ? ", "%"+author+"%")
	}
	session.OrderBy("ID DESC")
	session.Limit(int(length), int(start))

	records := make([]model.JobInfo, 0)
	count, err := session.FindAndCount(&records)
	if err != nil {
		errMsg := "Get job info query db err " + err.Error()
		log.Println(errMsg)
		return nil, 0, errors.New(errMsg)
	}
	return records, count, nil
}

func GetJobGroupList(start, length int32, appname, title string) ([]model.JobGroup, int64, error) {
	session := db.DbEngine.Table("xxl_job_group")
	if utf8.RuneCountInString(appname) > 0 {
		session = session.Where("appname like ? ", "%"+appname+"%")
	}

	if utf8.RuneCountInString(title) > 0 {
		session = session.Where("title like ? ", "%"+title+"%")
	}

	session.OrderBy("ID DESC")
	session.Limit(int(length), int(start))

	records := make([]model.JobGroup, 0)
	count, err := session.FindAndCount(&records)
	if err != nil {
		errMsg := "Get job group query db err " + err.Error()
		log.Println(errMsg)
		return nil, 0, errors.New(errMsg)
	}
	return records, count, nil
}

func GetJobUserList(start, length int32, username string, role int32) ([]model.JobUser, int64, error) {
	session := db.DbEngine.Table("xxl_job_user")
	if utf8.RuneCountInString(username) > 0 {
		session = session.Where("username like ? ", "%"+username+"%")
	}

	if role > -1 {
		session = session.Where("role = ? ", role)
	}

	session.OrderBy("username ASC ")
	session.Limit(int(length), int(start))

	records := make([]model.JobUser, 0)
	count, err := session.FindAndCount(&records)
	if err != nil {
		errMsg := "Get job user query db err " + err.Error()
		log.Println(errMsg)
		return nil, 0, errors.New(errMsg)
	}
	return records, count, nil
}

func GetJobLogList(start, length, jobGroup, jobId, logStatus int32, filterTime string) ([]model.JobLog, int64, error) {

	session := db.DbEngine.Table("xxl_job_log")
	if jobId == 0 && jobGroup > 0 {
		session = session.Where("job_group = ? ", jobGroup)
	}

	if jobId > 0 {
		session = session.Where("job_id = ? ", jobId)
	}

	if logStatus == 1 {
		session = session.Where("handle_code = 200")
	} else if logStatus == 2 {
		session = session.Where("trigger_code NOT IN (0, 200) OR handle_code NOT IN (0, 200))")
	} else if logStatus == 3 {
		session = session.Where("trigger_code = 200 AND handle_code = 0")
	}

	filterTime = strings.TrimSpace(filterTime)
	if utf8.RuneCountInString(filterTime) > 0 {
		splits := strings.Split(filterTime, " - ")

		var startTime, endTime time.Time
		var err error
		if len(splits) == 2 {
			timeFormat := "2006-01-02 15:04:05"
			startTime, err = time.ParseInLocation(timeFormat, splits[0], time.Local)
			endTime, err = time.ParseInLocation(timeFormat, splits[1], time.Local)
		}

		if err == nil && !startTime.After(endTime) {
			session = session.Where("trigger_time >= ? AND trigger_time <= ?", startTime, endTime)
		}
	}

	session.OrderBy("trigger_time DESC")
	session.Limit(int(length), int(start))

	records := make([]model.JobLog, 0)
	count, err := session.FindAndCount(&records)
	if err != nil {
		errMsg := "Get job user query db err " + err.Error()
		log.Println(errMsg)
		return nil, 0, errors.New(errMsg)
	}
	return records, count, nil
}

func GetJobLog(jogLogId int32) (*model.JobLog, error) {
	records := make([]model.JobLog, 0)
	err := db.DbEngine.Table("xxl_job_log").Where("id = ?", jogLogId).Find(&records)
	if err != nil {
		errMsg := "Get job log query db err " + err.Error()
		log.Println(errMsg)
		return nil, err
	}

	if len(records) < 1 {
		return nil, nil
	}
	return &records[0], nil
}

func GetAllJobsByGroup(jobGroup int32) ([]model.JobInfo, error) {
	records := make([]model.JobInfo, 0)
	if jobGroup <= 0 {
		return records, nil
	}

	session := db.DbEngine.Table("xxl_job_info")
	if jobGroup > 0 {
		session = session.Where("job_group = ? ", jobGroup)
	}

	_, err := session.FindAndCount(&records)
	if err != nil {
		errMsg := "Get job info query db err " + err.Error()
		log.Println(errMsg)
		return records, errors.New(errMsg)
	}
	return records, nil
}

func GetJobsByGroup(jobGroup int32) ([]model.JobInfo, error) {
	records := make([]model.JobInfo, 0)
	if jobGroup <= 0 {
		return records, nil
	}

	session := db.DbEngine.Table("xxl_job_info")
	if jobGroup > 0 {
		session = session.Where("job_group = ? ", jobGroup)
	}

	_, err := session.FindAndCount(&records)
	if err != nil {
		errMsg := "Get job info query db err " + err.Error()
		log.Println(errMsg)
		return records, errors.New(errMsg)
	}
	return records, nil
}

func GetJobReport(startTime, endTime time.Time) ([]model.JobLogReport, int64, error) {

	session := db.DbEngine.Table("xxl_job_log_report").
		Where("trigger_day between ? and ? ", startTime, endTime).
		OrderBy("trigger_day ASC")

	records := make([]model.JobLogReport, 0)
	count, err := session.FindAndCount(&records)
	if err != nil {
		errMsg := "Get job log report db err " + err.Error()
		log.Println(errMsg)
		return nil, 0, errors.New(errMsg)
	}
	return records, count, nil
}

func SaveUser(user *model.JobUser) error {
	if user == nil {
		return errors.New("Invalid user ")
	}

	if utf8.RuneCountInString(user.Password) > 0 {
		hash := md5.New()
		_, err := hash.Write([]byte(user.Password))
		if err != nil {
			return err
		}
		passwordMd5 := hex.EncodeToString(hash.Sum(nil))
		user.Password = passwordMd5
	}

	_, err := db.DbEngine.Table("xxl_job_user").Insert(user)
	if err != nil {
		errMsg := "Update job user err " + err.Error()
		log.Println(errMsg)
		return err
	}
	return nil
}

func UpdateUserPassword(id int64, password string) error {
	hash := md5.New()
	_, err := hash.Write([]byte(password))
	if err != nil {
		return err
	}
	passwordMd5 := hex.EncodeToString(hash.Sum(nil))
	_, err = db.DbEngine.Exec("UPDATE xxl_job_user SET password = ? WHERE id = ?", passwordMd5, id)
	if err != nil {
		errMsg := "Update job user password err " + err.Error()
		log.Println(errMsg)
		return err
	}
	return nil
}

func UpdateUser(user *model.JobUser) error {
	if utf8.RuneCountInString(user.Password) > 0 {
		hash := md5.New()
		_, err := hash.Write([]byte(user.Password))
		if err != nil {
			return err
		}
		passwordMd5 := hex.EncodeToString(hash.Sum(nil))
		user.Password = passwordMd5
	}

	_, err := db.DbEngine.Table("xxl_job_user").ID(user.Id).Update(user)
	if err != nil {
		errMsg := "Update job user err " + err.Error()
		log.Println(errMsg)
		return err
	}
	return nil
}

func ClearJobLogByTime(jobGroup, jobId int64, clearBeforeTime time.Time) error {

	if (jobGroup <= 0 && jobId <= 0) || clearBeforeTime.Unix() <= 0 {
		return errors.New("invalid param")
	}

	maxLoopCount := 100 * 10000
	for maxLoopCount > 0 {
		session := db.DbEngine.Table("xxl_job_log")
		if jobGroup > 0 {
			session = session.Where("job_group = ? ", jobGroup)
		}

		if jobId > 0 {
			session = session.Where("job_id = ? ", jobId)
		}

		session = session.Where("trigger_time <= ", clearBeforeTime)
		session = session.OrderBy("id ASC")
		session = session.Limit(1000, 0)
		session = session.Select("id")

		records := make([]model.JobLog, 0)
		_, err := session.FindAndCount(&records)
		if err != nil {
			errMsg := "Get job user query db err " + err.Error()
			log.Println(errMsg)
			return errors.New(errMsg)
		}

		if err := DeleteJobLog(records); err != nil {
			errMsg := "delete job log err " + err.Error()
			log.Println(errMsg)
			return errors.New(errMsg)
		}

		maxLoopCount = maxLoopCount - 1000
		if len(records) <= 0 || maxLoopCount < 0 {
			return nil
		}
	}
	return nil
}

func ClearJobLogByCount(jobGroup, jobId int64, count int32) error {

	if (jobGroup <= 0 && jobId <= 0) || count <= 0 {
		return errors.New("invalid param")
	}

	remainCount := count
	for remainCount > 0 {
		session := db.DbEngine.Table("xxl_job_log")
		if jobGroup > 0 {
			session = session.Where("job_group = ? ", jobGroup)
		}

		if jobId > 0 {
			session = session.Where("job_id = ? ", jobId)
		}
		session = session.OrderBy("id ASC")
		session = session.Limit(1000, 0)
		session = session.Select("id")

		records := make([]model.JobLog, 0)
		_, err := session.FindAndCount(&records)
		if err != nil {
			errMsg := "Get job user query db err " + err.Error()
			log.Println(errMsg)
			return errors.New(errMsg)
		}

		if err := DeleteJobLog(records); err != nil {
			errMsg := "delete job log err " + err.Error()
			log.Println(errMsg)
			return errors.New(errMsg)
		}

		remainCount = remainCount - 1000
		if len(records) <= 0 {
			return nil
		}
	}
	return nil
}

func DeleteJobLog(records []model.JobLog) error {
	var err error
	recordsCount := len(records)
	if recordsCount >= 2 {
		_, err = db.DbEngine.Exec("DELETE FROM xxl_job_log WHERE id >= ? and id <= ? ", records[0].Id, records[recordsCount-1].Id)
	} else if recordsCount > 0 {
		_, err = db.DbEngine.Exec("DELETE FROM xxl_job_log WHERE id = ? ", records[0].Id)
	}
	if err != nil {
		errMsg := "delete job log err " + err.Error()
		log.Println(errMsg)
		return errors.New(errMsg)
	}
	return nil
}

func RemoveUser(id int32) error {
	result, err := db.DbEngine.Exec("DELETE xxl_job_user WHERE id = ?", id)
	log.Println(result)
	if err != nil {
		errMsg := "Remove job user password err " + err.Error()
		log.Println(errMsg)
		return err
	}
	return nil
}

func StopJobInfo(id int32) error {
	records := make([]model.JobInfo, 0)
	err := db.DbEngine.Table("xxl_job_info").Where("id = ?", id).Find(&records)
	if err != nil {
		err := "Get job info query db err " + err.Error()
		log.Println(err)
		return errors.New(err)
	}

	if len(records) < 1 {
		return nil
	}

	_, err = db.DbEngine.Exec("UPDATE xxl_job_info SET trigger_status = ?, trigger_last_time = ?, trigger_next_time = ?, update_time = ? WHERE id = ?", 0, int64(0), int64(0), time.Now(), int64(id))
	if err != nil {
		return err
	}
	return nil
}

func StartJobInfo(id int32) error {
	records := make([]model.JobInfo, 0)
	err := db.DbEngine.Table("xxl_job_info").Where("id = ?", id).Find(&records)
	if err != nil {
		err := "Get job info query db err " + err.Error()
		log.Println(err)
		return errors.New(err)
	}

	if len(records) < 1 {
		return nil
	}

	record := records[0]
	if record.ScheduleType == misc.NONE {
		return errors.New(config.I18n["schedule_type_none_limit_start"])
	}

	triggerLastTime, triggerNextTime, err := RefreshNextTriggerTime(record)
	if err != nil {
		return errors.New(config.I18n["schedule_type"] + config.I18n["system_unvalid"])
	}

	_, err = db.DbEngine.Exec("UPDATE xxl_job_info SET trigger_status = ?, trigger_last_time = ?, trigger_next_time = ?, update_time = ? WHERE id = ?", 1, triggerLastTime, triggerNextTime, time.Now(), int64(id))
	if err != nil {
		return err
	}
	return nil
}

func RefreshNextTriggerTime(record model.JobInfo) (triggerLastTime int64, triggerNextTime int64, err error) {
	switch record.ScheduleType {
	case "NONE":
		return 0, 0, nil

	case "CRON":
		cron, err := misc.ParseCronExpression(record.ScheduleConf)
		if err != nil {
			log.Println("Parse cron.go ", record.ScheduleConf, " error")
			break
		}

		lastTriggerTime := record.TriggerLastTime
		triggerLastTime = record.TriggerNextTime
		triggerNextTime = cron.NextTime(time.Unix(lastTriggerTime/1000, 0)).Unix() * 1000
		break

	case "FIX_RATE":
		scheduleIntervalTimeMs, err := strconv.Atoi(record.ScheduleConf)
		if err != nil {
			log.Println("Convert schedule conf ", record.ScheduleConf, err)
		}
		triggerLastTime = record.TriggerNextTime
		triggerNextTime = record.TriggerNextTime + int64(scheduleIntervalTimeMs)*1000
		break

	default:
		return triggerLastTime, triggerNextTime, errors.New("invalid schedule type " + record.ScheduleType)
	}
	return triggerLastTime, triggerNextTime, nil
}
