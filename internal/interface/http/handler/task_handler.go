package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"mini-sirus/internal/usecase/dto"
	"mini-sirus/internal/usecase/port/input"
	"mini-sirus/internal/usecase/task"
	"net/http"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	triggerTaskUC *task.TriggerTaskUseCase
	createTaskUC  *task.CreateTaskUseCase
	queryTaskUC   *task.QueryTaskUseCase
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(
	triggerTaskUC *task.TriggerTaskUseCase,
	createTaskUC *task.CreateTaskUseCase,
	queryTaskUC *task.QueryTaskUseCase,
) *TaskHandler {
	return &TaskHandler{
		triggerTaskUC: triggerTaskUC,
		createTaskUC:  createTaskUC,
		queryTaskUC:   queryTaskUC,
	}
}

// 确保实现了接口
var _ input.TaskService = (*TaskServiceImpl)(nil)

// TaskServiceImpl 任务服务实现
type TaskServiceImpl struct {
	triggerTaskUC *task.TriggerTaskUseCase
	createTaskUC  *task.CreateTaskUseCase
	queryTaskUC   *task.QueryTaskUseCase
}

// NewTaskServiceImpl 创建任务服务实现
func NewTaskServiceImpl(
	triggerTaskUC *task.TriggerTaskUseCase,
	createTaskUC *task.CreateTaskUseCase,
	queryTaskUC *task.QueryTaskUseCase,
) *TaskServiceImpl {
	return &TaskServiceImpl{
		triggerTaskUC: triggerTaskUC,
		createTaskUC:  createTaskUC,
		queryTaskUC:   queryTaskUC,
	}
}

// TriggerTask 触发任务
func (s *TaskServiceImpl) TriggerTask(ctx context.Context, input dto.TriggerTaskInput) error {
	return s.triggerTaskUC.Execute(ctx, input)
}

// CreateTask 创建任务
func (s *TaskServiceImpl) CreateTask(ctx context.Context, input dto.CreateTaskInput) (*dto.TaskOutput, error) {
	return s.createTaskUC.Execute(ctx, input)
}

// QueryTask 查询任务
func (s *TaskServiceImpl) QueryTask(ctx context.Context, input dto.QueryTaskInput) (*dto.TaskOutput, error) {
	return s.queryTaskUC.Execute(ctx, input)
}

// QueryTasksByUser 查询用户的任务列表
func (s *TaskServiceImpl) QueryTasksByUser(ctx context.Context, userID int64) ([]*dto.TaskOutput, error) {
	return s.queryTaskUC.ExecuteList(ctx, userID)
}

// HTTP Handler methods

// HandleCreateTask 处理创建任务请求
func (h *TaskHandler) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input dto.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	output, err := h.createTaskUC.Execute(r.Context(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Create task failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": output,
	})
}

// HandleQueryTask 处理查询任务请求
func (h *TaskHandler) HandleQueryTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从查询参数获取 taskID
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	var taskID int64
	fmt.Sscanf(taskIDStr, "%d", &taskID)

	input := dto.QueryTaskInput{
		TaskID: taskID,
	}

	output, err := h.queryTaskUC.Execute(r.Context(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query task failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": output,
	})
}

// HandleTriggerTask 处理触发任务请求
func (h *TaskHandler) HandleTriggerTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input dto.TriggerTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := h.triggerTaskUC.Execute(r.Context(), input); err != nil {
		http.Error(w, fmt.Sprintf("Trigger task failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"msg":  "success",
	})
}

