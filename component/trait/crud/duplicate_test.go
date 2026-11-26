package crud

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
)

// TestIsDuplicateEntryError 验证唯一索引冲突错误识别：
// MySQL 1062、SQLite UNIQUE 约束命中，其余错误与 nil 不命中
//
// 参数：
//   - t: 测试上下文
func TestIsDuplicateEntryError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil 错误", nil, false},
		{"MySQL 1062 唯一索引冲突", &mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'yanfubu' for key 'idx_oa_project_code'"}, true},
		{"MySQL 其他错误码", &mysql.MySQLError{Number: 1045, Message: "Access denied"}, false},
		{"SQLite UNIQUE 约束冲突", fmt.Errorf("constraint failed: UNIQUE constraint failed: oa_project.code (1299)"), true},
		{"普通业务错误", errors.New("项目编码不能为空"), false},
		{"包装后的 MySQL 1062", fmt.Errorf("创建失败: %w", &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}), true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsDuplicateEntryError(c.err); got != c.want {
				t.Fatalf("IsDuplicateEntryError(%v) = %v, 期望 %v", c.err, got, c.want)
			}
		})
	}
}
