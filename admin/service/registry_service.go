package service

import (
	"errors"
	"log"
	"os"
	"time"
	"villcore.com/admin/db"
	"villcore.com/common/api"
	"villcore.com/common/model"
)

func init() {
	log.SetOutput(os.Stdout)
}

func RegisterExecutor(param *api.RegistryParam) error {

	registryGroup, registryKey, registryVal := param.RegistryGroup, param.RegistryKey, param.RegistryValue
	if registryGroup == "" || registryKey == "" || registryVal == "" {
		return errors.New("Invalid param ")
	}

	// 1. try to update
	result, err := db.DbEngine.Exec("UPDATE xxl_job_registry SET update_time = ? WHERE registry_group = ? AND registry_key = ? AND registry_value = ? ", time.Now(), registryGroup, registryKey, registryVal)
	if err != nil {
		log.Printf("Update registry %v error %v \n", param, err)
		return errors.New("Update registry error ")
	}

	affectRows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Update registry %v error %v \n", param, err)
		return errors.New("Update registry error ")
	}

	if affectRows > 0 {
		log.Println("Update registry success ", param)
		return nil
	}

	// 2. try to insert
	// save & update
	jobRegistry := model.JobRegistry{
		RegistryGroup: registryGroup,
		RegistryKey:   registryKey,
		RegistryValue: registryVal,
		UpdateTime:    time.Now(),
	}
	_, err = db.DbEngine.Table("xxl_job_registry").InsertOne(&jobRegistry)
	if err != nil {
		log.Printf("Insert registry %v error %v \n", param, err)
		return errors.New("Insert registry error ")
	}
	log.Println("Insert registry success ", param)
	return nil
}

func RemoveRegisterExecutor(param *api.RegistryParam) error {

	registryGroup, registryKey, registryVal := param.RegistryGroup, param.RegistryKey, param.RegistryValue
	if registryGroup == "" || registryKey == "" || registryVal == "" {
		return errors.New("Invalid param ")
	}

	// 1. try to update
	result, err := db.DbEngine.Exec("DELETE xxl_job_registry WHERE registry_group = ? AND registry_key = ? AND registry_value = ? ", registryGroup, registryKey, registryVal)
	if err != nil {
		log.Printf("Delete registry %v error %v \n", param, err)
		return errors.New("Remove registry error ")
	}

	affectRows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Remove registry %v error %v \n", param, err)
		return errors.New("Remove registry error ")
	}

	if affectRows > 0 {
		log.Println("Update registry success ", param)
		return nil
	}
	return nil
}

func GetDeadJobRegistry(timeout int64) ([]model.JobRegistry, error) {
	checkedTime := time.Now().Add(time.Duration(-1 * int64(time.Millisecond) * int64(timeout)))
	records := make([]model.JobRegistry, 0)
	err := db.DbEngine.Table("xxl_job_registry").
		Where("update_time < ? ", checkedTime).
		Find(&records)
	return records, err
}

func RemoveRegistry(id int64) error {
	result, err := db.DbEngine.Exec("DELETE FROM xxl_job_registry WHERE id = ? ", id)
	if err != nil {
		log.Printf("Delete registry %v error %v \n", id, err)
		return errors.New("Remove registry error ")
	}

	affectRows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Remove registry %v error %v \n", id, err)
		return errors.New("Remove registry error ")
	}

	if affectRows > 0 {
		log.Println("Remove registry success ", id)
		return nil
	}
	return nil
}

func GetAliveJobRegistry(timeFormNow int64) ([]model.JobRegistry, error) {
	checkedTime := time.Now().Add(-1 * time.Duration(int64(time.Millisecond)*int64(timeFormNow)))
	records := make([]model.JobRegistry, 0)
	err := db.DbEngine.Table("xxl_job_registry").
		Where("update_time >= ? ", checkedTime).
		Find(&records)
	return records, err
}
