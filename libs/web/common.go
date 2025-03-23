package web

import (
	"log/slog"
	"sync"
)

// HealthResponse type
type HealthResponse struct {
	Status string `json:"status"`
}

type FrameworkOptions struct {
	ListenAddr     string
	OtelEnabled    bool
	LogLevelConfig *slog.LevelVar
}

type WebServer struct {
	MU               sync.Mutex
	Running          bool
	FrameworkOptions FrameworkOptions
}
