package command

import (
	"fmt"
	"github.com/levigross/grequests"
	"github.com/shirou/gopsutil/load"
	log "github.com/sirupsen/logrus"
	"gopkg.in/tucnak/telebot.v2"
	"runtime"
	"source.hitokoto.cn/hitokoto/telegram_bot/src/build"
	"strconv"
	"strings"
	"time"
)

func Status(b *telebot.Bot) {
	b.Handle("/status", func(m *telebot.Message) {
		response, err := grequests.Get("https://status.hitokoto.cn", nil)
		if err != nil {
			log.Errorf("尝试获取统计数据时出现错误，错误信息： %s\n", err)
			_, err = b.Send(m.Chat, "很抱歉，尝试获取数据时发生错误。")
			if err != nil {
				log.Errorf("尝试发送消息时出现错误，错误信息：%s \n", err)
			}
			return
		}
		data := &hitokotoStatusApiV1Response{}
		err = response.JSON(data)
		if err != nil {
			log.Errorf("尝试解析统计数据时发生错误，错误信息： %s", err)
			_, err = b.Send(m.Chat, "很抱歉，尝试解析数据时发生错误。")
			if err != nil {
				log.Errorf("尝试发送消息时出现错误，错误信息：%s \n", err)
			}
			return
		}

		// 读取系统负载
		load, err := load.Avg()
		if err != nil {
			log.Errorf("尝试解析系统负载时发生错误，错误信息： %s", err)
			_, err = b.Send(m.Chat, "很抱歉，尝试解析系统负载时发生错误。")
			if err != nil {
				log.Errorf("尝试发送消息时出现错误，错误信息：%s \n", err)
			}
			return
		}
		// log.Debug(data)
		_, err = b.Send(m.Chat, fmt.Sprintf(`*[一言统计信息]*
句子总数： %s
现存分类： %s
服务负载： %s
内存占用： %s MB
每分请求： %s
每时请求： %s
当日请求： %s

*[调试信息]*
当前时间： %s
操作系统： %s
设备架构： %s
系统负载： %s
程序版本： v%s
运行环境： %s
编译时间： %s
编译哈希： %s
`,
			strconv.Itoa(data.Status.Hitokoto.Total),
			strings.Join(data.Status.Hitokoto.Categroy, ","),
			loadToString(data.Status.Load[0])+","+loadToString(data.Status.Load[1])+","+loadToString(data.Status.Load[2]),
			loadToString(data.Status.Memory),
			strconv.FormatUint(data.Requests.All.PastMinute, 10),
			strconv.FormatUint(data.Requests.All.PastHour, 10),
			strconv.FormatUint(data.Requests.All.PastDay, 10),
			time.Now().Format("2006年1月2日 15:04:05"),
			runtime.GOOS,
			runtime.GOARCH,
			loadToString(load.Load1)+","+loadToString(load.Load5)+","+loadToString(load.Load15),
			build.Version,
			runtime.Version(),
			build.BuildTime,
			build.GitCommit,
		),
			&telebot.SendOptions{
				ParseMode: "Markdown",
			},
		)
	})
}

type hitokotoStatusApiV1Response struct { // 因为不需要使用全部数据，所以这里就只解析部分了
	Status   status   `json:"status"`
	Requests requests `json:"requests"`
}

type status struct {
	Load     []float64 `json:"load"`
	Memory   float64   `json:"memory"`
	Hitokoto hitokoto  `json:"hitokoto"`
}

type hitokoto struct {
	Total    int      `json:"total"`
	Categroy []string `json:"categroy"` // 别在意这个，Api 写的时候就打错了...
}

type requests struct {
	All all `json:"all"`
}

type all struct {
	Total      uint64 `json:"total"`
	PastMinute uint64 `json:"pastMinute"`
	PastHour   uint64 `json:"pastHour"`
	PastDay    uint64 `json:"pastDay"`
}

func loadToString(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
