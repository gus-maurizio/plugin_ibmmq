package main

import (
		"encoding/json"
		"errors"
		"fmt"
		"github.com/shirou/gopsutil/disk"
		log "github.com/sirupsen/logrus"
		"github.com/prometheus/client_golang/prometheus"
		"github.com/prometheus/client_golang/prometheus/promhttp"
		"net/http"		
		"time"
)

//	Define the metrics we wish to expose
var fsData = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "sreagent_fs",
		Help: "Host Filesystem utilization",
	}, []string{"fsname", "metric"} )



var PluginConfig 	map[string]map[string]map[string]interface{}
var PluginData 		map[string]interface{}



func PluginMeasure() ([]byte, []byte, float64) {
	// Get measurement of IOCounters
	osfs, _ := disk.Partitions(false)
	
	for _, fs := range osfs {
		// Update metrics related to the plugin
		mpoint 		:= fs.Mountpoint
		fsdata, _ 	:= disk.Usage(mpoint)
		PluginData[mpoint]  = fsdata
		fsData.With(prometheus.Labels{"fsname":  mpoint, "metric": "usedpercent"}).Set(fsdata.UsedPercent)
		fsData.With(prometheus.Labels{"fsname":  mpoint, "metric": "freespaceGB"}).Set(float64(fsdata.Free)/(1024.0*1024.0*1024.0))
	}
	myMeasure, _ := json.Marshal(PluginData)
	return myMeasure, []byte(""), float64(time.Now().UnixNano())/1e9
}

func PluginAlert(measure []byte) (string, string, bool, error) {
	// log.WithFields(log.Fields{"MyMeasure": string(MyMeasure[:]), "measure": string(measure[:])}).Info("PluginAlert")
	// var m 			interface{}
	// err := json.Unmarshal(measure, &m)
	// if err != nil { return "unknown", "", true, err }
	alertMsg := ""
	alertLvl := ""
	alertFlag := false
	alertErr := errors.New("no error")
	// Check each FS for potential issues with usage
	FSCHECK:
	for fsid, fsdata := range(PluginData) {
		_, present := PluginConfig["alert"][fsid]["low"].(float64)
		if !present {continue FSCHECK}
		switch {
			case fsdata.(*disk.UsageStat).UsedPercent < PluginConfig["alert"][fsid]["low"].(float64):
				alertLvl  = "warn"
				alertMsg  += fmt.Sprintf("FS %s below low design point: %f%% ",fsid, fsdata.(*disk.UsageStat).UsedPercent)
				alertFlag = true
				alertErr  = errors.New("low fs")
			case fsdata.(*disk.UsageStat).UsedPercent > PluginConfig["alert"][fsid]["engineered"].(float64):
				alertLvl  = "fatal"
				alertMsg  += fmt.Sprintf("FS %s above engineered point: %f%% ",fsid,fsdata.(*disk.UsageStat).UsedPercent)
				alertFlag = true
				alertErr  = errors.New("excessive fs use")
				// return now, looks bad
				return alertMsg, alertLvl, alertFlag, alertErr
			case fsdata.(*disk.UsageStat).UsedPercent > PluginConfig["alert"][fsid]["design"].(float64):
				alertLvl  = "warn"
				alertMsg  += fmt.Sprintf("FS %s above design point: %f ",fsid, fsdata.(*disk.UsageStat).UsedPercent)
				alertFlag = true
				alertErr  = errors.New("moderately high fs")
		}	
	}
	return alertMsg, alertLvl, alertFlag, alertErr
}

func InitPlugin(config string) {
	if PluginData == nil {
		PluginData = make(map[string]interface{}, 20)
	}
	if PluginConfig == nil {
		PluginConfig = make(map[string]map[string]map[string]interface{}, 20)
	}
	err := json.Unmarshal([]byte(config), &PluginConfig)
	if err != nil {
		log.WithFields(log.Fields{"config": config}).Error("failed to unmarshal config")
	}
	// Register metrics with prometheus
	prometheus.MustRegister(fsData)

	log.WithFields(log.Fields{"pluginconfig": PluginConfig, "plugindata": PluginData}).Info("InitPlugin")
}

func main() {
	config  := 	`
				{
					"alert": 
					{
						"/":
						{
							"low": 			2,
							"design": 		46.0,
							"engineered":	77.0
						},
						"/Volumes/TOSHIBA-001":
						{
							"low": 			22,
							"design": 		40.0,
							"engineered":	75.0
						}
				    }
				}
				`

	//--------------------------------------------------------------------------//
	// time to start a prometheus metrics server
	// and export any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":8999", nil)
	}()
	//--------------------------------------------------------------------------//

	InitPlugin(config)
	log.WithFields(log.Fields{"PluginConfig": PluginConfig}).Info("InitPlugin")
	tickd := 3 * time.Second
	for i := 1; i <= 3; i++ {
		measure, measureraw, measuretimestamp := PluginMeasure()
		alertmsg, alertlvl, isAlert, err := PluginAlert(measure)
		//fmt.Printf("Iteration #%d tick %d %v \n", i, tick,PluginData)
		log.WithFields(log.Fields{"timestamp": measuretimestamp,
			"measure":    string(measure[:]),
			"measureraw": string(measureraw[:]),
			"PluginData": PluginData,
			"alertMsg":   alertmsg,
			"alertLvl":   alertlvl,
			"isAlert":    isAlert,
			"AlertErr":   err,
		}).Info("Tick")
		time.Sleep(tickd)
	}
}
