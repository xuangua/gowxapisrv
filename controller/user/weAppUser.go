package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	// "github.com/kataras/iris"
	"github.com/shen100/golang123/config"
	"github.com/shen100/golang123/controller/common"
	"github.com/shen100/golang123/model"
	"github.com/shen100/golang123/utils"
)

// WeAppLogin 微信小程序登录
// func WeAppLogin(ctx *iris.Context) {
func WeAppLogin(ctx *gin.Context) {
	SendErrJSON := common.SendErrJSON
	// code := ctx.FormValue("code")
	code := ctx.Query("user") //code, _ := ctx.Get("user")
	if code == "" {
		SendErrJSON("code不能为空", ctx)
		return
	}
	appID := config.WeAppConfig.AppID
	secret := config.WeAppConfig.Secret
	CodeToSessURL := config.WeAppConfig.CodeToSessURL
	CodeToSessURL = strings.Replace(CodeToSessURL, "{appid}", appID, -1)
	CodeToSessURL = strings.Replace(CodeToSessURL, "{secret}", secret, -1)
	CodeToSessURL = strings.Replace(CodeToSessURL, "{code}", code, -1)

	resp, err := http.Get(CodeToSessURL)
	if err != nil {
		fmt.Println(err.Error())
		SendErrJSON("error", ctx)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		SendErrJSON("error", ctx)
		return
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println(err.Error())
		SendErrJSON("error", ctx)
		return
	}

	if _, ok := data["session_key"]; !ok {
		fmt.Println("session_key 不存在")
		fmt.Println(data)
		SendErrJSON("error", ctx)
		return
	}

	var openID string
	var sessionKey string
	openID = data["openid"].(string)
	sessionKey = data["session_key"].(string)

	session := sessions.Default(ctx)
	session.Set("weAppOpenID", openID)
	session.Set("weAppSessionKey", sessionKey)
	session.Save()
	fmt.Println("wx user access [weAppOpenID: %s, weAppSessionKey: %d ]has been saved to session\n", openID, sessionKey)

	resData := make(map[string]interface{})
	resData[config.ServerConfig.SessionID] = openID
	ctx.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  resData,
	})
}

// SetWeAppUserInfo 设置小程序用户加密信息
func SetWeAppUserInfo(ctx *gin.Context) {
	SendErrJSON := common.SendErrJSON
	type EncryptedUser struct {
		EncryptedData string `json:"encryptedData"`
		IV            string `json:"iv"`
	}

	var weAppUser EncryptedUser
	if err := ctx.ShouldBindWith(&weAppUser, binding.JSON); err != nil {
		SendErrJSON("invalid binding Json data", ctx)
		return
	}
	// if ctx.ReadJSON(&weAppUser) != nil {
	// 	SendErrJSON("参数错误", ctx)
	// 	return
	// }

	session := sessions.Default(ctx)
	sessionKey := session.Get("weAppSessionKey")
	sessionKeyStr, ok := sessionKey.(string)
	if ok {
		/* act on str */
		if sessionKeyStr == "" {
			SendErrJSON("session error", ctx)
			return
		}
	} else {
		/* not string */
		SendErrJSON("session error", ctx)
		return
	}

	userInfoStr, err := utils.DecodeWeAppUserInfo(weAppUser.EncryptedData, sessionKeyStr, weAppUser.IV)
	if err != nil {
		fmt.Println(err.Error())
		SendErrJSON("error", ctx)
		return
	}

	var wxAppUser model.WeAppUser
	if err := json.Unmarshal([]byte(userInfoStr), &wxAppUser); err != nil {
		SendErrJSON("error", ctx)
		return
	}

	session.Set("weAppUser", wxAppUser)
	resData := make(map[string]interface{})
	// resData[config.ServerConfig.SessionID] = session.ID()
	ctx.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  resData,
	})
	return
}

// YesterdayRegisterUser 昨日注册的用户数
func YesterdayRegisterUser(ctx *gin.Context) {
	var user model.WxUser
	count := user.YesterdayRegisterUser()
	resData := make(map[string]interface{})
	resData["count"] = count
	ctx.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  resData,
	})
}

// TodayRegisterUser 今日注册的用户数
func TodayRegisterUser(ctx *gin.Context) {
	var user model.WxUser
	count := user.TodayRegisterUser()
	resData := make(map[string]interface{})
	resData["count"] = count
	ctx.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  resData,
	})
}

// Latest30Day 近30天，每天注册的新用户数
// func Latest30Day(ctx *gin.Context) {
// 	var users model.UserPerDay
// 	result := users.Latest30Day()
// 	var data iris.Map
// 	if result == nil {
// 		data = iris.Map{
// 			"users": [0]int{},
// 		}
// 	} else {
// 		data = iris.Map{
// 			"users": result,
// 		}
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{
// 		"errNo": model.ErrorCode.SUCCESS,
// 		"msg":   "success",
// 		"data":  data,
// 	})
// }

// Analyze 用户分析
func Analyze(ctx *gin.Context) {
	var user model.WxUser
	now := time.Now()
	nowSec := now.Unix()              //秒
	yesterdaySec := nowSec - 24*60*60 //秒
	yesterday := time.Unix(yesterdaySec, 0)

	yesterdayCount := user.PurchaseUserByDate(yesterday)
	todayCount := user.PurchaseUserByDate(now)
	yesterdayRegisterCount := user.YesterdayRegisterUser()
	todayRegisterCount := user.TodayRegisterUser()

	data := make(map[string]interface{})
	data["todayNewUser"] = todayRegisterCount
	data["yesterdayNewUser"] = yesterdayRegisterCount
	data["todayPurchaseUser"] = todayCount
	data["yesterdayPurchaseUser"] = yesterdayCount
	// data := iris.Map{
	// 	"todayNewUser":          todayRegisterCount,
	// 	"yesterdayNewUser":      yesterdayRegisterCount,
	// 	"todayPurchaseUser":     todayCount,
	// 	"yesterdayPurchaseUser": yesterdayCount,
	// }

	ctx.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  data,
	})
}
