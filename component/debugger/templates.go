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
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .header h1 { color: #2c3e50; margin-bottom: 10px; }
        .header .stats { display: flex; gap: 20px; flex-wrap: wrap; }
        .stat-item { background: #f8f9fa; padding: 10px 15px; border-radius: 6px; border-left: 4px solid #3498db; }
        .stat-item .label { font-size: 12px; color: #666; }
        .stat-item .value { font-size: 18px; font-weight: bold; color: #2c3e50; }
        
        .filters { background: #fff; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
        .filter-form { display: flex; gap: 10px; flex-wrap: wrap; align-items: center; }
        .filter-form input, .filter-form select { padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; }
        .filter-form button { background: #3498db; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
        
        .logs-table { background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .table-header { background: #f8f9fa; padding: 15px; border-bottom: 1px solid #eee; display: grid; grid-template-columns: 100px minmax(200px, 1fr) 80px 100px 120px 100px; gap: 10px; font-weight: bold; }
        .log-row { padding: 15px; border-bottom: 1px solid #eee; display: grid; grid-template-columns: 100px minmax(200px, 1fr) 80px 100px 120px 100px; gap: 10px; align-items: center; }
        .log-row:hover { background: #f8f9fa; }
        .log-row:last-child { border-bottom: none; }
        .url { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 100%; }
        .method { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; }
        .method-get { background: #d4edda; color: #155724; }
        .method-post { background: #d1ecf1; color: #0c5460; }
        .method-put { background: #fff3cd; color: #856404; }
        .method-delete { background: #f8d7da; color: #721c24; }
        .status-code { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; }
        .status-2xx { background: #d4edda; color: #155724; }
        .status-3xx { background: #fff3cd; color: #856404; }
        .status-4xx { background: #f8d7da; color: #721c24; }
        .status-5xx { background: #f5c6cb; color: #721c24; }
        .duration { color: #666; font-size: 12px; }
        .actions a { color: #3498db; text-decoration: none; margin-right: 10px; }
        .actions a:hover { text-decoration: underline; }
        
        .pagination { display: flex; justify-content: center; gap: 5px; margin-top: 20px; flex-wrap: wrap; }
        .pagination a, .pagination span { padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; text-decoration: none; color: #333; min-width: 40px; text-align: center; }
        .pagination a:hover { background: #f8f9fa; }
        .pagination .current { background: #3498db; color: white; border-color: #3498db; }
        .pagination .disabled { color: #999; cursor: not-allowed; }
        .pagination .ellipsis { padding: 8px 6px; color: #999; }
        
        @media (max-width: 768px) {
            .pagination { gap: 3px; }
            .pagination a, .pagination span { padding: 6px 8px; min-width: 36px; font-size: 14px; }
            .pagination .ellipsis { padding: 6px 4px; }
        }
        
        .search-box { margin-bottom: 20px; }
        .search-box form { display: flex; gap: 10px; }
        .search-box input { flex: 1; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
        .search-box button { background: #27ae60; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        
        .nav { display: flex; gap: 20px; margin-bottom: 20px; }
        .nav a { color: #3498db; text-decoration: none; padding: 10px 15px; border-radius: 4px; }
        .nav a.active { background: #3498db; color: white; }
        
        @media (max-width: 768px) {
            .table-header, .log-row { grid-template-columns: 1fr; gap: 5px; }
            .header .stats { flex-direction: column; }
            .filter-form { flex-direction: column; align-items: stretch; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
            <div class="stats">
                {{if .Stats}}
                <div class="stat-item">
                    <div class="label">总请求数</div>
                    <div class="value">{{.Stats.total_requests}}</div>
                </div>
                <div class="stat-item">
                    <div class="label">平均响应时间</div>
                    <div class="value">{{.Stats.avg_duration}}ms</div>
                </div>
                <div class="stat-item">
                    <div class="label">错误率</div>
                    <div class="value">{{.Stats.error_rate}}%</div>
                </div>
                <div class="stat-item">
                    <div class="label">存储大小</div>
                    <div class="value">{{.Stats.storage_size}}</div>
                </div>
                {{end}}
            </div>
        </div>
        
        <div class="nav">
            <a href="{{.BasePath}}/list" class="active">日志列表</a>
            <a href="{{.BasePath}}/search">搜索</a>
        </div>
        
        <div class="search-box">
            <form action="{{.BasePath}}/search" method="get">
                <input type="text" name="q" placeholder="搜索日志内容..." value="">
                <button type="submit">搜索</button>
            </form>
        </div>
        
        <div class="filters">
            <form class="filter-form" method="get">
                <select name="method">
                    <option value="">所有方法</option>
                    <option value="GET" {{if eq .Filters.method "GET"}}selected{{end}}>GET</option>
                    <option value="POST" {{if eq .Filters.method "POST"}}selected{{end}}>POST</option>
                    <option value="PUT" {{if eq .Filters.method "PUT"}}selected{{end}}>PUT</option>
                    <option value="DELETE" {{if eq .Filters.method "DELETE"}}selected{{end}}>DELETE</option>
                </select>
                <input type="number" name="status_code" placeholder="状态码" value="{{.Filters.status_code}}">
                <input type="text" name="url" placeholder="URL包含" value="{{.Filters.url}}">
                <button type="submit">筛选</button>
                <a href="{{.BasePath}}/list" style="margin-left: auto;">清除筛选</a>
            </form>
        </div>
        
        <div class="logs-table">
            <div class="table-header">
                <div>时间</div>
                <div>URL</div>
                <div>方法</div>
                <div>状态码</div>
                <div>耗时</div>
                <div>操作</div>
            </div>
            
            {{range .Entries}}
            <div class="log-row">
                <div class="timestamp">{{.Timestamp.Format "2006-01-02 15:04:05"}}</div>
                <div class="url" title="{{.URL}}">{{.URL}}</div>
                <div class="method method-{{lower .Method}}">{{.Method}}</div>
                <div class="status-code status-{{if ge .StatusCode 200}}{{if lt .StatusCode 300}}2xx{{else if lt .StatusCode 400}}3xx{{else if lt .StatusCode 500}}4xx{{else}}5xx{{end}}{{end}}">{{.StatusCode}}</div>
                <div class="duration">{{.Duration.Milliseconds}}ms</div>
                <div class="actions">
                    <a href="{{$.BasePath}}/detail/{{.ID}}">详情</a>
                    <a href="{{$.BasePath}}/api/logs/{{.ID}}" target="_blank">JSON</a>
                </div>
            </div>
            {{else}}
            <div class="log-row" style="text-align: center; padding: 40px;">
                暂无日志记录
            </div>
            {{end}}
        </div>
        
        {{if .Pagination}}
        <div class="pagination">
            {{if .Pagination.HasPrev}}
            <a href="{{.BasePath}}/list?page={{.Pagination.PrevPage}}&pageSize={{.Pagination.PageSize}}{{if .Filters.method}}&method={{.Filters.method}}{{end}}{{if .Filters.status_code}}&status_code={{.Filters.status_code}}{{end}}{{if .Filters.start_time}}&start_time={{.Filters.start_time}}{{end}}{{if .Filters.end_time}}&end_time={{.Filters.end_time}}{{end}}{{if .Filters.url}}&url={{.Filters.url}}{{end}}">上一页</a>
            {{else}}
            <span class="disabled">上一页</span>
            {{end}}
            
            {{$page := .Pagination.Page}}
            {{$totalPages := .Pagination.TotalPages}}
            {{$basePath := .BasePath}}
            {{$pageSize := .Pagination.PageSize}}
            {{$filters := .Filters}}
            
            {{/* 智能分页显示逻辑 */}}
            {{if le $totalPages 7}}
                {{/* 总页数小于等于7时，显示所有页码 */}}
                {{range $i := seq 1 $totalPages}}
                {{if eq $i $page}}
                <span class="current">{{$i}}</span>
                {{else}}
                <a href="{{$basePath}}/list?page={{$i}}&pageSize={{$pageSize}}{{if $filters.method}}&method={{$filters.method}}{{end}}{{if $filters.status_code}}&status_code={{$filters.status_code}}{{end}}{{if $filters.start_time}}&start_time={{$filters.start_time}}{{end}}{{if $filters.end_time}}&end_time={{$filters.end_time}}{{end}}{{if $filters.url}}&url={{$filters.url}}{{end}}">{{$i}}</a>
                {{end}}
                {{end}}
            {{else}}
                {{/* 总页数大于7时，使用智能分页 */}}
                {{if gt $page 4}}
                    <a href="{{$basePath}}/list?page=1&pageSize={{$pageSize}}{{if $filters.method}}&method={{$filters.method}}{{end}}{{if $filters.status_code}}&status_code={{$filters.status_code}}{{end}}{{if $filters.start_time}}&start_time={{$filters.start_time}}{{end}}{{if $filters.end_time}}&end_time={{$filters.end_time}}{{end}}{{if $filters.url}}&url={{$filters.url}}{{end}}">1</a>
                    {{if gt $page 5}}
                    <span class="ellipsis">...</span>
                    {{end}}
                {{end}}
                
                {{/* 显示当前页附近的页码 */}}
                {{$start := max 1 (sub $page 2)}}
                {{$end := min $totalPages (add $page 2)}}
                {{range $i := seq $start $end}}
                {{if eq $i $page}}
                <span class="current">{{$i}}</span>
                {{else}}
                <a href="{{$basePath}}/list?page={{$i}}&pageSize={{$pageSize}}{{if $filters.method}}&method={{$filters.method}}{{end}}{{if $filters.status_code}}&status_code={{$filters.status_code}}{{end}}{{if $filters.start_time}}&start_time={{$filters.start_time}}{{end}}{{if $filters.end_time}}&end_time={{$filters.end_time}}{{end}}{{if $filters.url}}&url={{$filters.url}}{{end}}">{{$i}}</a>
                {{end}}
                {{end}}
                
                {{if lt $page (sub $totalPages 3)}}
                    {{if lt $page (sub $totalPages 4)}}
                    <span class="ellipsis">...</span>
                    {{end}}
                    <a href="{{$basePath}}/list?page={{$totalPages}}&pageSize={{$pageSize}}{{if $filters.method}}&method={{$filters.method}}{{end}}{{if $filters.status_code}}&status_code={{$filters.status_code}}{{end}}{{if $filters.start_time}}&start_time={{$filters.start_time}}{{end}}{{if $filters.end_time}}&end_time={{$filters.end_time}}{{end}}{{if $filters.url}}&url={{$filters.url}}{{end}}">{{$totalPages}}</a>
                {{end}}
            {{end}}
            
            {{if .Pagination.HasNext}}
            <a href="{{.BasePath}}/list?page={{.Pagination.NextPage}}&pageSize={{.Pagination.PageSize}}{{if .Filters.method}}&method={{.Filters.method}}{{end}}{{if .Filters.status_code}}&status_code={{.Filters.status_code}}{{end}}{{if .Filters.start_time}}&start_time={{.Filters.start_time}}{{end}}{{if .Filters.end_time}}&end_time={{.Filters.end_time}}{{end}}{{if .Filters.url}}&url={{.Filters.url}}{{end}}">下一页</a>
            {{else}}
            <span class="disabled">下一页</span>
            {{end}}
        </div>
        {{end}}
    </div>
    
    <script>
        function lower(str) {
            return str ? str.toLowerCase() : '';
        }
        
        function seq(start, end) {
            const result = [];
            for (let i = start; i <= end; i++) {
                result.push(i);
            }
            return result;
        }
        
        function max(a, b) {
            return a > b ? a : b;
        }
        
        function min(a, b) {
            return a < b ? a : b;
        }
        
        function sub(a, b) {
            return a - b;
        }
        
        function add(a, b) {
            return a + b;
        }
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
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .header h1 { color: #2c3e50; margin-bottom: 10px; }
        .back-link { color: #3498db; text-decoration: none; margin-bottom: 10px; display: inline-block; }
        .back-link:hover { text-decoration: underline; }
        
        .detail-sections { display: grid; gap: 20px; }
        .section { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .section h2 { color: #2c3e50; margin-bottom: 15px; padding-bottom: 10px; border-bottom: 1px solid #eee; }
        
        .basic-info { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
        .info-item { display: flex; flex-direction: column; }
        .info-label { font-size: 12px; color: #666; margin-bottom: 5px; }
        .info-value { font-size: 14px; color: #333; word-break: break-all; }
        
        .headers-table, .params-table { width: 100%; border-collapse: collapse; }
        .headers-table th, .params-table th { background: #f8f9fa; padding: 10px; text-align: left; border-bottom: 1px solid #eee; }
        .headers-table td, .params-table td { padding: 10px; border-bottom: 1px solid #eee; }
        .headers-table tr:last-child td, .params-table tr:last-child td { border-bottom: none; }
        
        .json-viewer { background: #f8f9fa; border: 1px solid #eee; border-radius: 4px; padding: 15px; max-height: 400px; overflow: auto; font-family: 'Courier New', monospace; font-size: 12px; }
        .json-viewer pre { margin: 0; white-space: pre-wrap; }
        
        .tab-container { margin-top: 20px; }
        .tabs { display: flex; border-bottom: 1px solid #eee; margin-bottom: 15px; }
        .tab { padding: 10px 20px; cursor: pointer; border: 1px solid transparent; border-bottom: none; margin-bottom: -1px; }
        .tab.active { background: #fff; border-color: #eee; border-bottom-color: #fff; border-radius: 4px 4px 0 0; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        
        .method-badge, .status-badge { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .method-get { background: #d4edda; color: #155724; }
        .method-post { background: #d1ecf1; color: #0c5460; }
        .status-2xx { background: #d4edda; color: #155724; }
        .status-4xx { background: #f8d7da; color: #721c24; }
        .status-5xx { background: #f5c6cb; color: #721c24; }
        
        @media (max-width: 768px) {
            .basic-info { grid-template-columns: 1fr; }
            .headers-table, .params-table { font-size: 12px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="javascript:history.back()" class="back-link" id="back-link">← 返回上一页</a>
        <a href="{{.BasePath}}/list" class="back-link" id="fallback-link" style="display: none;">← 返回日志列表</a>
        
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        
        {{if .Entry}}
        <div class="detail-sections">
            <!-- 基本信息 -->
            <div class="section">
                <h2>基本信息</h2>
                <div class="basic-info">
                    <div class="info-item">
                        <div class="info-label">请求ID</div>
                        <div class="info-value">{{.Entry.ID}}</div>
                    </div>
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
                        <div class="info-value status-badge status-{{if ge .Entry.StatusCode 200}}{{if lt .Entry.StatusCode 300}}2xx{{else if lt .Entry.StatusCode 400}}3xx{{else if lt .Entry.StatusCode 500}}4xx{{else}}5xx{{end}}{{end}}">{{.Entry.StatusCode}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">耗时</div>
                        <div class="info-value">{{.Entry.Duration.Milliseconds}}ms</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">客户端IP</div>
                        <div class="info-value">{{.Entry.ClientIP}}</div>
                    </div>
                </div>
            </div>
            
            <!-- URL和参数 -->
            <div class="section">
                <h2>请求信息</h2>
                <div class="basic-info">
                    <div class="info-item">
                        <div class="info-label">URL</div>
                        <div class="info-value">{{.Entry.URL}}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">User Agent</div>
                        <div class="info-value">{{.Entry.UserAgent}}</div>
                    </div>
                </div>
                
                {{if .Entry.RequestHeaders}}
                <div style="margin-top: 15px;">
                    <h3>请求头</h3>
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
                {{end}}
                
                {{if .Entry.QueryParams}}
                <div style="margin-top: 15px;">
                    <h3>查询参数</h3>
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
                {{end}}
            </div>
            
            <!-- 请求体和响应体 -->
            <div class="section">
                <h2>请求和响应</h2>
                <div class="tab-container">
                    <div class="tabs">
                        <div class="tab active" onclick="showTab('request')">请求体</div>
                        <div class="tab" onclick="showTab('response')">响应体</div>
                    </div>
                    
                    <div id="request" class="tab-content active">
                        <div class="json-viewer">
                            <pre>{{.Entry.RequestBody | html}}</pre>
                        </div>
                    </div>
                    
                    <div id="response" class="tab-content">
                        <div class="json-viewer">
                            <pre>{{.Entry.ResponseBody | html}}</pre>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        {{else}}
        <div class="section">
            <h2>日志记录不存在</h2>
            <p>请求的日志记录不存在或已被删除。</p>
        </div>
        {{end}}
    </div>
    
    <script>
        // 页面加载时检查历史记录
        document.addEventListener('DOMContentLoaded', function() {
            const backLink = document.getElementById('back-link');
            const fallbackLink = document.getElementById('fallback-link');
            
            // 检查是否有历史记录可以返回
            if (history.length <= 1) {
                // 没有历史记录，显示备用链接
                backLink.style.display = 'none';
                fallbackLink.style.display = 'inline-block';
            }
            
            // 为返回链接添加点击事件处理
            backLink.addEventListener('click', function(e) {
                e.preventDefault();
                
                // 尝试返回上一页
                if (history.length > 1) {
                    history.back();
                } else {
                    // 如果没有历史记录，跳转到列表页
                    window.location.href = '{{.BasePath}}/list';
                }
            });
        });
        
        function lower(str) {
            return str ? str.toLowerCase() : '';
        }
        
        function showTab(tabName) {
            // 隐藏所有标签内容
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // 移除所有标签的激活状态
            document.querySelectorAll('.tab').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // 显示选中的标签内容
            document.getElementById(tabName).classList.add('active');
            
            // 激活选中的标签
            event.target.classList.add('active');
        }
    </script>
</body>
</html>`

// searchTemplate 搜索页面模板
const searchTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .header h1 { color: #2c3e50; margin-bottom: 10px; }
        
        .search-box { margin-bottom: 20px; }
        .search-box form { display: flex; gap: 10px; }
        .search-box input { flex: 1; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
        .search-box button { background: #27ae60; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        
        .logs-table { background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .table-header { background: #f8f9fa; padding: 15px; border-bottom: 1px solid #eee; display: grid; grid-template-columns: 100px minmax(200px, 1fr) 80px 100px 120px 100px; gap: 10px; font-weight: bold; }
        .log-row { padding: 15px; border-bottom: 1px solid #eee; display: grid; grid-template-columns: 100px minmax(200px, 1fr) 80px 100px 120px 100px; gap: 10px; align-items: center; }
        .log-row:hover { background: #f8f9fa; }
        .log-row:last-child { border-bottom: none; }
        .method { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; }
        .method-get { background: #d4edda; color: #155724; }
        .method-post { background: #d1ecf1; color: #0c5460; }
        .method-put { background: #fff3cd; color: #856404; }
        .method-delete { background: #f8d7da; color: #721c24; }
        .status-code { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-align: center; }
        .status-2xx { background: #d4edda; color: #155724; }
        .status-3xx { background: #fff3cd; color: #856404; }
        .status-4xx { background: #f8d7da; color: #721c24; }
        .status-5xx { background: #f5c6cb; color: #721c24; }
        .duration { color: #666; font-size: 12px; }
        .actions a { color: #3498db; text-decoration: none; margin-right: 10px; }
        .actions a:hover { text-decoration: underline; }
        
        .pagination { display: flex; justify-content: center; gap: 10px; margin-top: 20px; }
        .pagination a, .pagination span { padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; text-decoration: none; color: #333; }
        .pagination a:hover { background: #f8f9fa; }
        .pagination .current { background: #3498db; color: white; border-color: #3498db; }
        .pagination .disabled { color: #999; cursor: not-allowed; }
        
        .nav { display: flex; gap: 20px; margin-bottom: 20px; }
        .nav a { color: #3498db; text-decoration: none; padding: 10px 15px; border-radius: 4px; }
        .nav a.active { background: #27ae60; color: white; }
        
        @media (max-width: 768px) {
            .table-header, .log-row { grid-template-columns: 1fr; gap: 5px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        
        <div class="nav">
            <a href="{{.BasePath}}/list">日志列表</a>
            <a href="{{.BasePath}}/search" class="active">搜索</a>
        </div>
        
        <div class="search-box">
            <form action="{{.BasePath}}/search" method="get">
                <input type="text" name="q" placeholder="搜索日志内容..." value="{{.Keyword}}">
                <button type="submit">搜索</button>
            </form>
        </div>
        
        <div class="logs-table">
            <div class="table-header">
                <div>时间</div>
                <div>URL</div>
                <div>方法</div>
                <div>状态码</div>
                <div>耗时</div>
                <div>操作</div>
            </div>
            
            {{range .Entries}}
            <div class="log-row">
                <div class="timestamp">{{.Timestamp.Format "2006-01-02 15:04:05"}}</div>
                <div class="url" title="{{.URL}}">{{.URL}}</div>
                <div class="method method-{{lower .Method}}">{{.Method}}</div>
                <div class="status-code status-{{if ge .StatusCode 200}}{{if lt .StatusCode 300}}2xx{{else if lt .StatusCode 400}}3xx{{else if lt .StatusCode 500}}4xx{{else}}5xx{{end}}{{end}}">{{.StatusCode}}</div>
                <div class="duration">{{.Duration.Milliseconds}}ms</div>
                <div class="actions">
                    <a href="{{$.BasePath}}/detail/{{.ID}}">详情</a>
                    <a href="{{$.BasePath}}/api/logs/{{.ID}}" target="_blank">JSON</a>
                </div>
            </div>
            {{else}}
            {{if .Keyword}}
            <div class="log-row" style="text-align: center; padding: 40px;">
                未找到匹配 "{{.Keyword}}" 的日志记录
            </div>
            {{else}}
            <div class="log-row" style="text-align: center; padding: 40px;">
                请输入搜索关键词
            </div>
            {{end}}
            {{end}}
        </div>
        
        {{if and .Pagination .Keyword}}
        <div class="pagination">
            {{if .Pagination.HasPrev}}
            <a href="{{.BasePath}}/search?q={{.Keyword}}&page={{.Pagination.PrevPage}}&pageSize={{.Pagination.PageSize}}">上一页</a>
            {{else}}
            <span class="disabled">上一页</span>
            {{end}}
            
            {{$page := .Pagination.Page}}
            {{$totalPages := .Pagination.TotalPages}}
            {{$basePath := .BasePath}}
            {{$pageSize := .Pagination.PageSize}}
            {{$keyword := .Keyword}}
            
            {{range $i := seq 1 $totalPages}}
            {{if eq $i $page}}
            <span class="current">{{$i}}</span>
            {{else}}
            <a href="{{$basePath}}/search?q={{$keyword}}&page={{$i}}&pageSize={{$pageSize}}">{{$i}}</a>
            {{end}}
            {{end}}
            
            {{if .Pagination.HasNext}}
            <a href="{{.BasePath}}/search?q={{.Keyword}}&page={{.Pagination.NextPage}}&pageSize={{.Pagination.PageSize}}">下一页</a>
            {{else}}
            <span class="disabled">下一页</span>
            {{end}}
        </div>
        {{end}}
    </div>
    
    <script>
        function lower(str) {
            return str ? str.toLowerCase() : '';
        }
        
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
