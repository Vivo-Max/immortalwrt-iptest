package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"golang.org/x/net/proxy"
)

const (
	timeout     = 1 * time.Second // 超时时间
	maxDuration = 2 * time.Second // 最大持续时间
)

var (
	Path         = flag.String("path", "ip.txt", "指定包含IP地址的文件或目录")          // IP地址文件或目录
	outFile      = flag.String("outfile", "ip.csv", "输出文件名称")                              // 输出文件名称
	maxThreads   = flag.Int("max", 100, "并发请求最大协程数")                                       // 最大协程数
	speedTest    = flag.Int("speedtest", 5, "下载测速协程数量,设为0禁用测速")                            // 下载测速协程数量
	speedLimit   = flag.Int("int", 0, "最低下载速度(MB/s)")                                   // 最低下载速度
    speedTestURL = flag.String("url", "speed.cloudflare.com/__down?bytes=500000000", "测速文件地址") // 测速文件地址
	enableTLS    = flag.Bool("tls", true, "是否启用TLS")                                       // TLS是否启用
	TCPurl       = flag.String("tcpurl", "www.speedtest.net", "TCP请求地址")                   // TCP请求地址
	ports = flag.String("ports", "", "指定仅输出这些端口的结果，用逗号分隔，空表示不过滤")

	telegramToken   = flag.String("telegram_token", "", "Telegram Bot TOKEN")
	telegramChatID  = flag.String("telegram_chat_id", "", "Telegram Chat ID")
	presetProxy     = flag.String("preset_proxy", "", "预设SOCKS5代理列表")

	telegramClientCache *http.Client
	clientCacheMutex    sync.Mutex
)

type result struct {
	ip          string        // IP地址
	port        int           // 端口
	dataCenter  string        // 数据中心
	region      string        // 地区
	cca1        string         // 国家代码	
	cca2        string         // 国家
	city        string        // 城市
	latency     string        // 延迟
	tcpDuration time.Duration // TCP请求延迟
}

type speedtestresult struct {
	result
	downloadSpeed float64 // 下载速度
}

type location struct {
	Iata   string  `json:"iata"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Cca1   string  `json:"cca1"`
	Cca2   string  `json:"cca2"`
	Region string  `json:"region"`
	City   string  `json:"city"`
}

type telegramAPIResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

// 国家代码到国旗的映射
func getCountryFlag(cca1 string) string {
	flagMap := map[string]string{
		"AD": "🇦🇩", "AE": "🇦🇪", "AF": "🇦🇫", "AG": "🇦🇬", "AI": "🇦🇮", "AL": "🇦🇱", "AM": "🇦🇲", "AO": "🇦🇴",
		"AQ": "🇦🇶", "AR": "🇦🇷", "AS": "🇦🇸", "AT": "🇦🇹", "AU": "🇦🇺", "AW": "🇦🇼", "AX": "🇦🇽", "AZ": "🇦🇿",
		"BA": "🇧🇦", "BB": "🇧🇧", "BD": "🇧🇩", "BE": "🇧🇪", "BF": "🇧🇫", "BG": "🇧🇬", "BH": "🇧🇭", "BI": "🇧🇮",
		"BJ": "🇧🇯", "BL": "🇧🇱", "BM": "🇧🇲", "BN": "🇧🇳", "BO": "🇧🇴", "BQ": "🇧🇶", "BR": "🇧🇷", "BS": "🇧🇸",
		"BT": "🇧🇹", "BV": "🇧🇻", "BW": "🇧🇼", "BY": "🇧🇾", "BZ": "🇧🇿", "CA": "🇨🇦", "CC": "🇨🇨", "CD": "🇨🇩",
		"CF": "🇨🇫", "CG": "🇨🇬", "CH": "🇨🇭", "CI": "🇨🇮", "CK": "🇨🇰", "CL": "🇨🇱", "CM": "🇨🇲", "CN": "🇨🇳",
		"CO": "🇨🇴", "CR": "🇨🇷", "CU": "🇨🇺", "CV": "🇨🇻", "CW": "🇨🇼", "CX": "🇨🇽", "CY": "🇨🇾", "CZ": "🇨🇿",
		"DE": "🇩🇪", "DJ": "🇩🇯", "DK": "🇩🇰", "DM": "🇩🇲", "DO": "🇩🇴", "DZ": "🇩🇿", "EC": "🇪🇨", "EE": "🇪🇪",
		"EG": "🇪🇬", "EH": "🇪🇭", "ER": "🇪🇷", "ES": "🇪🇸", "ET": "🇪🇹", "FI": "🇫🇮", "FJ": "🇫🇯", "FK": "🇫🇰",
		"FM": "🇫🇲", "FO": "🇫🇴", "FR": "🇫🇷", "GA": "🇬🇦", "GB": "🇬🇧", "GD": "🇬🇩", "GE": "🇬🇪", "GF": "🇬🇫",
		"GG": "🇬🇬", "GH": "🇬🇭", "GI": "🇬🇮", "GL": "🇬🇱", "GM": "🇬🇲", "GN": "🇬🇳", "GP": "🇬🇵", "GQ": "🇬🇶",
		"GR": "🇬🇷", "GS": "🇬🇸", "GT": "🇬🇹", "GU": "🇬🇺", "GW": "🇬🇼", "GY": "🇬🇾", "HK": "🇭🇰", "HM": "🇭🇲",
		"HN": "🇭🇳", "HR": "🇭🇷", "HT": "🇭🇹", "HU": "🇭🇺", "ID": "🇮🇩", "IE": "🇮🇪", "IL": "🇮🇱", "IM": "🇮🇲",
		"IN": "🇮🇳", "IO": "🇮🇴", "IQ": "🇮🇶", "IR": "🇮🇷", "IS": "🇮🇸", "IT": "🇮🇹", "JE": "🇯🇪", "JM": "🇯🇲",
		"JO": "🇯🇴", "JP": "🇯🇵", "KE": "🇰🇪", "KG": "🇰🇬", "KH": "🇰🇭", "KI": "🇰🇮", "KM": "🇰🇲", "KN": "🇰🇳",
		"KP": "🇰🇵", "KR": "🇰🇷", "KW": "🇰🇼", "KY": "🇰🇾", "KZ": "🇰🇿", "LA": "🇱🇦", "LB": "🇱🇧", "LC": "🇱🇨",
		"LI": "🇱🇮", "LK": "🇱🇰", "LR": "🇱🇷", "LS": "🇱🇸", "LT": "🇱🇹", "LU": "🇱🇺", "LV": "🇱🇻", "LY": "🇱🇾",
		"MA": "🇲🇦", "MC": "🇲🇨", "MD": "🇲🇩", "ME": "🇲🇪", "MF": "🇲🇫", "MG": "🇲🇬", "MH": "🇲🇷", "MK": "🇲🇰",
		"ML": "🇲🇱", "MM": "🇲🇲", "MN": "🇲🇳", "MO": "🇲🇴", "MP": "🇲🇵", "MQ": "🇲🇶", "MR": "🇲🇷", "MS": "🇲🇸",
		"MT": "🇲🇹", "MU": "🇲🇺", "MV": "🇲🇻", "MW": "🇲🇼", "MX": "🇲🇽", "MY": "🇲🇾", "MZ": "🇲🇿", "NA": "🇳🇦",
		"NC": "🇳🇨", "NE": "🇳🇪", "NF": "🇳🇫", "NG": "🇳🇬", "NI": "🇳🇮", "NL": "🇳🇱", "NO": "🇳🇴", "NP": "🇳🇵",
		"NR": "🇳🇷", "NU": "🇳🇺", "NZ": "🇳🇿", "OM": "🇴🇲", "PA": "🇵🇦", "PE": "🇵🇪", "PF": "🇵🇫", "PG": "🇵🇬",
		"PH": "🇵🇭", "PK": "🇵🇰", "PL": "🇵🇱", "PM": "🇵🇲", "PN": "🇵🇳", "PR": "🇵🇷", "PS": "🇵🇸", "PT": "🇵🇹",
		"PW": "🇵🇼", "PY": "🇵🇾", "QA": "🇶🇦", "RE": "🇷🇪", "RO": "🇷🇴", "RS": "🇷🇸", "RU": "🇷🇺", "RW": "🇷🇼",
		"SA": "🇸🇦", "SB": "🇸🇧", "SC": "🇸🇨", "SD": "🇸🇩", "SE": "🇸🇪", "SG": "🇸🇬", "SH": "🇸🇭", "SI": "🇸🇮",
		"SJ": "🇸🇯", "SK": "🇸🇰", "SL": "🇸🇱", "SM": "🇸🇲", "SN": "🇸🇳", "SO": "🇸🇴", "SR": "🇸🇷", "SS": "🇸🇸",
		"ST": "🇸🇹", "SV": "🇸🇻", "SX": "🇸🇽", "SY": "🇸🇾", "SZ": "🇸🇿", "TC": "🇹🇨", "TD": "🇹🇩", "TF": "🇹🇫",
		"TG": "🇹🇬", "TH": "🇹🇭", "TJ": "🇹🇯", "TK": "🇹🇰", "TL": "🇹🇱", "TM": "🇹🇲", "TN": "🇹🇳", "TO": "🇹🇴",
		"TR": "🇹🇷", "TT": "🇹🇹", "TV": "🇹🇻", "TW": "🇹🇼", "TZ": "🇹🇿", "UA": "🇺🇦", "UG": "🇺🇬", "UM": "🇺🇲",
		"US": "🇺🇸", "UY": "🇺🇾", "UZ": "🇺🇿", "VA": "🇻🇦", "VC": "🇻🇨", "VE": "🇻🇪", "VG": "🇻🇬", "VI": "🇻🇮",
		"VN": "🇻🇳", "VU": "🇻🇺", "WF": "🇼🇫", "WS": "🇼🇸", "XK": "🇽🇰", "YE": "🇾🇪", "YT": "🇾🇹", "ZA": "🇿🇦",
		"ZM": "🇿🇲", "ZW": "🇿🇼", "UNKNOWN": "🌐",
	}

	if flag, ok := flagMap[cca1]; ok {
		return flag
	}
	return "🏳️" // 默认未知国旗
}

// 尝试提升文件描述符的上限
func increaseMaxOpenFiles() {
	fmt.Println("正在尝试提升文件描述符的上限...")
	cmd := exec.Command("bash", "-c", "ulimit -n 10000")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("提升文件描述符上限时出现错误: %v\n", err)
	} else {
		fmt.Printf("文件描述符上限已提升!\n")
	}
}

// maskBotToken 脱敏 Telegram Bot Token
func maskBotToken(logText string) string {
	re := regexp.MustCompile(`(bot)\d+:[a-zA-Z0-9_-]+`)
	return re.ReplaceAllString(logText, "${1}********************")
}

// escapeMarkdownV2 对字符串进行转义以符合MarkdownV2规范
func escapeMarkdownV2(text string) string {
	var escaped bytes.Buffer
	for _, r := range text {
		switch r {
		case '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!':
			escaped.WriteRune('\\')
			escaped.WriteRune(r)
		default:
			escaped.WriteRune(r)
		}
	}
	return escaped.String()
}

// createTelegramClientWithProxy 创建带代理的Telegram客户端
func createTelegramClientWithProxy(proxyURL string) (*http.Client, error) {
	var transport *http.Transport
	var err error

	if proxyURL == "" {
		transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 3 * time.Second,
			}).DialContext,
		}
	} else {
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("解析代理URL失败: %v", err)
		}

		dialer := &net.Dialer{
			Timeout: 3 * time.Second,
		}

		switch parsedURL.Scheme {
		case "socks5", "socks5h":
			var auth *proxy.Auth
			if parsedURL.User != nil {
				password, _ := parsedURL.User.Password()
				auth = &proxy.Auth{User: parsedURL.User.Username(), Password: password}
			}
			socks5Dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, dialer)
			if err != nil {
				return nil, fmt.Errorf("创建SOCKS5代理失败: %v", err)
			}
			transport = &http.Transport{
				DialContext: socks5Dialer.(proxy.ContextDialer).DialContext,
			}
		default:
			return nil, fmt.Errorf("不支持的代理协议: %s", parsedURL.Scheme)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// 测试代理连接
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", *telegramToken)
	fmt.Printf("尝试连接 %s (代理: %s)\n", maskBotToken(url), proxyURL)
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("代理验证失败: %v (耗时: %v)", err, time.Since(start))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("代理验证失败,HTTP状态码: %d, 响应: %s (耗时: %v)", resp.StatusCode, string(body), time.Since(start))
	}
	fmt.Printf("连接成功 (代理: %s, 耗时: %v)\n", proxyURL, time.Since(start))
	return client, nil
}

// getTelegramClient 获取可用的Telegram客户端
func getTelegramClient() *http.Client {
	clientCacheMutex.Lock()
	defer clientCacheMutex.Unlock()

	if telegramClientCache != nil {
		fmt.Println("使用缓存的Telegram客户端")
		return telegramClientCache
	}

	// 解析代理列表
	proxyList := strings.Split(*presetProxy, ",")
	for _, proxyURL := range proxyList {
		proxyURL = strings.TrimSpace(proxyURL)
		if proxyURL == "" {
			continue
		}
		fmt.Printf("尝试通过代理 %s 连接Telegram API...\n", proxyURL)
		client, err := createTelegramClientWithProxy(proxyURL)
		if err == nil {
			fmt.Printf("通过代理 %s 建立Telegram会话\n", proxyURL)
			telegramClientCache = client
			return client
		}
		fmt.Printf("代理 %s 连接Telegram失败: %v\n", proxyURL, err)
	}

	// 所有代理失败,尝试直连
	fmt.Println("所有预设代理失败,尝试直连...")
	client, err := createTelegramClientWithProxy("")
	if err == nil {
		fmt.Println("直连Telegram API成功")
		telegramClientCache = client
		return client
	}
	fmt.Printf("直连Telegram API失败: %v\n", err)
	return nil
}

// sendTelegramMessage 发送Telegram消息(带重试)
func sendTelegramMessage(message string) bool {
    if *telegramToken == "" {
        fmt.Println("未配置Telegram Bot Token,跳过消息推送")
        return false
    }
    // 从环境变量中读取 chat_ids
    chatIDs := strings.Split(os.Getenv("CHAT_IDS"), " ")
    client := getTelegramClient()
    if client == nil {
        fmt.Println("无法建立网络连接,跳过Telegram消息推送")
        return false
    }
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", *telegramToken)
    escapedMessage := escapeMarkdownV2(message)
    escapedMessage = strings.ReplaceAll(escapedMessage, "\\*", "*") // 保留粗体
    escapedMessage = strings.ReplaceAll(escapedMessage, "\\`", "`") // 保留代码块
    payload := map[string]string{
        "text": escapedMessage,
        "parse_mode": "MarkdownV2",
    }
    const maxRetries = 3
    for _, chatID := range chatIDs {
        payload["chat_id"] = chatID
        jsonPayload, _ := json.Marshal(payload)
        for attempt := 1; attempt <= maxRetries; attempt++ {
            resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
            if err != nil {
                fmt.Printf("Telegram消息推送失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
                if attempt == maxRetries {
                    clientCacheMutex.Lock()
                    telegramClientCache = nil
                    clientCacheMutex.Unlock()
                    fmt.Println("Telegram客户端失效,清除缓存")
                    return false
                }
                sleepDuration := time.Duration(math.Pow(2, float64(attempt)) * float64(time.Second))
                time.Sleep(sleepDuration)
                continue
            }
            defer resp.Body.Close()
            var apiResp telegramAPIResponse
            if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil || !apiResp.Ok {
                fmt.Printf("Telegram API错误 (尝试 %d/%d): %s\n", attempt, maxRetries, apiResp.Description)
                if attempt == maxRetries {
                    clientCacheMutex.Lock()
                    telegramClientCache = nil
                    clientCacheMutex.Unlock()
                    fmt.Println("Telegram客户端失效,清除缓存")
                    return false
                }
                time.Sleep(time.Duration(5) * time.Second)
                continue
            }
            fmt.Println("Telegram消息推送成功")
            break
        }
    }
    return true
}

// sendTelegramFile 发送Telegram文件（带重试）
func sendTelegramFile(filePath string) bool {
    if *telegramToken == "" {
        fmt.Println("未配置Telegram Bot Token,跳过文件推送")
        return false
    }
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        fmt.Printf("文件 %s 不存在,跳过推送\n", filepath.Base(filePath))
        return false
    }
    fileInfo, _ := os.Stat(filePath)
    if fileInfo.Size() == 0 {
        fmt.Printf("文件 %s 为空,删除并跳过推送\n", filepath.Base(filePath))
        os.Remove(filePath)
        return false
    }
    // 从环境变量中读取 chat_ids
    chatIDs := strings.Split(os.Getenv("CHAT_IDS"), " ")
    client := getTelegramClient()
    if client == nil {
        fmt.Println("无法建立网络连接,跳过Telegram文件推送")
        return false
    }
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", *telegramToken)
    const maxRetries = 3
    for _, chatID := range chatIDs {
        body := &bytes.Buffer{}
        writer := multipart.NewWriter(body)
        err := writer.WriteField("chat_id", chatID)
        if err != nil {
            fmt.Printf("写入chat_id字段失败: %v\n", err)
            return false
        }
        file, err := os.Open(filePath)
        if err != nil {
            fmt.Printf("无法打开文件 %s: %v\n", filePath, err)
            return false
        }
        part, err := writer.CreateFormFile("document", filepath.Base(filePath))
        if err != nil {
            file.Close()
            fmt.Printf("创建multipart表单失败: %v\n", err)
            return false
        }
        _, err = io.Copy(part, file)
        file.Close()
        if err != nil {
            fmt.Printf("复制文件到表单失败: %v\n", err)
            return false
        }
        writer.Close()
        req, err := http.NewRequest("POST", url, body)
        if err != nil {
            fmt.Printf("创建HTTP请求失败: %v\n", err)
            return false
        }
        req.Header.Set("Content-Type", writer.FormDataContentType())
        for attempt := 1; attempt <= maxRetries; attempt++ {
            resp, err := client.Do(req)
            if err != nil {
                fmt.Printf("文件 %s 推送失败 (尝试 %d/%d): %v\n", filepath.Base(filePath), attempt, maxRetries, err)
                if attempt == maxRetries {
                    clientCacheMutex.Lock()
                    telegramClientCache = nil
                    clientCacheMutex.Unlock()
                    fmt.Println("Telegram客户端失效,清除缓存")
                    return false
                }
                sleepDuration := time.Duration(math.Pow(2, float64(attempt)) * float64(time.Second))
                time.Sleep(sleepDuration)
                continue
            }
            defer resp.Body.Close()
            var apiResp telegramAPIResponse
            if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil || !apiResp.Ok {
                fmt.Printf("Telegram API错误 (尝试 %d/%d): %s\n", attempt, maxRetries, apiResp.Description)
                if attempt == maxRetries {
                    clientCacheMutex.Lock()
                    telegramClientCache = nil
                    clientCacheMutex.Unlock()
                    fmt.Println("Telegram客户端失效,清除缓存")
                    return false
                }
                time.Sleep(time.Duration(5) * time.Second)
                continue
            }
            fmt.Printf("文件 %s 推送成功\n", filepath.Base(filePath))
            break
        }
    }
    return true
}

// gracefulExit 优雅退出
func gracefulExit(msg string, code int) {
	if msg != "" {
		fmt.Println(msg)
		if *telegramToken != "" && *telegramChatID != "" {
			sendTelegramMessage(msg)
			sendTelegramMessage("*🎉 程序运行结束*")
		}
	}
	os.Exit(code)
}

func main() {
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			gracefulExit(fmt.Sprintf("*⚠️ 程序崩溃*\n%v", r), 1)
		}
	}()

	// 从环境变量读取多个 chat_id
	chatIDs := strings.Split(os.Getenv("CHAT_IDS"), " ")
	if *telegramToken != "" && len(chatIDs) > 0 {
		sendTelegramMessage("*🚀 开始延迟/速度测试*") // 推送开始测试的消息
	}

	startTime := time.Now()
	osType := runtime.GOOS
	// 如果是linux系统,尝试提升文件描述符的上限
	// 判断是否以root用户运行

	if osType == "linux" && os.Getuid() == 0 {
		increaseMaxOpenFiles()
	}

	var locations []location
/////////////////////////////
	body := `[{"iata":"TIA","lat":41.4146995544,"lon":19.7206001282,"cca2":"阿尔巴尼亚","cca1":"AL","region":"欧洲","city":"地拉那"},{"iata":"KHN","lat":41.4146995544,"lon":19.7206001282,"cca2":"中国","cca1":"CN","region":"亚洲","city":"南昌"},{"iata":"ALG","lat":36.6910018921,"lon":3.2154099941,"cca2":"阿尔及利亚","cca1":"DZ","region":"非洲","city":"阿尔及尔"},{"iata":"ORN","lat":35.6911,"lon":-0.6416,"cca2":"阿尔及利亚","cca1":"DZ","region":"非洲","city":"奥兰"},{"iata":"LAD","lat":-8.8583698273,"lon":13.2312002182,"cca2":"安哥拉","cca1":"AO","region":"非洲","city":"罗安达"},{"iata":"EZE","lat":-34.8222,"lon":-58.5358,"cca2":"阿根廷","cca1":"AR","region":"南美","city":"布宜诺斯艾利斯"},{"iata":"COR","lat":-31.31,"lon":-64.208333,"cca2":"阿根廷","cca1":"AR","region":"南美","city":"科尔多瓦"},{"iata":"NQN","lat":-38.9490013123,"lon":-68.1557006836,"cca2":"阿根廷","cca1":"AR","region":"南美","city":"内乌肯"},{"iata":"EVN","lat":40.1473007202,"lon":44.3959007263,"cca2":"亚美尼亚","cca1":"AM","region":"中东","city":"埃里温"},{"iata":"ADL","lat":-34.9431729,"lon":138.5335637,"cca2":"澳大利亚","cca1":"AU","region":"大洋洲","city":"阿德莱德"},{"iata":"BNE","lat":-27.3841991425,"lon":153.117004394,"cca2":"澳大利亚","cca1":"AU","region":"大洋洲","city":"布里斯班"},{"iata":"CBR","lat":-35.3069000244,"lon":149.1950073242,"cca2":"澳大利亚","cca1":"AU","region":"大洋洲","city":"堪培拉"},{"iata":"HBA","lat":-42.883209,"lon":147.331665,"cca2":"澳大利亚","cca1":"AU","region":"大洋洲","city":"霍巴特"},{"iata":"MEL","lat":-37.6733016968,"lon":144.843002319,"cca2":"澳大利亚","cca1":"AU","region":"大洋洲","city":"墨尔本"},{"iata":"PER","lat":-31.9402999878,"lon":115.967002869,"cca2":"澳大利亚","cca1":"AU","region":"大洋洲","city":"珀斯"},{"iata":"SYD","lat":-33.9460983276,"lon":151.177001953,"cca2":"澳大利亚","cca1":"AU","region":"大洋洲","city":"悉尼"},{"iata":"VIE","lat":48.1102981567,"lon":16.5697002411,"cca2":"奥地利","cca1":"AT","region":"欧洲","city":"维也纳"},{"iata":"LLK","lat":38.7463989258,"lon":48.8180007935,"cca2":"阿塞拜疆","cca1":"AZ","region":"中东","city":"Astara"},{"iata":"GYD","lat":40.4674987793,"lon":50.0466995239,"cca2":"阿塞拜疆","cca1":"AZ","region":"中东","city":"巴库"},{"iata":"BAH","lat":26.2707996368,"lon":50.6335983276,"cca2":"巴林","cca1":"BH","region":"中东","city":"麦纳麦"},{"iata":"CGP","lat":22.2495995,"lon":91.8133011,"cca2":"孟加拉国","cca1":"BD","region":"亚太","city":"吉大港"},{"iata":"DAC","lat":23.843347,"lon":90.397783,"cca2":"孟加拉国","cca1":"BD","region":"亚太","city":"达卡"},{"iata":"JSR","lat":23.1837997437,"lon":89.1607971191,"cca2":"孟加拉国","cca1":"BD","region":"亚太","city":"杰索尔"},{"iata":"MSQ","lat":53.9006,"lon":27.599,"cca2":"白俄罗斯","cca1":"BY","region":"欧洲","city":"明斯克"},{"iata":"BRU","lat":50.9014015198,"lon":4.4844398499,"cca2":"比利时","cca1":"BE","region":"欧洲","city":"布鲁塞尔"},{"iata":"PBH","lat":27.4712,"lon":89.6339,"cca2":"不丹","cca1":"BT","region":"亚太","city":"廷布"},{"iata":"GBE","lat":-24.6282,"lon":25.9231,"cca2":"不丹","cca1":"BW","region":"非洲","city":"博茨瓦纳"},{"iata":"QWJ","lat":-22.738,"lon":-47.334,"cca2":"巴西","cca1":"BR","region":"南美","city":"亚美利加纳"},{"iata":"BEL","lat":-1.4563,"lon":-48.5013,"cca2":"巴西","cca1":"BR","region":"南美","city":"贝伦"},{"iata":"CNF","lat":-19.624444,"lon":-43.971944,"cca2":"巴西","cca1":"BR","region":"南美","city":"贝洛奥里宗特"},{"iata":"BNU","lat":-26.89245,"lon":-49.07696,"cca2":"巴西","cca1":"BR","region":"南美","city":"布鲁梅瑙"},{"iata":"BSB","lat":-15.79824,"lon":-47.90859,"cca2":"巴西","cca1":"BR","region":"南美","city":"利亚"},{"iata":"CFC","lat":-26.7762,"lon":-51.0125,"cca2":"巴西","cca1":"BR","region":"南美","city":"卡萨多尔"},{"iata":"VCP","lat":-22.90662,"lon":-47.08576,"cca2":"巴西","cca1":"BR","region":"南美","city":"坎皮纳斯"},{"iata":"CAW","lat":-21.698299408,"lon":-41.301700592,"cca2":"巴西","cca1":"BR","region":"南美","city":"坎普斯"},{"iata":"CGB","lat":-15.59611,"lon":-56.09667,"cca2":"巴西","cca1":"BR","region":"南美","city":"库亚巴"},{"iata":"CWB","lat":-25.5284996033,"lon":-49.1758003235,"cca2":"巴西","cca1":"BR","region":"南美","city":"库里蒂巴"},{"iata":"FLN","lat":-27.6702785492,"lon":-48.5525016785,"cca2":"巴西","cca1":"BR","region":"南美","city":"弗洛里亚诺波利斯"},{"iata":"FOR","lat":-3.7762799263,"lon":-38.5326004028,"cca2":"巴西","cca1":"BR","region":"南美","city":"福塔莱萨"},{"iata":"GYN","lat":-16.69727,"lon":-49.26851,"cca2":"巴西","cca1":"BR","region":"南美","city":"戈亚尼亚"},{"iata":"ITJ","lat":-27.6116676331,"lon":-48.6727790833,"cca2":"巴西","cca1":"BR","region":"南美","city":"伊塔雅伊"},{"iata":"JOI","lat":-26.304408,"lon":-48.846383,"cca2":"巴西","cca1":"BR","region":"南美","city":"若茵维莱"},{"iata":"JDO","lat":-7.2242,"lon":-39.313,"cca2":"巴西","cca1":"BR","region":"南美","city":"北茹阿泽鲁"},{"iata":"MAO","lat":-3.11286,"lon":-60.01949,"cca2":"巴西","cca1":"BR","region":"南美","city":"马瑙斯"},{"iata":"POA","lat":-29.9944000244,"lon":-51.1713981628,"cca2":"巴西","cca1":"BR","region":"南美","city":"阿雷格里港"},{"iata":"REC","lat":-8.1264896393,"lon":-34.9235992432,"cca2":"巴西","cca1":"BR","region":"南美","city":"累西腓"},{"iata":"RAO","lat":-21.1363887787,"lon":-47.7766685486,"cca2":"巴西","cca1":"BR","region":"南美","city":"里贝朗普雷图"},{"iata":"GIG","lat":-22.8099994659,"lon":-43.2505569458,"cca2":"巴西","cca1":"BR","region":"南美","city":"里约热内卢"},{"iata":"SSA","lat":-12.9086112976,"lon":-38.3224983215,"cca2":"巴西","cca1":"BR","region":"南美","city":"萨尔瓦多"},{"iata":"SJP","lat":-20.807157,"lon":-49.378994,"cca2":"巴西","cca1":"BR","region":"南美","city":"圣若泽"},{"iata":"SJK","lat":-23.1791,"lon":-45.8872,"cca2":"巴西","cca1":"BR","region":"南美","city":"圣若泽杜斯坎普斯"},{"iata":"GRU","lat":-23.4355564117,"lon":-46.4730567932,"cca2":"巴西","cca1":"BR","region":"南美","city":"圣保罗"},{"iata":"SOD","lat":-23.54389,"lon":-46.63445,"cca2":"巴西","cca1":"BR","region":"南美","city":"索罗卡巴"},{"iata":"NVT","lat":-26.8251,"lon":-49.2695,"cca2":"巴西","cca1":"BR","region":"南美","city":"圣卡塔琳娜"},{"iata":"UDI","lat":-18.8836116791,"lon":-48.225276947,"cca2":"巴西","cca1":"BR","region":"南美","city":"乌贝兰迪亚"},{"iata":"VIX","lat":-20.64871,"lon":-41.90857,"cca2":"巴西","cca1":"BR","region":"南美","city":"维多利亚"},{"iata":"BWN","lat":4.903052,"lon":114.939819,"cca2":"文莱","cca1":"BN","region":"亚太","city":"斯里巴加湾"},{"iata":"SOF","lat":42.6966934204,"lon":23.4114360809,"cca2":"保加利亚","cca1":"BG","region":"欧洲","city":"索菲亚"},{"iata":"OUA","lat":12.3531999588,"lon":-1.5124200583,"cca2":"布基纳法索","cca1":"BF","region":"非洲","city":"瓦加杜古"},{"iata":"PNH","lat":11.5466003418,"lon":104.84400177,"cca2":"柬埔寨","cca1":"KH","region":"亚太","city":"金边"},{"iata":"YYC","lat":51.113899231,"lon":-114.019996643,"cca2":"加拿大","cca1":"CA","region":"北美洲","city":"卡尔加里"},{"iata":"YVR","lat":49.193901062,"lon":-123.183998108,"cca2":"加拿大","cca1":"CA","region":"北美洲","city":"温哥华"},{"iata":"YWG","lat":49.9099998474,"lon":-97.2398986816,"cca2":"加拿大","cca1":"CA","region":"北美洲","city":"温尼伯"},{"iata":"YOW","lat":45.3224983215,"lon":-75.6691970825,"cca2":"加拿大","cca1":"CA","region":"北美洲","city":"渥太华"},{"iata":"YYZ","lat":43.6772003174,"lon":-79.6305999756,"cca2":"加拿大","cca1":"CA","region":"北美洲","city":"多伦多"},{"iata":"YUL","lat":45.4706001282,"lon":-73.7407989502,"cca2":"加拿大","cca1":"CA","region":"北美洲","city":"蒙特利尔"},{"iata":"YXE","lat":52.1707992554,"lon":-106.699996948,"cca2":"加拿大","cca1":"CA","region":"北美洲","city":"萨斯卡通"},{"iata":"ARI","lat":-18.348611,"lon":-70.338889,"cca2":"智利","cca1":"CL","region":"南美","city":"阿里卡"},{"iata":"CCP","lat":-36.8201,"lon":-73.0444,"cca2":"智利","cca1":"CL","region":"南美","city":"康塞普西翁"},{"iata":"SCL","lat":-33.3930015564,"lon":-70.7857971191,"cca2":"智利","cca1":"CL","region":"南美","city":"圣地亚哥"},{"iata":"BOG","lat":4.70159,"lon":-74.1469,"cca2":"哥伦比亚","cca1":"CO","region":"南美","city":"波哥大"},{"iata":"MDE","lat":6.16454,"lon":-75.4231,"cca2":"哥伦比亚","cca1":"CO","region":"南美","city":"麦德林"},{"iata":"FIH","lat":-4.3857498169,"lon":15.4446001053,"cca2":"刚果","cca1":"CD","region":"非洲","city":"金沙萨"},{"iata":"SJO","lat":9.9938602448,"lon":-84.2088012695,"cca2":"哥斯达黎加","cca1":"CR","region":"南美","city":"圣何塞"},{"iata":"ZAG","lat":45.7429008484,"lon":16.0687999725,"cca2":"克罗地亚","cca1":"HR","region":"欧洲","city":"萨格勒布"},{"iata":"CUR","lat":12.1888999939,"lon":-68.9598007202,"cca2":"库拉索岛","cca1":"CW","region":"南美","city":"库拉索岛"},{"iata":"LCA","lat":34.8750991821,"lon":33.6249008179,"cca2":"塞浦路斯","cca1":"CY","region":"欧洲","city":"尼科西亚"},{"iata":"PRG","lat":50.1007995605,"lon":14.2600002289,"cca2":"捷克","cca1":"CZ","region":"欧洲","city":"布拉格"},{"iata":"CPH","lat":55.6179008484,"lon":12.6560001373,"cca2":"丹麦","cca1":"DK","region":"欧洲","city":"哥本哈根"},{"iata":"JIB","lat":11.5473003387,"lon":43.1595001221,"cca2":"吉布提","cca1":"DJ","region":"非洲","city":"吉布提"},{"iata":"SDQ","lat":18.4297008514,"lon":-69.6688995361,"cca2":"多米尼加","cca1":"DO","region":"北美洲","city":"圣多明各"},{"iata":"GYE","lat":-2.1894,"lon":-79.8891,"cca2":"厄瓜多尔","cca1":"EC","region":"南美","city":"瓜亚基尔"},{"iata":"UIO","lat":-0.1291666667,"lon":-78.3575,"cca2":"厄瓜多尔","cca1":"EC","region":"南美","city":"基多"},{"iata":"TLL","lat":59.4132995605,"lon":24.8327999115,"cca2":"爱沙尼亚","cca1":"EE","region":"欧洲","city":"塔林"},{"iata":"HEL","lat":60.317199707,"lon":24.963300705,"cca2":"芬兰","cca1":"FI","region":"欧洲","city":"赫尔辛基"},{"iata":"LYS","lat":45.7263,"lon":5.0908,"cca2":"法国","cca1":"FR","region":"欧洲","city":"里昂"},{"iata":"MRS","lat":43.439271922,"lon":5.2214241028,"cca2":"法国","cca1":"FR","region":"欧洲","city":"马赛"},{"iata":"CDG","lat":49.0127983093,"lon":2.5499999523,"cca2":"法国","cca1":"FR","region":"欧洲","city":"巴黎"},{"iata":"PPT","lat":-17.5536994934,"lon":-149.606994629,"cca2":"玻利尼西亚","cca1":"PF","region":"大洋洲","city":"塔希提岛"},{"iata":"TBS","lat":41.6692008972,"lon":44.95470047,"cca2":"格鲁吉亚","cca1":"GE","region":"欧洲","city":"第比利斯"},{"iata":"TXL","lat":52.5597000122,"lon":13.2876996994,"cca2":"德国","cca1":"DE","region":"欧洲","city":"柏林"},{"iata":"DUS","lat":51.2895011902,"lon":6.7667798996,"cca2":"德国","cca1":"DE","region":"欧洲","city":"杜塞尔多夫"},{"iata":"FRA","lat":50.0264015198,"lon":8.543129921,"cca2":"德国","cca1":"DE","region":"欧洲","city":"法兰克福"},{"iata":"HAM","lat":53.6304016113,"lon":9.9882297516,"cca2":"德国","cca1":"DE","region":"欧洲","city":"汉堡"},{"iata":"MUC","lat":48.3538017273,"lon":11.7861003876,"cca2":"德国","cca1":"DE","region":"欧洲","city":"慕尼黑"},{"iata":"STR","lat":48.783333,"lon":9.183333,"cca2":"德国","cca1":"DE","region":"欧洲","city":"斯图加特"},{"iata":"ACC","lat":5.614818,"lon":-0.205874,"cca2":"加纳","cca1":"GH","region":"非洲","city":"阿克拉"},{"iata":"ATH","lat":37.9364013672,"lon":23.9444999695,"cca2":"希腊","cca1":"GR","region":"欧洲","city":"雅典"},{"iata":"SKG","lat":40.5196990967,"lon":22.9708995819,"cca2":"希腊","cca1":"GR","region":"欧洲","city":"塞萨洛尼基"},{"iata":"GND","lat":12.007116,"lon":-61.7882288,"cca2":"格林纳达","cca1":"GD","region":"南美","city":"圣乔治"},{"iata":"GUM","lat":13.4834003448,"lon":144.796005249,"cca2":"关岛","cca1":"GU","region":"亚太","city":"阿加尼亚"},{"iata":"GUA","lat":14.5832996368,"lon":-90.5274963379,"cca2":"危地马拉","cca1":"GT","region":"北美洲","city":"危地马拉"},{"iata":"GEO","lat":6.825648,"lon":-58.163756,"cca2":"圭亚那","cca1":"GY","region":"南美","city":"乔治城"},{"iata":"PAP","lat":18.5799999237,"lon":-72.2925033569,"cca2":"海地","cca1":"HT","region":"北美洲","city":"太子港"},{"iata":"TGU","lat":14.0608,"lon":-87.2172,"cca2":"洪都拉斯","cca1":"HN","region":"南美","city":"洪都拉斯"},{"iata":"HKG","lat":22.3089008331,"lon":113.915000916,"cca2":"香港","cca1":"HK","region":"亚太","city":"香港"},{"iata":"BUD","lat":47.4369010925,"lon":19.2555999756,"cca2":"匈牙利","cca1":"HU","region":"欧洲","city":"布达佩斯"},{"iata":"KEF","lat":63.9850006104,"lon":-22.6056003571,"cca2":"冰岛","cca1":"IS","region":"欧洲","city":"雷克雅未克"},{"iata":"AMD","lat":23.0225,"lon":72.5714,"cca2":"印度","cca1":"IN","region":"亚太","city":"艾哈迈达巴德"},{"iata":"BLR","lat":13.7835719,"lon":76.6165937,"cca2":"印度","cca1":"IN","region":"亚太","city":"班加罗尔"},{"iata":"BBI","lat":20.2961,"lon":85.8245,"cca2":"印度","cca1":"IN","region":"亚太","city":"布巴内斯瓦尔"},{"iata":"IXC","lat":30.673500061,"lon":76.7884979248,"cca2":"印度","region":"亚太","city":"昌迪加尔"},{"iata":"MAA","lat":12.9900054932,"lon":80.1692962646,"cca2":"印度","cca1":"IN","region":"亚太","city":"金奈"},{"iata":"HYD","lat":17.2313175201,"lon":78.4298553467,"cca2":"印度","cca1":"IN","region":"亚太","city":"海得拉巴"},{"iata":"CNN","lat":11.915858,"lon":75.55094,"cca2":"印度","cca1":"IN","region":"亚太","city":"坎纳诺尔"},{"iata":"KNU","lat":26.4499,"lon":80.3319,"cca2":"印度","cca1":"IN","region":"亚太","city":"坎普尔"},{"iata":"COK","lat":9.9312,"lon":76.2673,"cca2":"印度","cca1":"IN","region":"亚太","city":"高知城"},{"iata":"CCU","lat":22.6476933,"lon":88.4349249,"cca2":"印度","cca1":"IN","region":"亚太","city":"加尔各答"},{"iata":"BOM","lat":19.0886993408,"lon":72.8678970337,"cca2":"印度","cca1":"IN","region":"亚太","city":"孟买"},{"iata":"NAG","lat":21.1610714,"lon":79.0024702,"cca2":"印度","cca1":"IN","region":"亚太","city":"那格浦尔"},{"iata":"DEL","lat":28.5664997101,"lon":77.1031036377,"cca2":"印度","cca1":"IN","region":"亚太","city":"新德里"},{"iata":"PAT","lat":25.591299057,"lon":85.0879974365,"cca2":"印度","cca1":"IN","region":"亚太","city":"巴特那"},{"iata":"DPS","lat":-8.748169899,"lon":115.1669998169,"cca2":"印尼","cca1":"ID","region":"亚太","city":"登巴萨"},{"iata":"CGK","lat":-6.1275229,"lon":106.6515118,"cca2":"印尼","cca1":"ID","region":"亚太","city":"雅加达"},{"iata":"JOG","lat":-7.7881798744,"lon":110.4319992065,"cca2":"印尼","cca1":"ID","region":"亚太","city":"日惹特区"},{"iata":"BGW","lat":33.2625007629,"lon":44.2346000671,"cca2":"伊拉克","cca1":"IQ","region":"中东","city":"巴格达"},{"iata":"BSR","lat":30.5491008759,"lon":47.6621017456,"cca2":"伊拉克","cca1":"IQ","region":"中东","city":"巴士拉"},{"iata":"EBL","lat":36.1901,"lon":43.993,"cca2":"伊拉克","cca1":"IQ","region":"中东","city":"阿尔比尔"},{"iata":"NJF","lat":31.989722,"lon":44.404167,"cca2":"伊拉克","cca1":"IQ","region":"中东","city":"纳杰夫"},{"iata":"XNH","lat":30.9358005524,"lon":46.0900993347,"cca2":"伊拉克","cca1":"IQ","region":"中东","city":"纳西里耶"},{"iata":"ISU","lat":35.5668,"lon":45.4161,"cca2":"伊拉克","cca1":"IQ","region":"中东","city":"苏莱曼尼亚"},{"iata":"ORK","lat":51.8413009644,"lon":-8.491109848,"cca2":"爱尔兰","cca1":"IE","region":"欧洲","city":"科克"},{"iata":"DUB","lat":53.4212989807,"lon":-6.270070076,"cca2":"爱尔兰","cca1":"IE","region":"欧洲","city":"都柏林"},{"iata":"HFA","lat":32.78492,"lon":34.96069,"cca2":"以色列","cca1":"IL","region":"中东","city":"海法"},{"iata":"TLV","lat":32.0113983154,"lon":34.8866996765,"cca2":"以色列","cca1":"IL","region":"中东","city":"特拉维夫"},{"iata":"MXP","lat":45.6305999756,"lon":8.7281103134,"cca2":"意大利","cca1":"IT","region":"欧洲","city":"米兰"},{"iata":"PMO","lat":38.16114,"lon":13.31546,"cca2":"意大利","cca1":"IT","region":"欧洲","city":"巴勒莫"},{"iata":"FCO","lat":41.8045005798,"lon":12.2508001328,"cca2":"意大利","cca1":"IT","region":"欧洲","city":"罗马"},{"iata":"KIN","lat":17.9951,"lon":-76.7846,"cca2":"牙买加","cca1":"JM","region":"北美洲","city":"金斯顿"},{"iata":"FUK","lat":33.5902,"lon":130.4017,"cca2":"日本","cca1":"JP","region":"亚太","city":"福冈"},{"iata":"OKA","lat":26.1958,"lon":127.646,"cca2":"日本","cca1":"JP","region":"亚太","city":"那霸"},{"iata":"KIX","lat":34.4272994995,"lon":135.244003296,"cca2":"日本","cca1":"JP","region":"亚太","city":"大阪"},{"iata":"NRT","lat":35.7647018433,"lon":140.386001587,"cca2":"日本","cca1":"JP","region":"亚太","city":"东京"},{"iata":"AMM","lat":31.7226009369,"lon":35.9931983948,"cca2":"约旦","cca1":"JO","region":"中东","city":"安曼"},{"iata":"ALA","lat":43.3521003723,"lon":77.0404968262,"cca2":"哈萨克斯坦","cca1":"KZ","region":"亚太","city":"阿拉木图"},{"iata":"MBA","lat":-4.0348300934,"lon":39.5942001343,"cca2":"肯尼亚","cca1":"KE","region":"非洲","city":"蒙巴萨"},{"iata":"NBO","lat":-1.319239974,"lon":36.9277992249,"cca2":"肯尼亚","cca1":"KE","region":"非洲","city":"内罗毕"},{"iata":"ICN","lat":37.4691009521,"lon":126.450996399,"cca2":"韩国","cca1":"KR","region":"亚太","city":"首尔"},{"iata":"KWI","lat":29.226600647,"lon":47.9688987732,"cca1":"科威特","cca1":"KW","region":"中东","city":"科威特城"},{"iata":"VTE","lat":17.9757,"lon":102.5683,"cca2":"老挝","cca1":"LA","region":"亚太","city":"万象"},{"iata":"RIX","lat":56.9235992432,"lon":23.9710998535,"cca2":"拉脱维亚","cca1":"LV","region":"欧洲","city":"里加"},{"iata":"BEY","lat":33.8208999634,"lon":35.4883995056,"cca2":"黎巴嫩","cca1":"LB","region":"中东","city":"贝鲁特"},{"iata":"VNO","lat":54.6341018677,"lon":25.2858009338,"cca2":"立陶宛","cca1":"LT","region":"欧洲","city":"维尔纽斯"},{"iata":"LUX","lat":49.6265983582,"lon":6.211520195,"cca2":"卢森堡","cca1":"LU","region":"欧洲","city":"卢森堡"},{"iata":"MFM","lat":22.1495990753,"lon":113.592002869,"cca2":"澳门","cca1":"MO","region":"亚太","city":"澳门"},{"iata":"TNR","lat":-18.91368,"lon":47.53613,"cca2":"马达加斯加","cca1":"MG","region":"非洲","city":"塔那那利佛"},{"iata":"JHB","lat":1.635848,"lon":103.665943,"cca2":"马来西亚","cca1":"MY","region":"亚太","city":"柔佛州"},{"iata":"KUL","lat":2.745579958,"lon":101.709999084,"cca2":"马来西亚","cca1":"MY","region":"亚太","city":"吉隆坡"},{"iata":"MLE","lat":4.1748,"lon":73.50888,"cca2":"马尔代夫","cca1":"MV","region":"亚太","city":"马累"},{"iata":"MRU","lat":-20.4302005768,"lon":57.6836013794,"cca2":"毛里求斯","cca1":"MU","region":"非洲","city":"路易港"},{"iata":"GDL","lat":20.5217990875,"lon":-103.3109970093,"cca2":"墨西哥","cca1":"MX","region":"北美洲","city":"瓜达拉哈拉"},{"iata":"MEX","lat":19.4363002777,"lon":-99.0720977783,"cca2":"墨西哥","cca1":"MX","region":"北美洲","city":"墨西哥"},{"iata":"QRO","lat":20.6173000336,"lon":-100.185997009,"cca2":"墨西哥","cca1":"MX","region":"北美洲","city":"克雷塔羅"},{"iata":"KIV","lat":46.9277000427,"lon":28.9309997559,"cca2":"摩尔多瓦","cca1":"MD","region":"欧洲","city":"基希讷乌"},{"iata":"ULN","lat":47.8431015015,"lon":106.766998291,"cca2":"蒙古","cca1":"MN","region":"亚太","city":"蒙古"},{"iata":"CMN","lat":33.3675003052,"lon":-7.5899701118,"cca2":"摩洛哥","cca1":"MA","region":"非洲","city":"卡萨布兰卡"},{"iata":"MPM","lat":-25.9207992554,"lon":32.5726013184,"cca2":"莫桑比克","cca1":"MZ","region":"非洲","city":"马普托"},{"iata":"MDL","lat":21.7051697,"lon":95.9695206,"cca2":"缅甸","cca1":"MM","region":"亚太","city":"曼德勒"},{"iata":"RGN","lat":16.9073009491,"lon":96.1332015991,"cca2":"缅甸","cca1":"MM","region":"亚太","city":"仰光"},{"iata":"KTM","lat":27.6965999603,"lon":85.3591003418,"cca2":"尼泊尔","cca1":"NP","region":"亚太","city":"加德满都"},{"iata":"AMS","lat":52.3086013794,"lon":4.7638897896,"cca2":"荷兰","cca1":"NL","region":"欧洲","city":"阿姆斯特丹"},{"iata":"NOU","lat":-22.0146007538,"lon":166.212997436,"cca2":"新喀里多尼亚","cca1":"NC","region":"大洋洲","city":"努美阿"},{"iata":"AKL","lat":-37.0080986023,"lon":174.792007446,"cca2":"新西兰","cca1":"NZ","region":"大洋洲","city":"奥克兰"},{"iata":"CHC","lat":-43.4893989563,"lon":172.5319976807,"cca2":"新西兰","cca1":"NZ","region":"大洋洲","city":"克赖斯特彻"},{"iata":"LOS","lat":6.5773701668,"lon":3.321160078,"cca2":"尼日利亚","cca1":"NG","region":"非洲","city":"拉各斯"},{"iata":"OSL","lat":60.193901062,"lon":11.100399971,"cca2":"挪威","cca1":"NO","region":"欧洲","city":"奥斯陆"},{"iata":"MCT","lat":23.5932998657,"lon":58.2844009399,"cca2":"阿曼","cca1":"OM","region":"中东","city":"马斯喀特"},{"iata":"ISB","lat":33.6166992188,"lon":73.0991973877,"cca2":"巴基斯坦","cca1":"PK","region":"亚太","city":"伊斯兰堡"},{"iata":"KHI","lat":24.9064998627,"lon":67.1607971191,"cca2":"巴基斯坦","cca1":"PK","region":"亚太","city":"卡拉奇"},{"iata":"LHE","lat":31.5216007233,"lon":74.4036026001,"cca2":"巴基斯坦","cca1":"PK","region":"亚太","city":"拉合尔"},{"iata":"ZDM","lat":32.2719,"lon":35.0194,"cca2":"巴勒斯坦","cca1":"PS","region":"中东","city":"拉姆安拉"},{"iata":"PTY","lat":9.0713596344,"lon":-79.3834991455,"cca2":"巴拿马","cca1":"PA","region":"南美","city":"巴拿马城"},{"iata":"ASU","lat":-25.2399997711,"lon":-57.5200004578,"cca2":"巴拉圭","cca1":"PY","region":"南美","city":"亚松森"},{"iata":"LIM","lat":-12.021900177,"lon":-77.1143035889,"cca2":"秘鲁","cca1":"PE","region":"南美","city":"利马"},{"iata":"CGY","lat":8.4156198502,"lon":124.611000061,"cca2":"菲律宾","cca1":"PH","region":"亚太","city":"哥打巴托市"},{"iata":"CEB","lat":10.3074998856,"lon":123.978996277,"cca2":"菲律宾","cca1":"PH","region":"亚太","city":"宿务"},{"iata":"MNL","lat":14.508600235,"lon":121.019996643,"cca2":"菲律宾","cca1":"PH","region":"亚太","city":"马尼拉"},{"iata":"WAW","lat":52.1656990051,"lon":20.9671001434,"cca2":"波兰","cca1":"PL","region":"欧洲","city":"华沙"},{"iata":"LIS","lat":38.7812995911,"lon":-9.1359195709,"cca2":"葡萄牙","cca1":"PT","region":"欧洲","city":"里斯本"},{"iata":"DOH","lat":25.2605946,"lon":51.6137665,"cca2":"卡塔尔","cca1":"QA","region":"中东","city":"多哈"},{"iata":"RUN","lat":-20.8871002197,"lon":55.5102996826,"cca2":"留尼汪","cca1":"RE","region":"非洲","city":"圣但尼"},{"iata":"OTP","lat":44.5722007751,"lon":26.1021995544,"cca2":"罗马尼亚","cca1":"RO","region":"欧洲","city":"布加勒斯特"},{"iata":"KHV","lat":48.5279998779,"lon":135.18800354,"cca2":"俄罗斯","cca1":"RU","region":"亚太","city":"哈巴罗夫斯克"},{"iata":"KJA","lat":56.0153,"lon":92.8932,"cca2":"俄罗斯","cca1":"RU","region":"亚太","city":"克拉斯诺亚尔斯克"},{"iata":"DME","lat":55.4087982178,"lon":37.9062995911,"cca2":"俄罗斯","cca1":"RU","region":"欧洲","city":"莫斯科"},{"iata":"LED","lat":59.8003005981,"lon":30.2625007629,"cca2":"俄罗斯","cca1":"RU","region":"欧洲","city":"圣彼得堡"},{"iata":"KLD","lat":56.8587,"lon":35.9176,"cca2":"俄罗斯","cca1":"RU","region":"欧洲","city":"特维尔"},{"iata":"SVX","lat":56.8431,"lon":60.6454,"cca2":"俄罗斯","cca1":"RU","region":"亚太","city":"叶卡捷琳堡"},{"iata":"KGL","lat":-1.9686299563,"lon":30.1394996643,"cca2":"卢旺达","cca1":"RW","region":"非洲","city":"基加利"},{"iata":"DMM","lat":26.471200943,"lon":49.7979011536,"cca2":"沙特阿拉伯","cca1":"SA","region":"中东","city":"达曼"},{"iata":"JED","lat":21.679599762,"lon":39.15650177,"cca2":"沙特阿拉伯","cca1":"SA","region":"中东","city":"吉达"},{"iata":"RUH","lat":24.9575996399,"lon":46.6987991333,"cca2":"沙特阿拉伯","cca1":"SA","region":"中东","city":"利雅得"},{"iata":"DKR","lat":14.7412099,"lon":-17.4889771,"cca2":"塞内加尔","cca1":"SN","region":"非洲","city":"达喀尔"},{"iata":"BEG","lat":44.8184013367,"lon":20.3090991974,"cca2":"塞尔维亚","cca1":"RS","region":"欧洲","city":"贝尔格莱德"},{"iata":"SIN","lat":1.3501900434,"lon":103.994003296,"cca2":"新加坡","cca1":"SG","region":"亚太","city":"新加坡"},{"iata":"BTS","lat":48.1486,"lon":17.1077,"cca2":"斯洛伐克","cca1":"SK","region":"欧洲","city":"布拉迪斯拉发"},{"iata":"CPT","lat":-33.9648017883,"lon":18.6016998291,"cca2":"南非","cca1":"ZA","region":"非洲","city":"开普敦"},{"iata":"DUR","lat":-29.6144444444,"lon":31.1197222222,"cca2":"南非","cca1":"ZA","region":"非洲","city":"德班"},{"iata":"JNB","lat":-26.133333,"lon":28.25,"cca2":"南非","cca1":"ZA","region":"非洲","city":"约翰内斯堡"},{"iata":"BCN","lat":41.2971000671,"lon":2.0784599781,"cca2":"西班牙","cca1":"ES","region":"欧洲","city":"巴塞罗那"},{"iata":"MAD","lat":40.4936,"lon":-3.56676,"cca2":"西班牙","cca1":"ES","region":"欧洲","city":"马德里"},{"iata":"CMB","lat":7.1807599068,"lon":79.8841018677,"cca2":"斯里兰卡","cca1":"LK","region":"亚太","city":"科伦坡"},{"iata":"PBM","lat":5.452831,"lon":-55.187783,"cca2":"苏里南","cca1":"SR","region":"南美","city":"帕拉马里博"},{"iata":"GOT","lat":57.6627998352,"lon":12.279800415,"cca2":"瑞典","cca1":"SE","region":"欧洲","city":"哥德堡"},{"iata":"ARN","lat":59.6519012451,"lon":17.9186000824,"cca2":"瑞典","cca1":"SE","region":"欧洲","city":"斯德哥尔摩"},{"iata":"GVA","lat":46.2380981445,"lon":6.1089501381,"cca2":"瑞士","cca1":"CH","region":"欧洲","city":"日内瓦"},{"iata":"ZRH","lat":47.4646987915,"lon":8.5491695404,"cca2":"瑞士","cca1":"CH","region":"欧洲","city":"苏黎世"},{"iata":"KHH","lat":22.5771007538,"lon":120.3499984741,"cca2":"台湾","cca1":"TW","region":"亚太","city":"高雄"},{"iata":"TPE","lat":25.0776996613,"lon":121.233001709,"cca2":"台湾","cca1":"TW","region":"亚太","city":"台北"},{"iata":"DAR","lat":-6.8781099319,"lon":39.2025985718,"cca2":"坦桑尼亚","cca1":"TZ","region":"非洲","city":"达累斯萨拉姆"},{"iata":"BKK","lat":13.6810998917,"lon":100.747001648,"cca2":"泰国","cca1":"TH","region":"亚太","city":"曼谷"},{"iata":"CNX","lat":18.7667999268,"lon":98.962600708,"cca2":"泰国","cca1":"TH","region":"亚太","city":"清迈"},{"iata":"URT","lat":9.1325998306,"lon":99.135597229,"cca2":"泰国","cca1":"TH","region":"亚太","city":"素叻府"},{"iata":"TUN","lat":36.8510017395,"lon":10.2271995544,"cca2":"突尼斯","cca1":"TN","region":"非洲","city":"突尼斯"},{"iata":"IST","lat":40.9768981934,"lon":28.8145999908,"cca2":"土耳其","cca1":"TR","region":"欧洲","city":"伊斯坦布尔"},{"iata":"ADB","lat":38.32377,"lon":27.14317,"cca2":"土耳其","cca1":"TR","region":"欧洲","city":"伊兹密尔"},{"iata":"KBP","lat":50.3450012207,"lon":30.8946990967,"cca2":"乌克兰","cca1":"UA","region":"欧洲","city":"基辅"},{"iata":"DXB","lat":25.2527999878,"lon":55.3643989563,"cca2":"阿联酋","cca1":"AE","region":"中东","city":"迪拜"},{"iata":"EDI","lat":55.9500007629,"lon":-3.3724999428,"cca2":"英国","cca1":"GB","region":"欧洲","city":"爱丁堡"},{"iata":"LHR","lat":51.4706001282,"lon":-0.4619410038,"cca2":"英国","cca1":"GB","region":"欧洲","city":"伦敦"},{"iata":"MAN","lat":53.3536987305,"lon":-2.2749500275,"cca2":"英国","cca1":"GB","region":"欧洲","city":"Manchester"},{"iata":"MGM","lat":32.30059814,"lon":-86.39399719,"cca2":"美国","cca1":"US","region":"北美洲","city":"蒙哥马利"},{"iata":"PHX","lat":33.434299469,"lon":-112.012001038,"cca2":"美国","cca1":"US","region":"北美洲","city":"凤凰城"},{"iata":"LAX","lat":33.94250107,"lon":-118.4079971,"cca2":"美国","cca1":"US","region":"北美洲","city":"洛杉矶"},{"iata":"SMF","lat":38.695400238,"lon":-121.591003418,"cca2":"美国","cca1":"US","region":"北美洲","city":"萨克拉门托"},{"iata":"SAN","lat":32.7336006165,"lon":-117.190002441,"cca2":"美国","cca1":"US","region":"北美洲","city":"圣地亚哥"},{"iata":"SFO","lat":37.6189994812,"lon":-122.375,"cca2":"美国","cca1":"US","region":"北美洲","city":"旧金山"},{"iata":"SJC","lat":37.3625984192,"lon":-121.929000855,"cca2":"美国","cca1":"US","region":"北美洲","city":"圣何塞"},{"iata":"DEN","lat":39.8616981506,"lon":-104.672996521,"cca2":"美国","cca1":"US","region":"北美洲","city":"丹佛"},{"iata":"JAX","lat":30.4941005707,"lon":-81.6878967285,"cca2":"美国","cca1":"US","region":"北美洲","city":"杰克逊维尔"},{"iata":"MIA","lat":25.7931995392,"lon":-80.2906036377,"cca2":"美国","cca1":"US","region":"北美洲","city":"迈阿密"},{"iata":"TLH","lat":30.3964996338,"lon":-84.3503036499,"cca2":"美国","cca1":"US","region":"北美洲","city":"塔拉哈西"},{"iata":"TPA","lat":27.9755001068,"lon":-82.533203125,"cca2":"美国","cca1":"US","region":"北美洲","city":"坦帕市"},{"iata":"ATL","lat":33.6366996765,"lon":-84.4281005859,"cca2":"美国","cca1":"US","region":"北美洲","city":"亚特兰大"},{"iata":"HNL","lat":21.3187007904,"lon":-157.9219970703,"cca2":"美国","cca1":"US","region":"北美洲","city":"檀香山"},{"iata":"ORD","lat":41.97859955,"lon":-87.90480042,"cca2":"美国","cca1":"US","region":"北美洲","city":"芝加哥"},{"iata":"IND","lat":39.717300415,"lon":-86.2944030762,"cca2":"美国","cca1":"US","region":"北美洲","city":"印第安纳波利斯"},{"iata":"BGR","lat":44.8081,"lon":-68.795,"cca2":"美国","cca1":"US","region":"北美洲","city":"班格尔"},{"iata":"BOS","lat":42.36429977,"lon":-71.00520325,"cca2":"美国","cca1":"US","region":"北美洲","city":"波士顿"},{"iata":"DTW","lat":42.2123985291,"lon":-83.3534011841,"cca2":"美国","cca1":"US","region":"北美洲","city":"底特律"},{"iata":"MSP","lat":44.8819999695,"lon":-93.2218017578,"cca2":"美国","cca1":"US","region":"北美洲","city":"明尼阿波利斯"},{"iata":"MCI","lat":39.2975997925,"lon":-94.7138977051,"cca2":"美国","cca1":"US","region":"北美洲","city":"堪萨斯城"},{"iata":"STL","lat":38.7486991882,"lon":-90.3700027466,"cca2":"美国","cca1":"US","region":"北美洲","city":"圣路易斯"},{"iata":"OMA","lat":41.3031997681,"lon":-95.8940963745,"cca2":"美国","cca1":"US","region":"北美洲","city":"奥马哈"},{"iata":"LAS","lat":36.08010101,"lon":-115.1520004,"cca2":"美国","cca1":"US","region":"北美洲","city":"拉斯维加斯"},{"iata":"EWR","lat":40.6925010681,"lon":-74.1687011719,"cca2":"美国","cca1":"US","region":"北美洲","city":"纽瓦克"},{"iata":"ABQ","lat":35.0844,"lon":-106.6504,"cca2":"美国","cca1":"US","region":"北美洲","city":"阿尔伯克基"},{"iata":"BUF","lat":42.94049835,"lon":-78.73220062,"cca2":"美国","cca1":"US","region":"北美洲","city":"布法罗"},{"iata":"CLT","lat":35.2140007019,"lon":-80.9430999756,"cca2":"美国","cca1":"US","region":"北美洲","city":"夏洛特敦"},{"iata":"CMH","lat":39.9980010986,"lon":-82.8918991089,"cca2":"美国","cca1":"US","region":"北美洲","city":"哥伦布"},{"iata":"PDX","lat":45.58869934,"lon":-122.5979996,"cca2":"美国","cca1":"US","region":"北美洲","city":"波特兰"},{"iata":"PHL","lat":39.8718986511,"lon":-75.2410964966,"cca2":"美国","cca1":"US","region":"北美洲","city":"费城"},{"iata":"PIT","lat":40.49150085,"lon":-80.23290253,"cca2":"美国","cca1":"US","region":"北美洲","city":"匹兹堡"},{"iata":"FSD","lat":43.540819819502,"lon":-96.65511577730963,"cca2":"美国","cca1":"US","region":"北美洲","city":"苏瀑布"},{"iata":"MEM","lat":35.0424003601,"lon":-89.9766998291,"cca2":"美国","cca1":"US","region":"北美洲","city":"孟菲斯"},{"iata":"BNA","lat":36.1245002747,"lon":-86.6781997681,"cca2":"美国","cca1":"US","region":"北美洲","city":"纳什维尔"},{"iata":"AUS","lat":30.1975,"lon":-97.6664,"cca2":"美国","cca1":"US","region":"北美洲","city":"奥斯汀"},{"iata":"DFW","lat":32.8968009949,"lon":-97.0380020142,"cca2":"美国","cca1":"US","region":"北美洲","city":"达拉斯"},{"iata":"IAH","lat":29.9843997955,"lon":-95.3414001465,"cca2":"美国","cca1":"US","region":"北美洲","city":"休斯顿"},{"iata":"MFE","lat":26.17580032,"lon":-98.23860168,"cca2":"美国","cca1":"US","region":"北美洲","city":"麦卡伦"},{"iata":"SLC","lat":40.7883987427,"lon":-111.977996826,"cca2":"美国","cca1":"US","region":"北美洲","city":"盐湖城"},{"iata":"IAD","lat":38.94449997,"lon":-77.45580292,"cca2":"美国","cca1":"US","region":"北美洲","city":"阿什本"},{"iata":"ORF","lat":36.8945999146,"lon":-76.2012023926,"cca2":"美国","cca1":"US","region":"北美洲","city":"诺福克"},{"iata":"RIC","lat":37.5051994324,"lon":-77.3197021484,"cca2":"美国","cca1":"US","region":"北美洲","city":"里士满"},{"iata":"SEA","lat":47.4490013123,"lon":-122.308998108,"cca2":"美国","cca1":"US","region":"北美洲","city":"西雅图"},{"iata":"TAS","lat":41.257900238,"lon":69.2811965942,"cca2":"乌兹别克斯坦","cca1":"UZ","region":"亚太","city":"塔什干"},{"iata":"HAN","lat":21.221200943,"lon":105.806999206,"cca2":"越南","cca1":"VN","region":"亚太","city":"河内"},{"iata":"SGN","lat":10.8187999725,"lon":106.652000427,"cca2":"越南","cca1":"VN","region":"亚太","city":"胡志明市"},{"iata":"HRE","lat":-17.9318008423,"lon":31.0928001404,"cca2":"津巴布韦","cca1":"ZW","region":"非洲","city":"哈拉雷"}]`//位置数据已清除不用输出，后续手动补全。

	err := json.Unmarshal([]byte(body), &locations)
	if err != nil {
		gracefulExit(fmt.Sprintf("*⚠️ 错误*\n解析locations JSON失败: %v", err), 1)
	}
	
////////////////////////////////////////////

	locationMap := make(map[string]location)
	for _, loc := range locations {
		locationMap[loc.Iata] = loc
	}

	ips, err := readIPs(*Path)
	if err != nil {
		gracefulExit(fmt.Sprintf("*⚠️ 错误*\n无法从文件中读取 IP: %v", err), 1)
	}

	var wg sync.WaitGroup
	wg.Add(len(ips))

	resultChan := make(chan result, len(ips))

	thread := make(chan struct{}, *maxThreads)

	var count int
	total := len(ips)

	for _, ip := range ips {
		thread <- struct{}{}
		go func(ip string) {
			defer func() {
				<-thread
				wg.Done()
				count++
				percentage := float64(count) / float64(total) * 100
				fmt.Printf("已完成: %d 总数: %d 已完成: %.2f%%\r", count, total, percentage)
				if count == total {
					fmt.Printf("已完成: %d 总数: %d 已完成: %.2f%%\n", count, total, percentage)
				}
			}()

			parts := strings.Fields(ip)
			if len(parts) != 2 {
				fmt.Printf("IP地址格式错误: %s\n", ip)
				return
			}
			ipAddr := parts[0]
			portStr := parts[1]

			port, err := strconv.Atoi(portStr)
			if err != nil {
				fmt.Printf("端口格式错误: %s\n", portStr)
				return
			}

			dialer := &net.Dialer{
				Timeout:   timeout,
				KeepAlive: 0,
			}
			start := time.Now()
			conn, err := dialer.Dial("tcp", net.JoinHostPort(ipAddr, strconv.Itoa(port)))
			if err != nil {
				return
			}
			defer conn.Close()

			tcpDuration := time.Since(start)
			start = time.Now()

			client := http.Client{
				Transport: &http.Transport{
					Dial: func(network, addr string) (net.Conn, error) {
						return conn, nil
					},
				},
				Timeout: timeout,
			}

			var protocol string
			if *enableTLS {
				protocol = "https://"
			} else {
				protocol = "http://"
			}
			requestURL := protocol + *TCPurl + "/cdn-cgi/trace"

			req, _ := http.NewRequest("GET", requestURL, nil)

			// 添加用户代理
			req.Header.Set("User-Agent", "Mozilla/5.0")
			req.Close = true
			resp, err := client.Do(req)
			if err != nil {
				return
			}

			duration := time.Since(start)
			if duration > maxDuration {
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return
			}

			if strings.Contains(string(body), "uag=Mozilla/5.0") {
				if matches := regexp.MustCompile(`colo=([A-Z]+)`).FindStringSubmatch(string(body)); len(matches) > 1 {
					dataCenter := matches[1]
					loc, ok := locationMap[dataCenter]
					if ok {
						fmt.Printf("发现有效IP %s 端口 %d 位置信息 %s 延迟 %d 毫秒\n", ipAddr, port, loc.City, tcpDuration.Milliseconds())
						resultChan <- result{
							ip:          ipAddr,
							port:        port,
							dataCenter:  dataCenter,
							region:      loc.Region,
							cca1:        loc.Cca1,
							cca2:        loc.Cca2,
							city:        loc.City,
							latency:     fmt.Sprintf("%d ms", tcpDuration.Milliseconds()),
							tcpDuration: tcpDuration,
						}
					} else {
						fmt.Printf("发现有效IP %s 端口 %d 位置信息未知 延迟 %d 毫秒\n", ipAddr, port, tcpDuration.Milliseconds())
						resultChan <- result{
							ip:          ipAddr,
							port:        port,
							dataCenter:  dataCenter,
							region:      "",
							cca1:        "",
							cca2:        "",
							city:        "",
							latency:     fmt.Sprintf("%d ms", tcpDuration.Milliseconds()),
							tcpDuration: tcpDuration,
						}
					}
				}
			}
		}(ip)
	}

	wg.Wait()
	close(resultChan)

	if len(resultChan) == 0 {
		fmt.Println("没有发现有效的IP")
		if *telegramToken != "" && len(chatIDs) > 0 {
			sendTelegramMessage("*⚠️ 无检测结果*")
		}
		return
	}
	var results []speedtestresult
	if *speedTest > 0 {
		fmt.Printf("开始测速\n")
		var wg2 sync.WaitGroup
		wg2.Add(*speedTest)
		count = 0
		total := len(resultChan)
		results = []speedtestresult{}
		for i := 0; i < *speedTest; i++ {
			thread <- struct{}{}
			go func() {
				defer func() {
					<-thread
					wg2.Done()
				}()
				for res := range resultChan {

					downloadSpeed := getDownloadSpeed(res.ip, res.port)
					results = append(results, speedtestresult{result: res, downloadSpeed: downloadSpeed})

					count++
					percentage := float64(count) / float64(total) * 100
					fmt.Printf("已完成: %.2f%%\r", percentage)
					if count == total {
						fmt.Printf("已完成: %.2f%%\033[0\n", percentage)
					}
				}
			}()
		}
		wg2.Wait()
	} else {
		for res := range resultChan {
			results = append(results, speedtestresult{result: res})
		}
	}

	if *speedTest > 0 {
		sort.Slice(results, func(i, j int) bool {
			return results[i].downloadSpeed > results[j].downloadSpeed
		})
	} else {
		sort.Slice(results, func(i, j int) bool {
			return results[i].result.tcpDuration < results[j].result.tcpDuration
		})
	}

	file, err := os.Create(*outFile)
	if err != nil {
		gracefulExit(fmt.Sprintf("*⚠️ 错误*\n无法创建文件: %v", err), 1)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	// 写入头部
    if *speedTest > 0 {
        writer.Write([]string{
            "IP地址", "端口", "TLS", "数据中心", "地区", "国家代码", "国家", "城市", "网络延迟", "下载速度MB/s",
        })
    } else {
        writer.Write([]string{
            "IP地址", "端口", "TLS", "数据中心", "地区", "国家代码", "国家", "城市", "网络延迟",
        })
    }
	allowedPorts := make(map[int]bool)
	if *ports != "" {
		portStrs := strings.Split(*ports, ",")
		for _, pStr := range portStrs {
			p, err := strconv.Atoi(strings.TrimSpace(pStr))
			if err == nil && p > 0 && p < 65536 {
				allowedPorts[p] = true
			}
		}
	}
	// 写入数据
    for _, res := range results {
		if len(allowedPorts) > 0 && !allowedPorts[res.result.port] {
			continue
		}
        if *speedTest > 0 && res.downloadSpeed >= float64(*speedLimit) {
            writer.Write([]string{
                res.result.ip, strconv.Itoa(res.result.port), strconv.FormatBool(*enableTLS), res.result.dataCenter,
                res.result.region, res.result.cca1, res.result.cca2, res.result.city, res.result.latency,
                fmt.Sprintf("%.2f", res.downloadSpeed),
            })
        } else if *speedTest == 0 {
            writer.Write([]string{
                res.result.ip, strconv.Itoa(res.result.port), strconv.FormatBool(*enableTLS), res.result.dataCenter,
                res.result.region, res.result.cca1, res.result.cca2, res.result.city, res.result.latency,
            })
        }
    }
	writer.Flush()
	fmt.Printf("成功将结果写入文件 %s，耗时 %d秒\n", *outFile, time.Since(startTime)/time.Second)

	// 生成检测报告
	var report strings.Builder
	duration := time.Since(startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	countryCount := make(map[string]int)
	countryNameMap := make(map[string]string)
	for _, res := range results {
		cca1 := res.result.cca1
		if cca1 == "" {
			cca1 = "UNKNOWN"
		}
		countryCount[cca1]++
		if _, ok := countryNameMap[cca1]; !ok {
			countryNameMap[cca1] = res.result.cca2
		}
	}
	var countries []string
	for cca1 := range countryCount {
		countries = append(countries, cca1)
	}
	sort.Strings(countries)

	var totalLatency float64
	var minLatency, maxLatency float64
	if len(results) > 0 {
		minLatency = float64(results[0].result.tcpDuration.Milliseconds())
		maxLatency = minLatency
		for _, res := range results {
			latencyMs := float64(res.result.tcpDuration.Milliseconds())
			totalLatency += latencyMs
			if latencyMs < minLatency {
				minLatency = latencyMs
			}
			if latencyMs > maxLatency {
				maxLatency = latencyMs
			}
		}
	}
	avgLatency := 0.0
	if len(results) > 0 {
		avgLatency = totalLatency / float64(len(results))
	}

	var avgSpeed, minSpeed, maxSpeed float64
	if *speedTest > 0 && len(results) > 0 {
		for _, res := range results {
			avgSpeed += res.downloadSpeed
			if minSpeed == 0 || res.downloadSpeed < minSpeed {
				minSpeed = res.downloadSpeed
			}
			if res.downloadSpeed > maxSpeed {
				maxSpeed = res.downloadSpeed
			}
		}
		avgSpeed /= float64(len(results))
	} else {
		avgSpeed, minSpeed, maxSpeed = 0, 0, 0
	}

	cstZone := time.FixedZone("CST", 8*3600)
	startTimeLocal := startTime.In(cstZone)

	if len(results) > 0 {
		fmt.Fprintf(&report, "*✅ 延迟/速度测试完成*\n")
		fmt.Fprintf(&report, "⏰ 开始时间: %s\n", startTimeLocal.Format("2006/01/02 15:04:05"))
		fmt.Fprintf(&report, "⏰ 运行耗时: %02d时 %02d分 %02d秒\n", hours, minutes, seconds)
		fmt.Fprintf(&report, "  - 总计测试IP: %d\n", total)
		fmt.Fprintf(&report, "  - 有效IP: %d\n", len(results))
		fmt.Fprintf(&report, "*🌍 国家分布*\n")
		for _, cca1 := range countries {
			name := countryNameMap[cca1]
			if name == "" {
				name = cca1
			}
			fmt.Fprintf(&report, "- %s %s (%d个)\n", getCountryFlag(cca1), name, countryCount[cca1])
		}
		fmt.Fprintf(&report, "*📈 延迟统计*\n")
		fmt.Fprintf(&report, "  - 均值: %.2fms\n", avgLatency)
		fmt.Fprintf(&report, "  - 最低: %.2fms\n", minLatency)
		fmt.Fprintf(&report, "  - 最高: %.2fms\n", maxLatency)
		fmt.Fprintf(&report, "*⚡️ 速度统计*\n")
		if *speedTest > 0 {
			fmt.Fprintf(&report, "  - 均值: %.2f MB/s\n", avgSpeed)
			fmt.Fprintf(&report, "  - 最高: %.2f MB/s\n", maxSpeed)
			fmt.Fprintf(&report, "  - 最低: %.2f MB/s\n", minSpeed)
		} else {
			fmt.Fprintf(&report, "  - 均值: 待测\n")
			fmt.Fprintf(&report, "  - 最高: 待测\n")
			fmt.Fprintf(&report, "  - 最低: 待测\n")
		}
	} else {
		fmt.Fprintf(&report, "*⚠️ 无检测结果*\n")
		fmt.Fprintf(&report, "⏰ 运行耗时: %02d时 %02d分 %02d秒\n", hours, minutes, seconds)
		fmt.Fprintf(&report, "  - 总计测试IP: %d\n", total)
		fmt.Fprintf(&report, "  - 有效IP: 0\n")
	}

	fmt.Println("生成检测报告:\n" + report.String())
	// 推送到 Telegram
    if *telegramToken != "" && len(chatIDs) > 0 {        
            sendTelegramMessage(report.String())
            fileInfo, err := os.Stat(*outFile)
            if err == nil && fileInfo.Size() > 0 {
                sendTelegramFile(*outFile)
            } else {
                fmt.Printf("测试结果文件 %s 不存在或为空\n", *outFile)
                sendTelegramMessage(fmt.Sprintf("*⚠️ 错误*\n测试结果文件 `%s` 不存在或为空", escapeMarkdownV2(*outFile)))
            }        
        sendTelegramMessage("*🎉 程序运行结束*")
    }
}

// readIPs 函数根据提供的路径（文件或目录）读取IP地址
func readIPs(path string) ([]string, error) {
	var ips []string
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("无法获取文件/目录信息: %w", err)
	}

	if fileInfo.IsDir() {
		// 如果是目录，遍历所有文件
		err := filepath.WalkDir(path, func(filePath string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				// 排除临时文件和隐藏文件
				if strings.HasSuffix(d.Name(), "~") || strings.HasPrefix(d.Name(), ".") {
					return nil
				}

				newIps, err := readIPsFromFile(filePath)
				if err != nil {
					fmt.Printf("读取文件 %s 时出错: %v\n", filePath, err)
					return nil // 继续处理下一个文件
				}

				// 目录遍历模式下：添加/修改输出
				fmt.Printf("正在读取文件: %s 解析到 %d 条\n", filePath, len(newIps))

				ips = append(ips, newIps...)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// 如果是文件，直接读取
		newIps, err := readIPsFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取文件 %s 时出错: %w", path, err)
		}

		// 单文件模式下：添加/修改输出
		fmt.Printf("正在读取文件: %s 解析到 %d 条\n", path, len(newIps))

		ips = newIps
	}

	// 记录去重前的总数
	totalCount := len(ips)

	// 对IP端口对进行去重
	uniqueIPs := make(map[string]bool)
	var resultIPs []string
	for _, ip := range ips {
		if _, ok := uniqueIPs[ip]; !ok {
			uniqueIPs[ip] = true
			resultIPs = append(resultIPs, ip)
		}
	}

	// 计算并打印去重结果【使用粗实线边框】
	uniqueCount := len(resultIPs)
	duplicateCount := totalCount - uniqueCount

	// 定义框的宽度和内容
	boxWidth := 50

	// 格式化输出字符串
	content := fmt.Sprintf("总计 %d 条，最终 %d 条，去重 %d 条。", totalCount, uniqueCount, duplicateCount)

	// 计算终端显示宽度 (假设中文字符占 2 栏)
	contentDisplayWidth := 0
	for _, r := range content {
		if r <= 127 {
			contentDisplayWidth += 1
		} else {
			contentDisplayWidth += 2
		}
	}

	// 定义新的粗实线字符
	const (
		cornerTopLeft     = "┏"
		cornerTopRight    = "┓"
		cornerBottomLeft  = "┗"
		cornerBottomRight = "┛"
		lineHorizontal    = "━" // 粗横线
		lineVertical      = "┃" // 粗竖线
	)
	
	// 创建分隔线
	line := strings.Repeat(lineHorizontal, boxWidth)

	// 辅助函数：居中对齐内容
	if contentDisplayWidth > boxWidth {
		contentDisplayWidth = boxWidth
	}

	totalPadding := boxWidth - contentDisplayWidth
	leftPaddingLength := totalPadding / 2
	rightPaddingLength := totalPadding - leftPaddingLength 

	leftPadding := strings.Repeat(" ", leftPaddingLength)
	rightPadding := strings.Repeat(" ", rightPaddingLength)

	// 打印信息框
	fmt.Printf("\n%s%s%s\n", cornerTopLeft, line, cornerTopRight)
	fmt.Printf("%s%s%s%s%s\n", lineVertical, leftPadding, content, rightPadding, lineVertical)
	fmt.Printf("%s%s%s\n\n", cornerBottomLeft, line, cornerBottomRight)

	return resultIPs, nil
}

// readIPsFromFile 从单个文件中逐行读取IP地址和端口，支持多种格式
func readIPsFromFile(filePath string) ([]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    ips := make([]string, 0)

    // 如果是CSV文件，单独解析
    if strings.ToLower(filepath.Ext(filePath)) == ".csv" {
        reader := csv.NewReader(file)
        reader.TrimLeadingSpace = true

        // 自动判断分隔符（先试逗号，失败再试制表符）
        records, err := reader.ReadAll()
        if err != nil {
            file.Seek(0, 0)
            reader = csv.NewReader(file)
            reader.Comma = '\t'
            records, err = reader.ReadAll()
            if err != nil {
                return nil, err
            }
        }

        // 去掉UTF-8 BOM
        if len(records) > 0 && len(records[0]) > 0 && strings.HasPrefix(records[0][0], "\ufeff") {
            records[0][0] = strings.TrimPrefix(records[0][0], "\ufeff")
        }

        for i, record := range records {
            if i == 0 {
                continue // 跳过标题行
            }
            if len(record) < 3 {
                continue
            }
            ipAddr := strings.TrimSpace(record[0]) // 第1列 ip
            portStr := strings.TrimSpace(record[1]) // 第2列 port
            if ipAddr == "" || portStr == "" {
                continue
            }
            port, err := strconv.Atoi(portStr)
            if err == nil && port > 0 && port < 65536 {
                ip := fmt.Sprintf("%s %d", ipAddr, port)
                ips = append(ips, ip)
            }
        }
        return ips, nil
    }

    scanner := bufio.NewScanner(file)

    type jsonIP struct {
        IP   string `json:"ip"`
        Port string `json:"port"`
    }

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue // 跳过空行和注释
        }

        // 支持解析代理链接 (如 vless://..., vmess://..., trojan://..., ss://...)
        // 支持 IPv4、[IPv6]:port、IPv6 port、域名、域名:port、无端口默认443
        proxyPattern := regexp.MustCompile(`@(\[[0-9a-fA-F:]+\]|[0-9]{1,3}(?:\.[0-9]{1,3}){3}|[a-zA-Z0-9.-]+)(?::(\d{1,5}))?`)
        if matches := proxyPattern.FindStringSubmatch(line); len(matches) >= 2 {
            host := strings.Trim(matches[1], "[]")
            port := "443" // 默认端口
            if len(matches) == 3 && matches[2] != "" {
                port = matches[2]
            }
            if p, err := strconv.Atoi(port); err == nil && p > 0 && p < 65536 {
                ips = append(ips, fmt.Sprintf("%s %d", host, p))
                continue
            }
        }

        // 兼容 [IPv6]:port 或 [IPv6]、域名:port 或 域名 格式（冒号分隔，无端口默认443）
        hostColonPattern := regexp.MustCompile(`^(\[[0-9a-fA-F:]+]|[a-zA-Z0-9.-]+)(?::(\d{1,5}))?$`)
        if matches := hostColonPattern.FindStringSubmatch(line); len(matches) >= 2 {
            host := strings.Trim(matches[1], "[]")
            port := "443" // 默认端口
            if len(matches) == 3 && matches[2] != "" {
                port = matches[2]
            }
            if p, err := strconv.Atoi(port); err == nil && p > 0 && p < 65536 {
                ips = append(ips, fmt.Sprintf("%s %d", host, p))
                continue
            }
        }

        // 兼容 IPv6 port 或 域名 port 格式（空格分隔）
        hostSpacePattern := regexp.MustCompile(`^([0-9a-fA-F:]+|[a-zA-Z0-9.-]+)\s+(\d{1,5})$`)
        if matches := hostSpacePattern.FindStringSubmatch(line); len(matches) == 3 {
            host := matches[1]
            port := matches[2]
            if p, err := strconv.Atoi(port); err == nil && p > 0 && p < 65536 {
                ips = append(ips, fmt.Sprintf("%s %d", host, p))
                continue
            }
        }

        // 兼容无端口的 [IPv6] 或 域名（默认443）
        hostOnlyPattern := regexp.MustCompile(`^(\[[0-9a-fA-F:]+]|[a-zA-Z0-9.-]+)$`)
        if matches := hostOnlyPattern.FindStringSubmatch(line); len(matches) == 2 {
            host := strings.Trim(matches[1], "[]")
            port := 443
            ips = append(ips, fmt.Sprintf("%s %d", host, port))
            continue
        }

		// 支持 "open tcp PORT IP TIMESTAMP" 格式
		openTcpPattern := regexp.MustCompile(`open tcp (\d+) ([\d.]+|\[[0-9a-fA-F:]+]|[a-zA-Z0-9.-]+) \d+`)
		if matches := openTcpPattern.FindStringSubmatch(line); len(matches) == 3 {
			portStr := matches[1]
			host := strings.Trim(matches[2], "[]")
			if p, err := strconv.Atoi(portStr); err == nil && p > 0 && p < 65536 {
				ips = append(ips, fmt.Sprintf("%s %d", host, p))
				continue
			}
		}

        // 支持 "IP:port | ..." 格式
        ipPipePattern := regexp.MustCompile(`^([\d.]+|\[[0-9a-fA-F:]+]|[a-zA-Z0-9.-]+):(\d{1,5}) \|`)
        if matches := ipPipePattern.FindStringSubmatch(line); len(matches) == 3 {
            host := strings.Trim(matches[1], "[]")
            port := matches[2]
            if p, err := strconv.Atoi(port); err == nil && p > 0 && p < 65536 {
                ips = append(ips, fmt.Sprintf("%s %d", host, p))
                continue
            }
        }

        var ipAddr string
        var portStr string

        // 先尝试解析为JSON
        var jip jsonIP
        if err := json.Unmarshal([]byte(line), &jip); err == nil && jip.IP != "" && jip.Port != "" {
            ipAddr = jip.IP
            portStr = jip.Port
        } else {
            // 尝试解析各种格式
            if strings.Contains(line, ":") {
                lastColon := strings.LastIndex(line, ":")
                if lastColon != -1 {
                    ipAddr = line[:lastColon]
                    portAndComment := line[lastColon+1:]
                    if strings.Contains(portAndComment, "#") {
                        parts := strings.SplitN(portAndComment, "#", 2)
                        portStr = parts[0]
                    } else {
                        portStr = portAndComment
                    }
                }
            } else if strings.Contains(line, "：") {
                lastColon := strings.LastIndex(line, "：")
                if lastColon != -1 {
                    ipAddr = line[:lastColon]
                    portAndComment := line[lastColon+3:] // 全角冒号占3个字节
                    if strings.Contains(portAndComment, "#") {
                        parts := strings.SplitN(portAndComment, "#", 2)
                        portStr = parts[0]
                    } else {
                        portStr = portAndComment
                    }
                }
            } else if strings.Contains(line, ",") {
                parts := strings.Split(line, ",")
                if len(parts) >= 2 {
                    ipAddr = parts[0]
                    portStr = parts[1]
                }
            } else {
                lineWithSpace := strings.ReplaceAll(line, "：", " ")
                parts := strings.Fields(lineWithSpace)
                if len(parts) >= 2 {
                    ipAddr = parts[0]
                    portStr = parts[1]
                }
            }
        }

        if ipAddr != "" && portStr != "" {
            ipAddr = strings.Trim(strings.Trim(ipAddr, "[]"), " \t")
            portStr = strings.TrimSpace(portStr)
            port, err := strconv.Atoi(portStr)
            if err == nil && port > 0 && port < 65536 {
                ip := fmt.Sprintf("%s %d", ipAddr, port)
                ips = append(ips, ip)
            } else {
                fmt.Printf("跳过无效行(格式错误): %s\n", line)
            }
        } else {
            fmt.Printf("跳过无效行(无法解析): %s\n", line)
        }
    }
    return ips, scanner.Err()
}


// 测速函数
func getDownloadSpeed(ip string, port int) float64 {
	var protocol string
	if *enableTLS {
		protocol = "https://"
	} else {
		protocol = "http://"
	}
	speedTestURL := protocol + *speedTestURL
	// 创建请求
	req, _ := http.NewRequest("GET", speedTestURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// 创建TCP连接
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 0,
	}
	conn, err := dialer.Dial("tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		return 0
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	fmt.Printf("正在测试IP %s 端口 %d\n", ip, port)
	startTime := time.Now()
	// 创建HTTP客户端
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return conn, nil
			},
		},
		//设置单个IP测速最长时间为5秒
		Timeout: 5 * time.Second,
	}
	// 发送请求
	req.Close = true
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("IP %s 端口 %d 测速无效\n", ip, port)
		return 0
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 复制响应体到/dev/null，并计算下载速度
	written, _ := io.Copy(io.Discard, resp.Body)
	duration := time.Since(startTime)
	speed := float64(written) / duration.Seconds() / 1024 / 1024

	// 输出结果
	fmt.Printf("IP %s 端口 %d 下载速度 %.2f MB/s\n", ip, port, speed)
	return speed
}