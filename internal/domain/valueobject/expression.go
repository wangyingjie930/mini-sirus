package valueobject

// Expression 表达式值对象
type Expression struct {
	value string
}

// NewExpression 创建表达式
func NewExpression(value string) Expression {
	return Expression{value: value}
}

// Value 获取表达式值
func (e Expression) Value() string {
	return e.value
}

// IsEmpty 判断表达式是否为空
func (e Expression) IsEmpty() bool {
	return e.value == ""
}

// Equals 比较两个表达式是否相等
func (e Expression) Equals(other Expression) bool {
	return e.value == other.value
}

// ExpressionArguments 表达式参数
type ExpressionArguments map[string]interface{}

// Get 获取参数值
func (e ExpressionArguments) Get(key string) (interface{}, bool) {
	val, exists := e[key]
	return val, exists
}

// Set 设置参数值
func (e ExpressionArguments) Set(key string, value interface{}) {
	e[key] = value
}

// Has 判断是否包含某个参数
func (e ExpressionArguments) Has(key string) bool {
	_, exists := e[key]
	return exists
}

