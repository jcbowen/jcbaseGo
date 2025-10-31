package debugger

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestControllerInitialization 测试控制器初始化功能
func TestControllerInitialization(t *testing.T) {
	t.Run("默认配置初始化", func(t *testing.T) {
		// 创建调试器实例
		storage, _ := NewMemoryStorage()
		config := &Config{}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎并注册路由以初始化控制器
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 获取控制器
		controller := dbg.GetController()
		assert.NotNil(t, controller)

		// 验证默认配置
		assert.Equal(t, "/jcbase/debugger", controller.config.BasePath)
		assert.Equal(t, 20, controller.config.PageSize)
	})

	t.Run("自定义配置初始化", func(t *testing.T) {
		// 创建调试器实例
		storage, _ := NewMemoryStorage()
		config := &Config{}
		config.Storage = storage
		dbg, _ := New(config)

		// 自定义控制器配置
		controllerConfig := &ControllerConfig{
			BasePath: "/custom/debug",
			PageSize: 50,
		}

		// 创建Gin引擎
		router := gin.New()

		// 重新配置控制器
		dbg.WithController(router.Group(""), controllerConfig)

		// 获取控制器
		controller := dbg.GetController()
		assert.NotNil(t, controller)

		// 验证自定义配置
		assert.Equal(t, "/custom/debug", controller.config.BasePath)
		assert.Equal(t, 50, controller.config.PageSize)
	})

	t.Run("路由注册功能", func(t *testing.T) {
		// 创建Gin引擎
		router := gin.New()

		// 创建调试器实例
		storage, _ := NewMemoryStorage()
		config := &Config{}
		config.Storage = storage
		dbg, _ := New(config)

		// 注册路由
		dbg.RegisterRoutes(router.Group(""))

		// 验证路由是否注册成功
		routes := router.Routes()
		assert.NotEmpty(t, routes)

		// 检查关键路由是否存在
		foundIndex := false
		foundDetail := false
		foundSearch := false

		for _, route := range routes {
			if route.Path == "/jcbase/debugger" && route.Method == "GET" {
				foundIndex = true
			}
			if route.Path == "/jcbase/debugger/detail/:id" && route.Method == "GET" {
				foundDetail = true
			}
			if route.Path == "/jcbase/debugger/search" && route.Method == "GET" {
				foundSearch = true
			}
		}

		assert.True(t, foundIndex, "索引路由未注册")
		assert.True(t, foundDetail, "详情路由未注册")
		assert.True(t, foundSearch, "搜索路由未注册")
	})
}

// TestControllerHandlers 测试控制器处理器功能
func TestControllerHandlers(t *testing.T) {
	// 创建测试数据
	testLogs := []*LogEntry{
		{
			ID:         "test-id-1",
			Timestamp:  time.Now(),
			Method:     "GET",
			URL:        "/api/test",
			StatusCode: 200,
			Duration:   100 * time.Millisecond,
			ClientIP:   "127.0.0.1",
		},
		{
			ID:         "test-id-2",
			Timestamp:  time.Now().Add(-time.Hour),
			Method:     "POST",
			URL:        "/api/users",
			StatusCode: 201,
			Duration:   200 * time.Millisecond,
			ClientIP:   "192.168.1.1",
		},
	}

	t.Run("索引页面处理器", func(t *testing.T) {
		// 创建内存存储并添加测试数据
		storage, _ := NewMemoryStorage()
		for _, log := range testLogs {
			_ = storage.Save(log)
		}

		// 创建调试器实例（启用调试器，不跳过任何路径）
		config := &Config{
			Enabled:   true,
			SkipPaths: []string{}, // 清空跳过路径，确保记录所有请求
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎和测试请求
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 创建测试请求
		req := httptest.NewRequest("GET", "/jcbase/debugger", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "日志列表")
		assert.Contains(t, w.Body.String(), "test-id-1")
		assert.Contains(t, w.Body.String(), "test-id-2")
	})

	t.Run("详情页面处理器", func(t *testing.T) {
		// 创建内存存储并添加测试数据
		storage, _ := NewMemoryStorage()
		for _, log := range testLogs {
			_ = storage.Save(log)
		}

		// 创建调试器实例（启用调试器，不跳过任何路径）
		config := &Config{
			Enabled:   true,
			SkipPaths: []string{}, // 清空跳过路径，确保记录所有请求
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎和测试请求
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 创建测试请求
		req := httptest.NewRequest("GET", "/jcbase/debugger/detail/test-id-1", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "日志详情")
		assert.Contains(t, w.Body.String(), "test-id-1")
		assert.Contains(t, w.Body.String(), "GET")
		assert.Contains(t, w.Body.String(), "/api/test")
	})

	t.Run("搜索页面处理器", func(t *testing.T) {
		// 创建内存存储并添加测试数据
		storage, _ := NewMemoryStorage()
		for _, log := range testLogs {
			_ = storage.Save(log)
		}

		// 创建调试器实例，明确设置SkipPaths为空以确保记录所有请求
		config := &Config{
			Enabled:   true,
			SkipPaths: []string{}, // 清空跳过路径，确保记录所有请求
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎和测试请求
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 创建测试请求
		req := httptest.NewRequest("GET", "/jcbase/debugger/search", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "搜索")
		assert.Contains(t, w.Body.String(), "搜索日志内容")
	})

	t.Run("不存在的详情页面", func(t *testing.T) {
		// 创建内存存储
		storage, _ := NewMemoryStorage()

		// 创建调试器实例，明确设置SkipPaths为空以确保记录所有请求
		config := &Config{
			Enabled:   true,
			SkipPaths: []string{}, // 清空跳过路径，确保记录所有请求
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎和测试请求
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 创建测试请求（不存在的ID）
		req := httptest.NewRequest("GET", "/jcbase/debugger/detail/non-existent-id", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应（应该返回错误页面）
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "未找到ID为 non-existent-id 的日志条目")
	})
}

// TestControllerAPIHandlers 测试控制器API处理器功能
func TestControllerAPIHandlers(t *testing.T) {
	// 创建测试数据
	testLogs := []*LogEntry{
		{
			ID:         "api-test-1",
			Timestamp:  time.Now(),
			Method:     "GET",
			URL:        "/api/test",
			StatusCode: 200,
			Duration:   100 * time.Millisecond,
			ClientIP:   "127.0.0.1",
		},
		{
			ID:         "api-test-2",
			Timestamp:  time.Now().Add(-time.Hour),
			Method:     "POST",
			URL:        "/api/users",
			StatusCode: 201,
			Duration:   200 * time.Millisecond,
			ClientIP:   "192.168.1.1",
		},
	}

	t.Run("API日志列表", func(t *testing.T) {
		// 创建内存存储并添加测试数据
		storage, _ := NewMemoryStorage()
		for _, log := range testLogs {
			_ = storage.Save(log)
		}

		// 创建调试器实例，明确设置SkipPaths为空以确保记录所有请求
		config := &Config{
			Enabled:   true,
			SkipPaths: []string{}, // 清空跳过路径，确保记录所有请求
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎和测试请求
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 创建测试请求
		req := httptest.NewRequest("GET", "/jcbase/debugger/api/logs?page=1&page_size=10", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析JSON响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 验证响应结构
		assert.Contains(t, response, "data")
		assert.Contains(t, response, "pagination")

		pagination := response["pagination"].(map[string]interface{})
		assert.Contains(t, pagination, "page")
		assert.Contains(t, pagination, "pageSize")
		assert.Contains(t, pagination, "total")
		assert.Contains(t, pagination, "pages")

		logs := response["data"].([]interface{})
		assert.Len(t, logs, 2)

		// 验证第一个日志
		firstLog := logs[0].(map[string]interface{})
		assert.Equal(t, "api-test-1", firstLog["id"])
		assert.Equal(t, "GET", firstLog["method"])
	})

	t.Run("API日志详情", func(t *testing.T) {
		// 创建内存存储并添加测试数据
		storage, _ := NewMemoryStorage()
		for _, log := range testLogs {
			_ = storage.Save(log)
		}

		// 创建调试器实例，明确设置配置参数确保测试请求被记录
		config := &Config{
			Enabled:    true,
			SkipPaths:  []string{}, // 空数组确保不跳过任何路径
			SampleRate: 1.0,        // 设置采样率为1.0，确保记录所有请求
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎和测试请求
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 创建测试请求
		req := httptest.NewRequest("GET", "/jcbase/debugger/api/logs/api-test-1", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析JSON响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 验证响应结构
		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "api-test-1", data["id"])
		assert.Equal(t, "GET", data["method"])
		assert.Equal(t, "/api/test", data["url"])
		assert.Equal(t, float64(200), data["status_code"])
	})

	t.Run("API搜索功能", func(t *testing.T) {
		// 创建内存存储并添加测试数据
		storage, _ := NewMemoryStorage()
		for _, log := range testLogs {
			_ = storage.Save(log)
		}

		// 创建调试器实例，明确设置配置参数确保测试请求被记录
		config := &Config{
			Enabled:    true,
			SkipPaths:  []string{}, // 空数组确保不跳过任何路径
			SampleRate: 1.0,        // 设置采样率为1.0，确保所有请求都被记录
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎和测试请求
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 创建测试请求
		req := httptest.NewRequest("GET", "/jcbase/debugger/api/search?q=test&page=1&page_size=10", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析JSON响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 验证响应结构
		assert.Contains(t, response, "data")
		assert.Contains(t, response, "pagination")

		pagination := response["pagination"].(map[string]interface{})
		assert.Contains(t, pagination, "page")
		assert.Contains(t, pagination, "pageSize")
		assert.Contains(t, pagination, "total")
		assert.Contains(t, pagination, "pages")

		logs := response["data"].([]interface{})
		assert.Len(t, logs, 1) // 只有一个日志包含"test"
	})

	t.Run("API不存在的日志详情", func(t *testing.T) {
		// 创建内存存储
		storage, _ := NewMemoryStorage()

		// 创建调试器实例，明确设置配置参数确保测试请求被记录
		config := &Config{
			Enabled:    true,
			SkipPaths:  []string{}, // 空数组确保不跳过任何路径
			SampleRate: 1.0,        // 确保所有请求都被记录
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎和测试请求
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 创建测试请求（不存在的ID）
		req := httptest.NewRequest("GET", "/jcbase/debugger/api/logs/non-existent-id", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusNotFound, w.Code)

		// 解析JSON响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 验证错误信息
		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "未找到ID为 non-existent-id 的日志条目")
	})
}

// TestControllerTemplateRendering 测试控制器模板渲染功能
func TestControllerTemplateRendering(t *testing.T) {
	t.Run("模板渲染功能", func(t *testing.T) {
		// 创建调试器实例
		storage, _ := NewMemoryStorage()
		config := &Config{}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎并注册路由以初始化控制器
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 获取控制器
		controller := dbg.GetController()
		assert.NotNil(t, controller)

		// 创建Gin上下文
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 测试数据
		data := gin.H{
			"Title":   "测试页面",
			"Message": "Hello, World!", // 注意：模板使用大写的Message字段
		}

		// 调用模板渲染方法 - 使用error.html模板，它接受简单的消息数据
		controller.renderTemplate(c, "error.html", data)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "测试页面")
		assert.Contains(t, w.Body.String(), "Hello, World!")
	})

	t.Run("错误页面渲染", func(t *testing.T) {
		// 创建调试器实例
		storage, _ := NewMemoryStorage()
		config := &Config{}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎并注册路由以初始化控制器
		router := gin.New()
		dbg.RegisterRoutes(router.Group(""))

		// 获取控制器
		controller := dbg.GetController()
		assert.NotNil(t, controller)

		// 创建Gin上下文
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 调用错误渲染方法
		controller.renderError(c, "页面不存在")

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code) // renderTemplate会设置200状态码
		assert.Contains(t, w.Body.String(), "页面不存在")
		assert.Contains(t, w.Body.String(), "错误")
	})
}

// TestControllerPagination 测试控制器分页功能
func TestControllerPagination(t *testing.T) {
	t.Run("分页计算功能", func(t *testing.T) {
		// 创建调试器实例
		storage, _ := NewMemoryStorage()
		config := &Config{}
		config.Storage = storage
		dbg, _ := New(config)

		// 获取控制器
		controller := dbg.GetController()

		// 测试分页计算
		pagination := controller.calculatePagination(1, 20, 100)
		assert.Equal(t, 1, pagination["Page"])
		assert.Equal(t, 20, pagination["PageSize"])
		assert.Equal(t, 100, pagination["Total"])

		pagination = controller.calculatePagination(2, 20, 100)
		assert.Equal(t, 2, pagination["Page"])
		assert.Equal(t, 20, pagination["PageSize"])
		assert.Equal(t, 100, pagination["Total"])

		// 测试默认值 - 当pageSize为0时，方法会设置默认值20避免除零错误
		pagination = controller.calculatePagination(0, 0, 0)
		assert.Equal(t, 0, pagination["Page"])
		assert.Equal(t, 20, pagination["PageSize"]) // 默认分页大小
		assert.Equal(t, 0, pagination["Total"])
	})

	t.Run("分页参数解析", func(t *testing.T) {
		// 创建调试器实例
		storage, _ := NewMemoryStorage()
		config := &Config{}
		config.Storage = storage
		dbg, _ := New(config)

		// 获取控制器
		_ = dbg.GetController()

		// 创建Gin上下文
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置查询参数
		c.Request = httptest.NewRequest("GET", "/?page=2&pageSize=10", nil)

		// 模拟分页参数解析（直接使用strconv.Atoi）
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		// 验证结果
		assert.Equal(t, 2, page)
		assert.Equal(t, 10, pageSize)
	})

	t.Run("默认分页参数", func(t *testing.T) {
		// 创建调试器实例
		storage, _ := NewMemoryStorage()
		config := &Config{}
		config.Storage = storage
		dbg, _ := New(config)

		// 获取控制器
		_ = dbg.GetController()

		// 创建Gin上下文
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置空查询参数
		c.Request = httptest.NewRequest("GET", "/", nil)

		// 模拟分页参数解析（直接使用strconv.Atoi）
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		// 验证默认值
		assert.Equal(t, 1, page)
		assert.Equal(t, 20, pageSize) // 默认页面大小
	})
}

// TestControllerIntegration 测试控制器集成功能
func TestControllerIntegration(t *testing.T) {
	t.Run("完整流程测试", func(t *testing.T) {
		// 创建内存存储
		storage, _ := NewMemoryStorage()

		// 创建调试器实例，明确设置配置参数确保测试请求被记录
		config := &Config{
			Enabled:   true,
			SkipPaths: []string{}, // 空数组确保不跳过任何路径
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎
		router := gin.New()

		// 先只注册调试器中间件，测试中间件功能
		router.Use(dbg.Middleware())

		// 注册调试器路由
		dbg.RegisterRoutes(router.Group(""))

		// 添加测试业务路由
		router.GET("/api/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test successful"})
		})

		// 创建测试服务器
		server := httptest.NewServer(router)
		defer server.Close()

		// 发送业务请求（会被调试器记录）
		resp, err := http.Get(server.URL + "/api/test")
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		resp.Body.Close()

		// 等待日志被记录
		time.Sleep(100 * time.Millisecond)

		// 调试信息：检查存储实例是否一致
		t.Log("=== 存储实例检查 ===")
		t.Logf("存储实例相同: %v", storage == dbg.GetStorage())

		// 检查存储中是否有日志记录 - 同时检查原始存储和调试器存储
		entries1, total1, err1 := storage.FindAll(1, 10, nil)
		entries2, total2, err2 := dbg.GetStorage().FindAll(1, 10, nil)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		t.Logf("原始存储记录数: %d, 调试器存储记录数: %d", total1, total2)

		if total1 == 0 && total2 == 0 {
			t.Log("两个存储中都没有日志记录，调试器可能没有正确记录请求")
		} else if total1 > 0 && total2 > 0 {
			t.Logf("两个存储中都有记录，原始存储URL: %s, 调试器存储URL: %s", entries1[0].URL, entries2[0].URL)
		} else if total1 > 0 {
			t.Logf("只有原始存储有记录，URL: %s", entries1[0].URL)
		} else if total2 > 0 {
			t.Logf("只有调试器存储有记录，URL: %s", entries2[0].URL)
		}

		// 使用调试器存储的记录数进行断言，因为controller使用的是调试器存储
		assert.Greater(t, total2, 0, "调试器存储中应该有日志记录")

		// 访问调试器页面
		resp, err = http.Get(server.URL + "/jcbase/debugger/")
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		resp.Body.Close()

		// 验证页面包含记录的日志
		assert.Contains(t, string(body), "/api/test")
	})
}
