package main

import (
	"context"
	"github.com/core-go/config"
	"github.com/core-go/core"
	"github.com/core-go/log/convert"
	mid "github.com/core-go/log/middleware"
	"github.com/core-go/log/strings"
	"github.com/core-go/log/zap"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"net/http"

	"go-service/internal/app"
)

func main() {
	var conf app.Config
	var ctx = context.Background()
	err := config.Load(&conf, "configs/config")
	if err != nil {
		panic(err)
	}
	conf.MiddleWare.Constants = convert.ToCamelCase(conf.MiddleWare.Constants)
	conf.MiddleWare.Map = convert.ToCamelCase(conf.MiddleWare.Map)
	r := mux.NewRouter()
	log.Initialize(conf.Log)

	r.Use(func(handler http.Handler) http.Handler {
		return mid.BuildContextWithMask(handler, MaskLog)
	})
	logger := mid.NewLogger()
	if log.IsInfoEnable() {
		r.Use(mid.Logger(conf.MiddleWare, log.InfoFields, logger))
	}
	r.Use(mid.Recover(log.PanicMsg))
	err = app.Route(r, ctx, conf)
	if err != nil {
		panic(err)
	}
	log.Info(ctx, core.ServerInfo(conf.Server))
	server := core.CreateServer(conf.Server, r)
	if err = server.ListenAndServe(); err != nil {
		log.Error(ctx, err.Error())
	}
}
func MaskLog(name, s string) string {
	if name == "mobileNo" {
		return strings.Mask(s, 2, 2, "x")
	} else {
		return strings.Mask(s, 0, 5, "x")
	}
}
