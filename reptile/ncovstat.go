package reptile

import (
	"bytes"
	"encoding/json"
	"fmt"

	client "github.com/smiecj/go_common/http"
)

type NCovCityStatus struct {
	Province   string
	City       string
	Sick       int64
	Confirming int64
	Cure       int64
	Death      int64
}

func (ncovStatus NCovCityStatus) String() string {
	return fmt.Sprintf("省份：%s, 城市名字：%s, 确认中人数：%d, 患病人数：%d, 治愈人数：%d, 死亡人数：%d",
		ncovStatus.Province, ncovStatus.City, ncovStatus.Confirming, ncovStatus.Sick, ncovStatus.Cure, ncovStatus.Death)
}

func GetNcovStatus() []*NCovCityStatus {
	ncovStatusResultArr := make([]*NCovCityStatus, 0)
	contentBytes := client.DoGetRequest("https://3g.dxy.cn/newh5/view/pneumonia", nil)
	if len(contentBytes) == 0 {
		fmt.Printf("解析丁香医生网站数据失败！请检查数据: %s\n", string(contentBytes))
		return ncovStatusResultArr
	}
	// 通过爬虫解析有问题，可能是因为特殊字符，会导致解析后的数据有字符丢失，这里直接按照byte数组来解析
	byteStartIndex := bytes.Index(contentBytes, []byte("[{\"provinceName\""))
	if -1 == byteStartIndex {
		fmt.Printf("解析丁香医生网站数据失败！请检查数据: %s\n", string(contentBytes))
		return ncovStatusResultArr
	}
	byteLastPos := bytes.Index(contentBytes[byteStartIndex+1:], []byte("}catch(e){}"))
	var ncovStatusObj interface{}
	json.Unmarshal(contentBytes[byteStartIndex:byteStartIndex+byteLastPos+1], &ncovStatusObj)
	fmt.Printf("%v\n", ncovStatusObj)
	ncovStatusArr, ok := ncovStatusObj.([]interface{})
	if !ok {
		fmt.Printf("解析丁香医生网站数据失败！请检查数据: %s\n", string(contentBytes))
		return ncovStatusResultArr
	}
	for _, provinceStatusObj := range ncovStatusArr {
		provinceStatusMap := provinceStatusObj.(map[string]interface{})
		provinceName := provinceStatusMap["provinceName"].(string)

		cityStatusArr := provinceStatusMap["cities"].([]interface{})
		for _, cityStatusObj := range cityStatusArr {
			cityStatusMap := cityStatusObj.(map[string]interface{})
			currentNcovStatus := new(NCovCityStatus)
			currentNcovStatus.Province = provinceName
			currentNcovStatus.City = cityStatusMap["cityName"].(string)
			currentNcovStatus.Confirming = int64(cityStatusMap["suspectedCount"].(float64))
			currentNcovStatus.Sick = int64(cityStatusMap["confirmedCount"].(float64))
			currentNcovStatus.Cure = int64(cityStatusMap["curedCount"].(float64))
			currentNcovStatus.Death = int64(cityStatusMap["deadCount"].(float64))
			ncovStatusResultArr = append(ncovStatusResultArr, currentNcovStatus)
		}
	}
	return ncovStatusResultArr
}
