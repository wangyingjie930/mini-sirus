package output

import (
	"context"
	"mini-sirus/internal/domain/valueobject"

	"github.com/Knetic/govaluate"
)

// RuleEngine 规则引擎输出端口
// 定义规则引擎的抽象接口，具体实现在 adapter 层
type RuleEngine interface {
	// Evaluate 执行表达式求值
	// expr: 表达式字符串
	// functions: 自定义函数集
	// args: 表达式参数
	// 返回: 求值结果（布尔值）和错误
	Evaluate(
		ctx context.Context,
		expr string,
		functions map[string]govaluate.ExpressionFunction,
		args valueobject.ExpressionArguments,
	) (bool, error)

	// RegisterFunction 注册自定义函数
	RegisterFunction(name string, fn govaluate.ExpressionFunction) error

	// GetRegisteredFunctions 获取所有注册的函数
	GetRegisteredFunctions() map[string]govaluate.ExpressionFunction
}

