package api

type ReturnT struct {
	Code    int32       `json:"code"`
	Msg     string      `json:"msg"`
	Content interface{} `json:"content"`
}

func NewSuccessReturnT(content interface{}) ReturnT {
	return ReturnT{Code: 200, Content: content}
}

func NewFailReturnT(content interface{}) ReturnT {
	return ReturnT{Code: 500, Content: content}
}

type IdleBeatParam struct {
	JobId int32 `json:"jobId"`
}

type TriggerParam struct {
	JobId                 int32  `json:"jobId"`
	ExecutorHandler       string `json:"executorHandler"`
	ExecutorParams        string `json:"executorParams"`
	ExecutorBlockStrategy string `json:"executorBlockStrategy"`
	ExecutorTimeout       int64  `json:"executorTimeout"`

	LogId       int64 `json:"logId"`
	LogDateTime int64 `json:"logDateTime"`

	GlueType       string `json:"glueType"`
	GlueSource     string `json:"glueSource"`
	GlueUpdateTime int64  `json:"glueUpdateTime"`

	BroadcastIndex int32 `json:"broadcastIndex"`
	BroadcastTotal int32 `json:"broadcastTotal"`
}

type KillParam struct {
	JobId int32 `json:"jobId"`
}

type LogParam struct {
	LogId       int64 `json:"logId"`
	LogDateTime int64 `json:"logDateTime"`
	FromLineNum int32 `json:"fromLineNum"`
}

type HandleCallbackParam struct {
	LogId      int64  `json:"logId"`
	LogDateTim int64  `json:"logDateTim"`
	HandleCode int32  `json:"handleCode"`
	HandleMsg  string `json:"handleMsg"`
}

type RegistryParam struct {
	RegistryGroup string `json:"registryGroup"`
	RegistryKey   string `json:"registryKey"`
	RegistryValue string `json:"registryValue"`
}
