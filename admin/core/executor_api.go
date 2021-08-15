package core

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"villcore.com/common/api"
)

func init() {
	log.SetOutput(os.Stdout)
}

const XxlJobAccessToken = "XXL-JOB-ACCESS-TOKEN"

var httpClient = &http.Client{}

type ServerExecutorApiImpl struct {
	accessToken   string
	timeoutSecond int32
}

func (receiver *ServerExecutorApiImpl) run(hostUrl string, triggerParam *api.TriggerParam) api.ReturnT {
	ctx := context.Background()
	jsonBytes, err := json.Marshal(triggerParam)
	if err != nil {
		return api.NewFailReturnT("Marshal run trigger param failed ")
	}

	request, err := http.NewRequest("POST", hostUrl+"run", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return api.NewFailReturnT("Create api/run trigger request failed ")
	}

	request.Header.Add(XxlJobAccessToken, receiver.accessToken)
	resp, err := httpClient.Do(request.WithContext(ctx))
	if err != nil {
		return api.NewFailReturnT("Do post api/run trigger request failed ")
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return api.NewFailReturnT("Read api/run trigger response body failed ")
	}

	result := api.ReturnT{}
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return api.NewFailReturnT("Parse api/run trigger response body failed ")
	}
	return result
}
