package executor

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

type ClientAdminApiImpl struct {
	hostUrl       string
	accessToken   string
	timeoutSecond int32
}

func (receiver *ClientAdminApiImpl) Callback(callbackParamSlice []api.HandleCallbackParam) api.ReturnT {
	return api.NewSuccessReturnT(nil)
}

func (receiver *ClientAdminApiImpl) Registry(registryParam api.RegistryParam) api.ReturnT {
	// ctx, cancel := context.WithTimeout(context.Background(), 3000)
	// defer cancel()

	ctx := context.Background()
	jsonBytes, err := json.Marshal(registryParam)
	if err != nil {
		return api.NewFailReturnT("Marshal registry param failed ")
	}

	request, err := http.NewRequest("POST", receiver.hostUrl+"/api/registry", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return api.NewFailReturnT("Create api/registry request failed ")
	}

	log.Printf("Request : %v", request)
	request.Header.Add(XxlJobAccessToken, receiver.accessToken)
	resp, err := httpClient.Do(request.WithContext(ctx))
	if err != nil {
		log.Printf("Do post api/registry request failed %v", err)
		return api.NewFailReturnT("Do post api/registry request failed ")
	}

	log.Printf("Do post api/registry request success ")
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return api.NewFailReturnT("Read api/registry response body failed ")
	}

	result := api.ReturnT{}
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return api.NewFailReturnT("Parse api/registry response body failed ")
	}
	return result
}

func (receiver *ClientAdminApiImpl) RegistryRemove(registryParam api.RegistryParam) api.ReturnT {
	ctx, cancel := context.WithTimeout(context.Background(), 3000)
	defer cancel()

	jsonBytes, err := json.Marshal(registryParam)
	if err != nil {
		return api.NewFailReturnT("Marshal registry param failed ")
	}

	request, err := http.NewRequest("POST", receiver.hostUrl+"api/registryRemove", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return api.NewFailReturnT("Create api/registryRemove request failed ")
	}

	request.Header.Add(XxlJobAccessToken, receiver.accessToken)
	resp, err := httpClient.Do(request.WithContext(ctx))
	if err != nil {
		return api.NewFailReturnT("Do post api/registryRemove request failed ")
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return api.NewFailReturnT("Read api/registryRemove response body failed ")
	}

	result := api.ReturnT{}
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return api.NewFailReturnT("Parse api/registryRemove response body failed ")
	}
	return result
}
