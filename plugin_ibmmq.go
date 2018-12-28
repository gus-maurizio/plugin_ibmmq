package main

import (
		"crypto/tls"
		"encoding/json"
		"errors"
		"fmt"
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
var PluginData 		map[string][]QueueData

var Tr  			*http.Transport
var Cl  			*http.Client

type QueueData 		struct {
	Name			string			`json:"name"`
	Status			struct	{
		CurrDepth	float64			`json:"currentDepth"`
		LastGet		string			`json:"lastGet"`
		LastPut		string			`json:"lastPut"`
		MRLog		string			`json:"mediaRecoveryLogExtent"`
		MRate 		string			`json:"monitoringRate"`
		OMAge		float64			`json:"oldestMessageAge"`
		OnQueueTime struct {
			LSample	float64			`json:"longSamplePeriod"`
			SSample	float64			`json:"shortSamplePeriod"`
		}							`json:"onQueueTime"`
		OpenInput	float64 		`json:"openInputCount"`
		OpenOutput	float64 		`json:"openOutputCount"`
		Uncommitted	float64 		`json:"uncommittedMessages"`
	}								`json:"status"`
	QType			string 			`json:"type"`
}

func PluginMeasure() ([]byte, []byte, float64) {
	// Use MQ REST API
	var data 	[]byte
	response, err := Cl.Get("https://admin:passw0rd@localhost:9443/ibmmq/rest/v1/admin/qmgr/IBMQM1/queue?name=DEV.QUEUE*&status=*")
	//response, err := Cl.Get("https://admin:passw0rd@localhost:9443/ibmmq/rest/v1/admin/qmgr/IBMQM1/queue/DEV.QUEUE.1?status=*")
    if err != nil {
    	log.WithFields(log.Fields{"error": err}).Error("Error in REST API")
    	data 				= []byte("{}")
    	PluginData["queue"] = []QueueData{} 
    } else {
        data, _ = ioutil.ReadAll(response.Body)
        json.Unmarshal(data, &PluginData)
        // log.WithFields(log.Fields{"data": string(data)}).Info("read")
    }
	for _, queueData := range(PluginData["queue"]) {
		log.WithFields(log.Fields{"queuedata": queueData}).Debug("queue")
		mqData.With(prometheus.Labels{"queue":  queueData.Name, "metric": "queuedepth"}).Set(queueData.Status.CurrDepth)
		mqData.With(prometheus.Labels{"queue":  queueData.Name, "metric": "oldestMessageAgeSec"}).Set(queueData.Status.OMAge)
		mqData.With(prometheus.Labels{"queue":  queueData.Name, "metric": "onQueueTimeLongMSec"}).Set(queueData.Status.OnQueueTime.LSample/1e3)
		mqData.With(prometheus.Labels{"queue":  queueData.Name, "metric": "OnQueueTimeShortMSec"}).Set(queueData.Status.OnQueueTime.SSample/1e3)
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
	if len(PluginData["queue"]) == 0 {
		alertLvl  = "fatal"
		alertMsg  += fmt.Sprintf("No DATA from MQ")
		alertFlag = true
		alertErr  = errors.New("NO DATA from MQ")
	}
	// Check each FS for potential issues with usage
	QUEUECHECK:
	for _, queueData := range(PluginData["queue"]) {
		log.WithFields(log.Fields{"queuedata": queueData}).Debug("queue") 
		_, present := PluginConfig["alert"][queueData.Name]["low"].(float64)
		if !present {continue QUEUECHECK}
		log.WithFields(log.Fields{"queuedata": queueData}).Info("queue found in alert") 
		switch {
			case queueData.Status.CurrDepth < PluginConfig["alert"][queueData.Name]["low"].(float64):
				alertLvl  = "warn"
				alertMsg  += fmt.Sprintf("Queue %s below low design point: %f%% ",queueData.Name, queueData.Status.CurrDepth)
				alertFlag = true
				alertErr  = errors.New("low Queue")
			case queueData.Status.CurrDepth > PluginConfig["alert"][queueData.Name]["engineered"].(float64):
				alertLvl  = "fatal"
				alertMsg  += fmt.Sprintf("Queue %s above engineered point: %f%% ",queueData.Name,queueData.Status.CurrDepth)
				alertFlag = true
				alertErr  = errors.New("excessive Queue use")
				// return now, looks bad
				return alertMsg, alertLvl, alertFlag, alertErr
			case queueData.Status.CurrDepth > PluginConfig["alert"][queueData.Name]["design"].(float64):
				alertLvl  = "warn"
				alertMsg  += fmt.Sprintf("Queue %s above design point: %f ",queueData.Name, queueData.Status.CurrDepth)
				alertFlag = true
				alertErr  = errors.New("moderately high queue")
		}	
	}
	return alertMsg, alertLvl, alertFlag, alertErr
}

func InitPlugin(config string) {
	if PluginData == nil {
		PluginData = make(map[string][]QueueData, 2)
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
	log.WithFields(log.Fields{"PluginConfig": PluginConfig}).Debug("InitPlugin")
	tickd := 1000 * time.Millisecond
	for i := 1; i <= 10; i++ {
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
