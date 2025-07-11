package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== PHP 自定义函数示例 ===")

	// 示例1: 基本自定义函数
	fmt.Println("\n1. 基本自定义函数:")

	// 定义简单的数学计算函数
	mathFunctions := `
function calculateSum($a, $b) {
    return $a + $b;
}

function calculateProduct($a, $b) {
    return $a * $b;
}

function calculateAverage($numbers) {
    if (empty($numbers)) {
        return 0;
    }
    return array_sum($numbers) / count($numbers);
}
`
	// 将函数写入临时文件（这里只是演示，实际使用中需要将函数添加到PHP脚本中）
	fmt.Println("定义数学计算函数:")
	fmt.Println(mathFunctions)

	// 示例2: 字符串处理函数
	fmt.Println("\n2. 字符串处理函数:")

	stringFunctions := `
function formatName($firstName, $lastName) {
    return ucfirst(strtolower($firstName)) . ' ' . ucfirst(strtolower($lastName));
}

function generateSlug($text) {
    $text = strtolower($text);
    $text = preg_replace('/[^a-z0-9\s-]/', '', $text);
    $text = preg_replace('/[\s-]+/', '-', $text);
    return trim($text, '-');
}

function truncateText($text, $length = 100, $suffix = '...') {
    if (strlen($text) <= $length) {
        return $text;
    }
    return substr($text, 0, $length) . $suffix;
}
`
	fmt.Println("定义字符串处理函数:")
	fmt.Println(stringFunctions)

	// 示例3: 数组处理函数
	fmt.Println("\n3. 数组处理函数:")

	arrayFunctions := `
function filterArray($array, $callback) {
    return array_filter($array, $callback);
}

function sortArrayByKey($array, $key, $direction = 'asc') {
    usort($array, function($a, $b) use ($key, $direction) {
        if ($direction === 'asc') {
            return $a[$key] <=> $b[$key];
        } else {
            return $b[$key] <=> $a[$key];
        }
    });
    return $array;
}

function flattenArray($array) {
    $result = [];
    foreach ($array as $item) {
        if (is_array($item)) {
            $result = array_merge($result, flattenArray($item));
        } else {
            $result[] = $item;
        }
    }
    return $result;
}
`
	fmt.Println("定义数组处理函数:")
	fmt.Println(arrayFunctions)

	// 示例4: 日期时间处理函数
	fmt.Println("\n4. 日期时间处理函数:")

	dateFunctions := `
function formatDate($date, $format = 'Y-m-d H:i:s') {
    if (is_string($date)) {
        $date = new DateTime($date);
    }
    return $date->format($format);
}

function getDateDifference($date1, $date2) {
    $d1 = new DateTime($date1);
    $d2 = new DateTime($date2);
    $diff = $d1->diff($d2);
    return [
        'days' => $diff->days,
        'hours' => $diff->h,
        'minutes' => $diff->i,
        'seconds' => $diff->s
    ];
}

function isWeekend($date) {
    $dayOfWeek = date('N', strtotime($date));
    return $dayOfWeek >= 6;
}
`
	fmt.Println("定义日期时间处理函数:")
	fmt.Println(dateFunctions)

	// 示例5: 文件处理函数
	fmt.Println("\n5. 文件处理函数:")

	fileFunctions := `
function getFileInfo($filePath) {
    if (!file_exists($filePath)) {
        return null;
    }
    
    return [
        'name' => basename($filePath),
        'size' => filesize($filePath),
        'modified' => date('Y-m-d H:i:s', filemtime($filePath)),
        'extension' => pathinfo($filePath, PATHINFO_EXTENSION),
        'mime_type' => mime_content_type($filePath)
    ];
}

function readFileContent($filePath) {
    if (!file_exists($filePath)) {
        return null;
    }
    return file_get_contents($filePath);
}

function writeFileContent($filePath, $content) {
    return file_put_contents($filePath, $content);
}
`
	fmt.Println("定义文件处理函数:")
	fmt.Println(fileFunctions)

	// 示例6: 验证函数
	fmt.Println("\n6. 验证函数:")

	validationFunctions := `
function validateEmail($email) {
    return filter_var($email, FILTER_VALIDATE_EMAIL) !== false;
}

function validatePhone($phone) {
    return preg_match('/^1[3-9]\d{9}$/', $phone);
}

function validatePassword($password) {
    $errors = [];
    
    if (strlen($password) < 8) {
        $errors[] = '密码长度至少8位';
    }
    
    if (!preg_match('/[A-Z]/', $password)) {
        $errors[] = '密码必须包含大写字母';
    }
    
    if (!preg_match('/[a-z]/', $password)) {
        $errors[] = '密码必须包含小写字母';
    }
    
    if (!preg_match('/[0-9]/', $password)) {
        $errors[] = '密码必须包含数字';
    }
    
    return empty($errors) ? true : $errors;
}
`
	fmt.Println("定义验证函数:")
	fmt.Println(validationFunctions)

	// 示例7: 加密解密函数
	fmt.Println("\n7. 加密解密函数:")

	cryptoFunctions := `
function simpleEncrypt($data, $key) {
    $method = 'AES-256-CBC';
    $iv = openssl_random_pseudo_bytes(openssl_cipher_iv_length($method));
    $encrypted = openssl_encrypt($data, $method, $key, 0, $iv);
    return base64_encode($iv . $encrypted);
}

function simpleDecrypt($encryptedData, $key) {
    $method = 'AES-256-CBC';
    $data = base64_decode($encryptedData);
    $ivLength = openssl_cipher_iv_length($method);
    $iv = substr($data, 0, $ivLength);
    $encrypted = substr($data, $ivLength);
    return openssl_decrypt($encrypted, $method, $key, 0, $iv);
}

function generateHash($data, $algorithm = 'sha256') {
    return hash($algorithm, $data);
}
`
	fmt.Println("定义加密解密函数:")
	fmt.Println(cryptoFunctions)

	// 示例8: 数据库模拟函数
	fmt.Println("\n8. 数据库模拟函数:")

	databaseFunctions := `
function mockQuery($sql, $params = []) {
    // 模拟数据库查询
    $users = [
        ['id' => 1, 'name' => '张三', 'email' => 'zhangsan@example.com'],
        ['id' => 2, 'name' => '李四', 'email' => 'lisi@example.com'],
        ['id' => 3, 'name' => '王五', 'email' => 'wangwu@example.com']
    ];
    
    if (strpos($sql, 'SELECT') !== false) {
        return $users;
    }
    
    return ['affected_rows' => 1];
}

function mockInsert($table, $data) {
    return [
        'id' => rand(1000, 9999),
        'affected_rows' => 1,
        'insert_id' => rand(1000, 9999)
    ];
}

function mockUpdate($table, $data, $where) {
    return ['affected_rows' => 1];
}

function mockDelete($table, $where) {
    return ['affected_rows' => 1];
}
`
	fmt.Println("定义数据库模拟函数:")
	fmt.Println(databaseFunctions)

	// 示例9: 缓存函数
	fmt.Println("\n9. 缓存函数:")

	cacheFunctions := `
function cacheSet($key, $value, $ttl = 3600) {
    $cacheFile = '/tmp/cache_' . md5($key) . '.json';
    $data = [
        'value' => $value,
        'expires' => time() + $ttl
    ];
    return file_put_contents($cacheFile, json_encode($data));
}

function cacheGet($key) {
    $cacheFile = '/tmp/cache_' . md5($key) . '.json';
    
    if (!file_exists($cacheFile)) {
        return null;
    }
    
    $data = json_decode(file_get_contents($cacheFile), true);
    
    if ($data['expires'] < time()) {
        unlink($cacheFile);
        return null;
    }
    
    return $data['value'];
}

function cacheDelete($key) {
    $cacheFile = '/tmp/cache_' . md5($key) . '.json';
    if (file_exists($cacheFile)) {
        return unlink($cacheFile);
    }
    return false;
}
`
	fmt.Println("定义缓存函数:")
	fmt.Println(cacheFunctions)

	// 示例10: 日志函数
	fmt.Println("\n10. 日志函数:")

	logFunctions := `
function writeLog($message, $level = 'INFO', $file = '/tmp/app.log') {
    $timestamp = date('Y-m-d H:i:s');
    $logEntry = "[$timestamp] [$level] $message" . PHP_EOL;
    return file_put_contents($file, $logEntry, FILE_APPEND | LOCK_EX);
}

function logInfo($message) {
    return writeLog($message, 'INFO');
}

function logError($message) {
    return writeLog($message, 'ERROR');
}

function logWarning($message) {
    return writeLog($message, 'WARNING');
}

function logDebug($message) {
    return writeLog($message, 'DEBUG');
}
`
	fmt.Println("定义日志函数:")
	fmt.Println(logFunctions)

	// 示例11: 实际调用演示
	fmt.Println("\n11. 实际调用演示:")

	// 注意：这些调用需要在实际的PHP环境中才能工作
	// 这里只是演示如何调用自定义函数

	fmt.Println("调用示例（需要在实际PHP环境中运行）:")
	fmt.Println("- calculateSum(10, 20)")
	fmt.Println("- formatName('john', 'doe')")
	fmt.Println("- validateEmail('test@example.com')")
	fmt.Println("- generateSlug('Hello World!')")
	fmt.Println("- formatDate('2023-12-25')")

	// 示例12: 函数组合使用
	fmt.Println("\n12. 函数组合使用:")

	combinedFunctions := `
function processUserData($userData) {
    // 验证数据
    if (!validateEmail($userData['email'])) {
        return ['error' => '邮箱格式不正确'];
    }
    
    // 格式化姓名
    $userData['full_name'] = formatName($userData['first_name'], $userData['last_name']);
    
    // 生成用户名
    $userData['username'] = generateSlug($userData['full_name']);
    
    // 记录日志
    logInfo("处理用户数据: " . $userData['full_name']);
    
    // 保存到数据库
    $result = mockInsert('users', $userData);
    
    return [
        'success' => true,
        'user_id' => $result['insert_id'],
        'data' => $userData
    ];
}

function calculateStatistics($numbers) {
    $stats = [
        'count' => count($numbers),
        'sum' => array_sum($numbers),
        'average' => calculateAverage($numbers),
        'min' => min($numbers),
        'max' => max($numbers)
    ];
    
    $stats['range'] = $stats['max'] - $stats['min'];
    
    return $stats;
}
`
	fmt.Println("定义组合函数:")
	fmt.Println(combinedFunctions)

	fmt.Println("\n=== 自定义函数示例完成 ===")
	fmt.Println("\n注意：这些自定义函数需要在PHP环境中定义后才能调用。")
	fmt.Println("在实际使用中，您需要将这些函数添加到PHP脚本文件中。")
}
