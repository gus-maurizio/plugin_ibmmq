package main

import (
	"fmt"
	"testing"
	"time"
)

var config string = `
				{
					"alert": 
					{
						"/":
						{
							"low": 			2,
							"design": 		60.0,
							"engineered":	80.0
						},
						"/Volumes/TOSHIBA-001":
						{
							"low": 			22,
							"design": 		40.0,
							"engineered":	75.0
						}
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
