package main

import (
	"ddd_demo/internal/events"
	"ddd_demo/pkg/grpcx"
)

type App struct {
	consumers []events.Consumer
	server    *grpcx.Server
}
