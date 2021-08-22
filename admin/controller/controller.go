package controller

import (
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"villcore.com/admin/config"
	"villcore.com/admin/core"
	"villcore.com/admin/service"
	"villcore.com/common/api"
	"villcore.com/common/model"
)

const TokenCookieName = "XXL_JOB_LOGIN_IDENTITY"

func Registry(c *gin.Context) {
	var registryParam api.RegistryParam
	if err := c.BindJSON(&registryParam); err != nil {
		log.Println("Request param invalid ", registryParam)
		c.JSON(http.StatusOK, api.NewFailReturnT("Request param invalid"))
	}

	log.Println("Registry ", registryParam)
	if err := service.RegisterExecutor(&registryParam); err != nil {
		c.JSON(http.StatusOK, api.NewFailReturnT(err.Error()))
	}
	c.JSON(http.StatusOK, api.NewSuccessReturnT(""))
}

func RegistryRemove(c *gin.Context) {
	c.JSON(http.StatusOK, api.NewSuccessReturnT(""))
}

func Index(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	dashboardInfoMap, _ := service.GetDashboardInfo()
	c.HTML(http.StatusOK, "index.html", map[string]interface{}{
		"contextPath":   config.ServerConfig.ServerContextPath,
		"page":          "index",
		"userRole":      user.Role,
		"user":          user,
		"dashboardInfo": dashboardInfoMap,
		"I18n":          config.I18n,
		"I18nJson":      config.I18nJson,
	})
}

func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", map[string]interface{}{
		"contextPath": config.ServerConfig.ServerContextPath,
		"page":        "login",
		"I18n":        config.I18n,
		"I18nJson":    config.I18nJson,
	})
}

func Login(c *gin.Context) {
	username, _ := c.GetPostForm("userName")
	password, _ := c.GetPostForm("password")
	if utf8.RuneCountInString(username) <= 0 || utf8.RuneCountInString(password) <= 0 {
		c.JSON(http.StatusOK, api.ReturnT{Code: 500, Msg: config.I18n["login_param_empty"], Content: nil})
		return
	}

	token, err := service.DoLogin(username, password)
	if err != nil {
		c.JSON(http.StatusOK, api.ReturnT{Code: 500, Msg: config.I18n["login_param_unvalid"], Content: nil})
		return
	}

	maxCookieAge := math.MaxInt32
	ifRemember, _ := c.GetPostForm("ifRemember")
	if ifRemember == "ok" {
		// add cookie
		maxCookieAge = math.MaxInt32
	}
	c.SetCookie(TokenCookieName, token, maxCookieAge, config.ServerConfig.ServerContextPath, "", false, true)
	c.JSON(http.StatusOK, api.NewSuccessReturnT(token))
}

func Logout(c *gin.Context) {
	c.SetCookie(TokenCookieName, "token", 0, config.ServerConfig.ServerContextPath, "", false, true)
	c.JSON(http.StatusOK, api.NewSuccessReturnT(""))
}

func JobInfo(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	var jobGroupList []model.JobGroup
	if user.Role == 1 {
		jobGroupList, _ = service.GetAllJobGroup()
	} else {
		records := make([]model.JobGroup, 0)
		if utf8.RuneCountInString(user.Permission) > 0 {
			for _, id := range strings.Split(user.Permission, ",") {
				jobGroupId, _ := strconv.ParseInt(id, 10, 64)
				jobGroup, _ := service.GetJobGroup(jobGroupId)
				records = append(records, *jobGroup)
			}
		}
	}

	jobGroupId, _ := getInt32OrDefault(c, "jobGroup", 0)
	c.HTML(http.StatusOK, "jobinfo.html", map[string]interface{}{
		"contextPath":  config.ServerConfig.ServerContextPath,
		"page":         "jobinfo",
		"userRole":     user.Role,
		"user":         user,
		"jobGroupList": jobGroupList,
		"jobGroup":     jobGroupId,
		"I18n":         config.I18n,
		"I18nJson":     config.I18nJson,
	})
}

func JobInfoPageList(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	start, _ := getInt32OrDefault(c, "start", 0)
	length, _ := getInt32OrDefault(c, "length", 10)
	jobGroup, _ := getInt32OrDefault(c, "jobGroup", 0)
	triggerStatus, _ := getInt32OrDefault(c, "triggerStatus", 0)
	jobDesc, _ := getStringOrDefault(c, "jobDesc", "")
	executorHandler, _ := getStringOrDefault(c, "executorHandler", "")
	author, _ := getStringOrDefault(c, "author", "")

	records, totalCount, err := service.GetJobInfoList(start, length, jobGroup, triggerStatus, jobDesc, executorHandler, author)
	if err != nil {
		log.Println("Get job info list error ", err)
		records = make([]model.JobInfo, 0)
		totalCount = 0
	}

	pageListInfo := map[string]interface{}{
		"recordsTotal":    totalCount,
		"recordsFiltered": totalCount,
		"data":            records,
	}

	c.JSON(http.StatusOK, pageListInfo)
}

func StartJobInfo(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	id, _ := getInt32OrDefault(c, "id", 0)
	_ = service.StartJobInfo(id)
	c.JSON(http.StatusOK, api.NewSuccessReturnT(nil))
}

func StopJobInfo(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	id, _ := getInt32OrDefault(c, "id", 0)
	_ = service.StopJobInfo(id)
	c.JSON(http.StatusOK, api.NewSuccessReturnT(nil))
}

func getInt32OrDefault(c *gin.Context, param string, defaultVal int32) (int32, error) {
	paramStr, contains := c.GetPostForm(param)
	if !contains {
		return defaultVal, nil
	}

	val, err := strconv.Atoi(paramStr)
	if err != nil {
		return defaultVal, err
	}
	return int32(val), nil
}

func getInt64OrDefault(c *gin.Context, param string, defaultVal int64) (int64, error) {
	paramStr, contains := c.GetPostForm(param)
	if !contains {
		return defaultVal, nil
	}

	val, err := strconv.Atoi(paramStr)
	if err != nil {
		return defaultVal, err
	}
	return int64(val), nil
}

func getStringOrDefault(c *gin.Context, param string, defaultVal string) (string, error) {
	paramStr, contains := c.GetPostForm(param)
	if !contains {
		return defaultVal, nil
	}
	return paramStr, nil
}

func JobGroup(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	c.HTML(http.StatusOK, "jobgroup.index.html", map[string]interface{}{
		"contextPath": config.ServerConfig.ServerContextPath,
		"page":        "jobgroup",
		"userRole":    user.Role,
		"user":        user,
		"I18n":        config.I18n,
		"I18nJson":    config.I18nJson,
	})
}

func JobGroupPageList(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	start, _ := getInt32OrDefault(c, "start", 0)
	length, _ := getInt32OrDefault(c, "length", 10)
	appname, _ := getStringOrDefault(c, "appname", "")
	title, _ := getStringOrDefault(c, "title", "")

	records, totalCount, err := service.GetJobGroupList(start, length, appname, title)
	if err != nil {
		log.Println("Get job group list error ", err)
		records = make([]model.JobGroup, 0)
		totalCount = 0
	}

	voRecords := make([]*JobGroupVO, 0)
	for _, record := range records {
		voRecords = append(voRecords, createJobGroupVO(&record))
	}
	pageListInfo := map[string]interface{}{
		"recordsTotal":    totalCount,
		"recordsFiltered": totalCount,
		"data":            voRecords,
	}
	c.JSON(http.StatusOK, pageListInfo)
}

func LoadJobGroupById(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	id, _ := getInt32OrDefault(c, "id", 0)
	jobGroup, err := service.GetJobGroup(int64(id))
	if err != nil {
		log.Println("Get job group list error ", err)
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(""))
	} else {
		c.JSON(http.StatusOK, api.NewSuccessReturnT(createJobGroupVO(jobGroup)))
	}
}

type JobGroupVO struct {
	Id           int64     `json:"id"`
	AppName      string    `json:"appname"`
	Title        string    `json:"title"`
	AddressType  int32     `json:"addressType"`
	RegistryList []string  `json:"registryList"`
	AddressList  string    `json:"addressList"`
	UpdateTime   time.Time `json:"updateTime"`
}

func createJobGroupVO(jobGroup *model.JobGroup) *JobGroupVO {
	return &JobGroupVO{
		Id:          jobGroup.Id,
		AppName:     jobGroup.AppName,
		Title:       jobGroup.Title,
		AddressType: jobGroup.AddressType,
		UpdateTime:  jobGroup.UpdateTime,
		RegistryList: func() []string {
			if utf8.RuneCountInString(jobGroup.AddressList) <= 0 {
				return make([]string, 0)
			}
			return strings.Split(jobGroup.AddressList, ",")
		}(),
	}
}

func User(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	c.HTML(http.StatusOK, "user.index.html", map[string]interface{}{
		"contextPath": config.ServerConfig.ServerContextPath,
		"page":        "jobgroup",
		"userRole":    user.Role,
		"user":        user,
		"I18n":        config.I18n,
		"I18nJson":    config.I18nJson,
	})
}

func UserPageList(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	start, _ := getInt32OrDefault(c, "start", 0)
	length, _ := getInt32OrDefault(c, "length", 10)
	username, _ := getStringOrDefault(c, "username", "")
	role, _ := getInt32OrDefault(c, "role", 0)

	records, totalCount, err := service.GetJobUserList(start, length, username, role)
	if err != nil {
		log.Println("Get job group list error ", err)
		records = make([]model.JobUser, 0)
		totalCount = 0
	}

	pageListInfo := map[string]interface{}{
		"recordsTotal":    totalCount,
		"recordsFiltered": totalCount,
		"data":            records,
	}
	c.JSON(http.StatusOK, pageListInfo)
}

func HelpPage(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	c.HTML(http.StatusOK, "help.html", map[string]interface{}{
		"contextPath": config.ServerConfig.ServerContextPath,
		"page":        "help",
		"userRole":    user.Role,
		"user":        user,
		"I18n":        config.I18n,
		"I18nJson":    config.I18nJson,
	})
}

func JobLogPage(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	doResp := func(jobGroupList []model.JobGroup, jobInfo *model.JobInfo) {
		c.HTML(http.StatusOK, "joblog.index.html", map[string]interface{}{
			"contextPath":  config.ServerConfig.ServerContextPath,
			"page":         "joblog",
			"userRole":     user.Role,
			"user":         user,
			"jobInfo":      jobInfo,
			"jobGroupList": jobGroupList,
			"I18n":         config.I18n,
			"I18nJson":     config.I18nJson,
		})
	}

	jobId, _ := getInt32OrDefault(c, "jobId", 0)
	jobInfo, _ := service.GetJobInfo(jobId)

	if user.Role == 1 {
		records, err := service.GetAllJobGroup()
		if err != nil {
			log.Println("Get all job group error ", err.Error())
		}
		doResp(records, jobInfo)
	} else {
		records := make([]model.JobGroup, 0)
		if utf8.RuneCountInString(user.Permission) > 0 {
			for _, id := range strings.Split(user.Permission, ",") {
				jobGroupId, _ := strconv.ParseInt(id, 10, 64)
				jobGroup, _ := service.GetJobGroup(jobGroupId)
				records = append(records, *jobGroup)
			}
		}
		doResp(records, jobInfo)
	}
}

func JobLogPageList(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	start, _ := getInt32OrDefault(c, "start", 0)
	length, _ := getInt32OrDefault(c, "length", 10)
	jobGroup, _ := getInt32OrDefault(c, "jobGroup", 0)
	jobId, _ := getInt32OrDefault(c, "jobId", 0)
	logStatus, _ := getInt32OrDefault(c, "logStatus", 0)
	filterTime, _ := getStringOrDefault(c, "filterTime", "")
	records, totalCount, err := service.GetJobLogList(start, length, jobGroup, jobId, logStatus, filterTime)
	if err != nil {
		log.Println("Get job group list error ", err)
		records = make([]model.JobLog, 0)
		totalCount = 0
	}

	doResp := func(data []JobLogVO) {
		pageListInfo := map[string]interface{}{
			"recordsTotal":    totalCount,
			"recordsFiltered": totalCount,
			"data":            data,
		}
		c.JSON(http.StatusOK, pageListInfo)
	}

	if user.Role != 1 {
		if utf8.RuneCountInString(user.Permission) > 0 {
			for _, id := range strings.Split(user.Permission, ",") {
				if !strings.Contains(id, strconv.FormatInt(int64(jobGroup), 10)) {
					doResp(make([]JobLogVO, 0))
					return
				}
			}
		} else {
			doResp(make([]JobLogVO, 0))
			return
		}
	}

	voRecords := make([]JobLogVO, 0)
	for _, r := range records {
		voRecords = append(voRecords, JobLogVO{
			r,
			func() string {
				if r.HandleTime.Unix() <= 0 {
					return ""
				} else {
					return r.HandleTime.Format("2006-01-02 15:04:05")
				}
			}(),
		})
	}
	doResp(voRecords)
}

type JobLogVO struct {
	model.JobLog
	HandleTimeStr string `json:"handleTime"`
}

func GetJobsByGroup(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	jobGroup, _ := getInt32OrDefault(c, "jobGroup", 0)
	records, err := service.GetJobsByGroup(jobGroup)
	if err != nil {
		log.Println("Get job group list error ", err)
	}
	c.JSON(http.StatusOK, api.NewSuccessReturnT(records))
}

func LogDetailPage(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	var jobLog *model.JobLog
	if id, _ := strconv.ParseInt(c.Query("id"), 10, 32); id > 0 {
		jobLog, _ = service.GetJobLog(int32(id))
	}

	if jobLog == nil {
		c.HTML(http.StatusOK, "common.exception.html", map[string]interface{}{
			"contextPath":  config.ServerConfig.ServerContextPath,
			"I18nJson":     config.I18nJson,
			"exceptionMsg": config.I18n["joblog_logid_unvalid"],
		})
		return
	}

	c.HTML(http.StatusOK, "joblog.detail.html", map[string]interface{}{
		"contextPath": config.ServerConfig.ServerContextPath,
		"userRole":    user.Role,
		"user":        user,
		"I18n":        config.I18n,
		"I18nJson":    config.I18nJson,

		"triggerCode":     jobLog.TriggerCode,
		"handleCode":      jobLog.HandleCode,
		"executorAddress": jobLog.ExecutorAddress,
		"triggerTime":     jobLog.TriggerTime,
		"logId":           jobLog.Id,
	})
}

func LogDetailCat(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	c.JSON(http.StatusOK, api.NewSuccessReturnT("NOT SUPPORTED"))
}

func LogKill(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	c.JSON(http.StatusOK, api.NewSuccessReturnT("NOT SUPPORTED"))
}

func ClearLog(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	jobGroup, _ := getInt64OrDefault(c, "jobGroup", 0)
	jobId, _ := getInt64OrDefault(c, "jobId", 0)
	clearType, _ := getInt64OrDefault(c, "type", 0)

	var cleanLogStartTime = time.Unix(0, 0)
	var cleanLogCount = 0

	switch clearType {
	case 1:
		cleanLogStartTime = time.Now().Add(time.Hour * 24 * 30 * -1)
		break
	case 2:
		cleanLogStartTime = time.Now().Add(time.Hour * 24 * 30 * -3)
		break
	case 3:
		cleanLogStartTime = time.Now().Add(time.Hour * 24 * 30 * -6)
		break
	case 4:
		cleanLogStartTime = time.Now().Add(time.Hour * 24 * 30 * -12)
		break
	case 5:
		cleanLogCount = 1000
		break
	case 6:
		cleanLogCount = 10000
		break
	case 7:
		cleanLogCount = 10000
		break
	case 8:
		cleanLogCount = 100000
		break
	case 9:
		cleanLogCount = 0
		break
	default:
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["joblog_clean_type_unvalid"]))
		return
	}

	if cleanLogStartTime.Unix() > 0 {
		_ = service.ClearJobLogByTime(jobGroup, jobId, cleanLogStartTime)
	}

	if cleanLogCount > 0 {
		_ = service.ClearJobLogByCount(jobGroup, jobId, int32(cleanLogCount))
	}
	c.JSON(http.StatusOK, api.NewSuccessReturnT(""))
}

func ChartInfo(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	startDateParam, _ := getStringOrDefault(c, "startDate", "")
	endDateParam, _ := getStringOrDefault(c, "endDate", "")
	if utf8.RuneCountInString(startDateParam) <= 0 || utf8.RuneCountInString(endDateParam) <= 0 {
		c.JSON(http.StatusOK, api.NewFailReturnT("Invalid time param "))
		return
	}

	timeFormat := "2006-01-02 15:04:05"
	startTime, err := time.ParseInLocation(timeFormat, strings.TrimSpace(startDateParam), time.Local)
	endTime, err := time.ParseInLocation(timeFormat, strings.TrimSpace(endDateParam), time.Local)
	if err != nil || startTime.After(endTime) {
		c.JSON(http.StatusOK, api.NewFailReturnT("Invalid time param "))
		return
	}

	records, _, err := service.GetJobReport(startTime, endTime)
	if err != nil || startTime.After(endTime) {
		c.JSON(http.StatusOK, api.NewFailReturnT("Invalid time param "))
		return
	}

	var triggerDayList []string
	var triggerDayCountRunningList []int32
	var triggerDayCountSucList []int32
	var triggerDayCountFailList []int32
	var triggerCountRunningTotal, triggerCountSucTotal, triggerCountFailTotal int32

	if len(records) > 0 {
		for _, record := range records {
			triggerDayList = append(triggerDayList, record.TriggerDay.Format("2006-01-02"))
			triggerDayCountRunningList = append(triggerDayCountRunningList, record.RunningCount)
			triggerDayCountSucList = append(triggerDayCountSucList, record.SucCount)
			triggerDayCountFailList = append(triggerDayCountFailList, record.FailCount)
			triggerCountRunningTotal = triggerCountRunningTotal + record.RunningCount
			triggerCountSucTotal = triggerCountSucTotal + record.SucCount
			triggerCountFailTotal = triggerCountFailTotal + record.FailCount
		}
	} else {
		for i := -6; i <= 0; i++ {
			triggerDayList = append(triggerDayList, time.Now().Add(-6*time.Hour*24).Format("2006-01-02"))
			triggerDayCountRunningList = append(triggerDayCountRunningList, 0)
			triggerDayCountSucList = append(triggerDayCountSucList, 0)
			triggerDayCountFailList = append(triggerDayCountFailList, 0)
		}
	}

	// START END
	c.JSON(http.StatusOK, api.NewSuccessReturnT(map[string]interface{}{
		"triggerDayList":             triggerDayList,
		"triggerDayCountRunningList": triggerDayCountRunningList,
		"triggerDayCountSucList":     triggerDayCountSucList,
		"triggerDayCountFailList":    triggerDayCountFailList,
		"triggerCountRunningTotal":   triggerCountRunningTotal,
		"triggerCountSucTotal":       triggerCountSucTotal,
		"triggerCountFailTotal":      triggerCountFailTotal,
	}))
}

func AddUser(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	loginUser, err := service.GetUserFromToken(token)
	if err != nil || loginUser == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	if loginUser.Role != 1 {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["user_update_loginuser_limit"]))
		return
	}

	user := &model.JobUser{}
	err = c.BindJSON(user)
	if err != nil {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["user_update_loginuser_limit"]))
		return
	}

	usernameStrLen := utf8.RuneCountInString(user.Username)
	if usernameStrLen <= 0 {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["system_please_input"]+config.I18n["user_username"]))
		return
	}
	user.Username = strings.TrimSpace(user.Username)

	if usernameStrLen < 4 || usernameStrLen > 20 {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["system_lengh_limit"]+"[4-20]"))
		return
	}

	passwordStrLen := utf8.RuneCountInString(user.Password)
	if passwordStrLen <= 0 {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["system_please_input"]+config.I18n["user_password"]))
		return
	}
	user.Password = strings.TrimSpace(user.Password)

	existUser, err := service.GetUserByUsername(user.Username)
	if existUser != nil {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["user_username_repeat"]))
		return
	}

	_ = service.SaveUser(user)
	c.JSON(http.StatusOK, api.NewSuccessReturnT(nil))
}

func UpdateUserPwd(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	password, _ := getStringOrDefault(c, "password", "")
	passwordCount := utf8.RuneCountInString(password)
	if passwordCount < 4 || passwordCount > 20 {
		c.JSON(http.StatusOK, api.NewFailReturnT(config.I18n["system_lengh_limit"]+"[4-20]"))
		return
	}

	_ = service.UpdateUserPassword(user.Id, password)
	c.JSON(http.StatusOK, api.NewSuccessReturnT("NOT SUPPORTED"))
}

func RemoveUser(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	id, _ := getInt32OrDefault(c, "id", 0)
	if user.Id == int64(id) || user.Role != 1 {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["user_update_loginuser_limit"]))
		return
	}

	_ = service.RemoveUser(id)
	c.JSON(http.StatusOK, api.NewSuccessReturnT(nil))
}

func UpdateUser(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	loginUser, err := service.GetUserFromToken(token)
	if err != nil || loginUser == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	if loginUser.Role != 1 {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["user_update_loginuser_limit"]))
		return
	}

	user := &model.JobUser{}
	err = c.BindJSON(user)
	if err != nil {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["user_update_loginuser_limit"]))
		return
	}

	if loginUser.Username == user.Username {
		c.JSON(http.StatusOK, api.NewFailReturnTWithMsg(config.I18n["user_update_loginuser_limit"]))
		return
	}

	passwordCount := utf8.RuneCountInString(user.Password)
	if passwordCount < 4 || passwordCount > 20 {
		c.JSON(http.StatusOK, api.NewFailReturnT(config.I18n["system_lengh_limit"]+"[4-20]"))
		return
	}

	_ = service.UpdateUser(user)
	c.JSON(http.StatusOK, api.NewSuccessReturnT(nil))
}

func TriggerJobInfo(c *gin.Context) {
	token, err := c.Cookie(TokenCookieName)
	if err != nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	user, err := service.GetUserFromToken(token)
	if err != nil || user == nil {
		c.Redirect(http.StatusFound, "login")
		return
	}

	id, _ := getInt32OrDefault(c, "id", 0)
	executorParam, _ := getStringOrDefault(c, "executorParam", "")
	addressList, _ := getStringOrDefault(c, "addressList", "")
	err = core.TriggerJob(&core.TriggerJobParam{
		JobId:                 id,
		TriggerType:           "MANUAL",
		FailRetryCount:        -1,
		ExecutorShardingParam: "",
		ExecutorParam:         executorParam,
		AddressList:           addressList,
	})

	if err != nil {
		log.Println("Get job info list error ", err)
	}
	c.JSON(http.StatusOK, api.NewSuccessReturnT(nil))
}
