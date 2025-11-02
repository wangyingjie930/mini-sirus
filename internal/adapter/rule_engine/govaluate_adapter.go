package rule_engine

import (
	"context"
	"errors"
	"fmt"
	"mini-sirus/internal/domain/valueobject"

	"github.com/Knetic/govaluate"
)

// GovaluateAdapter 规则引擎适配器（基于 govaluate）
type GovaluateAdapter struct {
	functions map[string]govaluate.ExpressionFunction
}

// NewGovaluateAdapter 创建规则引擎适配器
func NewGovaluateAdapter() *GovaluateAdapter {
	return &GovaluateAdapter{
		functions: make(map[string]govaluate.ExpressionFunction),
	}
}

// Evaluate 执行表达式求值
func (a *GovaluateAdapter) Evaluate(
	ctx context.Context,
	expr string,
	functions map[string]govaluate.ExpressionFunction,
	args valueobject.ExpressionArguments,
) (bool, error) {
	if expr == "" {
		// 空表达式默认返回 true
		return true, nil
	}

	// 合并函数（优先使用传入的函数）
	mergedFunctions := make(map[string]govaluate.ExpressionFunction)
	for k, v := range a.functions {
		mergedFunctions[k] = v
	}
	for k, v := range functions {
		mergedFunctions[k] = v
	}

	// 创建表达式
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(expr, mergedFunctions)
	if err != nil {
		return false, fmt.Errorf("parse expression failed: %w", err)
	}

	// 执行求值
	result, err := expression.Evaluate(map[string]interface{}(args))
	if err != nil {
		return false, fmt.Errorf("evaluate expression failed: %w", err)
	}

	// 转换为布尔值
	reach, ok := result.(bool)
	if !ok {
		return false, errors.New("expression result must be bool")
	}

	return reach, nil
}

// RegisterFunction 注册自定义函数
func (a *GovaluateAdapter) RegisterFunction(name string, fn govaluate.ExpressionFunction) error {
	if name == "" {
		return errors.New("function name cannot be empty")
	}
	if fn == nil {
		return errors.New("function cannot be nil")
	}

	a.functions[name] = fn
	return nil
}

// GetRegisteredFunctions 获取所有注册的函数
func (a *GovaluateAdapter) GetRegisteredFunctions() map[string]govaluate.ExpressionFunction {
	// 返回副本，避免外部修改
	functions := make(map[string]govaluate.ExpressionFunction)
	for k, v := range a.functions {
		functions[k] = v
	}
	return functions
}

