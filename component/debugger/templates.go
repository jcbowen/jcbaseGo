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
        .header h1 { color: #2c3e50; margin-bottom: 10px; font-size: 24px; word-break: break-word; }
        .header .stats { display: flex; gap: 20px; flex-wrap: wrap; }
        .stat-item { background: #f8f9fa; padding: 10px 15px; border-radius: 6px; border-left: 4px solid #3498db; }
        .stat-item .label { font-size: 12px; color: #666; }
        .stat-item .value { font-size: 18px; font-weight: bold; color: #2c3e50; }
        
        .filters { background: #fff; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .filter-form { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 16px; align-items: end; }
        .filter-form input, .filter-form select { 
            padding: 10px 12px; 
            border: 1px solid #ddd; 
            border-radius: 6px; 
            font-size: 14px; 
            transition: border-color 0.2s ease;
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
        }
        .filter-form button:hover { background: #2980b9; }
        .filter-form a { 
            color: #666; 
            text-decoration: none; 
            font-size: 14px; 
            padding: 10px 0;
            text-align: center;
        }
        .filter-form a:hover { color: #3498db; }
        
        .logs-table { background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .table-container { overflow-x: auto; -webkit-overflow-scrolling: touch; }
        .table-content { min-width: 800px; }
        .table-header { background: #f8f9fa; padding: 15px; border-bottom: 1px solid #eee; display: grid; grid-template-columns: minmax(140px, 220px) 160px 100px 120px 140px 100px 100px minmax(250px, 1fr); gap: 16px; font-weight: bold; font-size: 14px; }
        .log-row { padding: 15px; border-bottom: 1px solid #eee; display: grid; grid-template-columns: minmax(140px, 220px) 160px 100px 120px 140px 100px 100px minmax(250px, 1fr); gap: 16px; align-items: center; font-size: 14px; }
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
        
        .search-box { margin-bottom: 20px; }
        .search-box form { display: flex; gap: 10px; }
        .search-box input { flex: 1; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
        .search-box button { background: #27ae60; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        
        .nav { display: flex; gap: 20px; margin-bottom: 20px; }
        .nav a { color: #3498db; text-decoration: none; padding: 10px 15px; border-radius: 4px; }
        .nav a.active { background: #3498db; color: white; }
        
        @media (max-width: 768px) {
            .container { padding: 10px; }
            .header { padding: 15px; margin-bottom: 15px; }
            .header h1 { font-size: 20px; }
            .header .stats { flex-direction: column; gap: 10px; }
            .stat-item { padding: 8px 12px; }
            .stat-item .value { font-size: 16px; }
            
            .filter-form { flex-direction: column; align-items: stretch; gap: 8px; }
            .filter-form input, .filter-form select { width: 100%; }
            
            .table-container { overflow-x: auto; -webkit-overflow-scrolling: touch; }
            .table-content { min-width: 600px; }
            .table-header, .log-row { 
                grid-template-columns: minmax(80px, 120px) 100px 60px 80px 90px 60px 60px minmax(120px, 1fr); 
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
                grid-template-columns: minmax(70px, 100px) 80px 50px 70px 80px 50px 50px minmax(100px, 1fr); 
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
            .header h1 { font-size: 28px; }
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
        </div>
        
        <div class="search-box">
            <form action="{{.BasePath}}/list" method="get">
                <input type="text" name="q" placeholder="搜索日志内容..." value="{{.Keyword}}">
                <button type="submit">搜索</button>
                {{if .Keyword}}
                <a href="{{.BasePath}}/list" style="margin-left: 10px; color: #666; text-decoration: none; padding: 10px 15px; border: 1px solid #ddd; border-radius: 4px;">清除搜索</a>
                {{end}}
            </form>
        </div>
        
        <div class="filters">
            <form class="filter-form" method="get">
                <select name="method" onchange="this.form.submit()">
                    <option value="">所有方法</option>
                    <option value="GET" {{if eq .Filters.method "GET"}}selected{{end}}>GET</option>
                    <option value="POST" {{if eq .Filters.method "POST"}}selected{{end}}>POST</option>
                    <option value="PUT" {{if eq .Filters.method "PUT"}}selected{{end}}>PUT</option>
                    <option value="DELETE" {{if eq .Filters.method "DELETE"}}selected{{end}}>DELETE</option>
                </select>
                <select name="status_code" onchange="this.form.submit()">
                    <option value="">所有状态码</option>
                    <option value="200" {{if eq .Filters.status_code "200"}}selected{{end}}>200 - 成功</option>
                    <option value="201" {{if eq .Filters.status_code "201"}}selected{{end}}>201 - 已创建</option>
                    <option value="204" {{if eq .Filters.status_code "204"}}selected{{end}}>204 - 无内容</option>
                    <option value="301" {{if eq .Filters.status_code "301"}}selected{{end}}>301 - 永久重定向</option>
                    <option value="302" {{if eq .Filters.status_code "302"}}selected{{end}}>302 - 临时重定向</option>
                    <option value="400" {{if eq .Filters.status_code "400"}}selected{{end}}>400 - 错误请求</option>
                    <option value="401" {{if eq .Filters.status_code "401"}}selected{{end}}>401 - 未授权</option>
                    <option value="403" {{if eq .Filters.status_code "403"}}selected{{end}}>403 - 禁止访问</option>
                    <option value="404" {{if eq .Filters.status_code "404"}}selected{{end}}>404 - 未找到</option>
                    <option value="500" {{if eq .Filters.status_code "500"}}selected{{end}}>500 - 服务器错误</option>
                    <option value="502" {{if eq .Filters.status_code "502"}}selected{{end}}>502 - 网关错误</option>
                    <option value="503" {{if eq .Filters.status_code "503"}}selected{{end}}>503 - 服务不可用</option>
                </select>
                <input type="text" name="client_ip" placeholder="IP地址" value="{{.Filters.client_ip}}">
                <input type="text" name="url" placeholder="URL包含" value="{{.Filters.url}}">
                <button type="submit">筛选</button>
                <a href="{{.BasePath}}/list" style="margin-left: auto;">清除筛选</a>
            </form>
        </div>
        
        <div class="logs-table">
            <div class="table-container">
                <div class="table-content">
                    <div class="table-header">
                        <div>请求Id</div>
                        <div>时间</div>
                        <div>耗时</div>
                        <div>存储大小</div>
                        <div>IP</div>
                        <div>方法</div>
                        <div>状态</div>
                        <div>URL</div>
                    </div>
                    
                    {{range .Entries}}
                    <div class="log-row">
                        <div class="request-id"><a href="{{$.BasePath}}/detail/{{.ID}}" title="查看详情">{{.ID}}</a></div>
                        <div class="timestamp">{{.Timestamp.Format "2006-01-02 15:04:05"}}</div>
                        <div class="duration">{{.Duration.Milliseconds}}ms</div>
                        <div class="storage-size">{{.StorageSize}}</div>
                        <div class="client-ip">{{.ClientIP}}</div>
                        <div class="method method-{{lower .Method}}">{{.Method}}</div>
                        <div class="status-code status-{{if ge .StatusCode 200}}{{if lt .StatusCode 300}}2xx{{else if lt .StatusCode 400}}3xx{{else if lt .StatusCode 500}}4xx{{else}}5xx{{end}}{{end}}">{{.StatusCode}}</div>
                        <div class="url" title="{{.URL}}">{{.URL}}</div>
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
        <div class="pagination">
            {{if .Pagination.HasPrev}}
            <a href="{{.BasePath}}/list?page={{.Pagination.PrevPage}}&pageSize={{.Pagination.PageSize}}{{if .Keyword}}&q={{.Keyword}}{{end}}{{if .Filters.method}}&method={{.Filters.method}}{{end}}{{if .Filters.status_code}}&status_code={{.Filters.status_code}}{{end}}{{if .Filters.client_ip}}&client_ip={{.Filters.client_ip}}{{end}}{{if .Filters.start_time}}&start_time={{.Filters.start_time}}{{end}}{{if .Filters.end_time}}&end_time={{.Filters.end_time}}{{end}}{{if .Filters.url}}&url={{.Filters.url}}{{end}}">上一页</a>
            {{else}}
            <span class="disabled">上一页</span>
            {{end}}
            
            {{$page := .Pagination.Page}}
            {{$totalPages := .Pagination.TotalPages}}
            {{$basePath := .BasePath}}
            {{$pageSize := .Pagination.PageSize}}
            {{$filters := .Filters}}
            {{$keyword := .Keyword}}
            
            {{/* 智能分页显示逻辑 */}}
            {{if le $totalPages 7}}
                {{/* 总页数小于等于7时，显示所有页码 */}}
                {{range $i := seq 1 $totalPages}}
                {{if eq $i $page}}
                <span class="current">{{$i}}</span>
                {{else}}
                <a href="{{$basePath}}/list?page={{$i}}&pageSize={{$pageSize}}{{if $keyword}}&q={{$keyword}}{{end}}{{if $filters.method}}&method={{$filters.method}}{{end}}{{if $filters.status_code}}&status_code={{$filters.status_code}}{{end}}{{if $filters.client_ip}}&client_ip={{$filters.client_ip}}{{end}}{{if $filters.start_time}}&start_time={{$filters.start_time}}{{end}}{{if $filters.end_time}}&end_time={{$filters.end_time}}{{end}}{{if $filters.url}}&url={{$filters.url}}{{end}}">{{$i}}</a>
                {{end}}
                {{end}}
            {{else}}
                {{/* 总页数大于7时，使用智能分页 */}}
                {{if gt $page 4}}
                    <a href="{{$basePath}}/list?page=1&pageSize={{$pageSize}}{{if $keyword}}&q={{$keyword}}{{end}}{{if $filters.method}}&method={{$filters.method}}{{end}}{{if $filters.status_code}}&status_code={{$filters.status_code}}{{end}}{{if $filters.client_ip}}&client_ip={{$filters.client_ip}}{{end}}{{if $filters.start_time}}&start_time={{$filters.start_time}}{{end}}{{if $filters.end_time}}&end_time={{$filters.end_time}}{{end}}{{if $filters.url}}&url={{$filters.url}}{{end}}">1</a>
                    {{if gt $page 5}}
                    <span class="ellipsis">...</span>
                    {{end}}
                {{end}}
                
                {{/* 显示当前页附近的页码 */}}
                {{$start := 1}}
                {{if gt $page 2}}
                    {{$start = sub $page 2}}
                {{end}}
                {{$end := $totalPages}}
                {{if lt $page $totalPages}}
                    {{if lt $page (sub $totalPages 2)}}
                        {{$end = add $page 2}}
                    {{end}}
                {{end}}
                {{range $i := seq $start $end}}
                {{if eq $i $page}}
                <span class="current">{{$i}}</span>
                {{else}}
                <a href="{{$basePath}}/list?page={{$i}}&pageSize={{$pageSize}}{{if $keyword}}&q={{$keyword}}{{end}}{{if $filters.method}}&method={{$filters.method}}{{end}}{{if $filters.status_code}}&status_code={{$filters.status_code}}{{end}}{{if $filters.client_ip}}&client_ip={{$filters.client_ip}}{{end}}{{if $filters.start_time}}&start_time={{$filters.start_time}}{{end}}{{if $filters.end_time}}&end_time={{$filters.end_time}}{{end}}{{if $filters.url}}&url={{$filters.url}}{{end}}">{{$i}}</a>
                {{end}}
                {{end}}
                
                {{if lt $page $totalPages}}
                    {{if lt $page (sub $totalPages 3)}}
                        {{if lt $page (sub $totalPages 4)}}
                        <span class="ellipsis">...</span>
                        {{end}}
                        <a href="{{$basePath}}/list?page={{$totalPages}}&pageSize={{$pageSize}}{{if $keyword}}&q={{$keyword}}{{end}}{{if $filters.method}}&method={{$filters.method}}{{end}}{{if $filters.status_code}}&status_code={{$filters.status_code}}{{end}}{{if $filters.client_ip}}&client_ip={{$filters.client_ip}}{{end}}{{if $filters.start_time}}&start_time={{$filters.start_time}}{{end}}{{if $filters.end_time}}&end_time={{$filters.end_time}}{{end}}{{if $filters.url}}&url={{$filters.url}}{{end}}">{{$totalPages}}</a>
                    {{end}}
                {{end}}
            {{end}}
            
            {{if .Pagination.HasNext}}
            <a href="{{.BasePath}}/list?page={{.Pagination.NextPage}}&pageSize={{.Pagination.PageSize}}{{if .Keyword}}&q={{.Keyword}}{{end}}{{if .Filters.method}}&method={{.Filters.method}}{{end}}{{if .Filters.status_code}}&status_code={{.Filters.status_code}}{{end}}{{if .Filters.client_ip}}&client_ip={{.Filters.client_ip}}{{end}}{{if .Filters.start_time}}&start_time={{.Filters.start_time}}{{end}}{{if .Filters.end_time}}&end_time={{.Filters.end_time}}{{end}}{{if .Filters.url}}&url={{.Filters.url}}{{end}}">下一页</a>
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
            word-break: break-word;
            line-height: 1.4;
            flex: 1;
        }
        
        .headers-table, .params-table { width: 100%; border-collapse: collapse; }
        .headers-table th, .params-table th { background: #f8f9fa; padding: 10px; text-align: left; border-bottom: 1px solid #eee; }
        .headers-table td, .params-table td { 
            padding: 10px; 
            border-bottom: 1px solid #eee; 
            max-width: 300px; 
            overflow: visible; 
            white-space: normal; 
            word-break: break-word;
            line-height: 1.4;
        }
        .headers-table td:first-child, .params-table td:first-child {
            font-weight: 600;
            background: #f8f9fa;
            min-width: 120px;
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
        
        .json-viewer { background: #f8f9fa; border: 1px solid #eee; border-radius: 4px; padding: 15px; max-height: 400px; overflow: auto; font-family: 'Courier New', monospace; font-size: 12px; }
        .json-viewer pre { margin: 0; white-space: pre-wrap; }
        
        .tab-container { margin-top: 20px; }
        .tabs { display: flex; border-bottom: 1px solid #eee; margin-bottom: 15px; }
        .tab { padding: 10px 20px; cursor: pointer; border: 1px solid transparent; border-bottom: none; margin-bottom: -1px; }
        .tab.active { background: #fff; border-color: #eee; border-bottom-color: #fff; border-radius: 4px 4px 0 0; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        
        .method-badge, .status-badge { padding: 8px 8px; border-radius: 4px; font-weight: bold; display: inline-block; min-width: 50px; }
        .method-get { background: #d4edda; color: #155724; }
        .method-post { background: #d1ecf1; color: #0c5460; }
        .method-put { background: #fff3cd; color: #856404; }
        .method-delete { background: #f8d7da; color: #721c24; }
        .method-patch { background: #e2e3e5; color: #383d41; }
        .method-head { background: #d1ecf1; color: #0c5460; }
        .method-options { background: #e2e3e5; color: #383d41; }
        .status-2xx { background: #d4edda; color: #155724; }
        .status-3xx { background: #d1ecf1; color: #0c5460; }
        .status-4xx { background: #f8d7da; color: #721c24; }
        .status-5xx { background: #f5c6cb; color: #721c24; }
        
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
        .log-message { font-size: 14px; color: #333; margin-bottom: 8px; }
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
            .info-value { font-size: 14px; line-height: 1.4; }
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
                min-width: 100px;
            }
            .json-viewer { padding: 10px; font-size: 12px; max-height: 300px; }
            .log-header { flex-direction: column; align-items: flex-start; gap: 6px; }
            .log-item { padding: 10px; margin-bottom: 8px; }
            .log-message { font-size: 14px; line-height: 1.4; }
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
            .info-value { font-size: 13px; line-height: 1.4; }
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
                min-width: 80px;
            }
            .json-viewer { padding: 8px; font-size: 11px; max-height: 250px; }
            .log-item { padding: 8px; margin-bottom: 8px; }
            .log-message { font-size: 13px; line-height: 1.4; }
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
                padding: 10px 12px; 
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
    </style>
</head>
<body>
    <div class="container">
        <a href="javascript:history.back()" class="back-link" id="back-link">← 返回上一页</a>
        <a href="{{.BasePath}}/list" class="back-link" id="fallback-link" style="display: none;">← 返回日志列表</a>
        
        <div class="header">
            <h1>{{.Title}} <a href="{{.BasePath}}/api/logs/{{.Entry.ID}}" target="_blank" class="json-view-link" title="查看JSON数据">[JSON]</a></h1>
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
                        <div class="info-value status-badge status-{{if ge .Entry.StatusCode 200}}{{if lt .Entry.StatusCode 300}}2xx{{else if lt .Entry.StatusCode 400}}3xx{{else if lt .Entry.StatusCode 500}}4xx{{else}}5xx{{end}}{{else}}4xx{{end}}">{{.Entry.StatusCode}}</div>
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
            
            <!-- 响应体 -->
            {{if .Entry.ResponseBody}}
            <div class="section">
                <h2>响应信息</h2>
                <div class="json-viewer">
                    <pre>{{.Entry.ResponseBody | html}}</pre>
                </div>
            </div>
            {{end}}

            <!-- Logger -->
            {{if .Entry.LoggerLogs}}
            <div class="section">
                <h2>Logger</h2>
                <div class="logger-logs">
                    {{range .Entry.LoggerLogs}}
                    <div class="log-item">
                        <div class="log-header">
                        <span class="log-level level-{{.Level}}">{{.Level}}</span>
                        <span class="log-timestamp">{{.Timestamp.Format "2006-01-02 15:04:05.000"}}</span>
                    </div>
                        <div class="log-message">{{.Message}}</div>
                        {{if .Fields}}
                        <div class="log-fields">
                            {{range $key, $value := .Fields}}
                            {{/* 过滤掉在基本信息区域已经展示过的字段 */}}
                            {{if and (ne $key "level") (ne $key "message") (ne $key "timestamp") (ne $key "request_id") (ne $key "method") (ne $key "url") (ne $key "client_ip")}}
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
    </script>
</body>
</html>`
