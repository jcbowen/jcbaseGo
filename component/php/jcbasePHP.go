package php

var TmpJcbasePHP = `#!/usr/bin/env php
<?php

(new Console)->run();

class Console
{
    public static $regOpts = [
        'help'    => [
            'short'     => 'h',
            'long'      => 'help',
            'desc'      => '输出帮助信息',
            'has_value' => false,
        ],
        'version' => [
            'short'     => 'v',
            'long'      => 'version',
            'desc'      => '输出版本信息',
            'has_value' => false,
        ],
        'func'    => [
            'short'     => 'f',
            'long'      => 'func',
            'desc'      => '要执行的函数名',
            'has_value' => true,
        ]
    ];

    public static $shortOpts = '';
    public static $longOpts = [];

    public static $argv = [];
    public static $opts = [];

    public static $args = [];

    public function __construct()
    {
        foreach (self::$regOpts as $opt) {
            if ($opt['short']) {
                self::$shortOpts .= $opt['short'];
                if ($opt['has_value']) {
                    self::$shortOpts .= ':';
                }
            }
            if ($opt['long']) {
                self::$longOpts[] = $opt['long'] . ($opt['has_value'] ? ':' : '');
            }
        }
    }

    public function run()
    {
        self::$argv = $_SERVER['argv'];
        self::$opts = getopt(self::$shortOpts, self::$longOpts);

        if (isset(self::$opts['h']) || isset(self::$opts['help'])) {
            self::stdout("Usage: " . self::$argv[0] . " [options] [...args]");
            foreach (self::$regOpts as $opt) {
                self::stdout("  -{$opt['short']}, --{$opt['long']} {$opt['desc']}");
            }
            exit();
        }

        if (isset(self::$opts['v']) || isset(self::$opts['version'])) {
            self::stdout("Version: 0.0.1");
            exit();
        }

        if (isset(self::$opts['f']) || isset(self::$opts['func'])) {
            // 将$argv中的参数转换为数组
            self::$args = array_slice(self::$argv, 1);
            // 去除长短option
            self::$args = array_filter(self::$args, function ($arg) {
                return !preg_match('/^(-\w|--\w+)/', $arg);
            });

            // 转换参数中的true/false为bool类型，转换参数中的数字为int类型，转换参数中的json为数组类型
            self::$args = array_map(function ($arg) {
                if (in_array($arg, ['true', 'false'])) {
                    return $arg === 'true';
                }
                if (is_numeric($arg)) {
                    return (int)$arg;
                }
                if (preg_match('/^\{.*\}$/', $arg)) {
                    return @json_decode($arg, true);
                }
                return $arg;
            }, self::$args);

            // 执行函数，传入参数，返回结果
            $func = self::$opts['f'] ?? self::$opts['func'];
            if (function_exists($func)) {
                $result = call_user_func_array($func, self::$args);
                if (is_array($result))
                    $result = stripslashes(json_encode($result, JSON_UNESCAPED_UNICODE));
//                if($result == null)
//                    $result = 'NULL';
                self::stdout((string)$result);
            } else {
                self::error("fatal:Function $func not exists");
            }

            exit();
        }

        self::error("fatal:Function not specified");
    }


    /**
     * Gets input from STDIN and returns a string right-trimmed for EOLs.
     *
     * @param bool $raw If set to true, returns the raw string without trimming
     * @return string the string read from stdin
     */
    public static function stdin(bool $raw = false): string
    {
        return $raw ? fgets(\STDIN) : rtrim(fgets(\STDIN), PHP_EOL);
    }

    /**
     * Prints a string to STDOUT.
     *
     * @param string $string the string to print
     * @return int|bool Number of bytes printed or false on error
     */
    public static function stdout(string $string)
    {
        return fwrite(\STDOUT, $string);
    }

    /**
     * Prints a string to STDERR.
     *
     * @param string $string the string to print
     * @return int|bool Number of bytes printed or false on error
     */
    public static function stderr(string $string)
    {
        return fwrite(\STDERR, $string);
    }

    /**
     * Asks the user for input. Ends when the user types a carriage return (PHP_EOL). Optionally, It also provides a
     * prompt.
     *
     * @param string|null $prompt the prompt to display before waiting for input (optional)
     * @return string the user's input
     */
    public static function input(string $prompt = null): string
    {
        if (isset($prompt)) {
            static::stdout($prompt);
        }

        return static::stdin();
    }

    /**
     * Prints text to STDOUT appended with a carriage return (PHP_EOL).
     *
     * @param string|null $string the text to print
     * @return int|bool number of bytes printed or false on error.
     */
    public static function output(string $string = null)
    {
        return static::stdout($string . PHP_EOL);
    }

    /**
     * Prints text to STDERR appended with a carriage return (PHP_EOL).
     *
     * @param string|null $string the text to print
     * @return int|bool number of bytes printed or false on error.
     */
    public static function error(string $string = null)
    {
        return static::stderr($string . PHP_EOL);
    }

    /**
     * Asks user to confirm by typing y or n.
     *
     * A typical usage looks like the following:
     *
     * if (Console::confirm("Are you sure?")) {
     *     echo "user typed yes\n";
     * } else {
     *     echo "user typed no\n";
     * }
     *
     * @param string $message to print out before waiting for user input
     * @param bool $default this value is returned if no selection is made.
     * @return bool whether user confirmed
     */
    public static function confirm(string $message, bool $default = false): bool
    {
        while (true) {
            static::stdout($message . ' (yes|no) [' . ($default ? 'yes' : 'no') . ']:');
            $input = trim(static::stdin());

            if (empty($input)) {
                return $default;
            }

            if (!strcasecmp($input, 'y') || !strcasecmp($input, 'yes')) {
                return true;
            }

            if (!strcasecmp($input, 'n') || !strcasecmp($input, 'no')) {
                return false;
            }
        }
    }

    /**
     * Gives the user an option to choose from. Giving '?' as an input will show
     * a list of options to choose from and their explanations.
     *
     * @param string $prompt the prompt message
     * @param array $options Key-value array of options to choose from. Key is what is inputed and used, value is
     * what's displayed to end user by help command.
     *
     * @return string An option character the user chose
     */
    public static function select(string $prompt, array $options = []): string
    {
        top:
        static::stdout("$prompt [" . implode(',', array_keys($options)) . ',?]: ');
        $input = static::stdin();
        if ($input === '?') {
            foreach ($options as $key => $value) {
                static::output(" $key - $value");
            }
            static::output(' ? - Show help');
            goto top;
        } elseif (!array_key_exists($input, $options)) {
            goto top;
        }

        return $input;
    }
}

`
