package jcbaseGo

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestUnmarshalYAML_DbStruct_LowercaseKeys 校验小写键在 v0.17.8 及之后继续生效。
// 这是 market 等下游项目的实际写法（bin/config.yaml 等三份模板全用小写键），
// v0.17.7 补驼峰 tag 后曾静默失效，本用例锁住兼容行为。
func TestUnmarshalYAML_DbStruct_LowercaseKeys(t *testing.T) {
	data := `
drivername: mysql
tableprefix: ""
parsetime: "True"
singulartable: true
host: 127.0.0.1
port: "3306"
dbname: market
`
	var cfg DbStruct
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if cfg.ParseTime != "True" {
		t.Errorf("ParseTime = %q, want \"True\"（小写键 parsetime 应兼容生效）", cfg.ParseTime)
	}
	if cfg.DriverName != "mysql" {
		t.Errorf("DriverName = %q, want \"mysql\"", cfg.DriverName)
	}
	if cfg.TablePrefix != "" {
		t.Errorf("TablePrefix = %q, want \"\"", cfg.TablePrefix)
	}
	if !cfg.SingularTable {
		t.Error("SingularTable = false, want true")
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != "3306" || cfg.Dbname != "market" {
		t.Errorf("无tag字段解析异常: host=%q port=%q dbname=%q", cfg.Host, cfg.Port, cfg.Dbname)
	}
}

// TestUnmarshalYAML_DbStruct_CamelKeys 校验驼峰规范键（v0.17.7 引入的写法）正常生效。
func TestUnmarshalYAML_DbStruct_CamelKeys(t *testing.T) {
	data := `
driverName: mysql
tablePrefix: ""
parseTime: "True"
singularTable: true
`
	var cfg DbStruct
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if cfg.ParseTime != "True" || cfg.DriverName != "mysql" || !cfg.SingularTable {
		t.Errorf("驼峰键解析异常: parseTime=%q driverName=%q singularTable=%v", cfg.ParseTime, cfg.DriverName, cfg.SingularTable)
	}
}

// TestUnmarshalYAML_DbStruct_ConflictingKeys 校验同一字段两种键名同现时报错，消除静默歧义。
func TestUnmarshalYAML_DbStruct_ConflictingKeys(t *testing.T) {
	data := `
parseTime: "True"
parsetime: "False"
`
	var cfg DbStruct
	err := yaml.Unmarshal([]byte(data), &cfg)
	if err == nil {
		t.Fatal("双键同现应返回错误，实际成功")
	}
	if !strings.Contains(err.Error(), "parseTime") || !strings.Contains(err.Error(), "parsetime") {
		t.Errorf("错误信息应包含冲突的两种键名, got: %v", err)
	}
}

// TestUnmarshalYAML_DbStruct_UnknownKeys 校验无法识别的键被忽略且不影响其他字段解析。
func TestUnmarshalYAML_DbStruct_UnknownKeys(t *testing.T) {
	data := `
parsetimex: "True"
parseTime: "False"
host: 127.0.0.1
`
	var cfg DbStruct
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		t.Fatalf("未知键不应导致解析失败: %v", err)
	}

	if cfg.ParseTime != "False" {
		t.Errorf("ParseTime = %q, want \"False\"（未知键 parsetimex 不应影响规范键）", cfg.ParseTime)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want \"127.0.0.1\"", cfg.Host)
	}
}

// TestUnmarshalYAML_DefaultConfigStruct 模拟 jcbaseGo 配置加载链路（DefaultConfigStruct 整体解析），
// 校验嵌套的 db 段小写键经归一化后正确生效，其余段不受影响。
func TestUnmarshalYAML_DefaultConfigStruct(t *testing.T) {
	data := `
db:
  drivername: mysql
  parsetime: "True"
  singulartable: true
redis:
  host: 127.0.0.1
  port: "6379"
project:
  name: market
`
	var cfg DefaultConfigStruct
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if cfg.Db.ParseTime != "True" {
		t.Errorf("Db.ParseTime = %q, want \"True\"", cfg.Db.ParseTime)
	}
	if cfg.Db.DriverName != "mysql" || !cfg.Db.SingularTable {
		t.Errorf("db 段解析异常: driverName=%q singularTable=%v", cfg.Db.DriverName, cfg.Db.SingularTable)
	}
	if cfg.Redis.Host != "127.0.0.1" || cfg.Redis.Port != "6379" {
		t.Errorf("redis 段解析异常: host=%q port=%q", cfg.Redis.Host, cfg.Redis.Port)
	}
	if cfg.Project.Name != "market" {
		t.Errorf("project 段解析异常: name=%q", cfg.Project.Name)
	}
}

// TestMarshalYAML_UsesCanonicalKeys 校验 Marshal 写回时使用规范键，
// 使 updateConfigFile 的启动重写天然完成「小写键 → 驼峰键」的一次性自动迁移。
func TestMarshalYAML_UsesCanonicalKeys(t *testing.T) {
	cfg := DbStruct{ParseTime: "True", DriverName: "mysql", TablePrefix: "t_", SingularTable: true}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}
	s := string(out)

	for _, want := range []string{"parseTime:", "driverName:", "tablePrefix:", "singularTable:"} {
		if !strings.Contains(s, want) {
			t.Errorf("写回输出应包含规范键 %q, got:\n%s", want, s)
		}
	}
	for _, avoid := range []string{"parsetime:", "drivername:", "tableprefix:", "singulartable:"} {
		if strings.Contains(s, avoid) {
			t.Errorf("写回输出不应包含小写键 %q, got:\n%s", avoid, s)
		}
	}
}

// TestUnmarshalYAML_SqlLiteStruct 校验 SqlLiteStruct（字段全部无 yaml tag，规范键即小写字段名）解析正常。
func TestUnmarshalYAML_SqlLiteStruct(t *testing.T) {
	data := `
dbFile: ./data/db/test.db
tablePrefix: t_
singularTable: false
`
	var cfg SqlLiteStruct
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if cfg.DbFile != "./data/db/test.db" || cfg.TablePrefix != "t_" || cfg.SingularTable {
		t.Errorf("SqlLiteStruct 解析异常: dbFile=%q tablePrefix=%q singularTable=%v", cfg.DbFile, cfg.TablePrefix, cfg.SingularTable)
	}
}

// TestUnmarshalYAML_NullNode 校验 db 段为空值（null 节点）时不 panic、保持零值。
func TestUnmarshalYAML_NullNode(t *testing.T) {
	var cfg DefaultConfigStruct
	if err := yaml.Unmarshal([]byte("db:\n"), &cfg); err != nil {
		t.Fatalf("null 节点不应导致解析失败: %v", err)
	}
	if cfg.Db != (DbStruct{}) {
		t.Errorf("null 节点应保持零值, got %+v", cfg.Db)
	}
}
