package core

import (
	"container/list"
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
}

type TimeTask func()

type TimeWheel struct {
	name         string
	mutex        sync.Mutex
	slotNum      int32
	slotTaskList []*list.List
	startTimeMs  int64
	maxTimeRange time.Duration
	reaperTicker *time.Timer
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewTimeWheel(name string, slotNum int32) (*TimeWheel, error) {
	if len(name) <= 0 {
		return nil, errors.New("Empty name " + name)
	}

	timeRange := time.Duration(int64(1) * int64(time.Second) * int64(slotNum))
	if timeRange <= 0 {
		errorMsg := "SlotNum = " + strconv.FormatInt(int64(slotNum), 10) + " maybe invalid "
		return nil, errors.New(errorMsg)
	}

	array := make([]*list.List, slotNum)
	for i := 0; i < int(slotNum); i++ {
		array[i] = list.New()
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &TimeWheel{
		name:         name,
		slotNum:      slotNum,
		slotTaskList: array,
		maxTimeRange: timeRange,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

func (tw *TimeWheel) Start() {
	tw.startReaper()
	log.Println("Start time wheel ", tw.name)
}

func (tw *TimeWheel) AddTimeTask(task TimeTask, runTime time.Time) error {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	now := time.Now()
	offsetDuration := runTime.Sub(now)
	if offsetDuration <= 0 {
		go task()
		return nil
	}

	durationAfterNow := offsetDuration
	if durationAfterNow.Milliseconds() > tw.maxTimeRange.Milliseconds() {
		return errors.New("Task run duration " + strconv.FormatInt(durationAfterNow.Milliseconds(), 10) + " exceed max time wheel range " + strconv.FormatInt(tw.maxTimeRange.Milliseconds(), 10))
	}
	return tw.addTimeTask(task, runTime, durationAfterNow)
}

func (tw *TimeWheel) addTimeTask(task TimeTask, runtime time.Time, durationAfterNow time.Duration) error {

	slotIndex := (runtime.Unix()) % int64(tw.slotNum)
	taskList := tw.slotTaskList[slotIndex]
	taskList.PushBack(task)
	log.Printf("Add time task at slot %v with duration %v ms ", slotIndex, strconv.FormatInt(durationAfterNow.Milliseconds(), 10))
	return nil
}

func (tw *TimeWheel) startReaper() {
	// start reaper ticker
	tw.reaperTicker = time.NewTimer(time.Second)

	// start reaper goroutine
	ctx, _ := context.WithCancel(tw.ctx)
	go func() {
		readyTimeTask := make([]TimeTask, 1, 100)

		for {
			select {
			case _ = <-ctx.Done():
				break
			case nowTime := <-tw.reaperTicker.C:
				slotIndex := (nowTime.Unix()) % int64(tw.slotNum)
				taskList := tw.slotTaskList[slotIndex]
				//log.Printf("Reaper slot index %v, task list %v running at %v \n", slotIndex, taskList, nowTime)

				tw.mutex.Lock()
				for element := taskList.Front(); element != nil; {
					next := element.Next()
					val := element.Value
					timeTask := val.(TimeTask)
					readyTimeTask = append(readyTimeTask, timeTask)
					taskList.Remove(element)
					element = next
				}
				tw.mutex.Unlock()

				for seq, timeTask := range readyTimeTask {
					if timeTask != nil {
						go timeTask()
						log.Printf(">>>>>>>>>>>>Reaper slot index %v, task list %v running at %v \n", slotIndex, taskList, nowTime)
						readyTimeTask[seq] = nil
					}
				}
				readyTimeTask = readyTimeTask[0:0]

				currentTimeMillis := time.Now().UnixNano()
				offsetMillis := (time.Now().Unix()+1)*int64(time.Second) - currentTimeMillis
				tw.reaperTicker.Reset(time.Duration(offsetMillis))
			}
		}
	}()
	log.Println("Start reaper ")
}

func (tw *TimeWheel) Stop() {
	if tw.ctx != nil && tw.cancel != nil {
		tw.cancel()
	}
	log.Println("Stop time wheel ", tw.name)
}
