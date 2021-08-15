package core

import (
	"context"
	"github.com/go-xorm/xorm"
	"log"
	"os"
	"strconv"
	"time"
	"villcore.com/admin/db"
	"villcore.com/common/model"
)

func init() {
	log.SetOutput(os.Stdout)
}

type SimpleScheduler struct {
	name            string
	scanTimer       *time.Timer
	timeWheel       *TimeWheel
	registryMonitor *SimpleJobRegistryMonitor
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewSimpleScheduler(name string) *SimpleScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &SimpleScheduler{
		name:            name,
		registryMonitor: NewRegistryMonitor(),
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (receiver *SimpleScheduler) Start() error {
	log.Println("SimpleScheduler start ")
	if err := receiver.startScheduler(); err != nil {
		return err
	}

	if err := receiver.registryMonitor.Start(); err != nil {
		return err
	}
	return nil
}

func (receiver *SimpleScheduler) startScheduler() error {

	// start time wheel
	tw, err := NewTimeWheel(receiver.name, 5*60, 1*time.Second)
	if err != nil {
		return err
	}
	receiver.timeWheel = tw
	receiver.timeWheel.Start()

	// scan latest job goroutine
	go receiver.startScanLatestScheduleJob()

	return nil
}

func (receiver *SimpleScheduler) startScanLatestScheduleJob() {
	receiver.scanTimer = time.NewTimer(time.Second * 5)
	for {
		select {
		case <-receiver.ctx.Done():
			break

		case <-receiver.scanTimer.C:
			jobCount, err := receiver.scanLatestScheduleJob()
			currentTimeMillis := time.Now().UnixNano() / 10e6
			timeOffset := currentTimeMillis % int64(1000)
			if err != nil || jobCount <= 0 {
				log.Printf("Scan latest schedule job total count %v error %v ", jobCount, err)
				receiver.scanTimer.Reset(time.Duration(int64(time.Millisecond)*int64(5000) - timeOffset))
			} else {
				receiver.scanTimer.Reset(time.Duration(int64(time.Millisecond)*int64(1000) - timeOffset))
			}
			break
		}
	}
}

func (receiver *SimpleScheduler) scanLatestScheduleJob() (int, error) {
	count, err := db.DbEngine.Transaction(func(session *xorm.Session) (interface{}, error) {
		start := time.Now()
		defer func() {
			end := time.Now()
			log.Printf("Scan latest schedule job use time %v ms \n", end.Sub(start).Milliseconds())
		}()

		result, err := session.Exec("select * from xxl_job_lock where lock_name = 'schedule_lock' for update")
		if err != nil {
			log.Println("Try lock xxl_job_lock error ", err)
			return 0, err
		}

		log.Println("Try lock xxl_job_lock result ", result)
		records := make([]model.JobInfo, 0)
		nowTimeMillis := time.Now().Unix() * 1000
		err = session.Table("xxl_job_info").Where("trigger_status = ? AND trigger_next_time < ?", 1, nowTimeMillis).OrderBy("ID ASC").Find(&records)
		if err != nil {
			log.Println("Select xxl_job_info result error ", err)
			return 0, err
		}

		for _, record := range records {
			// 已经过期任务
			if nowTimeMillis > record.TriggerNextTime+int64(time.Second)*5 {
				log.Println("1_______________________________________________")

				// misfire
				switch record.MisfireStrategy {
				case "DO_NOTHING":
					break
				case "FIRE_ONCE_NOW":
					err := TriggerJob(&TriggerJobParam{
						JobId:                 int32(record.Id),
						TriggerType:           TRIGGER_MISFIRE,
						FailRetryCount:        -1,
						ExecutorShardingParam: "",
						ExecutorParam:         "",
						AddressList:           "",
					})
					if err != nil {
						log.Println("Trigger job ", record, " error ", err)
					}
					break
				}
			}

			// 本循环周期内任务
			if nowTimeMillis > record.TriggerNextTime {
				log.Println("2_______________________________________________")
				err := receiver.timeWheel.AddTimeTask(
					func() {
						err := TriggerJob(&TriggerJobParam{
							JobId:                 int32(record.Id),
							TriggerType:           TRIGGER_CRON,
							FailRetryCount:        -1,
							ExecutorShardingParam: "",
							ExecutorParam:         "",
							AddressList:           "",
						})
						if err != nil {
							log.Println("Trigger job ", record, " error ", err)
						}
					},
					time.Unix(0, record.TriggerNextTime*int64(time.Millisecond)),
				)

				if err != nil {
					log.Println("Add job ", record, " to time wheel error ", err)
				}
			}

			// 下一个循环周期内任务
			if nowTimeMillis > record.TriggerNextTime+int64(time.Second)*5 {
				log.Println("3_______________________________________________")
				err := receiver.timeWheel.AddTimeTask(
					func() {
						err := TriggerJob(&TriggerJobParam{
							JobId:                 int32(record.Id),
							TriggerType:           TRIGGER_CRON,
							FailRetryCount:        -1,
							ExecutorShardingParam: "",
							ExecutorParam:         "",
							AddressList:           "",
						})
						if err != nil {
							log.Println("Trigger job ", record, " error ", err)
						}
					},
					time.Unix(0, record.TriggerNextTime*int64(time.Millisecond)),
				)

				if err != nil {
					log.Println("Add job ", record, " to time wheel error ", err)
				}
			}

			// refreshNextTriggerTime
			switch record.ScheduleType {
			case "NONE":
				record.TriggerLastTime = 0
				record.TriggerNextTime = 0
				break

			case "CRON":
				cron, err := ParseCronExpression(record.ScheduleConf)
				if err != nil {
					log.Println("Parse cron.go ", record.ScheduleConf, " error")
					break
				}
				record.TriggerLastTime = record.TriggerNextTime
				record.TriggerNextTime = cron.NextTime(time.Now()).Unix() * 1000
				break

			case "FIX_RATE":
				scheduleIntervalTimeMs, err := strconv.Atoi(record.ScheduleConf)
				if err != nil {
					log.Println("Convert schedule conf ", record.ScheduleConf, err)
				}
				record.TriggerLastTime = record.TriggerNextTime
				record.TriggerNextTime = record.TriggerNextTime + int64(scheduleIntervalTimeMs)*1000
				break
			}

			record.UpdateTime = time.Now()
			rows, err := session.Exec("UPDATE xxl_job_info SET trigger_last_time = ?, trigger_next_time = ?, trigger_status = ? WHERE id = ? ", record.TriggerLastTime, record.TriggerNextTime, record.TriggerStatus, record.Id)
			log.Println("Update record ", record, " affect rows = ", rows, " err = ", err)
		}
		return len(records), err
	})
	return count.(int), err
}

func (receiver *SimpleScheduler) Stop() error {
	if receiver.registryMonitor != nil {
		_ = receiver.registryMonitor.Stop()
	}

	if receiver.scanTimer != nil {
		receiver.scanTimer.Stop()
	}

	log.Println("SimpleScheduler stop ")
	return nil
}
