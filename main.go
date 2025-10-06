package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v4"
)

func attrLocal(attr string) string {
	if attr == "A" {
		return "优势"
	}

	return "劣势"
}

type RandomInfo struct {
	Lock           *sync.Mutex
	EthBlockHash   string
	HashUsedCount  int
	LastHashUpdate time.Time
}

var randomInfo *RandomInfo = &RandomInfo{
	Lock:           &sync.Mutex{},
	EthBlockHash:   "",
	HashUsedCount:  0,
	LastHashUpdate: time.Unix(0, 0),
}

func parseDice(diceStr string) (int, int, int, string, error) {
	dIndex := strings.Index(diceStr, "d")
	if dIndex == -1 {
		return 0, 0, 0, "", fmt.Errorf("invalid dice format")
	}

	diceCountStr := diceStr[:dIndex]
	diceCount := 1
	if diceCountStr != "" {
		var err error
		diceCount, err = strconv.Atoi(diceCountStr)
		if err != nil {
			return 0, 0, 0, "", err
		}
	}

	// find end of diceFace
	faceEnd := dIndex + 1
	for faceEnd < len(diceStr) && (diceStr[faceEnd] >= '0' && diceStr[faceEnd] <= '9') {
		faceEnd++
	}

	diceFaceStr := diceStr[dIndex+1 : faceEnd]
	diceFace, err := strconv.Atoi(diceFaceStr)
	if err != nil {
		return 0, 0, 0, "", err
	}

	remaining := diceStr[faceEnd:]
	adder := 0
	re := regexp.MustCompile(`[+-]\d+`)
	matches := re.FindAllString(remaining, -1)
	adderStr := strings.Join(matches, "")
	for _, match := range matches {
		num, err := strconv.Atoi(match)
		if err != nil {
			return 0, 0, 0, "", err
		}
		adder += num
	}

	return diceCount, diceFace, adder, adderStr, nil
}

func updateEthBlockHash() {
	url := os.Getenv("ETH_RPC_URL")
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{"latest", true},
		"id":      1,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		logrus.WithField("err", err).Error("updateEthBlockHash: Failed to marshal JSON")
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.WithField("err", err).Error("updateEthBlockHash: Failed to fetch block")
		return
	}
	defer resp.Body.Close()
	var response map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logrus.WithField("err", err).Error("updateEthBlockHash: Failed to decode response")
		return
	}
	if result, ok := response["result"].(map[string]interface{}); ok {
		hash, okHash := result["hash"].(string)
		height, okHeight := result["number"].(string)

		if okHash && okHeight {
			randomInfo.EthBlockHash = hash
			randomInfo.LastHashUpdate = time.Now()
			randomInfo.HashUsedCount = 0
			logrus.WithFields(logrus.Fields{
				"Height": height,
				"Hash":   hash,
			}).Info("ETH Block Hash has been updated")
		} else {
			logrus.Error("updateEthBlockHash: Hash not found in response")
		}
	} else {
		logrus.Error("updateEthBlockHash: Result not found in response")
	}
}

func getRandomNumber(max int, count int) ([]int, error) {
	randomInfo.Lock.Lock()
	defer randomInfo.Lock.Unlock()

	// if time diff > 12s
	if time.Since(randomInfo.LastHashUpdate) >= 12*time.Second {
		updateEthBlockHash()
	}

	seed, err := strconv.ParseInt(randomInfo.EthBlockHash[51:], 16, 64)
	if err != nil {
		return []int{}, err
	}

	// Set random seed
	seed = seed + int64(randomInfo.HashUsedCount)
	rng := rand.New(rand.NewSource(seed))

	respond := []int{}
	for range count {
		respond = append(respond, min(1, rng.Intn(max+1)))
	}

	randomInfo.HashUsedCount++

	logrus.WithFields(logrus.Fields{
		"seed":   seed,
		"result": fmt.Sprintf("%v", respond),
	}).Info("Random number generated")

	return respond, nil
}

func main() {
	logger := logrus.New()

	err := godotenv.Load()
	if err != nil {
		logger.WithField("err", err).Warnf("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		logrus.Fatalf("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	tgApi := os.Getenv("TELEGRAM_API")
	if tgApi == "" {
		tgApi = "https://api.telegram.org"
	}
	logrus.Infof("Using API Endpoint: %v", tgApi)

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		URL:    tgApi,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		logger.WithField("err", err).Fatal("Failed to create bot")
	}

	logrus.WithFields(logrus.Fields{
		"ID":          bot.Me.ID,
		"Username":    bot.Me.Username,
		"DisplayName": bot.Me.FirstName,
	}).Info("Starting bot")

	bot.Handle(telebot.OnQuery, func(c telebot.Context) error {
		logrus.WithFields(logrus.Fields{
			"ID":       c.Query().Sender.ID,
			"Username": c.Query().Sender.Username,
		}).Info("User query")

		defaultResponse := []telebot.Result{
			&telebot.ArticleResult{
				Title:       "帮助 & 关于",
				Description: "使用方法 & 寻找帮助 & 报告错误",
				Text: "DnD DM\n" +
					"例子: 1d20 一个20面的色子 (1~20)\n" +
					"例子: 1d20+5 一个20面的色子+5 (6~25)\n" +
					"例子: 1d20>15 一个20面的色子(1~20) 大于15检定成功\n" +
					"例子: A 1d20>15 一个20面的色子(1~20) 带优势(扔2个取大) 大于15检定成功\n" +
					"例子: D 1d20>15 一个20面的色子(1~20) 带劣势(扔2个取小) 大于15检定成功\n" +
					"例子: A 1d20+2>15 一个20面的色子+2(2~25) 带优势(扔2个取大) 大于15检定成功\n" +
					"例子[带名字的检定]: 自定义名字 D 1d20>15 一个20面的色子(1~20) 带劣势(扔2个取小) 大于15检定成功\n" +
					"例子[建议选仅数字]: 4d8 4个8面的色子 (4~32)\n" +
					"属性检定：带 大成功(20) 和 大失败(1)\n" +
					"限制: 色子数量不能大于1000 & 面数不能大于500\n" +
					"Github pedxyuyuko/dnd_dicemaster",
			},
		}

		rawRequest := c.Query().Text
		if rawRequest == "" {
			rawRequest = "1d20>10"
		}

		diceInfo := strings.Split(rawRequest, " ")

		checkName := ""
		attr := ""
		rawDice := ""

		switch len(diceInfo) {
		case 1:
			rawDice = diceInfo[0]
		case 2:
			checkName = diceInfo[0]
			attr = diceInfo[0]
			rawDice = diceInfo[1]

			if checkName == "A" || checkName == "D" {
				checkName = ""
			} else {
				attr = ""
			}
		default:
			checkName = diceInfo[0]
			attr = diceInfo[1]
			rawDice = diceInfo[2]
		}

		rawDiceWithCompare := strings.Split(rawDice, ">")
		isChecking := len(rawDiceWithCompare) == 2
		rawDice = rawDiceWithCompare[0]

		compareValue := 0
		if isChecking {
			compareValue, _ = strconv.Atoi(rawDiceWithCompare[1])
		}

		diceCount, diceFace, adder, adderStr, err := parseDice(rawDice)
		if err != nil {
			// handle error, perhaps log or set defaults
			diceCount, diceFace, adder, adderStr = 1, 20, 0, ""
		}

		if diceCount > 1000 || diceFace > 500 {
			_ = c.Answer(&telebot.QueryResponse{
				Results: append([]telebot.Result{
					&telebot.ArticleResult{
						Title:       "数量限制",
						Description: "色子数量不能大于1000 & 面数不能大于500",
						Text:        fmt.Sprintf("[%s]: %s", rawRequest, err.Error()),
					},
				}, defaultResponse...),
				CacheTime: -1,
			})
		}

		// 2 个及以上的色子无法进行优劣势判定
		if diceCount > 1 {
			attr = ""
		}

		if attr != "" {
			diceCount = 2
		}

		diceRolled, err := getRandomNumber(diceFace, diceCount)
		if err != nil {
			_ = c.Answer(&telebot.QueryResponse{
				Results: append([]telebot.Result{
					&telebot.ArticleResult{
						Title:       "在获取随机数的时候发生了点错误",
						Description: "点击查看错误 (提交Issue pedxyuyuko/dnd_dicemaster)",
						Text:        fmt.Sprintf("[%s]: %s", rawRequest, err.Error()),
					},
				}, defaultResponse...),
				CacheTime: -1,
			})
			return err
		}

		finalDice := 0
		if attr == "A" {
			finalDice = slices.Max(diceRolled)
		}

		if attr == "D" {
			finalDice = slices.Min(diceRolled)
		}

		if attr == "" {
			for _, r := range diceRolled {
				finalDice += r
			}
		}

		finalValue := max(finalDice+adder, 1)

		respondText := fmt.Sprintf("🎲 %dd%d %v = %d", diceCount, diceFace, diceRolled, finalDice)
		if adderStr != "" {
			respondText = fmt.Sprintf("%s\n调整值: %s = %d", respondText, adderStr, adder)
		}
		respondText = fmt.Sprintf("%s\n最终结果: %d", respondText, finalValue)

		respondTextChecking := ""
		if attr != "" {
			respondTextChecking = fmt.Sprintf("(%s)", attrLocal(attr))
		}
		respondTextChecking = fmt.Sprintf("%s属性", respondTextChecking)
		if checkName != "" {
			respondTextChecking = fmt.Sprintf("%s [%s] ", respondTextChecking, checkName)
		}
		respondTextChecking = fmt.Sprintf("%s检定", respondTextChecking)
		respondTitleChecking := fmt.Sprintf("%s 掷🎲 %s", respondTextChecking, rawDice)
		if c.Query().Text == "" {
			respondTitleChecking = fmt.Sprintf("[属性检定] 掷🎲 %s", rawDice)
		}
		if diceFace == 20 {
			if finalDice == 1 {
				respondTextChecking = fmt.Sprintf("%s大失败(1)", respondTextChecking)
			} else if finalDice == 20 {
				respondTextChecking = fmt.Sprintf("%s大成功(20)", respondTextChecking)
			} else if finalValue >= compareValue {
				respondTextChecking = fmt.Sprintf("%s成功 %d>=%d", respondTextChecking, finalValue, compareValue)
			} else {
				respondTextChecking = fmt.Sprintf("%s失败 %d<%d", respondTextChecking, finalValue, compareValue)
			}
		} else if finalValue >= compareValue {
			respondTextChecking = fmt.Sprintf("%s成功 %d>=%d", respondTextChecking, finalValue, compareValue)
		} else {
			respondTextChecking = fmt.Sprintf("%s失败 %d<%d", respondTextChecking, finalValue, compareValue)
		}
		respondTextChecking = fmt.Sprintf("%s\n%s", respondTextChecking, respondText)

		return c.Answer(&telebot.QueryResponse{
			Results: append([]telebot.Result{
				&telebot.ArticleResult{
					Title:       respondTitleChecking,
					Description: "举例: [智力 A 1d20+1-2>15] 1个20面色子优势最终结果+1再-2 大于15通过检定",
					Text:        respondTextChecking,
				},
				&telebot.ArticleResult{
					Title:       fmt.Sprintf("[仅数字] 掷🎲 %s", rawDice),
					Description: "举例: [A 1d20+1-2] 1个20面色子优势最终结果+1再-2",
					Text:        respondText,
				},
			}, defaultResponse...),
			CacheTime: -1,
		})
	})

	logger.Info("Bot started")
	bot.Start()
}
