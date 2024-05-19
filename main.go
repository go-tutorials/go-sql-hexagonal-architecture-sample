package main

import (
	"context"
	"github.com/core-go/config"
	"github.com/core-go/core"
	mid "github.com/core-go/log/middleware"
	"github.com/core-go/log/strings"
	"github.com/core-go/log/zap"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"go-service/internal/app"
)

func main() {
	var cfg app.Config
	err := config.Load(&cfg, "configs/config")
	if err != nil {
		panic(err)
	}
	r := mux.NewRouter()

	log.Initialize(cfg.Log)
	r.Use(mid.BuildContext)
	logger := mid.NewMaskLogger(MaskLog, MaskLog)
	if log.IsInfoEnable() {
		r.Use(mid.Logger(cfg.MiddleWare, log.InfoFields, logger))
	}
	r.Use(mid.Recover(log.PanicMsg))

	ctx := context.Background()
	err = app.Route(ctx, r, cfg)
	if err != nil {
		panic(err)
	}
	log.Info(ctx, core.ServerInfo(cfg.Server))
	server := core.CreateServer(cfg.Server, r)
	if err = server.ListenAndServe(); err != nil {
		log.Error(ctx, err.Error())
	}
}

func MaskLog(obj map[string]interface{}){
	v, ok := obj["phone"]
	if ok {
		s, ok2 := v.(string)
		if ok2 && len(s) > 3 {
			obj["phone"] = strings.Mask(s, 0, 3, "*")
		}
	}
}
