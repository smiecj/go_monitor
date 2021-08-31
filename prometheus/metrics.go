package main

import (
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/smiecj/go_monitor/reptile"
)

func recordMetrics() {
	// 设置数据标题、标签和说明信息
	machineGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "produce_machine_status",
		Help: "当前生产环境机器的性能（模拟）"}, []string{"type"})
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for range ticker.C {
			machineGauge.With(prometheus.Labels{"type": "cpu"}).Set(rand.Float64() * 100)
			machineGauge.With(prometheus.Labels{"type": "memory"}).Set(rand.Float64() * 100)
		}
	}()

	// 记录微博热搜数据
	weiboHotDataGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "weibo_hotdata",
		Help: "微博热搜数据实时刷新"}, []string{"topic"})
	weiboTicker := time.NewTicker(15 * time.Second)
	go func() {
		for range weiboTicker.C {
			hotDataMap := reptile.GetHotTopicAndClickTime()
			for topic, hotNum := range hotDataMap {
				weiboHotDataGauge.With(prometheus.Labels{"topic": topic}).Set(float64(hotNum))
			}
		}
	}()

	// 记录新型肺炎感染情况
	ncovStatusGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "ncov_status",
		Help: "新型肺炎感染情况"}, []string{"province", "city", "type"})
	ncovTicker := time.NewTicker(time.Minute)
	go func() {
		for range ncovTicker.C {
			ncovStatusArr := reptile.GetNcovStatus()
			for _, ncovStatus := range ncovStatusArr {
				ncovStatusGauge.With(prometheus.Labels{"province": ncovStatus.Province,
					"city": ncovStatus.City, "type": "患病"}).Set(float64(ncovStatus.Sick))
				ncovStatusGauge.With(prometheus.Labels{"province": ncovStatus.Province,
					"city": ncovStatus.City, "type": "治愈"}).Set(float64(ncovStatus.Cure))
				ncovStatusGauge.With(prometheus.Labels{"province": ncovStatus.Province,
					"city": ncovStatus.City, "type": "待确认"}).Set(float64(ncovStatus.Confirming))
				ncovStatusGauge.With(prometheus.Labels{"province": ncovStatus.Province,
					"city": ncovStatus.City, "type": "死亡"}).Set(float64(ncovStatus.Death))
			}
		}
	}()
}

// counter 服务如果异常挂死，这个数据应该由用户来恢复，而不是prometheus 服务端
// prometheus 不关心用户上传的数据是否准确
func recordCount() {
	apiCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "produce_api_status",
		Help: "生产环境的接口调用情况（模拟）",
	}, []string{"url", "ret"})

	ticker := time.NewTicker(15 * time.Second)

	go func() {
		for range ticker.C {
			// 查询接口调用比较多，每次新增10-20次
			apiCounter.With(prometheus.Labels{"url": "/api/search", "ret": "200"}).Add(10 + math.Floor(10*(rand.Float64())))
			// 新增接口调用较少，每次新增0-10次
			apiCounter.With(prometheus.Labels{"url": "/api/add", "ret": "200"}).Add(math.Floor(10 * (rand.Float64())))
			// 查询接口偶尔因为网络问题会报错，每次新增0-5次
			apiCounter.With(prometheus.Labels{"url": "/api/search", "ret": "500"}).Add(math.Floor(5 * (rand.Float64())))
		}
	}()
}

// 用于记录平均数
func recordSummary() {
	apiSummary := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "produce_api_summary",
		Help: "生产环境的接口耗时汇总（模拟）",
		// Objectives 和众数的统计相关，两个参数都是百分比
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"url", "ret"})

	secondTicker := time.NewTicker(15 * time.Second)
	minuteTicker := time.NewTicker(time.Minute)

	go func() {
		for range secondTicker.C {
			// 接口的正常耗时：设置为1-10s范围内
			apiSummary.With(prometheus.Labels{"url": "/api/search", "ret": "200"}).Observe(math.Ceil(10 * (rand.Float64())))
			time.Sleep(15 * time.Second)
		}
	}()
	go func() {
		for range minuteTicker.C {
			// 异常接口耗时：60~120s
			apiSummary.With(prometheus.Labels{"url": "/api/search", "ret": "200"}).Observe(60 + math.Ceil(60*rand.Float64()))
		}
	}()
}

// 直方图
func recordHistogram() {
	apiHistogram := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "produce_api_histogram",
		Help:    "生产环境的接口耗时分位统计（模拟）",
		Buckets: []float64{0.01, 0.1, 1, 2, 5, 10, 30},
	}, []string{"l_url", "l_ret"})

	secondTicker := time.NewTicker(15 * time.Second)
	minuteTicker := time.NewTicker(time.Minute)

	go func() {
		for range secondTicker.C {
			// 接口的正常耗时：设置为1-10s范围内
			apiHistogram.With(prometheus.Labels{"l_url": "/api/search", "l_ret": "200"}).Observe(math.Ceil(10 * (rand.Float64())))
		}
	}()
	go func() {
		for range minuteTicker.C {
			// 异常接口耗时：60~120s
			apiHistogram.With(prometheus.Labels{"l_url": "/api/search", "l_ret": "200"}).Observe(60 + math.Ceil(60*rand.Float64()))
		}
	}()
}

func main() {
	recordMetrics()
	recordCount()
	recordSummary()
	recordHistogram()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
