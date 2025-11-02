package router

import (
	"mini-sirus/internal/interface/http/handler"
	"net/http"
)

// Router 路由器
type Router struct {
	mux         *http.ServeMux
	taskHandler *handler.TaskHandler
}

// NewRouter 创建路由器
func NewRouter(taskHandler *handler.TaskHandler) *Router {
	router := &Router{
		mux:         http.NewServeMux(),
		taskHandler: taskHandler,
	}

	router.registerRoutes()
	return router
}

// registerRoutes 注册路由
func (r *Router) registerRoutes() {
	// 任务相关路由
	r.mux.HandleFunc("/api/v1/task/create", r.taskHandler.HandleCreateTask)
	r.mux.HandleFunc("/api/v1/task/query", r.taskHandler.HandleQueryTask)
	r.mux.HandleFunc("/api/v1/task/trigger", r.taskHandler.HandleTriggerTask)

	// 健康检查
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// GetMux 获取原生 ServeMux
func (r *Router) GetMux() *http.ServeMux {
	return r.mux
}

