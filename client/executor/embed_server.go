package executor

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"villcore.com/common/api"
)

type EmbedServer interface {
	Start() error
	Stop() error
}

type EmbedHttpServer struct {
	config            *JobConfig
	clientExecutorBiz *ClientExecutorBizImpl
	httpServer        *http.Server
}

func NewHttpServer(config *JobConfig, clientExecutorBiz *ClientExecutorBizImpl) *EmbedHttpServer {
	return &EmbedHttpServer{
		config:            config,
		clientExecutorBiz: clientExecutorBiz,
	}
}

func (receiver *EmbedHttpServer) Start() error {

	httpServerMux := http.NewServeMux()
	for pattern, handler := range receiver.httpPatterns() {
		httpServerMux.Handle(pattern, handler)
	}

	receiver.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(int(receiver.config.Port)),
		Handler: httpServerMux,
	}
	go func() {
		if err := receiver.httpServer.ListenAndServe(); err != nil {
			return
		}
	}()
	return nil
}

func (receiver *EmbedHttpServer) Stop() error {
	if receiver.httpServer != nil {
		return receiver.httpServer.Close()
	}
	return nil
}

func (receiver *EmbedHttpServer) httpPatterns() map[string]http.Handler {
	return map[string]http.Handler{
		"/beat":     receiver.beatHttpHandler(),
		"/idleBeat": receiver.idleBeatHttpHandler(),
		"/run":      receiver.runHttpHandler(),
		"/kill":     receiver.killHttpHandler(),
		"/log":      receiver.logHttpHandler(),
	}
}

func (receiver *EmbedHttpServer) beatHttpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Recv request %v \n", req.RequestURI)
		if bytes, err := json.Marshal(api.NewSuccessReturnT(nil)); err == nil {
			log.Println("Write response ", string(bytes))
			_, _ = w.Write(bytes)
		} else {
			writeFailResponse(w, "Marshal error")
		}
	})
}

func (receiver *EmbedHttpServer) idleBeatHttpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Recv request %v \n", req.RequestURI)
		if bytes, err := json.Marshal(api.NewSuccessReturnT(nil)); err == nil {
			_, _ = w.Write(bytes)
		} else {
			writeFailResponse(w, "Marshal error")
		}
	})
}

func (receiver *EmbedHttpServer) runHttpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Recv request %v \n", req.RequestURI)
		bytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			writeFailResponse(w, "Request body invalid")
			return
		}

		var triggerParam api.TriggerParam
		err = json.Unmarshal(bytes, &triggerParam)
		if err != nil {
			log.Println("Parse trigger param error ", err)
			writeFailResponse(w, "Parse trigger param error ")
			return
		}

		receiver.clientExecutorBiz.Run(&triggerParam)
		if bytes, err := json.Marshal(api.NewSuccessReturnT(nil)); err == nil {
			log.Println("Write response ", string(bytes))
			_, _ = w.Write(bytes)
		} else {
			writeFailResponse(w, "Marshal error")
		}
	})
}

func (receiver *EmbedHttpServer) killHttpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Recv request %v \n", req.RequestURI)
		if bytes, err := json.Marshal(api.NewSuccessReturnT(nil)); err == nil {
			_, _ = w.Write(bytes)
		} else {
			writeFailResponse(w, "Marshal error")
		}
	})
}

func (receiver *EmbedHttpServer) logHttpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Recv request %v \n", req.RequestURI)
		bytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			writeFailResponse(w, "Request body invalid")
			return
		}

		var logParam api.LogParam
		err = json.Unmarshal(bytes, &logParam)
		if err != nil {
			log.Println("Parse log param error ", err)
			writeFailResponse(w, "Parse log param error ")
			return
		}
		log.Println("Recv log param ", logParam)
	})
}

func writeFailResponse(w http.ResponseWriter, msg string) {
	if bytes, err := json.Marshal(api.NewSuccessReturnT(msg)); err == nil {
		_, _ = w.Write(bytes)
	}
}
