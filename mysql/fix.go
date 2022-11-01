package mysql

import (
	"errors"
	"github.com/jcbowen/jcbaseGo/helper"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ----- TableSchema,Begin -----/

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

// TableSchema 获得指定数据表的结构
func TableSchema(c Context, tableName string) (*Schema, error) {

	if !(len(tableName) > 0) {
		return nil, errors.New("数据表名称不能为空")
	}

	tableSchema := &Schema{}

	// ------ tableStatus ------/
	type tableStatus struct {
		Name          string    `gorm:"name,omitempty"`
		Engine        string    `gorm:"engine,omitempty"`          // InnoDB
		Version       string    `gorm:"version,omitempty"`         // 10
		RowFormat     string    `gorm:"row_format,omitempty"`      // Dynamic
		Rows          string    `gorm:"rows,omitempty"`            // 32
		AvgRowLength  string    `gorm:"avg_row_length,omitempty"`  // 4608
		DataLength    string    `gorm:"data_length,omitempty"`     // 147456
		MaxDataLength string    `gorm:"max_data_length,omitempty"` // 0
		IndexLength   string    `gorm:"index_length,omitempty"`    // 49152
		DataFree      string    `gorm:"data_free,omitempty"`       // 0
		AutoIncrement string    `gorm:"auto_increment,omitempty"`  // 42
		CreateTime    time.Time `gorm:"create_time"`
		UpdateTime    time.Time `gorm:"update_time"`
		CheckTime     time.Time `gorm:"check_time"`
		Collation     string    `gorm:"collation,omitempty"` // utf8_general_ci
		Checksum      string    `gorm:"checksum,omitempty"`
		CreateOptions string    `gorm:"create_options,omitempty"`
		Comment       string    `gorm:"comment,omitempty"`
	}

	var result tableStatus
	c.Db.Raw("SHOW TABLE STATUS LIKE '" + tableName + "'").Scan(&result)
	if !(len(result.Name) > 0) {
		return nil, errors.New("没有找到数据表：" + tableName)
	}
	tableSchema.TableName = result.Name
	tableSchema.Charset = result.Collation
	tableSchema.Engine = result.Engine
	tableSchema.Increment = result.AutoIncrement

	// ------ tableCOLUMNS ------/
	type tableField struct {
		Field      string `gorm:"field,omitempty"`      // username
		Type       string `gorm:"type,omitempty"`       // varchar(50)
		Collation  string `gorm:"collation,omitempty"`  // utf8mb4_general_ci
		Null       string `gorm:"null,omitempty"`       // NO
		Key        string `gorm:"key,omitempty"`        // UNI
		Default    string `gorm:"default,omitempty"`    // 游客
		Extra      string `gorm:"extra,omitempty"`      // auto_increment
		Privileges string `gorm:"privileges,omitempty"` // select,insert,update,references
		Comment    string `gorm:"comment,omitempty"`    // 用户名
	}

	var result2 []tableField
	c.Db.Raw("SHOW FULL COLUMNS FROM " + tableName).Scan(&result2)
	Columns := make(map[string]*Column)
	for _, value := range result2 {
		temp := &Column{}
		itemType := strings.Split(value.Type, " ")
		temp.Name = value.Field
		itemPieces := strings.Split(itemType[0], "(")
		temp.Type = itemPieces[0]
		if len(itemPieces) > 1 {
			temp.Length = strings.TrimRight(itemPieces[1], ")")
		} else {
			temp.Length = ""
		}
		temp.Default = value.Default
		if value.Null != "NO" {
			temp.Null = true
		} else {
			temp.Null = false
		}
		if !(len(itemType) > 1) {
			temp.Signed = true
		} else {
			temp.Signed = false
		}
		if value.Extra == "auto_increment" {
			temp.Increment = true
		} else {
			temp.Increment = false
		}

		Columns[temp.Name] = temp
	}
	tableSchema.Fields = Columns

	// ------ tableIndex ------/
	type tableIndex struct {
		Table        string `gorm:"table,omitempty"`         // b_user
		NonUnique    string `gorm:"non_unique,omitempty"`    // 0
		KeyName      string `gorm:"key_name,omitempty"`      // PRIMARY
		SeqInIndex   string `gorm:"seq_in_index,omitempty"`  // 1
		ColumnName   string `gorm:"column_name,omitempty"`   // id
		Collation    string `gorm:"collation,omitempty"`     // A
		Cardinality  string `gorm:"cardinality,omitempty"`   // 32
		SubPart      string `gorm:"sub_part,omitempty"`      //
		Packed       string `gorm:"packed,omitempty"`        //
		Null         string `gorm:"null,omitempty"`          //
		IndexType    string `gorm:"index_type,omitempty"`    // BTREE
		Comment      string `gorm:"comment,omitempty"`       //
		IndexComment string `gorm:"index_comment,omitempty"` //
	}

	var result3 []tableIndex
	c.Db.Raw("SHOW INDEX FROM " + tableName).Scan(&result3)
	IndexItems := make(map[string]*Index)
	for _, value := range result3 {
		item := &Index{}

		item.Name = value.KeyName
		if value.KeyName == "PRIMARY" {
			item.Type = "primary"
		} else if value.NonUnique == "0" {
			item.Type = "unique"
		} else {
			item.Type = "index"
		}
		var _fields []string
		item.Fields = append(_fields, value.ColumnName)

		IndexItems[item.Name] = item
	}
	tableSchema.Indexes = IndexItems

	return tableSchema, nil
}

// ----- TableSchema,End -----/

// ----- SchemaCompare,Begin -----/

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

// SchemaCompare 比较两张表的结构差异
func SchemaCompare(table1 *Schema, table2 *Schema) *CompareDiffs {
	compareDiffs := &CompareDiffs{}
	cdFields := &CDFields{}
	cdIndexes := &CDIndexes{}
	cdDiffs := &CDDiffs{}
	if table1.TableName != table2.TableName {
		cdDiffs.TableName = true
	}
	if table1.Charset != table2.Charset {
		cdDiffs.Charset = true
	}
	if table1.Engine != table2.Engine {
		cdDiffs.Engine = true
	}
	compareDiffs.Diffs = cdDiffs

	fields1 := columnKeys(table1.Fields)
	fields2 := columnKeys(table2.Fields)

	// 统计fields差集的不同
	dif := helper.SetArrStr(fields1).ArrayDiff(fields2)
	if len(dif) > 0 {
		cdFields.Greater = helper.SetArrStr(dif).ArrayValue()
	}
	dif = helper.SetArrStr(fields2).ArrayDiff(fields1)
	if len(dif) > 0 {
		cdFields.Less = helper.SetArrStr(dif).ArrayValue()
	}

	// 统计fields交集的不同
	dif = []string{}
	intersects := helper.SetArrStr(fields1).ArrayIntersect(fields2)
	var fType = []string{
		"int", "tinyint", "smallint", "bigint",
	}
	if len(intersects) > 0 {
		for _, field := range intersects {
			if helper.InArray(table2.Fields[field].Type, fType) {
				table2.Fields[field].Length = ""
				table1.Fields[field].Length = ""
			}

			var (
				table1Json string
				table2Json string
			)
			helper.JsonStruct(table1.Fields[field]).ToString(&table1Json)
			helper.JsonStruct(table2.Fields[field]).ToString(&table2Json)
			table1Map := helper.JsonStr2Map(table1Json)
			table2Map := helper.JsonStr2Map(table2Json)

			var isDif bool
			for k, v := range table1Map {
				if v != table2Map[k] {
					isDif = true
					break
				}
			}
			if isDif {
				dif = append(dif, field)
			}
		}
	}
	if len(dif) > 0 {
		cdFields.Diff = helper.SetArrStr(dif).ArrayValue()
	}
	compareDiffs.Fields = cdFields

	// 统计indexes差集的不同
	indexes1 := indexesKeys(table1.Indexes)
	indexes2 := indexesKeys(table2.Indexes)
	dif = helper.SetArrStr(indexes1).ArrayDiff(indexes2)
	if len(dif) > 0 {
		cdIndexes.Greater = helper.SetArrStr(dif).ArrayValue()
	}
	dif = helper.SetArrStr(indexes2).ArrayDiff(indexes1)
	if len(dif) > 0 {
		cdIndexes.Less = helper.SetArrStr(dif).ArrayValue()
	}
	// 统计indexes交集的不同
	dif = []string{}
	intersects = helper.SetArrStr(indexes1).ArrayIntersect(indexes2)
	if len(intersects) > 0 {
		for _, index := range intersects {
			var (
				table1Json string
				table2Json string
			)
			helper.JsonStruct(table1.Indexes[index]).ToString(&table1Json)
			helper.JsonStruct(table2.Indexes[index]).ToString(&table2Json)
			table1Map := helper.JsonStr2Map(table1Json)
			table2Map := helper.JsonStr2Map(table2Json)

			var isDif bool
			for k, v := range table1Map {
				switch v.(type) {
				case string:
					if v != table2Map[k] {
						isDif = true
						break
					}
				default:
					v1 := reflect.ValueOf(v)
					v2 := reflect.ValueOf(table2Map[k])
					v1 = v1.Index(0)
					v2 = v2.Index(0)
					if helper.ToString(v1) != helper.ToString(v2) {
						isDif = true
						break
					}
				}
			}
			if isDif {
				dif = append(dif, index)
			}
		}
	}
	if len(dif) > 0 {
		cdIndexes.Diff = helper.SetArrStr(dif).ArrayValue()
	}
	compareDiffs.Indexes = cdIndexes

	return compareDiffs
}

// ----- SchemaCompare,End -----/

// ----- TableFixSql,Begin -----/

type TableFixSqlOpt struct {
	Table1           *Schema // 需要修复的数据表结构
	Table2           *Schema // 基准数据表结构
	Strict           bool    // 使用严格模式, 严格模式将会把表2完全变成表1的结构, 否则将只处理表2种大于表1的内容(多出的字段和索引)
	CompareTableName bool    // 是否比较数据表名称，如果名称不一致，将会根据基准表创建一张新的表
	TablePre         string  // 生成sql语句中的表前缀
	TablePreOld      string  // 原结构的表前缀
}

// TableFixSql 根据基准表生成修复差异的sql
func TableFixSql(opt TableFixSqlOpt) (sqlItems []string) {
	var sql string
	if opt.Table1 == nil {
		sqlItems = append(sqlItems, TableCreateSql(TableCreateSqlOpt{
			Table:       opt.Table2,
			TablePre:    opt.TablePre,
			TablePreOld: opt.TablePreOld,
		}))
		return
	}

	// 获取差异结构
	diff := SchemaCompare(opt.Table1, opt.Table2)
	if opt.CompareTableName && diff.Diffs.TableName {
		sqlItems = append(sqlItems, TableCreateSql(TableCreateSqlOpt{
			Table:       opt.Table2,
			TablePre:    opt.TablePre,
			TablePreOld: opt.TablePreOld,
		}))
		return
	}
	if diff.Diffs.Engine {
		sql = "ALTER TABLE `" + opt.Table1.TableName + "` ENGINE = " + opt.Table2.Engine
		sqlItems = append(sqlItems, sql)
	}
	if diff.Diffs.Charset {
		pieces := strings.Split(opt.Table2.Charset, "_")
		charset := pieces[0]
		sql = "ALTER TABLE `" + opt.Table1.TableName + "` DEFAULT CHARSET = " + charset
		sqlItems = append(sqlItems, sql)
	}

	var isIncrement *Column

	// diff.Fields 的相关处理
	if len(diff.Fields.Less) > 0 {
		for _, fieldName := range diff.Fields.Less {
			field := opt.Table2.Fields[fieldName]
			piece := BuildFieldSql(field)
			if len(field.Rename) > 0 && opt.Table1.Fields[field.Rename] != nil {
				sql = "ALTER TABLE `" + opt.Table1.TableName + "` CHANGE `" + field.Rename + "` `" + field.Name + "` " + piece
				delete(opt.Table1.Fields, field.Rename)
			} else {
				pos := ""
				if len(field.Position) > 0 {
					pos = " " + field.Position
				}
				sql = "ALTER TABLE `" + opt.Table1.TableName + "` ADD `" + field.Name + "` " + piece + pos
			}
			var primary *Column
			if strings.Index(sql, "AUTO_INCREMENT") != -1 {
				isIncrement = field
				sqlN, _ := helper.StrReplace("AUTO_INCREMENT", "", sql, -1)
				sql = helper.ToString(sqlN)
				for _, f := range opt.Table1.Fields {
					if f.Increment {
						primary = f
					}
				}
				if primary != nil {
					piece = BuildFieldSql(primary)
					if len(piece) > 0 {
						p, _ := helper.StrReplace("AUTO_INCREMENT", "", piece, -1)
						piece = helper.ToString(p)
					}
					sql2 := "ALTER TABLE `" + opt.Table1.TableName + "` CHANGE `" + primary.Name + "` `" + primary.Name + "` " + piece
					sqlItems = append(sqlItems, sql2)
				}
			}
			sqlItems = append(sqlItems, sql)
		}
	}
	if len(diff.Fields.Diff) > 0 {
		for _, fieldName := range diff.Fields.Diff {
			field := opt.Table2.Fields[fieldName]
			piece := BuildFieldSql(field)
			if opt.Table1.Fields[fieldName] != nil {
				sql = "ALTER TABLE `" + opt.Table1.TableName + "` CHANGE `" + field.Name + "` `" + field.Name + "` " + piece
				sqlItems = append(sqlItems, sql)
			}
		}
	}
	if opt.Strict && len(diff.Fields.Greater) > 0 {
		for _, fieldName := range diff.Fields.Greater {
			if opt.Table1.Fields[fieldName] != nil {
				sql = "ALTER TABLE `" + opt.Table1.TableName + "` DROP `" + fieldName + "`"
				sqlItems = append(sqlItems, sql)
			}
		}
	}

	// diff.Indexes 的相关处理
	if len(diff.Indexes.Less) > 0 {
		for _, indexName := range diff.Indexes.Less {
			index := opt.Table2.Indexes[indexName]
			piece := BuildIndexSql(index)
			sql = "ALTER TABLE `" + opt.Table1.TableName + "` ADD " + piece
			sqlItems = append(sqlItems, sql)
		}
	}
	if len(diff.Indexes.Diff) > 0 {
		for _, indexName := range diff.Indexes.Diff {
			index := opt.Table2.Indexes[indexName]
			piece := BuildIndexSql(index)
			sql = "ALTER TABLE `" + opt.Table1.TableName + "` DROP "
			sql2 := ""
			if "PRIMARY" == indexName {
				sql2 = " PRIMARY KEY "
			} else {
				sql2 = "INDEX " + indexName
			}
			sql = sql + sql2 + ", ADD " + piece
			sqlItems = append(sqlItems, sql)
		}
	}
	if opt.Strict && len(diff.Indexes.Greater) > 0 {
		for _, indexName := range diff.Indexes.Greater {
			sql = "ALTER TABLE `" + opt.Table1.TableName + "` DROP `" + indexName + "`"
			sqlItems = append(sqlItems, sql)
		}
	}

	if isIncrement != nil {
		piece := BuildFieldSql(isIncrement)
		sql = "ALTER TABLE `" + opt.Table1.TableName + "` CHANGE `" + isIncrement.Name + "` `" + isIncrement.Name + "` " + piece
		sqlItems = append(sqlItems, sql)
	}

	return
}

// ----- TableFixSql,End -----/

// ----- TableCreateSql,Begin -----/

type TableCreateSqlOpt struct {
	Table       *Schema
	TablePre    string // 生成sql语句中的表前缀
	TablePreOld string // 原结构的表前缀
}

// TableCreateSql 根据数据表结构生成建表语句
func TableCreateSql(opt TableCreateSqlOpt) (sql string) {
	pieces := strings.Split(opt.Table.Charset, "_")
	charset := pieces[0]
	engine := opt.Table.Engine
	tableName := opt.Table.TableName
	if len(opt.TablePre) > 0 && len(opt.TablePreOld) > 0 && opt.TablePre != opt.TablePreOld {
		newTableName, _ := helper.StrReplace(opt.TablePreOld, opt.TablePre, tableName, -1)
		tableName = helper.ToString(newTableName)
	}
	sql = "CREATE TABLE IF NOT EXISTS `" + tableName + "` (\n"

	fKeys := columnKeys(opt.Table.Fields)
	sort.Strings(fKeys)
	for i := 0; i < len(fKeys); i++ {
		fKey := fKeys[i]
		value := opt.Table.Fields[fKey]
		piece := BuildFieldSql(value)
		sql = sql + "`" + value.Name + "`" + piece + ",\n"
	}

	iKeys := indexesKeys(opt.Table.Indexes)
	sort.Strings(iKeys)
	for i := 0; i < len(iKeys); i++ {
		iKey := iKeys[i]
		value := opt.Table.Indexes[iKey]
		fields := strings.Join(value.Fields, "`,`")
		if "index" == value.Type {
			sql = sql + "KEY `" + value.Name + "` (`" + fields + "`),\n"
		}
		if "unique" == value.Type {
			sql = sql + "UNIQUE KEY `" + value.Name + "` (`" + fields + "`),\n"
		}
		if "primary" == value.Type {
			sql = sql + "PRIMARY KEY (`" + fields + "`),\n"
		}
	}
	sql = strings.TrimRight(sql, " ")
	sql = strings.TrimRight(sql, ",")

	sql = sql + ") ENGINE=" + engine + " DEFAULT CHARSET=" + charset + ";\n\n"

	return
}

// ----- TableCreateSql,End -----/

// BuildIndexSql 为数据表创建索引
func BuildIndexSql(index *Index) string {
	var pieceBuilder strings.Builder
	fields := strings.Join(index.Fields, ",")
	if index.Type == "index" {
		pieceBuilder.WriteString(" INDEX `" + index.Name + "` (`" + fields + "`)")
	}
	if index.Type == "unique" {
		pieceBuilder.WriteString(" UNIQUE `" + index.Name + "` (`" + fields + "`)")
	}
	if index.Type == "primary" {
		pieceBuilder.WriteString(" PRIMARY KEY (`" + fields + "`)")
	}
	return pieceBuilder.String()
}

// BuildFieldSql 创建一个完整字段
func BuildFieldSql(field *Column) string {
	var (
		length    = ""
		signed    = ""
		null      = ""
		_default  = ""
		increment = ""
	)
	fieldLength, err := strconv.ParseInt(field.Length, 10, 64)
	if err != nil {
		fieldLength = 0
	}
	if fieldLength > 0 {
		length = "(" + field.Length + ")"
	}
	fieldType := strings.ToLower(field.Type)
	types := []string{"decimal", "float", "dobule"}
	if strings.Index(fieldType, "int") != -1 || helper.InArray(fieldType, types) {
		if !field.Signed {
			signed = " unsigned"
		}
	}
	if !field.Null {
		null = " NOT NULL"
	}
	if len(field.Default) > 0 {
		_default = " DEFAULT '" + field.Default + "'"
	}
	if field.Increment {
		increment = " AUTO_INCREMENT"
	}

	return field.Type + length + signed + null + _default + increment
}

// TableSchemas 生成清空表内数据的sql语句
func TableSchemas(c Context, tableName string) (dump string) {
	sql := "SHOW CREATE TABLE " + tableName
	var result map[string]interface{}
	c.Db.Raw(sql).Scan(&result)

	dump = "DROP TABLE IF EXISTS " + tableName + "; "
	dump = dump + helper.ToString(result["Create Table"])

	return
}

// MakeInsertSql 获取某个表的insert语句
func MakeInsertSql(c Context, tableName string, start int, size int) (data string, result []map[string]interface{}) {
	var (
		keyBuilder strings.Builder
		tmpBuilder strings.Builder
		keys       string
	)
	c.Db.Table(tableName).Limit(size).Offset(start).Find(&result)
	if len(result) > 0 {
		for i := 0; i < len(result); i++ {
			item := result[i]
			if len(keys) < 1 {
				keyBuilder.WriteString("(")
			}
			tmpBuilder.WriteString("(")

			for k, v := range item {
				arr1 := []string{
					"\\", "\\0", "\n", "\r", "'", "\"", "\x1a",
				}
				arr2 := []string{
					"\\\\", "\\\\0", "\\n", "\\r", "\\'", "\\\"", "\\Z",
				}
				value, err := helper.StrReplace(arr1, arr2, v, -1)
				var str string
				if err != nil {
					str = helper.ToString(v)
				} else {
					str = helper.ToString(value)
				}
				if len(keys) < 1 {
					keyBuilder.WriteString("`" + k + "`,")
				}
				tmpBuilder.WriteString("'" + str + "',")
			}
			if len(keys) < 1 {
				keys = keyBuilder.String()
				keys = strings.TrimRight(keys, ",")
			}
			tmpBuilder.WriteString("),")
		}

		tmp := tmpBuilder.String()
		a1 := []string{",)"}
		a2 := []string{")"}
		replace, _ := helper.StrReplace(a1, a2, tmp, -1)
		tmp = helper.ToString(replace)
		tmp = strings.TrimRight(tmp, ",")
		data = "INSERT INTO `" + tableName + "` " + keys + ") VALUES " + tmp + ";"
	}
	defer func() {
		sqlDB, err := c.Db.DB()
		if err != nil {
			panic(err)
		}
		err = sqlDB.Close()
		if err != nil {
			panic(err)
		}
	}()

	return
}

// ----- 其他私有方法 -----/
func columnKeys(fields map[string]*Column) []string {
	var resp []string
	if len(fields) == 0 {
		return resp
	}

	for k := range fields {
		resp = append(resp, k)
	}

	return resp
}

func indexesKeys(fields map[string]*Index) []string {
	var resp []string
	if len(fields) == 0 {
		return resp
	}

	for k := range fields {
		resp = append(resp, k)
	}

	return resp
}
