package controller

import (
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"strconv"
	"unicode/utf8"
	"villcore.com/admin/config"
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
	})
}

func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", map[string]interface{}{
		"contextPath": config.ServerConfig.ServerContextPath,
		"page":        "login",
	})
}

func Login(c *gin.Context) {
	username, _ := c.GetPostForm("userName")
	password, _ := c.GetPostForm("password")
	if utf8.RuneCountInString(username) <= 0 || utf8.RuneCountInString(password) <= 0 {
		c.JSON(http.StatusInternalServerError, api.NewFailReturnT("login_param_empty"))
		return
	}

	token, err := service.DoLogin(username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.NewFailReturnT("login_param_empty"))
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

	c.HTML(http.StatusOK, "jobinfo.html", map[string]interface{}{
		"contextPath": config.ServerConfig.ServerContextPath,
		"page":        "jobinfo",
		"userRole":    1,
		"user":        user,
		"I18n":        config.I18n,
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

func getStringOrDefault(c *gin.Context, param string, defaultVal string) (string, error) {
	paramStr, contains := c.GetPostForm(param)
	if !contains {
		return defaultVal, nil
	}
	return paramStr, nil
}
