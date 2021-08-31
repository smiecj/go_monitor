package reptile

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	client "github.com/smiecj/go_common/http"
)

func GetHotTopicAndClickTime() map[string]int64 {
	hotDataMap := make(map[string]int64, 0)
	contentBytes := client.DoGetRequest("https://s.weibo.com/top/summary", nil)
	document, _ := goquery.NewDocumentFromReader(bytes.NewReader(contentBytes))
	document.Find(".td-02").Each(func(index int, selection *goquery.Selection) {
		// 第一个热搜表示实时刷新数据，没有热度数据，直接忽略
		if index == 0 {
			return
		}
		currentContent := selection.Text()
		//fmt.Printf("当前查询的热搜内容为: %s\n", currentContent)
		currentContent = strings.Replace(currentContent, " ", "", -1)
		contentSplitArr := strings.Split(currentContent, "\n")
		//fmt.Printf("分割之后，解析的热搜数据为: %v\n", contentSplitArr)
		hotNum, _ := strconv.ParseInt(contentSplitArr[2], 10, 64)
		hotDataMap[contentSplitArr[1]] = hotNum
	})
	return hotDataMap
}
