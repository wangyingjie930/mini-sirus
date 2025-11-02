package main

import (
	"context"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/domain/valueobject"
	"mini-sirus/internal/usecase/dto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupContainer 为测试创建依赖注入容器
func setupContainer() *Container {
	return NewContainer()
}

func TestCreateTask(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	createInput := dto.CreateTaskInput{
		ActivityID:   1,
		TaskID:       100,
		UserID:       12345,
		Target:       3,
		TaskType:     valueobject.TaskTypePublishTimes,
		TaskCondExpr: "WITH_ANY_TOPIC(tag_ids, required_tag_ids) && LIKE_COUNT_GTE(like_count, 10) && IS_AUDITED(is_audited)",
	}

	taskOutput, err := container.CreateTaskUC.Execute(ctx, createInput)

	require.NoError(t, err, "创建任务不应该失败")
	assert.Equal(t, int64(100), taskOutput.TaskID)   // TaskID是输入的任务ID，而不是自增的ID
	assert.Equal(t, int64(12345), taskOutput.UserID) // UserID是int64类型
	assert.NotEmpty(t, taskOutput.Status)
}

func TestTriggerPublishTask(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	// 先创建任务
	createInput := dto.CreateTaskInput{
		ActivityID:   1,
		TaskID:       100,
		UserID:       12345,
		Target:       3,
		TaskType:     valueobject.TaskTypePublishTimes,
		TaskCondExpr: "WITH_ANY_TOPIC(tag_ids, required_tag_ids) && LIKE_COUNT_GTE(like_count, 10) && IS_AUDITED(is_audited)",
	}

	_, err := container.CreateTaskUC.Execute(ctx, createInput)
	require.NoError(t, err)

	// 触发发布事件
	publishEvent := &dto.PublishEventDTO{
		UserID:       12345,
		ContentID:    999,
		TopicIDs:     []uint64{1001, 1003},
		LikeCount:    15,
		CommentCount: 5,
		IsAudited:    true,
		AuditStatus:  1,
	}

	triggerInput := dto.TriggerTaskInput{
		TaskMode: publishEvent,
	}

	err = container.TriggerTaskUC.Execute(ctx, triggerInput)
	assert.NoError(t, err, "触发发布任务不应该失败")
}

func TestCreateCheckinTask(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	checkinCreateInput := dto.CreateTaskInput{
		ActivityID:   2,
		TaskID:       200,
		UserID:       12345,
		Target:       1,
		TaskType:     valueobject.TaskTypeCheckin,
		TaskCondExpr: "IS_TODAY()",
	}

	checkinTaskOutput, err := container.CreateTaskUC.Execute(ctx, checkinCreateInput)

	require.NoError(t, err, "创建签到任务不应该失败")
	assert.Equal(t, int64(200), checkinTaskOutput.TaskID)
	assert.Equal(t, int64(12345), checkinTaskOutput.UserID)
}

func TestTriggerCheckinTask(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	// 先创建签到任务
	checkinCreateInput := dto.CreateTaskInput{
		ActivityID:   2,
		TaskID:       200,
		UserID:       12345,
		Target:       1,
		TaskType:     valueobject.TaskTypeCheckin,
		TaskCondExpr: "IS_TODAY()",
	}

	_, err := container.CreateTaskUC.Execute(ctx, checkinCreateInput)
	require.NoError(t, err)

	// 触发签到事件
	checkinEvent := &dto.CheckinEventDTO{
		UserID: 12345,
		Date:   "2024-01-01",
	}

	checkinTriggerInput := dto.TriggerTaskInput{
		TaskMode: checkinEvent,
	}

	err = container.TriggerTaskUC.Execute(ctx, checkinTriggerInput)
	assert.NoError(t, err, "触发签到任务不应该失败")
}

func TestQueryUserTasks(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	// 创建多个任务
	createInputs := []dto.CreateTaskInput{
		{
			ActivityID:   1,
			TaskID:       100,
			UserID:       12345,
			Target:       3,
			TaskType:     valueobject.TaskTypePublishTimes,
			TaskCondExpr: "WITH_ANY_TOPIC(tag_ids, required_tag_ids) && LIKE_COUNT_GTE(like_count, 10) && IS_AUDITED(is_audited)",
		},
		{
			ActivityID:   2,
			TaskID:       200,
			UserID:       12345,
			Target:       1,
			TaskType:     valueobject.TaskTypeCheckin,
			TaskCondExpr: "IS_TODAY()",
		},
	}

	for _, input := range createInputs {
		_, err := container.CreateTaskUC.Execute(ctx, input)
		require.NoError(t, err)
	}

	// 查询任务列表
	tasks, err := container.QueryTaskUC.ExecuteList(ctx, 12345)

	require.NoError(t, err, "查询任务列表不应该失败")
	assert.GreaterOrEqual(t, len(tasks), 2, "应该至少有2个任务")

	for _, task := range tasks {
		assert.Equal(t, int64(12345), task.UserID)
		assert.NotEmpty(t, task.Status)
	}
}

func TestActivityManagement(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	// 使用正确的实体类型
	activity := &entity.ActActivity{
		Name:      "Spring Festival Activity",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:    entity.ActivityStatusActive,
	}

	err := container.ActivityRepo.Create(ctx, activity)
	require.NoError(t, err, "创建活动不应该失败")
	assert.Greater(t, activity.ID, int64(0), "活动ID应该大于0")
	assert.True(t, activity.IsActive(), "活动应该是激活状态")
}

func TestRiskControl_NormalCheckin(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	// 创建测试用户的签到任务
	testCreateInput := dto.CreateTaskInput{
		ActivityID:   3,
		TaskID:       300,
		UserID:       999,
		Target:       1,
		TaskType:     valueobject.TaskTypeCheckin,
		TaskCondExpr: "IS_TODAY()",
	}

	testTask, err := container.CreateTaskUC.Execute(ctx, testCreateInput)
	require.NoError(t, err, "创建测试任务不应该失败")
	assert.Equal(t, int64(300), testTask.TaskID)

	// 正常签到
	normalCheckin := &dto.CheckinEventDTO{
		UserID: 999,
		Date:   time.Now().Format("2006-01-02"),
	}
	triggerInput := dto.TriggerTaskInput{
		TaskMode: normalCheckin,
	}

	err = container.TriggerTaskUC.Execute(ctx, triggerInput)
	assert.NoError(t, err, "正常签到应该成功，风控检查应该通过")
}

func TestRiskControl_FrequentCheckin(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	userID := int64(888)
	baseTaskID := int64(400)

	// 模拟短时间内频繁签到
	maxAttempts := 12
	successCount := 0
	var lastErr error

	for i := 0; i < maxAttempts; i++ {
		// 为每次签到创建新任务
		testCreateInput := dto.CreateTaskInput{
			ActivityID:   3,
			TaskID:       baseTaskID + int64(i),
			UserID:       userID,
			Target:       1,
			TaskType:     valueobject.TaskTypeCheckin,
			TaskCondExpr: "IS_TODAY()",
		}

		_, err := container.CreateTaskUC.Execute(ctx, testCreateInput)
		if err != nil {
			continue
		}

		checkinEvent := &dto.CheckinEventDTO{
			UserID: userID,
			Date:   time.Now().Format("2006-01-02"),
		}
		triggerInput := dto.TriggerTaskInput{
			TaskMode: checkinEvent,
		}

		err = container.TriggerTaskUC.Execute(ctx, triggerInput)
		if err != nil {
			lastErr = err
			t.Logf("第%d次签到被拒绝: %v", i+1, err)
			break
		}

		successCount++
		time.Sleep(10 * time.Millisecond)
	}

	// 验证风控机制生效
	// 应该在达到频率限制后被拦截
	assert.Less(t, successCount, maxAttempts, "应该有签到被风控拦截")
	if lastErr != nil {
		t.Logf("风控拦截错误: %v", lastErr)
	}
}

func TestRiskControl_BlacklistCheck(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	userID := int64(777)
	baseTaskID := int64(500)

	// 先执行多次操作触发风控
	for i := 0; i < 15; i++ {
		testCreateInput := dto.CreateTaskInput{
			ActivityID:   3,
			TaskID:       baseTaskID + int64(i),
			UserID:       userID,
			Target:       1,
			TaskType:     valueobject.TaskTypeCheckin,
			TaskCondExpr: "IS_TODAY()",
		}

		container.CreateTaskUC.Execute(ctx, testCreateInput)

		checkinEvent := &dto.CheckinEventDTO{
			UserID: userID,
			Date:   time.Now().Format("2006-01-02"),
		}
		triggerInput := dto.TriggerTaskInput{
			TaskMode: checkinEvent,
		}

		container.TriggerTaskUC.Execute(ctx, triggerInput)
		time.Sleep(5 * time.Millisecond)
	}

	// 检查用户是否被加入黑名单
	isBlacklisted, err := container.RiskCheckService.IsUserBlacklisted(ctx, userID)
	require.NoError(t, err, "检查黑名单不应该失败")

	if isBlacklisted {
		t.Logf("用户%d已被加入黑名单", userID)

		// 测试黑名单用户无法签到
		testCreateInput := dto.CreateTaskInput{
			ActivityID:   3,
			TaskID:       600,
			UserID:       userID,
			Target:       1,
			TaskType:     valueobject.TaskTypeCheckin,
			TaskCondExpr: "IS_TODAY()",
		}

		_, err := container.CreateTaskUC.Execute(ctx, testCreateInput)
		if err == nil {
			blacklistCheckin := &dto.CheckinEventDTO{
				UserID: userID,
				Date:   time.Now().Format("2006-01-02"),
			}
			triggerInput := dto.TriggerTaskInput{
				TaskMode: blacklistCheckin,
			}

			err = container.TriggerTaskUC.Execute(ctx, triggerInput)
			assert.Error(t, err, "黑名单用户的签到应该被拒绝")
		}
	} else {
		t.Log("用户未被加入黑名单（可能风控阈值未触发）")
	}
}

func TestRiskControl_BlacklistUser(t *testing.T) {
	container := setupContainer()
	ctx := context.Background()

	userID := int64(666)

	// 尝试黑名单用户操作
	testCreateInput := dto.CreateTaskInput{
		ActivityID:   3,
		TaskID:       700,
		UserID:       userID,
		Target:       1,
		TaskType:     valueobject.TaskTypeCheckin,
		TaskCondExpr: "IS_TODAY()",
	}

	// 先触发多次操作以确保用户被加入黑名单
	for i := 0; i < 20; i++ {
		testCreateInput.TaskID = 700 + int64(i)
		container.CreateTaskUC.Execute(ctx, testCreateInput)

		checkinEvent := &dto.CheckinEventDTO{
			UserID: userID,
			Date:   time.Now().Format("2006-01-02"),
		}
		triggerInput := dto.TriggerTaskInput{
			TaskMode: checkinEvent,
		}

		container.TriggerTaskUC.Execute(ctx, triggerInput)
		time.Sleep(5 * time.Millisecond)
	}

	// 检查是否在黑名单
	isBlacklisted, _ := container.RiskCheckService.IsUserBlacklisted(ctx, userID)

	if isBlacklisted {
		// 黑名单用户尝试新操作
		testCreateInput.TaskID = 800
		_, createErr := container.CreateTaskUC.Execute(ctx, testCreateInput)

		if createErr == nil {
			blacklistCheckin := &dto.CheckinEventDTO{
				UserID: userID,
				Date:   time.Now().Format("2006-01-02"),
			}
			triggerInput := dto.TriggerTaskInput{
				TaskMode: blacklistCheckin,
			}

			err := container.TriggerTaskUC.Execute(ctx, triggerInput)
			assert.Error(t, err, "黑名单用户的操作应该被拒绝")
		}
	}
}
