package main

import (
		"crypto/tls"
		"encoding/json"
		"errors"
		//"fmt"
		log "github.com/sirupsen/logrus"
		"github.com/prometheus/client_golang/prometheus"
		"github.com/prometheus/client_golang/prometheus/promhttp"
		"io/ioutil"
		"net/http"		
		"time"
)

//	Define the metrics we wish to expose
var mqData = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "sreagent_mq",
		Help: "Queue utilization",
	}, []string{"queue", "metric"} )



var PluginConfig 	map[string]map[string]map[string]interface{}
var PluginData 		map[string]interface{}

var Tr  			*http.Transport
var Cl  			*http.Client

func PluginMeasure() ([]byte, []byte, float64) {
	// Use MQ REST API
	var data 	[]byte
	// response, err := Cl.Get("https://admin:passw0rd@localhost:9443/ibmmq/rest/v1/admin/qmgr/IBMQM1/queue?name=DEV.QUEUE*&status=*")
	response, err := Cl.Get("https://admin:passw0rd@localhost:9443/ibmmq/rest/v1/admin/qmgr/IBMQM1/queue/DEV.QUEUE.1?status=*")
    if err != nil {
    	log.WithFields(log.Fields{"error": err}).Error("Error in REST API")
    } else {
        data, _ = ioutil.ReadAll(response.Body)
        json.Unmarshal(data, &PluginData)
        // log.WithFields(log.Fields{"data": string(data)}).Info("read")
    }
	return data, []byte(""), float64(time.Now().UnixNano())/1e9
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
	// // Check each FS for potential issues with usage
	// FSCHECK:
	// for fsid, mqData := range(PluginData) {
	// 	_, present := PluginConfig["alert"][fsid]["low"].(float64)
	// 	if !present {continue FSCHECK}
	// 	switch {
	// 		case mqData.(*disk.UsageStat).UsedPercent < PluginConfig["alert"][fsid]["low"].(float64):
	// 			alertLvl  = "warn"
	// 			alertMsg  += fmt.Sprintf("FS %s below low design point: %f%% ",fsid, mqData.(*disk.UsageStat).UsedPercent)
	// 			alertFlag = true
	// 			alertErr  = errors.New("low fs")
	// 		case mqData.(*disk.UsageStat).UsedPercent > PluginConfig["alert"][fsid]["engineered"].(float64):
	// 			alertLvl  = "fatal"
	// 			alertMsg  += fmt.Sprintf("FS %s above engineered point: %f%% ",fsid,mqData.(*disk.UsageStat).UsedPercent)
	// 			alertFlag = true
	// 			alertErr  = errors.New("excessive fs use")
	// 			// return now, looks bad
	// 			return alertMsg, alertLvl, alertFlag, alertErr
	// 		case mqData.(*disk.UsageStat).UsedPercent > PluginConfig["alert"][fsid]["design"].(float64):
	// 			alertLvl  = "warn"
	// 			alertMsg  += fmt.Sprintf("FS %s above design point: %f ",fsid, mqData.(*disk.UsageStat).UsedPercent)
	// 			alertFlag = true
	// 			alertErr  = errors.New("moderately high fs")
	// 	}	
	// }
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
	prometheus.MustRegister(mqData)
	Tr  = &http.Transport	{	TLSClientConfig: &tls.Config{ InsecureSkipVerify : true	},
							}
	Cl = &http.Client{Transport: Tr}

	log.WithFields(log.Fields{"pluginconfig": PluginConfig, "plugindata": PluginData}).Info("InitPlugin")
}

func main() {
	config  := 	`
				{
					"alert": 
					{
						"DEV.QUEUE.1":
						{
							"low": 			0,
							"design": 		46.0,
							"engineered":	77.0
						},
						"DEV.QUEUE.2":
						{
							"low": 			0,
							"design": 		10.0,
							"engineered":	175.0
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
