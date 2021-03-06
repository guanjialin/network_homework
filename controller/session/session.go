package session

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"network/global/constant"
	"network/global/logger"
	"network/global/session"
	"network/model/department"
	"network/model/loginlog"
	"network/model/user"
	"network/util/password"
	"strings"
	"time"
)

// TODO 可以考虑添加验证码
// Account校验
// 密码校验
func Login(c *gin.Context) {
	loginInfo := struct {
		Account  string `binding:"required"`
		Password string `binding:"required"`
	}{}

	if err := c.BindJSON(&loginInfo); err != nil {
		logger.Logger().Debug(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "账号和密码不能为空！"})
		return
	}

	if session.IsLogin(c) {
		c.Status(http.StatusOK)
		return
	}

	u := &user.User{
		Account:  strings.TrimSpace(loginInfo.Account),
		Password: password.New(strings.TrimSpace(loginInfo.Password)),
	}

	ok, err := u.Login()
	if err != nil {
		logger.Logger().Warn("query login info error:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误！"})
		return
	}

	depart := department.Department{ID: u.Department, Type: u.Type}
	if u.Type != constant.TypeUserAdministrator {
		depart, err = depart.Info()
		if err != nil {
			logger.Logger().Warn("query department error:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	if err := session.Login(c, loginInfo.Account); err != nil {
		logger.Logger().Warn("add session error:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userBriefInfo := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Type int8   `json:"type,omitempty"`
	}{ID: u.ID, Name: u.Name, Type: depart.Type}

	// 记录用户ID，方便日志拦截器使用
	c.Set(constant.KeyUserID, u.ID)

	c.JSON(http.StatusOK, &userBriefInfo)
}

func Logout(c *gin.Context) {
	if err := session.Logout(c); err != nil {
		logger.Logger().Warn("delete session error:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func LoginLog(c *gin.Context) {
	page := struct {
		Page  int `form:"page" binding:"required,gt=0"`
		Limit int `form:"limit" binding:"required,max=200"`
	}{}

	if err := c.BindQuery(&page); err != nil {
		logger.Logger().Debug(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "分页参数有误！"})
		return
	}

	logs, count, err := loginlog.List((page.Page-1)*page.Limit, page.Limit)
	if err != nil {
		logger.Logger().Warn("query login log error:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	type logInfo struct {
		ID          int64     `json:"id"`
		UserAccount string    `json:"user_account"`
		UserName    string    `json:"user_name"`
		IP          string    `json:"ip"`
		CreatedAt   time.Time `json:"created"`
	}

	logInfos := struct {
		Count int       `json:"count"`
		Logs  []logInfo `json:"logs"`
	}{}
	logInfos.Count = count

	for _, l := range logs {
		logInfos.Logs = append(logInfos.Logs, logInfo{
			ID:          l.ID,
			UserAccount: l.UserAccount,
			UserName:    l.UserName,
			IP:          l.IP,
			CreatedAt:   l.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, logInfos)
}
