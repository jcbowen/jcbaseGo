# db.go

数据库维护函数

### 函数说明

```go
package mysql

type Column struct {
	Name      string `json:"name"`   // 字段名
	Rename    string `json:"rename"` // 修改前的字段名
	Type      string `json:"type"`
	Length    string `json:"length,omitempty"`
	Default   string `json:"default"`
	Null      bool   `json:"null"`
	Signed    bool   `json:"signed"`
	Increment bool   `json:"increment"`
	Position  string `json:"position"` // 指定新增到什么位置，如 AFTER `updated_at`;
}

type Index struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Fields []string `json:"fields"`
}

type Schema struct {
	TableName string             `json:"tablename"`
	Charset   string             `json:"charset"`
	Engine    string             `json:"engine"`
	Increment string             `json:"increment"`
	Fields    map[string]*Column `json:"fields"`
	Indexes   map[string]*Index  `json:"indexes"`
}

type TableFixSqlOpt struct {
	Table1           *Schema // 需要修复的数据表结构
	Table2           *Schema // 基准数据表结构
	Strict           bool    // 使用严格模式, 严格模式将会把表2完全变成表1的结构, 否则将只处理表2种大于表1的内容(多出的字段和索引)
	CompareTableName bool    // 是否比较数据表名称，如果名称不一致，将会根据基准表创建一张新的表
	TablePre         string  // 生成sql语句中的表前缀
	TablePreOld      string  // 原结构的表前缀
}

type CDDiffs struct {
	Charset   bool `json:"charset"`
	TableName bool `json:"tablename"`
	Engine    bool `json:"engine"`
}

type CDFields struct {
	Less    []string `json:"less"`
	Diff    []string `json:"diff"`
	Greater []string `json:"greater"`
}

type CDIndexes struct {
	Less    []string `json:"less"`
	Diff    []string `json:"diff"`
	Greater []string `json:"greater"`
}

type CompareDiffs struct {
	Diffs   *CDDiffs   `json:"diffs"`
	Fields  *CDFields  `json:"fields"`
	Indexes *CDIndexes `json:"indexes"`
}

type TableCreateSqlOpt struct {
	Table       *Schema
	TablePre    string // 生成sql语句中的表前缀
	TablePreOld string // 原结构的表前缀
}

// TableSchema 获得指定表的结构
func TableSchema(tableName string) (*Schema, error)

// TableFixSql 根据基准表生成修复差异的sql
func TableFixSql(opt TableFixSqlOpt) (sqls []string)

// SchemaCompare 比较两张表的结构差异
// table1 表结构
// table2 表结构 基准表
func SchemaCompare(table1 *Schema, table2 *Schema) *CompareDiffs

// TableCreateSql 根据数据表结构生成建表语句
func TableCreateSql(opt TableCreateSqlOpt) (sql string)

// TableSchemas 生成清空表内数据的sql语句
func TableSchemas(tableName string) (dump string)

// MakeInsertSql 获取某个表的insert语句
func MakeInsertSql(tableName string, start int, size int) (data string, result []map[string]interface{})

// BuildIndexSql 为数据表创建索引
func BuildIndexSql(index *Index) string

// BuildFieldSql 创建一个完整字段
func BuildFieldSql(field *Column) string
```