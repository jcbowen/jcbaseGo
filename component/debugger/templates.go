package debugger

// 内联HTML模板定义

// indexTemplate 主页模板
const indexTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; }
        .container { max-width: 1600px; margin: 0 auto; padding: 20px; }
        .header { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .header-content { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
        .header h1 { color: #2c3e50; margin: 0; font-size: 24px; word-break: break-word; }
        .header-actions { display: flex; gap: 10px; }
        .download-btn {
            display: inline-block;
            padding: 10px 20px;
            background-color: #4CAF50;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 600;
            transition: background-color 0.2s ease;
        }
        .download-btn:hover {
            background-color: #45a049;
        }
        .time-range-row { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; align-items: end; }
        .time-range-row input { width: 100%; }
        .header .stats { display: flex; gap: 20px; flex-wrap: wrap; }
        .stat-item { background: #f8f9fa; padding: 10px 15px; border-radius: 6px; border-left: 4px solid #3498db; }
        .stat-item .label { font-size: 12px; color: #666; }
        .stat-item .value { font-size: 18px; font-weight: bold; color: #2c3e50; }
        
        .filters { background: #fff; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .filter-header { 
            margin-bottom: 16px; 
            display: flex; 
            justify-content: space-between; 
            align-items: center; 
        }
        .filter-header .filter-actions { 
            display: flex; 
            gap: 12px; 
            align-items: center; 
            margin: 0; 
            padding: 0; 
            border-top: none; 
        }
        .filter-header .filter-actions button { 
            background: #3498db; 
            color: white; 
            border: none; 
            padding: 8px 16px; 
            border-radius: 6px; 
            cursor: pointer; 
            font-size: 14px; 
            font-weight: 600; 
            transition: background-color 0.2s ease; 
            grid-column: auto; 
            justify-self: auto; 
        }
        .filter-header .filter-actions button:hover { 
            background: #2980b9; 
        }
        .filter-header .filter-actions a { 
            color: #666; 
            text-decoration: none; 
            font-size: 14px; 
            padding: 6px 12px; 
            border: 1px solid #ddd; 
            border-radius: 6px; 
            transition: all 0.2s ease; 
        }
        .filter-header .filter-actions a:hover { 
            color: #3498db; 
            border-color: #3498db; 
            background: #f8f9fa; 
        }
        .filter-header h3 { 
            margin: 0; 
            font-size: 16px; 
            font-weight: 600; 
            color: #2c3e50; 
            display: flex; 
            align-items: center; 
            gap: 8px;
        }
        .filter-header h3::before { 
            content: "📊"; 
            font-size: 14px; 
        }
        .filter-form { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; align-items: end; }
        .filter-group { 
            display: flex; 
            flex-direction: column; 
            gap: 8px; 
            padding: 16px; 
            background: #f8f9fa; 
            border-radius: 6px; 
            border: 1px solid #e9ecef;
        }
        .filter-group h4 { 
            margin: 0 0 8px 0; 
            font-size: 14px; 
            font-weight: 600; 
            color: #495057; 
            display: flex; 
            align-items: center; 
            gap: 6px;
        }
        .filter-group h4::before { 
            content: "🔍"; 
            font-size: 12px; 
        }
        .filter-group .filter-row { 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); 
            gap: 8px; 
            align-items: end;
        }
        .filter-form input, .filter-form select { 
            padding: 10px 12px; 
            border: 1px solid #ddd; 
            border-radius: 6px; 
            font-size: 14px; 
            transition: border-color 0.2s ease;
            background: #fff;
        }
        .filter-form input:focus, .filter-form select:focus { 
            outline: none; 
            border-color: #3498db; 
            box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
        }
        .filter-form button { 
            background: #3498db; 
            color: white; 
            border: none; 
            padding: 10px 20px; 
            border-radius: 6px; 
            cursor: pointer; 
            font-size: 14px;
            font-weight: 600;
            transition: background-color 0.2s ease;
            grid-column: 1 / -1;
            justify-self: start;
        }
        .filter-form button:hover { background: #2980b9; }
        .filter-actions { 
            display: flex; 
            gap: 12px; 
            align-items: center; 
            margin-top: 16px; 
            padding-top: 16px; 
            border-top: 1px solid #eee;
        }
        .filter-actions a { 
            color: #666; 
            text-decoration: none; 
            font-size: 14px; 
            padding: 8px 16px;
            border: 1px solid #ddd;
            border-radius: 6px;
            transition: all 0.2s ease;
        }
        .filter-actions a:hover { 
            color: #3498db; 
            border-color: #3498db;
            background: #f8f9fa;
        }
        
        .logs-table { background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .table-container { overflow-x: auto; -webkit-overflow-scrolling: touch; }
        .table-content { min-width: 800px; }
        .table-header { background: #f8f9fa; padding: 15px; border-bottom: 1px solid #eee; display: grid; grid-template-columns: minmax(140px, 220px) 160px 100px 120px 70px 100px minmax(110px, 150px) minmax(200px, 1fr); gap: 16px; font-weight: bold; font-size: 14px; }
        .log-row { padding: 15px; border-bottom: 1px solid #eee; display: grid; grid-template-columns: minmax(140px, 220px) 160px 100px 120px 70px 100px minmax(110px, 150px) minmax(200px, 1fr); gap: 16px; align-items: center; font-size: 14px; }
        .log-row:hover { background: #f8f9fa; }
        .log-row:last-child { border-bottom: none; }
        .no-data-row { 
            grid-column: 1 / -1; 
            text-align: center; 
            padding: 40px; 
            color: #666; 
            font-size: 16px; 
            background: #f8f9fa;
            border-radius: 4px;
            margin: 10px;
        }
        .request-id a { color: #3498db; text-decoration: none; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 100%; display: block; }
        .request-id a:hover { text-decoration: underline; }
        .url { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 100%; }
        .http-method { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; }
        .http-method.method-get { background: #d4edda; color: #155724; }
        .http-method.method-post { background: #d1ecf1; color: #0c5460; }
        .http-method.method-put { background: #fff3cd; color: #856404; }
        .http-method.method-delete { background: #f8d7da; color: #721c24; }
        .status-code { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; }
        .status-2xx { background: #d4edda; color: #155724; }
        .status-3xx { background: #fff3cd; color: #856404; }
        .status-4xx { background: #f8d7da; color: #721c24; }
        .status-5xx { background: #f5c6cb; color: #721c24; }
        
        /* 进程记录样式 */
        .process-status { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; }
        .process-status-running { background: #fff3cd; color: #856404; }
        .process-status-completed { background: #d4edda; color: #155724; }
        .process-status-failed { background: #f5c6cb; color: #721c24; }
        .process-status-cancelled { background: #f8d7da; color: #721c24; }
        
        /* 流式请求样式 */
        .streaming-badge { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; display: inline-block; width: fit-content; }
        .streaming-active { background: #e8f4fd; color: #1976d2; }
        .streaming-inactive { background: #f8f9fa; color: #666; }
        
        .process-details { display: flex; flex-direction: column; gap: 4px; }
        .process-name { font-weight: 600; color: #2c3e50; }
        .process-type { color: #666; background: #f8f9fa; padding: 2px 6px; border-radius: 3px; }
        
        .http-details { display: flex; flex-direction: column; gap: 4px; }
        .http-method-detail { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; display: inline-block; width: fit-content; }
        .url { color: #666; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .client-ip { color: #666; background: #f8f9fa; padding: 2px 6px; border-radius: 3px; }
        
        .duration { color: #666; font-size: 12px; }
        .request-id a { color: #3498db; text-decoration: none; font-weight: 600; }
        .request-id a:hover { text-decoration: underline; }
        .actions a { color: #3498db; text-decoration: none; margin-right: 10px; }
        .actions a:hover { text-decoration: underline; }
        
        .pagination { display: flex; justify-content: center; gap: 4px; margin-top: 20px; flex-wrap: wrap; align-items: center; }
        .pagination a, .pagination span { padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; text-decoration: none; color: #333; min-width: 40px; text-align: center; font-size: 14px; line-height: 1.2; }
        .pagination a:hover { background: #f8f9fa; }
        .pagination .current { background: #3498db; color: white; border-color: #3498db; }
        .pagination .disabled { color: #999; cursor: not-allowed; background: #f5f5f5; }
        .pagination .ellipsis { padding: 8px 6px; color: #999; min-width: auto; }
        .pagination .page-nav { font-weight: 600; }
        .pagination .page-number { transition: all 0.2s ease; }
        
        @media (max-width: 768px) {
            .pagination { gap: 3px; }
            .pagination a, .pagination span { padding: 6px 8px; min-width: 36px; font-size: 14px; }
            .pagination .ellipsis { padding: 6px 4px; }
        }
        

        
        .nav { display: flex; gap: 20px; margin-bottom: 20px; }
        .nav a { color: #3498db; text-decoration: none; padding: 10px 15px; border-radius: 4px; }
        .nav a.active { background: #3498db; color: white; }
        
        /* 记录类型标识 */
        .record-type-badge { padding: 4px 10px; border-radius: 4px; font-size: 12px; font-weight: bold; margin-left: 10px; vertical-align: middle; }
        .record-type-http { background: #d1ecf1; color: #0c5460; }
        .record-type-process { background: #fff3cd; color: #856404; }
        
        /* 列表页记录类型标签 */
        .type-badge-http { background: #f3e5f5; color: #7b1fa2; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .type-badge-process { background: #e8f4fd; color: #1976d2; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        
        @media (max-width: 768px) {
            .container { padding: 10px; }
            .header { padding: 15px; margin-bottom: 15px; }
            .header h1 { font-size: 20px; }
            .header .stats { flex-direction: column; gap: 10px; }
            .stat-item { padding: 8px 12px; }
            .stat-item .value { font-size: 16px; }
            
            .filter-header { padding: 12px 16px; }
            .filter-header h3 { font-size: 14px; }
            .filter-content { padding: 16px; }
            .filter-form { grid-template-columns: 1fr; gap: 12px; }
            .filter-group { padding: 12px; }
            .filter-group h4 { font-size: 13px; }
            .filter-row { grid-template-columns: 1fr; gap: 8px; }
            .filter-form input, .filter-form select { width: 100%; }
            .filter-actions { flex-direction: column; gap: 8px; }
            .filter-actions button, .filter-actions a { width: 100%; text-align: center; }
            
            .table-container { overflow-x: auto; -webkit-overflow-scrolling: touch; }
            .table-content { min-width: 600px; }
            .table-header, .log-row { 
                grid-template-columns: minmax(80px, 120px) 100px 60px 80px 90px 60px minmax(100px, 1fr) minmax(100px, 1fr); 
                gap: 6px; 
                font-size: 11px; 
            }
            .log-row { padding: 10px; }
            .method, .status-code { font-size: 10px; padding: 3px 6px; }
            .duration { font-size: 11px; }
            .actions a { margin-right: 5px; font-size: 12px; }
            
            .pagination { gap: 3px; }
            .pagination a, .pagination span { 
                padding: 6px 8px; 
                min-width: 32px; 
                font-size: 12px; 
            }
            .pagination .ellipsis { padding: 6px 4px; }
        }
        
        @media (max-width: 480px) {
            .table-content { min-width: 500px; }
            .table-header, .log-row { 
                grid-template-columns: minmax(70px, 100px) 80px 50px 70px 80px 50px minmax(80px, 1fr) minmax(80px, 1fr); 
                gap: 4px; 
            }
            .log-row { padding: 8px; }
            .header { padding: 12px; }
            .header h1 { font-size: 18px; }
            .stat-item .value { font-size: 14px; }
            .pagination a, .pagination span { 
                padding: 4px 6px; 
                min-width: 28px; 
                font-size: 11px; 
            }
        }
        
        /* 超大屏幕优化 */
        @media (min-width: 1600px) {
            .container { max-width: 1800px; padding: 30px; }
            .header { padding: 30px; }
            .header-content { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
            .header h1 { font-size: 28px; margin: 0; }
            .header-actions { display: flex; gap: 10px; }
            .download-btn {
                display: inline-block;
                padding: 10px 20px;
                background-color: #4CAF50;
                color: white;
                text-decoration: none;
                border-radius: 4px;
                font-size: 16px;
                transition: background-color 0.3s;
            }
            .download-btn:hover { background-color: #45a049; }
            .stat-item .value { font-size: 22px; }
            .table-header, .log-row { 
                grid-template-columns: minmax(160px, 250px) 180px 120px 140px 160px 120px 120px minmax(300px, 1fr); 
                gap: 20px; 
                font-size: 16px; 
            }
            .log-row { padding: 20px; }
            .filter-form { grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; }
            .filter-form input, .filter-form select { 
                padding: 12px 16px; 
                font-size: 16px; 
            }
            .filter-form button { 
                padding: 12px 24px; 
                font-size: 16px; 
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="header-content">
                <h1>{{.Title}}</h1>
                <div class="header-actions">
                    {{if .EnableMainLogger}}
                    <a href="{{.BasePath}}/api/download-main-logs" class="download-btn">
                        下载主进程日志
                    </a>
                    {{end}}
                </div>
            </div>
            <div class="stats">
                {{if .Stats}}
                <div class="stat-item">
                    <div class="label">总请求数</div>
                    <div class="value">{{.Stats.total_requests}}</div>
                </div>
                <div class="stat-item">
                    <div class="label">平均响应时间</div>
                    <div class="value">{{if .Stats.avg_duration_ms}}{{printf "%.2f" .Stats.avg_duration_ms}}ms{{else}}-{{end}}</div>
                </div>
                <div class="stat-item">
                    <div class="label">错误率</div>
                    <div class="value">{{if .Stats.error_rate_percent}}{{printf "%.2f" .Stats.error_rate_percent}}%{{else}}-{{end}}</div>
                </div>
                <div class="stat-item">
                    <div class="label">存储大小</div>
                    <div class="value">{{.Stats.storage_size}}</div>
                </div>
                <div class="stat-item">
                    <div class="label">流式请求数</div>
                    <div class="value">{{.Stats.streaming_request_count}}</div>
                </div>
                <div class="stat-item">
                    <div class="label">平均分块数</div>
                    <div class="value">{{.Stats.avg_streaming_chunks}}</div>
                </div>
                <div class="stat-item">
                    <div class="label">最大分块数</div>
                    <div class="value">{{.Stats.max_streaming_chunks}}</div>
                </div>
                {{end}}
            </div>
        </div>
        
        <div class="nav">
            <a href="{{.BasePath}}/list" class="active">日志列表</a>
        </div>
        
        <div class="filters">
            <form class="filter-form" method="get" id="filter-form" onsubmit="handleFilterSubmit(event)">
                <div class="filter-header">
                    <h3>筛选条件</h3>
                    <div class="filter-actions">
                        <button type="submit">筛选</button>
                        <a href="javascript:void(0)" onclick="resetFilters()">重置</a>
                    </div>
                </div>
                <!-- 基础筛选组 -->
                <div class="filter-group">
                    <h4>基础筛选</h4>
                    <div class="filter-row">
                        <select name="record_type" id="filter-record_type" onchange="handleFilterChange(this)">
                            <option value="">所有记录类型</option>
                            <option value="http">HTTP记录</option>
                            <option value="process">进程记录</option>
                        </select>
                        <select name="pageSize" id="filter-pageSize" onchange="handleFilterChange(this)">
                            <option value="10">10条/页</option>
                            <option value="20">20条/页</option>
                            <option value="50">50条/页</option>
                            <option value="100">100条/页</option>
                        </select>
                    </div>
                    <div class="filter-row">
                        <input type="text" name="q" id="filter-q" placeholder="搜索日志内容...">
                    </div>
                </div>
                
                <!-- HTTP记录筛选组 -->
                <div class="filter-group">
                    <h4>HTTP记录筛选</h4>
                    <div class="filter-row">
                        <select name="method" id="filter-method" onchange="handleFilterChange(this)">
                            <option value="">所有方法</option>
                            <option value="GET">GET</option>
                            <option value="POST">POST</option>
                            <option value="PUT">PUT</option>
                            <option value="DELETE">DELETE</option>
                        </select>
                        <select name="status_code" id="filter-status_code" onchange="handleFilterChange(this)">
                            <option value="">所有状态码</option>
                            <option value="200">200 - 成功</option>
                            <option value="201">201 - 已创建</option>
                            <option value="204">204 - 无内容</option>
                            <option value="301">301 - 永久重定向</option>
                            <option value="302">302 - 临时重定向</option>
                            <option value="400">400 - 错误请求</option>
                            <option value="401">401 - 未授权</option>
                            <option value="403">403 - 禁止访问</option>
                            <option value="404">404 - 未找到</option>
                            <option value="500">500 - 服务器错误</option>
                            <option value="502">502 - 网关错误</option>
                            <option value="503">503 - 服务不可用</option>
                        </select>
                    </div>
                    <div class="filter-row">
                        <input type="text" name="client_ip" id="filter-client_ip" placeholder="客户端IP地址">
                        <input type="text" name="host" id="filter-host" placeholder="域名包含">
                        <input type="text" name="url" id="filter-url" placeholder="URL路径包含">
                    </div>
                    <div class="filter-row time-range-row">
                        <input type="datetime-local" name="start_time" id="filter-start_time" placeholder="开始时间">
                        <input type="datetime-local" name="end_time" id="filter-end_time" placeholder="结束时间">
                    </div>
                    <div class="filter-row">
                        <select name="is_streaming" id="filter-is_streaming" onchange="handleFilterChange(this)">
                            <option value="">所有流式状态</option>
                            <option value="true">流式请求</option>
                            <option value="false">非流式请求</option>
                        </select>
                    </div>
                </div>
                
                <!-- 进程记录筛选组 -->
                <div class="filter-group">
                    <h4>进程记录筛选</h4>
                    <div class="filter-row">
                        <input type="text" name="process_name" id="filter-process_name" placeholder="进程名称">
                        <input type="text" name="process_id" id="filter-process_id" placeholder="进程ID">
                        <select name="process_status" id="filter-process_status" onchange="handleFilterChange(this)">
                            <option value="">所有进程状态</option>
                            <option value="running">运行中</option>
                            <option value="completed">已完成</option>
                            <option value="failed">失败</option>
                            <option value="cancelled">已取消</option>
                        </select>
                    </div>
                </div>
            </form>
            
            <script>
                // 初始化筛选表单值 - 优先从pageParams获取，后备从URL参数获取
                function initFilterForm() {
                    // 从URL参数获取值（作为后备）
                    const urlParams = new URLSearchParams(window.location.search);
                    
                    // 获取参数值的辅助函数（优先从pageParams，后备从URL）
                    function getParamValue(key, pageParamsValue) {
                        if (pageParamsValue && pageParamsValue !== 'null' && pageParamsValue !== '') {
                            return pageParamsValue;
                        }
                        const urlValue = urlParams.get(key);
                        return urlValue || null;
                    }
                    
                    // 确保 pageParams 存在
                    const filters = window.pageParams && window.pageParams.filters ? window.pageParams.filters : {};
                    const keyword = window.pageParams ? window.pageParams.keyword : null;
                    
                    // 设置搜索关键词（优先pageParams，后备URL参数q）
                    const qValue = getParamValue('q', keyword);
                    if (qValue) {
                        const qInput = document.getElementById('filter-q');
                        if (qInput) qInput.value = qValue;
                    }
                    
                    // 设置筛选条件（优先URL参数，其次pageParams，最后使用默认值）
                    function setFilterValue(elementId, key, pageParamsValue, defaultValue) {
                        const value = getParamValue(key, pageParamsValue) || defaultValue;
                        if (value) {
                            const element = document.getElementById(elementId);
                            if (element) element.value = value;
                        }
                    }

                    setFilterValue('filter-record_type', 'record_type', filters.record_type);
                    // pageSize 的后备值使用后端实际生效的 pageSize，避免 HTML 默认第一项（10）与后端默认（20）不一致
                    setFilterValue('filter-pageSize', 'pageSize', filters.pageSize, window.pageParams.pageSize);
                    setFilterValue('filter-method', 'method', filters.method);
                    setFilterValue('filter-status_code', 'status_code', filters.status_code);
                    setFilterValue('filter-client_ip', 'client_ip', filters.client_ip);
                    setFilterValue('filter-host', 'host', filters.host);
                    setFilterValue('filter-url', 'url', filters.url);
                    setFilterValue('filter-start_time', 'start_time', filters.start_time);
                    setFilterValue('filter-end_time', 'end_time', filters.end_time);
                    setFilterValue('filter-is_streaming', 'is_streaming', filters.is_streaming);
                    setFilterValue('filter-process_name', 'process_name', filters.process_name);
                    setFilterValue('filter-process_id', 'process_id', filters.process_id);
                    setFilterValue('filter-process_status', 'process_status', filters.process_status);
                }
                
                // 页面加载时初始化表单
                initFilterForm();
            </script>
        </div>
        
        <div class="logs-table">
            <div class="table-container">
                <div class="table-content">
                    <div class="table-header">
                        <div>记录Id</div>
                        <div>时间</div>
                        <div>耗时</div>
                        <div>存储大小</div>
                        <div>类型</div>
                        <div>状态</div>
                        <div>详细信息</div>
                        <div>URL/进程信息</div>
                    </div>
                    
                    {{range .Entries}}
                    <div class="log-row">
                        <div class="request-id"><a href="{{$.BasePath}}/detail/{{.ID}}" title="查看详情">{{.ID}}</a></div>
                        <div class="timestamp">{{.Timestamp.Format "2006-01-02 15:04:05"}}</div>
                        <div class="duration">{{formatDuration .Duration}}</div>
                        <div class="storage-size">{{.StorageSize}}</div>
                        <div class="record-type">
                            {{if eq .RecordType "process"}}
                            <span class="type-badge-process" title="进程记录">进程</span>
                            {{else}}
                            <span class="type-badge-http" title="HTTP记录">HTTP</span>
                            <div class="http-method method-{{lower .Method}}">{{.Method}}</div>
                            {{end}}
                        </div>
                        <div class="status-info">
                            {{if eq .RecordType "process"}}
                            <span class="process-status process-status-{{lower .Status}}" title="进程状态: {{.Status}}">{{.Status}}</span>
                            {{else}}
                            <span class="status-code status-{{if ge .StatusCode 200}}{{if lt .StatusCode 300}}2xx{{else if lt .StatusCode 400}}3xx{{else if lt .StatusCode 500}}4xx{{else}}5xx{{end}}{{else}}1xx{{end}}">{{.StatusCode}}</span>
                            {{end}}
                        </div>
                        <div class="details">
                            {{if eq .RecordType "process"}}
                            <div class="process-details">
                                <div class="process-name" title="进程名称: {{.ProcessName}}">{{.ProcessName}}</div>
                                {{if .ProcessType}}<div class="process-type" title="进程类型: {{.ProcessType}}">{{.ProcessType}}</div>{{end}}
                            </div>
                            {{else}}
                            <div class="http-details">
							{{if .IsStreamingResponse}}
							<div class="streaming-badge streaming-active" title="流式请求: {{.StreamingChunks}}个分块，分块大小: {{.StreamingChunkSize}}字节">流式</div>
							{{else}}
							<div class="streaming-badge streaming-inactive" title="非流式请求">-</div>
							{{end}}
                                {{if .ClientIP}}<div class="client-ip">{{.ClientIP}}</div>{{end}}
                                {{if .Host}}<div class="client-ip" title="域名">{{.Host}}</div>{{end}}
                            </div>
                            {{end}}
                        </div>
                        <div class="url-info">
                            {{if eq .RecordType "process"}}
                            <div class="process-info" title="进程ID: {{.ProcessID}}">PID: {{.ProcessID}}</div>
                            {{else}}
                            <div class="url" title="{{.URL}}">{{.URL}}</div>
                            {{end}}
                        </div>
                    </div>
                    {{else}}
                    <div class="no-data-row">
                        暂无日志记录
                    </div>
                    {{end}}
                </div>
            </div>
        </div>
        
        {{if .Pagination}}
        <div class="pagination" id="pagination">
            {{/* 分页内容将由JavaScript动态生成 */}}
        </div>
        
        {{/* 分页数据将在页面底部脚本中定义 */}}
        <script>
            window.paginationData = {
                page: {{.Pagination.Page}},
                pageSize: {{.Pagination.PageSize}},
                totalPages: {{.Pagination.TotalPages}},
                hasPrev: {{.Pagination.HasPrev}},
                hasNext: {{.Pagination.HasNext}},
                prevPage: {{.Pagination.PrevPage}},
                nextPage: {{.Pagination.NextPage}}
            };
        </script>
        {{end}}
    </div>
    
    <script>
        // 页面参数初始化 - 页面加载时从模板获取参数值
        window.pageParams = {
            basePath: '{{.BasePath}}',
            page: {{.Page}},
            pageSize: {{.PageSize}},
            keyword: {{if .Keyword}}{{jsString .Keyword}}{{else}}null{{end}},
            filters: {
                record_type: {{if .Filters.record_type}}'{{.Filters.record_type}}'{{else}}null{{end}},
                method: {{if .Filters.method}}'{{.Filters.method}}'{{else}}null{{end}},
                status_code: {{if .Filters.status_code}}'{{.Filters.status_code}}'{{else}}null{{end}},
                client_ip: {{if .Filters.client_ip}}'{{.Filters.client_ip}}'{{else}}null{{end}},
                host: {{if .Filters.host}}'{{.Filters.host}}'{{else}}null{{end}},
                url: {{if .Filters.url}}'{{.Filters.url}}'{{else}}null{{end}},
                process_name: {{if .Filters.process_name}}'{{.Filters.process_name}}'{{else}}null{{end}},
                process_id: {{if .Filters.process_id}}'{{.Filters.process_id}}'{{else}}null{{end}},
                process_status: {{if .Filters.process_status}}'{{.Filters.process_status}}'{{else}}null{{end}},
                is_streaming: {{if .Filters.is_streaming}}'{{.Filters.is_streaming}}'{{else}}null{{end}},
                pageSize: {{if .Filters.pageSize}}'{{.Filters.pageSize}}'{{else}}null{{end}},
                start_time: {{if .Filters.start_time}}'{{.Filters.start_time}}'{{else}}null{{end}},
                end_time: {{if .Filters.end_time}}'{{.Filters.end_time}}'{{else}}null{{end}}
            }
        };

        // 构建列表页URL - 从当前URL和表单获取参数，确保包含用户最新的筛选条件
        function buildListURL(targetPage) {
            const params = new URLSearchParams();

            // 1. 从当前URL复制所有参数（获取最新的筛选条件）
            const currentParams = new URLSearchParams(window.location.search);
            for (const [key, value] of currentParams.entries()) {
                if (value && value !== 'null' && value !== '') {
                    params.set(key, value);
                }
            }

            // 2. 从表单获取最新的筛选条件（覆盖URL中的旧值）
            const form = document.getElementById('filter-form');
            if (form) {
                const formData = new FormData(form);
                for (const [key, value] of formData.entries()) {
                    if (value && value.trim() !== '') {
                        params.set(key, value);
                    } else {
                        params.delete(key);
                    }
                }
            }

            // 3. 更新分页参数
            const pageNum = parseInt(targetPage, 10);
            if (!isNaN(pageNum) && pageNum > 0) {
                params.set('page', pageNum);
            } else {
                params.delete('page');
            }

            // pageSize 已经通过 filter-form 中的下拉框在步骤2中读取，
            // 不再使用初始值 window.pageParams.pageSize 覆盖，避免用户修改后分页被重置。

            const queryString = params.toString();
            return window.pageParams.basePath + '/list' + (queryString ? '?' + queryString : '');
        }

        // 跳转到指定页面
        function goToPage(page) {
            const url = buildListURL(page);
            // 保存当前列表页URL到sessionStorage，供详情页返回时使用
            sessionStorage.setItem('debugger_list_url', url);
            window.location.href = url;
        }

        // 处理筛选条件变化（下拉框onchange事件）
        // 当下拉框值变化时，自动收集表单中所有字段的最新值并提交，
        // 避免只提交当前下拉框而丢失用户已输入但未点击筛选按钮的文本/时间条件。
        function handleFilterChange(selectElement) {
            const params = new URLSearchParams();

            // 1. 从当前URL中获取最新参数（确保获取用户之前的筛选操作）
            const currentParams = new URLSearchParams(window.location.search);
            for (const [key, value] of currentParams.entries()) {
                if (value && value !== 'null' && value !== '') {
                    params.set(key, value);
                }
            }

            // 2. 从整个表单获取最新值（覆盖URL中的旧值）
            const form = document.getElementById('filter-form');
            if (form) {
                const formData = new FormData(form);
                for (const [key, value] of formData.entries()) {
                    if (value && value.trim() !== '') {
                        params.set(key, value);
                    } else {
                        params.delete(key);
                    }
                }
            }

            // 3. 如果当前变化的 select 不在 form 中（防御性处理），单独处理它
            const newValue = selectElement.value;
            if (newValue && newValue.trim() !== '') {
                params.set(selectElement.name, newValue);
            } else {
                params.delete(selectElement.name);
            }

            // 筛选条件变化时重置到第1页
            params.delete('page');

            const queryString = params.toString();
            const url = window.pageParams.basePath + '/list' + (queryString ? '?' + queryString : '');
            // 保存当前列表页URL到sessionStorage，供详情页返回时使用
            sessionStorage.setItem('debugger_list_url', url);
            window.location.href = url;
        }

        // 处理筛选表单提交 - 排除空值参数，保留URL中已有的其他参数
        function handleFilterSubmit(event) {
            event.preventDefault();
            const form = event.target;
            const formData = new FormData(form);
            const params = new URLSearchParams();

            // 1. 先复制当前URL中的所有参数（保留已有筛选条件）
            const currentParams = new URLSearchParams(window.location.search);
            for (const [key, value] of currentParams.entries()) {
                if (value && value !== 'null' && value !== '') {
                    params.set(key, value);
                }
            }

            // 2. 更新表单中的参数（表单值覆盖URL中的值）
            for (const [key, value] of formData.entries()) {
                if (value && value.trim() !== '') {
                    params.set(key, value);
                } else {
                    // 如果表单值为空，移除该参数
                    params.delete(key);
                }
            }

            // 筛选条件变化时重置到第1页
            params.delete('page');

            const queryString = params.toString();
            const url = window.pageParams.basePath + '/list' + (queryString ? '?' + queryString : '');
            // 保存当前列表页URL到sessionStorage，供详情页返回时使用
            sessionStorage.setItem('debugger_list_url', url);
            window.location.href = url;
        }

        // 重置筛选条件
        function resetFilters() {
            sessionStorage.removeItem('debugger_list_url');
            window.location.href = window.pageParams.basePath + '/list';
        }

        // 辅助函数：字符串转小写
        function lower(str) {
            return str ? str.toLowerCase() : '';
        }
        
        // 渲染分页
        function renderPagination() {
            if (!window.paginationData) return;

            const data = window.paginationData;

            // 只有一页时不显示分页
            if (data.totalPages <= 1) {
                const container = document.getElementById('pagination');
                if (container) container.innerHTML = '';
                return;
            }

            const container = document.getElementById('pagination');
            if (!container) return;

            let html = '';

            // 上一页
            if (data.hasPrev) {
                html += '<a href="' + buildListURL(data.prevPage) + '">上一页</a>';
            } else {
                html += '<span class="disabled">上一页</span>';
            }

            // 页码
            if (data.totalPages <= 7) {
                // 显示所有页码
                for (let i = 1; i <= data.totalPages; i++) {
                    if (i === data.page) {
                        html += '<span class="current">' + i + '</span>';
                    } else {
                        html += '<a href="' + buildListURL(i) + '">' + i + '</a>';
                    }
                }
            } else {
                // 智能分页
                if (data.page > 3) {
                    html += '<a href="' + buildListURL(1) + '">1</a>';
                    if (data.page > 4) {
                        html += '<span class="ellipsis">...</span>';
                    }
                }

                // 当前页附近的页码
                let start = Math.max(1, data.page - 2);
                let end = Math.min(data.totalPages, data.page + 2);

                // 调整范围确保显示5个页码
                if (data.page <= 3) {
                    end = Math.min(data.totalPages, 5);
                } else if (data.page >= data.totalPages - 2) {
                    start = Math.max(1, data.totalPages - 4);
                }

                for (let i = start; i <= end; i++) {
                    if (i === data.page) {
                        html += '<span class="current">' + i + '</span>';
                    } else {
                        html += '<a href="' + buildListURL(i) + '">' + i + '</a>';
                    }
                }

                if (data.page < data.totalPages - 2) {
                    if (data.page < data.totalPages - 3) {
                        html += '<span class="ellipsis">...</span>';
                    }
                    html += '<a href="' + buildListURL(data.totalPages) + '">' + data.totalPages + '</a>';
                }
            }

            // 下一页
            if (data.hasNext) {
                html += '<a href="' + buildListURL(data.nextPage) + '">下一页</a>';
            } else {
                html += '<span class="disabled">下一页</span>';
            }

            container.innerHTML = html;
        }
        
        // 页面加载时渲染分页
        renderPagination();
    </script>
</body>
</html>`

// errorTemplate 错误页面模板
const errorTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; }
        .container { max-width: 600px; margin: 100px auto; padding: 20px; }
        .error-box { background: #fff; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); text-align: center; }
        .error-icon { font-size: 48px; color: #e74c3c; margin-bottom: 20px; }
        .error-title { color: #2c3e50; margin-bottom: 10px; }
        .error-message { color: #666; margin-bottom: 20px; }
        .back-link { color: #3498db; text-decoration: none; }
        .back-link:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-box">
            <div class="error-icon">⚠️</div>
            <h1 class="error-title">{{.Title}}</h1>
            <p class="error-message">{{.Message}}</p>
            <a href="{{.BasePath}}/list" class="back-link">返回首页</a>
        </div>
    </div>
    
    <script>
        function seq(start, end) {
            const result = [];
            for (let i = start; i <= end; i++) {
                result.push(i);
            }
            return result;
        }
    </script>
</body>
</html>`

// detailTemplate 详情页面模板
const detailTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; }
        .container { max-width: 1600px; margin: 0 auto; padding: 20px; }
        .header { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .header h1 { color: #2c3e50; margin-bottom: 10px; word-break: break-word; }
        .back-link { color: #3498db; text-decoration: none; margin-bottom: 10px; display: inline-block; }
        .back-link:hover { text-decoration: underline; }
        
        .json-view-link { 
            color: #27ae60; 
            text-decoration: none; 
            font-size: 14px; 
            font-weight: normal; 
            margin-left: 10px; 
            padding: 2px 6px; 
            border: 1px solid #27ae60; 
            border-radius: 3px; 
            background: #f8fff9;
        }
        .json-view-link:hover { 
            background: #27ae60; 
            color: white; 
            text-decoration: none; 
        }
        
        .detail-sections { display: flex; flex-direction: column; gap: 20px; }
        .section { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .section h2 { color: #2c3e50; margin-bottom: 15px; padding-bottom: 10px; border-bottom: 1px solid #eee; }
        
        .basic-info { 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); 
            gap: 16px; 
            align-items: start;
        }
        .info-item { 
            display: flex; 
            flex-direction: column; 
            min-height: 60px;
        }
        .info-label { 
            font-size: 13px; 
            color: #666; 
            margin-bottom: 6px; 
            font-weight: 600;
            line-height: 1.3;
        }
        .info-value { 
            font-size: 14px; 
            color: #333; 
            word-wrap: break-word; 
            overflow-wrap: break-word; 
            word-break: break-word;
            line-height: 1.4;
            flex: 1;
            max-width: 100%;
        }
        
        .headers-table, .params-table { width: 100%; border-collapse: collapse; }
        .headers-table th, .params-table th { background: #f8f9fa; padding: 10px; text-align: left; border-bottom: 1px solid #eee; }
        .headers-table td, .params-table td { 
            padding: 10px; 
            border-bottom: 1px solid #eee; 
            max-width: 300px; 
            overflow: visible; 
            white-space: normal; 
            word-wrap: break-word; 
            overflow-wrap: break-word; 
            word-break: break-word;
            line-height: 1.4;
        }
        .headers-table td:first-child, .params-table td:first-child {
            font-weight: 600;
            background: #f8f9fa;
            width: 300px;
        }
        .headers-table tr:last-child td, .params-table tr:last-child td { border-bottom: none; }
        
        /* 表格容器，支持水平滚动 */
        .table-container { 
            overflow-x: auto; 
            margin-top: 15px; 
            border: 1px solid #eee;
            border-radius: 4px;
            max-width: 100%;
        }
        .table-container table { 
            min-width: 600px; 
            width: 100%;
            margin: 0;
        }
        
        .json-viewer { 
            background: #f8f9fa; 
            border: 1px solid #eee; 
            border-radius: 4px; 
            padding: 10px;
            font-family: 'Courier New', monospace; 
            font-size: 12px; 
            position: relative;
        }
        .json-viewer pre { 
            margin: 0; 
            white-space: pre-wrap; 
            word-wrap: break-word; 
            overflow-wrap: break-word; 
            word-break: break-word; 
            max-width: 100%; 
            line-height: 1.4;
        }
        
        /* JSON语法高亮样式 */
        .json-key { color: #881391; font-weight: bold; }
        .json-string { color: #c41a16; }
        .json-number { color: #1c00cf; }
        .json-boolean { color: #0d22aa; font-weight: bold; }
        .json-null { color: #808080; font-weight: bold; }
        .json-punctuation { color: #000000; }
        .json-collapse { cursor: pointer; color: #666; margin-right: 5px; }
        .json-collapsed { color: #999; font-style: italic; }
        .json-toggle { cursor: pointer; color: #666; margin-right: 5px; }
        .json-line { display: block; }
        
        .tab-container { margin-top: 20px; }
        .tabs { display: flex; border-bottom: 1px solid #eee; margin-bottom: 15px; }
        .tab { padding: 10px 20px; cursor: pointer; border: 1px solid transparent; border-bottom: none; margin-bottom: -1px; }
        .tab.active { background: #fff; border-color: #eee; border-bottom-color: #fff; border-radius: 4px 4px 0 0; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        
        .method-badge, .status-badge { padding: 8px 8px; border-radius: 4px; font-weight: bold; display: inline-block; min-width: 50px; }
        .method-badge.method-get { background: #d4edda; color: #155724; }
        .method-badge.method-post { background: #d1ecf1; color: #0c5460; }
        .method-badge.method-put { background: #fff3cd; color: #856404; }
        .method-badge.method-delete { background: #f8d7da; color: #721c24; }
        .method-badge.method-patch { background: #e2e3e5; color: #383d41; }
        .method-badge.method-head { background: #d1ecf1; color: #0c5460; }
        .method-badge.method-options { background: #e2e3e5; color: #383d41; }
        .status-2xx { background: #d4edda; color: #155724; }
        .status-3xx { background: #d1ecf1; color: #0c5460; }
        .status-4xx { background: #f8d7da; color: #721c24; }
        .status-5xx { background: #f5c6cb; color: #721c24; }
        
        /* 进程状态样式 */
        .process-status { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; display: inline-block; }
        .process-status-running { background: #fff3cd; color: #856404; }
        .process-status-completed { background: #d4edda; color: #155724; }
        .process-status-failed { background: #f5c6cb; color: #721c24; }
        .process-status-cancelled { background: #f8d7da; color: #721c24; }
        
        /* Logger日志样式 */
        .logger-logs { margin-top: 15px; }
        .log-item { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 4px; padding: 12px; margin-bottom: 10px; }
        .log-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
        .log-timestamp { font-size: 12px; color: #6c757d; }
        .log-level { padding: 2px 6px; border-radius: 3px; font-size: 11px; font-weight: bold; text-transform: uppercase; }
        .level-debug { background: #d1ecf1; color: #0c5460; }
        .level-info { background: #d4edda; color: #155724; }
        .level-warn { background: #fff3cd; color: #856404; }
        .level-error { background: #f8d7da; color: #721c24; }
        .log-message { 
            font-size: 14px; 
            color: #333; 
            margin-bottom: 8px; 
            word-wrap: break-word; 
            overflow-wrap: break-word; 
            word-break: break-word; 
            max-width: 100%; 
            line-height: 1.4;
        }
        .log-fields { display: flex; flex-wrap: wrap; gap: 8px; }
        .log-field { background: #e9ecef; padding: 2px 6px; border-radius: 3px; font-size: 11px; color: #495057; }
        
        .section:last-child { margin-bottom: 0; }
        
        @media (max-width: 768px) {
            .container { padding: 15px; }
            .header { padding: 15px; }
            .header h1 { font-size: 20px; margin-bottom: 15px; }
            .section { padding: 15px; }
            .section h2 { font-size: 18px; margin-bottom: 12px; }
            .section h3 { font-size: 16px; margin-bottom: 10px; }
            .basic-info { grid-template-columns: 1fr; gap: 12px; }
            .info-item { margin-bottom: 8px; }
            .info-label { font-size: 13px; font-weight: 600; }
            .info-value { 
                font-size: 14px; 
                line-height: 1.4; 
                word-wrap: break-word; 
                overflow-wrap: break-word; 
                word-break: break-word; 
                max-width: 100%; 
            }
            .headers-table, .params-table { font-size: 13px; }
            .headers-table th, .params-table th,
            .headers-table td, .params-table td { 
                padding: 8px 10px; 
                white-space: normal; 
                word-break: break-word;
                line-height: 1.4;
            }
            .headers-table td:first-child, .params-table td:first-child {
                font-weight: 600;
                background: #f8f9fa;
                width: 150px;
            }
            .json-viewer { padding: 10px; font-size: 12px; max-height: 300px; }
            .log-header { flex-direction: column; align-items: flex-start; gap: 6px; }
            .log-item { padding: 10px; margin-bottom: 8px; }
            .log-message { 
                font-size: 14px; 
                line-height: 1.4; 
                word-wrap: break-word; 
                overflow-wrap: break-word; 
                word-break: break-word; 
                max-width: 100%; 
            }
            .log-fields { gap: 4px; }
            .log-field { font-size: 11px; padding: 2px 6px; }
            
            /* 移动端表格容器优化 */
            .table-container {
                margin-top: 10px;
                border: 1px solid #ddd;
            }
            .table-container table {
                min-width: 500px;
            }
        }
        
        @media (max-width: 480px) {
            .container { padding: 10px; }
            .header { padding: 12px; margin-bottom: 15px; }
            .header h1 { font-size: 18px; margin-bottom: 12px; }
            .section { padding: 12px; }
            .section h2 { font-size: 16px; margin-bottom: 10px; }
            .section h3 { font-size: 14px; margin-bottom: 8px; }
            .basic-info { gap: 8px; }
            .info-item { margin-bottom: 6px; }
            .info-label { font-size: 12px; font-weight: 600; }
            .info-value { 
                font-size: 13px; 
                line-height: 1.4; 
                word-wrap: break-word; 
                overflow-wrap: break-word; 
                word-break: break-word; 
                max-width: 100%; 
            }
            .headers-table, .params-table { font-size: 12px; }
            .headers-table th, .params-table th,
            .headers-table td, .params-table td { 
                padding: 6px 8px; 
                white-space: normal; 
                word-break: break-word;
                line-height: 1.4;
            }
            .headers-table td:first-child, .params-table td:first-child {
                font-weight: 600;
                background: #f8f9fa;
                width: 150px;
            }
            .json-viewer { padding: 8px; font-size: 11px; max-height: 250px; }
            .log-item { padding: 8px; margin-bottom: 8px; }
            .log-message { 
                font-size: 13px; 
                line-height: 1.4; 
                word-wrap: break-word; 
                overflow-wrap: break-word; 
                word-break: break-word; 
                max-width: 100%; 
            }
            .log-field { font-size: 10px; padding: 1px 4px; }
            
            /* 超小屏幕表格优化 */
            .table-container {
                margin-top: 8px;
                border: 1px solid #ddd;
            }
            .table-container table {
                min-width: 400px;
            }
            
            /* 超小屏幕方法徽章优化 */
            .method-badge, .status-badge { 
                padding: 3px 6px; 
                font-size: 11px; 
                min-width: 40px;
            }
        }
        
        /* 超大屏幕优化 */
        @media (min-width: 1600px) {
            .container { max-width: 1800px; padding: 30px; }
            .header { padding: 30px; }
            .header h1 { font-size: 28px; margin-bottom: 15px; }
            .section { padding: 30px; }
            .section h2 { font-size: 24px; margin-bottom: 20px; }
            .section h3 { font-size: 20px; margin-bottom: 15px; }
            .basic-info { 
                grid-template-columns: repeat(auto-fit, minmax(350px, 1fr)); 
                gap: 20px; 
            }
            .info-item { min-height: 80px; }
            .info-label { font-size: 16px; margin-bottom: 8px; }
            .info-value { font-size: 18px; line-height: 1.5; }
            .headers-table, .params-table { font-size: 16px; }
            .headers-table th, .params-table th,
            .headers-table td, .params-table td { 
                padding: 15px 20px; 
                white-space: normal; 
                word-break: break-word;
                line-height: 1.5;
            }
            .headers-table td:first-child, .params-table td:first-child {
                min-width: 160px;
            }
            .json-viewer { 
                padding: 20px; 
                font-size: 16px; 
                max-height: 500px; 
                line-height: 1.5;
            }
            .method-badge, .status-badge { 
                padding: 14px 14px; 
                font-size: 16px; 
                min-width: 60px;
            }
            .back-link { font-size: 16px; margin-bottom: 15px; }
            .json-view-link { font-size: 16px; padding: 4px 8px; }
            
            /* 表格容器优化 */
            .table-container {
                margin-top: 20px;
                border: 1px solid #eee;
            }
            .table-container table {
                min-width: 800px;
            }
        }

        /* JSON 查看器工具栏样式 */
        .json-viewer-toolbar {
            display: flex;
            gap: 8px;
            margin-bottom: 10px;
            flex-wrap: wrap;
            align-items: center;
        }
        .json-viewer-toolbar button {
            background: #3498db;
            color: white;
            border: none;
            padding: 5px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
            transition: background-color 0.2s ease;
        }
        .json-viewer-toolbar button:hover {
            background: #2980b9;
        }
        .json-viewer-toolbar button.secondary {
            background: #6c757d;
        }
        .json-viewer-toolbar button.secondary:hover {
            background: #5a6268;
        }
        .json-viewer-toolbar button.success {
            background: #27ae60;
        }
        .json-viewer-toolbar button.success:hover {
            background: #219a52;
        }
        .json-viewer-toolbar .json-status {
            font-size: 12px;
            color: #666;
            margin-left: auto;
        }
        .json-viewer-toolbar .json-status.error {
            color: #e74c3c;
        }
        .json-viewer-toolbar .json-status.warning {
            color: #f57c00;
        }

        /* JSON 折叠查看器样式 */
        .json-tree-viewer {
            background: #f8f9fa;
            border: 1px solid #eee;
            border-radius: 4px;
            padding: 10px;
            max-height: 400px;
            overflow: auto;
            font-family: 'Courier New', Courier, monospace;
            font-size: 13px;
            line-height: 1.5;
        }
        .json-tree-viewer pre {
            margin: 0;
            white-space: pre-wrap;
            word-wrap: break-word;
            overflow-wrap: break-word;
            word-break: break-word;
        }
        .json-line {
            display: block;
            position: relative;
            padding-left: 14px;
        }
        .json-line:hover {
            background: rgba(52, 152, 219, 0.05);
        }
        .json-close-line {
            padding-left: 0;
        }
        .json-toggle {
            position: absolute;
            left: 0;
            top: 2px;
            width: 14px;
            height: 14px;
            line-height: 12px;
            text-align: center;
            cursor: pointer;
            color: #666;
            font-size: 10px;
            user-select: none;
            border-radius: 2px;
        }
        .json-toggle:hover {
            background: #e0e0e0;
        }
        .json-toggle::before {
            content: '▼';
        }
        .json-line.collapsed > .json-toggle::before {
            content: '▶';
        }
        .json-line.collapsed .json-children {
            display: none;
        }
        .json-line.collapsed .json-collapsed-preview {
            display: inline;
        }
        .json-collapsed-preview {
            display: none;
            color: #999;
            font-style: italic;
        }
        .json-children {
            display: block;
        }
        .json-key { color: #881391; font-weight: bold; }
        .json-string { color: #c41a16; }
        .json-number { color: #1c00cf; }
        .json-boolean { color: #0d22aa; font-weight: bold; }
        .json-null { color: #808080; font-weight: bold; }
        .json-punctuation { color: #000000; }
        .json-comment { color: #008000; font-style: italic; }
        .json-error-hint {
            background: #fff3cd;
            border: 1px solid #f57c00;
            border-radius: 4px;
            padding: 10px;
            margin-bottom: 10px;
            font-size: 12px;
            color: #856404;
        }
        .json-error-hint strong {
            color: #e65100;
        }

        /* 原始内容预览模式 */
        .json-raw-viewer {
            background: #f8f9fa;
            border: 1px solid #eee;
            border-radius: 4px;
            padding: 15px;
            max-height: 500px;
            overflow: auto;
            font-family: 'Courier New', Courier, monospace;
            font-size: 13px;
            line-height: 1.5;
            white-space: pre-wrap;
            word-wrap: break-word;
            overflow-wrap: break-word;
            word-break: break-word;
        }

        @media (max-width: 768px) {
            .json-tree-viewer,
            .json-raw-viewer {
                padding: 8px;
                font-size: 12px;
                max-height: 300px;
            }
            .json-viewer-toolbar {
                gap: 6px;
            }
            .json-viewer-toolbar button {
                padding: 4px 8px;
                font-size: 11px;
            }
            .json-viewer-toolbar .json-status {
                font-size: 11px;
                width: 100%;
                margin-left: 0;
                margin-top: 6px;
            }
        }
    </style>
    <script src="{{.BasePath}}/static/jsonc-parser.bundle.js"></script>
</head>
<body>
    <div class="container">
        <a href="javascript:void(0)" onclick="goBackToList()" class="back-link" id="back-link">← 返回日志列表</a>

        <div class="header">
            <h1>{{.Title}} <span class="record-type-badge record-type-{{.Entry.RecordType}}">{{if eq .Entry.RecordType "process"}}进程记录{{else}}HTTP记录{{end}}</span> <a href="{{.BasePath}}/api/logs/{{.Entry.ID}}" target="_blank" class="json-view-link" title="查看JSON数据">[JSON]</a></h1>
        </div>
        
        {{if .Entry}}
        <div class="detail-sections">
            <!-- 基本信息 -->
            <div class="section">
                <h2>基本信息</h2>
                <div class="basic-info">
                    <div class="info-item">
                        <div class="info-label">{{if eq .Entry.RecordType "process"}}进程ID{{else}}请求ID{{end}}</div>
                        <div class="info-value">{{.Entry.ID}}</div>
                    </div>
                    {{if eq .Entry.RecordType "process"}}
                    <!-- 进程记录专用信息 -->
                    <div class="info-item">
                        <div class="info-label">进程名称</div>
                        <div class="info-value">{{.Entry.ProcessName}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">进程类型</div>
                        <div class="info-value">{{.Entry.ProcessType}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">进程状态</div>
                        <div class="info-value status-badge process-status-{{lower .Entry.Status}}">{{.Entry.Status}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">开始时间</div>
                        <div class="info-value">{{.Entry.Timestamp.Format "2006-01-02 15:04:05"}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">结束时间</div>
                        <div class="info-value">{{if .Entry.EndTime.IsZero}}进行中{{else}}{{.Entry.EndTime.Format "2006-01-02 15:04:05"}}{{end}}</div>
                    </div>
                    {{else}}
                    <!-- HTTP记录信息 -->
                    <div class="info-item">
                        <div class="info-label">时间</div>
                        <div class="info-value">{{.Entry.Timestamp.Format "2006-01-02 15:04:05"}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">方法</div>
                        <div class="info-value method-badge method-{{lower .Entry.Method}}">{{.Entry.Method}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">状态码</div>
                        <div class="info-value status-badge status-{{if ge .Entry.StatusCode 200}}{{if lt .Entry.StatusCode 300}}2xx{{else if lt .Entry.StatusCode 400}}3xx{{else if lt .Entry.StatusCode 500}}4xx{{else}}5xx{{end}}{{else}}1xx{{end}}">{{.Entry.StatusCode}}</div>
                    </div>
                    {{end}}
                    <div class="info-item">
                        <div class="info-label">耗时</div>
                        <div class="info-value">{{formatDuration .Entry.Duration}}</div>
                    </div>
                    {{if ne .Entry.RecordType "process"}}
                    <div class="info-item">
                        <div class="info-label">客户端IP</div>
                        <div class="info-value">{{.Entry.ClientIP}}</div>
                    </div>
                    {{end}}
                    {{if ne .Entry.RecordType "process"}}
                    <!-- 流式请求信息 -->
                    <div class="info-item">
                        <div class="info-label">流式请求</div>
                        <div class="info-value">
                            {{if .Entry.IsStreamingResponse}}
                            <span class="streaming-badge streaming-active">是</span>
                            {{else}}
                            <span class="streaming-badge streaming-inactive">否</span>
                            {{end}}
                        </div>
                    </div>
                    {{if .Entry.IsStreamingResponse}}
                    <div class="info-item">
                        <div class="info-label">分块数量</div>
                        <div class="info-value">{{.Entry.StreamingChunks}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">分块大小</div>
                        <div class="info-value">{{.Entry.StreamingChunkSize}} bytes</div>
                    </div>
                    {{end}}
                    {{end}}
                </div>
            </div>
            
            <!-- 详细信息 -->
            {{if ne .Entry.RecordType "process"}}
            <!-- URL和参数 -->
            <div class="section">
                <h2>请求信息</h2>
                <div class="basic-info">
                    <div class="info-item">
                        <div class="info-label">URL</div>
                        <div class="info-value">{{.Entry.URL}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">域名</div>
                        <div class="info-value">{{.Entry.Host}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">User Agent</div>
                        <div class="info-value">{{.Entry.UserAgent}}</div>
                    </div>
                </div>
                
                {{if .Entry.RequestHeaders}}
                <div style="margin-top: 15px;">
                    <h3>请求头</h3>
                    <div class="table-container">
                        <table class="headers-table">
                            <thead>
                                <tr>
                                    <th>名称</th>
                                    <th>值</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range $key, $value := .Entry.RequestHeaders}}
                                <tr>
                                    <td>{{$key}}</td>
                                    <td>{{$value}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                </div>
                {{end}}
                
                {{if .Entry.QueryParams}}
                <div style="margin-top: 15px;">
                    <h3>查询参数</h3>
                    <div class="table-container">
                        <table class="params-table">
                            <thead>
                                <tr>
                                    <th>参数名</th>
                                    <th>值</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range $key, $value := .Entry.QueryParams}}
                                <tr>
                                    <td>{{$key}}</td>
                                    <td>{{$value}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                </div>
                {{end}}
                
                <!-- 请求体 -->
                {{if .Entry.RequestBody}}
                <div style="margin-top: 15px;">
                    <h3>请求体</h3>
                    <div class="json-viewer">
                        <pre>{{.Entry.RequestBody | html}}</pre>
                    </div>
                </div>
                {{end}}
            </div>
            {{end}}
            
            <!-- 响应信息 -->
            {{if or .Entry.ResponseBody .Entry.ResponseHeaders}}
            <div class="section">
                {{if eq .Entry.RecordType "process"}}
                <h2>进程输出</h2>
                {{else}}
                <h2>响应信息</h2>
                {{if .Entry.IsStreamingResponse}}
                <!-- 流式响应信息 -->
                <div style="margin-bottom: 20px;">
                    <h3>流式响应元数据</h3>
                    <div class="basic-info">
                        <div class="info-item">
                            <div class="info-label">最大分块数</div>
                            <div class="info-value">{{.Entry.MaxStreamingChunks}}</div>
                        </div>
                        <div class="info-item">
                            <div class="info-label">流式数据大小</div>
                            <div class="info-value">{{if .Entry.StreamingData}}{{.Entry.StreamingData | len}} bytes{{else}}0 bytes{{end}}</div>
                        </div>
                    </div>
                </div>
                {{if .Entry.StreamingData}}
                <div style="margin-bottom: 20px;">
                    <h3>流式响应数据</h3>
                    <div class="json-viewer">
                        <pre>{{.Entry.StreamingData | html}}</pre>
                    </div>
                </div>
                {{end}}
                {{end}}
                {{end}}

                {{if .Entry.ResponseHeaders}}
                <div style="margin-top: 15px;">
                    <h3>响应头</h3>
                    <div class="table-container">
                        <table class="headers-table">
                            <thead>
                                <tr>
                                    <th>名称</th>
                                    <th>值</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range $key, $value := .Entry.ResponseHeaders}}
                                <tr>
                                    <td>{{$key}}</td>
                                    <td>{{$value}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                </div>
                {{end}}

                {{if .Entry.ResponseBody}}
                <div style="margin-top: 15px;">
                    {{if eq .Entry.RecordType "process"}}
                    <h3>输出内容</h3>
                    {{else}}
                    <h3>响应体</h3>
                    {{end}}
                    <div class="json-viewer">
                        <pre>{{.Entry.ResponseBody | html}}</pre>
                    </div>
                </div>
                {{end}}
            </div>
            {{end}}

            <!-- 会话数据 -->
            {{if .Entry.SessionData}}
            <div class="section">
                <h2>会话数据</h2>
                <div class="json-viewer">
                    <pre>{{.Entry.SessionData | json}}</pre>
                </div>
            </div>
            {{end}}

            <!-- Logger -->
            {{if .Entry.LoggerLogs}}
            <div class="section">
                <h2>日志</h2>
                <div class="logger-logs">
                    {{range .Entry.LoggerLogs}}
                    <div class="log-item">
                        <div class="log-header">
                        <span class="log-level level-{{.Level}}">{{.Level}}</span>
                        <span class="log-timestamp">{{.Timestamp.Format "2006-01-02 15:04:05"}}</span>
                    </div>
                        <div class="log-message">
                            {{if isJSON .Message}}
                            <div class="json-viewer">
                                <pre>{{.Message}}</pre>
                            </div>
                            {{else}}
                            {{.Message}}
                            {{end}}
                        </div>
                        {{if .Fields}}
                        <div class="log-fields">
                            {{range $key, $value := .Fields}}
                            {{if and (ne $key "level") (ne $key "message") (ne $key "timestamp") (ne $key "request_id") (ne $key "method") (ne $key "url") (ne $key "host") (ne $key "client_ip") (ne $key "process_id") (ne $key "process_name") (ne $key "process_type")}}
                            <span class="log-field">{{$key}}: {{$value}}</span>
                            {{end}}
                            {{end}}
                        </div>
                        {{end}}
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}

        </div>
        {{else}}
        <div class="section">
            <h2>日志记录不存在</h2>
            <p>请求的日志记录不存在或已被删除。</p>
        </div>
        {{end}}
    </div>
    
    <script>
        // 页面加载时初始化所有 JSON 查看器
        document.addEventListener('DOMContentLoaded', function() {
            initJSONViewers();
        });

        function lower(str) {
            return str ? str.toLowerCase() : '';
        }

        // 返回日志列表，优先使用sessionStorage中的列表URL，其次使用浏览器返回，最后回退到列表页
        function goBackToList() {
            // 尝试从sessionStorage获取之前保存的列表页URL
            const savedListUrl = sessionStorage.getItem('debugger_list_url');
            if (savedListUrl) {
                window.location.href = savedListUrl;
                return;
            }
            // 如果有referrer且不是当前详情页，使用浏览器返回
            if (document.referrer && !document.referrer.includes('/detail/')) {
                history.back();
                return;
            }
            // 默认返回列表页（不带筛选条件）
            window.location.href = '{{.BasePath}}/list';
        }

        // 初始化页面上所有 .json-viewer 容器
        function initJSONViewers() {
            const containers = document.querySelectorAll('.json-viewer');
            containers.forEach(function(container) {
                const pre = container.querySelector('pre');
                if (!pre) return;

                const rawText = pre.textContent;
                if (!rawText || !rawText.trim()) return;

                // 仅处理看起来像 JSON 的内容
                const trimmed = rawText.trim();
                if (!trimmed.startsWith('{') && !trimmed.startsWith('[')) return;

                // 避免重复初始化
                if (container.dataset.jsonViewerInitialized === 'true') return;
                container.dataset.jsonViewerInitialized = 'true';

                try {
                    new JSONViewer(container, rawText).render();
                } catch (err) {
                    console.error('JSON 查看器初始化失败:', err);
                }
            });
        }

        // HTML 特殊字符转义
        function escapeHTML(text) {
            return text
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;');
        }

        // 使用 jsonc-parser 尝试解析文本，返回解析结果、错误列表及修复后的文本
        function parseWithRepair(text) {
            const errors = [];
            const options = { allowTrailingComma: true, allowEmptyContent: false };

            // 第一次解析：使用 jsonc-parser 的容错解析
            let value = jsoncParser.parse(text, errors, options);

            if (errors.length === 0 && value !== undefined) {
                return { value: value, errors: [], repairedText: null, rawText: text };
            }

            // 尝试去除注释后再解析
            let strippedText = text;
            try {
                strippedText = jsoncParser.stripComments(text);
            } catch (e) {
                strippedText = text;
            }

            if (strippedText !== text) {
                errors.length = 0;
                value = jsoncParser.parse(strippedText, errors, options);
                if (errors.length === 0 && value !== undefined) {
                    return { value: value, errors: [], repairedText: JSON.stringify(value, null, 2), rawText: text };
                }
            }

            // 仍有问题时，返回容错解析结果及修复后的文本
            if (value !== undefined) {
                return { value: value, errors: errors, repairedText: JSON.stringify(value, null, 2), rawText: text };
            }

            return { value: null, errors: errors, repairedText: null, rawText: text };
        }

        // 根据 AST 节点类型创建对应的高亮文本节点
        function createLiteralSpan(node) {
            if (!node) return createSpan('null', 'json-null');

            switch (node.type) {
                case 'string':
                    return createSpan(JSON.stringify(node.value), 'json-string');
                case 'number':
                    return createSpan(String(node.value), 'json-number');
                case 'boolean':
                    return createSpan(node.value ? 'true' : 'false', 'json-boolean');
                case 'null':
                    return createSpan('null', 'json-null');
                default:
                    return createSpan(String(node.value), 'json-string');
            }
        }

        function createSpan(text, className) {
            const span = document.createElement('span');
            span.className = className;
            span.textContent = text;
            return span;
        }

        // JSON 查看器类
        function JSONViewer(container, rawText) {
            this.container = container;
            this.rawText = rawText;
            this.parseResult = null;
            this.showingRaw = false;
            this.viewer = null;
            this.toolbarStatus = null;
            this.repairBtn = null;
            this.rawBtn = null;
        }

        JSONViewer.prototype.render = function() {
            this.container.innerHTML = '';
            this.parseResult = parseWithRepair(this.rawText);

            const toolbar = this.createToolbar();
            this.container.appendChild(toolbar);

            if (this.parseResult.errors.length > 0 && this.parseResult.repairedText) {
                this.container.appendChild(this.createRepairHint());
            }

            this.viewer = document.createElement('div');
            this.viewer.className = 'json-tree-viewer';
            this.container.appendChild(this.viewer);

            this.renderContent();
        };

        JSONViewer.prototype.createToolbar = function() {
            const self = this;
            const toolbar = document.createElement('div');
            toolbar.className = 'json-viewer-toolbar';

            const expandBtn = document.createElement('button');
            expandBtn.textContent = '全部展开';
            expandBtn.addEventListener('click', function() { self.expandAll(); });
            toolbar.appendChild(expandBtn);

            const collapseBtn = document.createElement('button');
            collapseBtn.textContent = '全部折叠';
            collapseBtn.addEventListener('click', function() { self.collapseAll(); });
            toolbar.appendChild(collapseBtn);

            const copyBtn = document.createElement('button');
            copyBtn.className = 'secondary';
            copyBtn.textContent = '复制';
            copyBtn.addEventListener('click', function() { self.copy(); });
            toolbar.appendChild(copyBtn);

            if (this.parseResult.repairedText) {
                this.repairBtn = document.createElement('button');
                this.repairBtn.className = 'success';
                this.repairBtn.textContent = '查看正常';
                this.repairBtn.addEventListener('click', function() {
                    self.showingRaw = false;
                    self.updateButtonStates();
                    self.updateStatus();
                    self.renderContent();
                });
                toolbar.appendChild(this.repairBtn);

                this.rawBtn = document.createElement('button');
                this.rawBtn.className = 'secondary';
                this.rawBtn.textContent = '查看原始';
                this.rawBtn.addEventListener('click', function() {
                    self.showingRaw = !self.showingRaw;
                    self.updateButtonStates();
                    self.updateStatus();
                    self.renderContent();
                });
                toolbar.appendChild(this.rawBtn);
            }

            const status = document.createElement('span');
            status.className = 'json-status';
            this.toolbarStatus = status;
            this.updateStatus();
            toolbar.appendChild(status);

            return toolbar;
        };

        JSONViewer.prototype.updateStatus = function() {
            if (!this.toolbarStatus) return;

            if (this.showingRaw) {
                this.toolbarStatus.textContent = '显示原始内容';
                this.toolbarStatus.className = 'json-status warning';
            } else if (this.parseResult.repairedText) {
                this.toolbarStatus.textContent = '显示自动修复后的内容';
                this.toolbarStatus.className = 'json-status warning';
            } else if (this.parseResult.errors.length > 0) {
                this.toolbarStatus.textContent = '解析存在 ' + this.parseResult.errors.length + ' 处问题，已尝试修复';
                this.toolbarStatus.className = 'json-status error';
            } else {
                this.toolbarStatus.textContent = 'JSON 格式正确';
                this.toolbarStatus.className = 'json-status';
            }
        };

        JSONViewer.prototype.updateButtonStates = function() {
            if (this.repairBtn) {
                this.repairBtn.textContent = this.showingRaw ? '查看修复后' : '查看正常';
            }
            if (this.rawBtn) {
                this.rawBtn.textContent = this.showingRaw ? '查看正常' : '查看原始';
            }
        };

        JSONViewer.prototype.createRepairHint = function() {
            const hint = document.createElement('div');
            hint.className = 'json-error-hint';

            let errorText = '';
            const maxErrors = 3;
            for (let i = 0; i < Math.min(this.parseResult.errors.length, maxErrors); i++) {
                const err = this.parseResult.errors[i];
                errorText += jsoncParser.printParseErrorCode(err.error);
                if (i < Math.min(this.parseResult.errors.length, maxErrors) - 1) {
                    errorText += '、';
                }
            }
            if (this.parseResult.errors.length > maxErrors) {
                errorText += ' 等';
            }

            hint.innerHTML = '<strong>提示：</strong>当前 JSON 存在格式问题（' + errorText + '），系统已自动修复。' +
                '可点击「查看原始」查看原始内容，点击「查看修复后」查看修复结果。';
            return hint;
        };

        JSONViewer.prototype.renderContent = function() {
            this.viewer.innerHTML = '';

            if (this.showingRaw) {
                const raw = document.createElement('pre');
                raw.className = 'json-raw-viewer';
                raw.textContent = this.rawText;
                this.viewer.appendChild(raw);
                this.viewer.className = 'json-tree-viewer';
                return;
            }

            // 默认优先显示修复后的 JSON（如果修复成功），用户可通过「查看原始」切换
            const textToRender = this.showingRaw
                ? this.rawText
                : (this.parseResult.repairedText || this.rawText);

            // 若修复失败，直接显示原始文本
            if (this.parseResult.errors.length > 0 && !this.parseResult.repairedText) {
                const raw = document.createElement('pre');
                raw.className = 'json-raw-viewer';
                raw.textContent = this.rawText;
                this.viewer.appendChild(raw);
                return;
            }

            try {
                const tree = jsoncParser.parseTree(textToRender, []);
                // parseTree 返回的根节点即为 JSON 值节点（object/array/string 等）
                if (tree && tree.type) {
                    this.renderNode(tree, 0, this.viewer);
                } else {
                    const fallback = document.createElement('pre');
                    fallback.textContent = textToRender;
                    this.viewer.appendChild(fallback);
                }
            } catch (err) {
                console.error('渲染 JSON 树失败:', err);
                const fallback = document.createElement('pre');
                fallback.textContent = textToRender;
                this.viewer.appendChild(fallback);
            }
        };

        JSONViewer.prototype.renderNode = function(node, level, parent) {
            const self = this;

            if (node.type === 'object') {
                this.renderCollapsibleContainer(node, parent, '{', '}', '... }',
                    function(childNode, childParent) {
                        // property 节点：key: value
                        const propLine = document.createElement('span');
                        propLine.className = 'json-line';

                        const keyNode = childNode.children[0];
                        const valueNode = childNode.children[1];

                        propLine.appendChild(createSpan(JSON.stringify(keyNode.value), 'json-key'));
                        propLine.appendChild(createSpan(': ', 'json-punctuation'));

                        if (valueNode && (valueNode.type === 'object' || valueNode.type === 'array')) {
                            self.renderNode(valueNode, level + 1, propLine);
                        } else {
                            propLine.appendChild(createLiteralSpan(valueNode));
                        }

                        childParent.appendChild(propLine);
                    });
            } else if (node.type === 'array') {
                this.renderCollapsibleContainer(node, parent, '[', ']', '... ]',
                    function(childNode, childParent) {
                        const itemLine = document.createElement('span');
                        itemLine.className = 'json-line';

                        if (childNode.type === 'object' || childNode.type === 'array') {
                            self.renderNode(childNode, level + 1, itemLine);
                        } else {
                            itemLine.appendChild(createLiteralSpan(childNode));
                        }

                        childParent.appendChild(itemLine);
                    });
            } else {
                const line = document.createElement('span');
                line.className = 'json-line';
                line.appendChild(createLiteralSpan(node));
                parent.appendChild(line);
            }
        };

        JSONViewer.prototype.renderCollapsibleContainer = function(node, parent, openSymbol, closeSymbol, previewText, renderChild) {
            const openLine = document.createElement('span');
            openLine.className = 'json-line json-collapsible';

            const toggle = document.createElement('span');
            toggle.className = 'json-toggle';
            toggle.addEventListener('click', function() {
                const isCollapsed = openLine.classList.toggle('collapsed');
                children.style.display = isCollapsed ? 'none' : 'block';
            });
            openLine.appendChild(toggle);
            openLine.appendChild(createSpan(openSymbol, 'json-punctuation'));

            const preview = document.createElement('span');
            preview.className = 'json-collapsed-preview';
            preview.textContent = ' ' + previewText;
            openLine.appendChild(preview);

            const children = document.createElement('span');
            children.className = 'json-children';

            if (node.children) {
                node.children.forEach(function(child, index) {
                    renderChild(child, children);
                    if (index < node.children.length - 1) {
                        const lastLine = children.lastElementChild;
                        if (lastLine) {
                            lastLine.appendChild(createSpan(',', 'json-punctuation'));
                        }
                    }
                });
            }

            const closeLine = document.createElement('span');
            closeLine.className = 'json-line json-close-line';
            closeLine.appendChild(createSpan(closeSymbol, 'json-punctuation'));

            // 将结束括号放入 children 中，折叠时与内容一起隐藏
            children.appendChild(closeLine);

            // 将 children 放入 openLine 内，配合 collapsed 类控制显示/隐藏
            openLine.appendChild(children);
            parent.appendChild(openLine);

        };

        JSONViewer.prototype.expandAll = function() {
            const lines = this.viewer.querySelectorAll('.json-line.collapsed');
            lines.forEach(function(line) {
                line.classList.remove('collapsed');
                const children = line.querySelector('.json-children');
                if (children) {
                    children.style.display = 'block';
                }
            });
        };

        JSONViewer.prototype.collapseAll = function() {
            const lines = this.viewer.querySelectorAll('.json-collapsible');
            lines.forEach(function(line) {
                line.classList.add('collapsed');
                const children = line.querySelector('.json-children');
                if (children) {
                    children.style.display = 'none';
                }
            });
        };



        JSONViewer.prototype.copy = function() {
            // 未显示原始内容时，优先复制修复后的 JSON
            let textToCopy = this.showingRaw
                ? this.rawText
                : (this.parseResult.repairedText || this.rawText);

            const self = this;
            if (navigator.clipboard && navigator.clipboard.writeText) {
                navigator.clipboard.writeText(textToCopy).then(function() {
                    self.showButtonFeedback(self.container.querySelector('.json-viewer-toolbar button.secondary'), '复制成功', '#27ae60');
                }).catch(function(err) {
                    console.error('Clipboard API复制失败:', err);
                    self.fallbackCopy(textToCopy);
                });
            } else {
                this.fallbackCopy(textToCopy);
            }
        };

        JSONViewer.prototype.fallbackCopy = function(text) {
            const textArea = document.createElement('textarea');
            textArea.value = text;
            textArea.style.position = 'fixed';
            textArea.style.left = '-9999px';
            textArea.style.top = '0';
            document.body.appendChild(textArea);
            textArea.focus();
            textArea.select();

            try {
                const successful = document.execCommand('copy');
                const color = successful ? '#27ae60' : '#e74c3c';
                const msg = successful ? '复制成功' : '复制失败';
                this.showButtonFeedback(this.container.querySelector('.json-viewer-toolbar button.secondary'), msg, color);
            } catch (err) {
                console.error('备用复制方法失败:', err);
                this.showButtonFeedback(this.container.querySelector('.json-viewer-toolbar button.secondary'), '复制失败', '#e74c3c');
            }

            document.body.removeChild(textArea);
        };

        JSONViewer.prototype.showButtonFeedback = function(button, text, color) {
            if (!button) return;
            const originalText = button.textContent;
            const originalColor = button.style.background;
            button.textContent = text;
            button.style.background = color;
            setTimeout(function() {
                button.textContent = originalText;
                button.style.background = originalColor;
            }, 2000);
        };
    </script>
</body>
</html>`
