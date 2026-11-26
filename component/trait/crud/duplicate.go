package crud

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// IsDuplicateEntryError 判断底层驱动错误是否为唯一索引冲突。
// 兜底场景：SaveBefore/SaveAfter 的显式查重与真正写入之间存在并发窗口，
// 此时数据库返回唯一索引错误，需要转译为用户可读的文案，避免透出原始 DB 报错。
//
// 兼容两类驱动：
//   - MySQL（go-sql-driver）：Error 1062 Duplicate entry
//   - SQLite（glebarez/sqlite 等）：错误串含 "UNIQUE constraint failed"
//
// 参数：
//   - err: GORM 返回的底层错误
//
// 返回值：
//   - bool: true 表示是唯一索引冲突错误
//
// 使用示例：
//
//	if IsDuplicateEntryError(err) {
//	    return errors.New("数据已存在（唯一字段冲突）")
//	}
func IsDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
