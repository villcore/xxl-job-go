package service

import (
	"errors"
	"log"
	"os"
	"time"
	"unicode/utf8"
	"villcore.com/admin/db"
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

func GetJobGroup(jobGroupId int32) (*model.JobGroup, error) {
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
