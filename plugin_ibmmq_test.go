package main

import (
	"fmt"
	"testing"
	"time"
)

var config string = `
				{
					"config": 
					{
						"mq":
						{
							"protocol":		"https",
							"authoriz":		"YWRtaW46cGFzc3cwcmQ=",
							"url":			"localhost:9443/ibmmq/rest/v1/admin/qmgr/IBMQM1/queue?name=DEV.QUEUE*&status=*"		
						}
				    },
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

func TestInit(t *testing.T) {
	fmt.Printf("Start TestInit\n")
	InitPlugin(config)
	fmt.Printf("Done TestInit %v\n",PluginConfig)
}

func TestMeasure(t *testing.T) {
	fmt.Printf("Start TestMeasure\n")

	for i := 1; i <= 5; i++ {
		m, mraw, ts := PluginMeasure()
		fmt.Printf("measure %v %v %v\n",string(m),string(mraw),ts)
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Printf("Done TestMeasure %v\n",PluginData)
}

func BenchmarkMeasureOnly(b *testing.B) {
    for i := 0; i < b.N; i++ {
		m, mraw, ts := PluginMeasure()
		if i % 15000 == 1 {fmt.Printf("measure %d %v %v %v\n",i,string(m),string(mraw),ts)}
		//time.Sleep(50 * time.Millisecond)
    }
}

func BenchmarkMeasureAlert(b *testing.B) {
    for i := 0; i < b.N; i++ {
		m, mraw, ts := PluginMeasure()
		alertmsg, alertlvl, isAlert, err := PluginAlert(m)
		if i % 15000 == 1 {fmt.Printf("alert %d %v %v %v\n%v %v %v %v\n\n",i,string(m),string(mraw),ts, alertmsg, alertlvl, isAlert, err)}
		//time.Sleep(50 * time.Millisecond)
    }
}
