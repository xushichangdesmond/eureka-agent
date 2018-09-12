package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xushichangdesmond/eureka-agent"
)

//var eurekaUrl = flag.String("eurekaUrl", "http://localhost:10001/eureka", "eureka url")
var configDir = flag.String("configDir", "config", "configDir")

func main() {

	flag.Parse()

	logger := log.New(os.Stdout, "eureka-agent", log.LstdFlags)

	pl, err := ioutil.ReadFile(filepath.Join(*configDir, "eurekaUrl"))
	if err != nil {
		logger.Fatalln("cannot read eurekaUrl", err)
	}
	eurekaUrls := strings.Split(string(pl), "\n")
	logger.Println("using eurekaUrls", eurekaUrls)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	quit := make(chan os.Signal)
	signal.Notify(quit)

	filepath.Walk(*configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Panicln("Error walking config directory", path, err)
			return nil
		}
		if filepath.Ext(path) == ".registration" {
			wg.Add(1)

			registration := &agent.Registration{}
			logger.Println("Processing", path)
			input, err := ioutil.ReadFile(path)
			if err != nil {
				logger.Fatalln(err)
			}
			if err := json.Unmarshal(input, registration); err != nil {
				logger.Fatalln(err)
			}

			go func() {
				run(logger, eurekaUrl, *registration, quit)
				wg.Done()
			}()
		}
		return nil
	})

}

func run(logger *log.Logger, eurekaUrl string, registration agent.Registration, quit <-chan os.Signal) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	tickerChannel := make(chan time.Time)
	go func() {
		tickerChannel <- time.Now()
	}()
	go func() {
		for t := range ticker.C {
			tickerChannel <- t
		}
	}()

	for {
		select {
		case <-tickerChannel:
			// check connectivity to instance
			logger.Println("Checking health of instance via GET", registration.Instance.HealthCheckUrl)
			_, err := http.Get(registration.Instance.HealthCheckUrl)
			if err != nil {
				logger.Println("error - ", err, " - deleting registration")
				url := eurekaUrl + "/apps/" + registration.Instance.App + "/" + registration.Instance.InstanceId
				req, err := http.NewRequest("DELETE", url, http.NoBody)
				if err != nil {
					logger.Println("invalid delete request ", err)
				}
				resp, err := http.DefaultClient.Do(req)
				logger.Println("Deregister response", resp, ",err", err)
				continue
			}

			url := eurekaUrl + "/apps/" + registration.Instance.App + "/" + registration.Instance.InstanceId
			logger.Println("Sending heartbeat to ", url)
			req, err := http.NewRequest("PUT", url, http.NoBody)
			if err != nil {
				logger.Println("invalid put request ", err)
			}
			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				logger.Println("error sending heartbeat", resp, err)
			}
			if resp.StatusCode == 404 {
				// need to register instance
				url = eurekaUrl + "/apps/" + registration.Instance.App
				logger.Println("Sending registration to ", url)
				registration.Instance.Status = "UP"
				registration.Instance.DataCenterInfo = agent.DataCenterInfo{
					Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
					Name:  "MyOwn",
				}
				body, err := json.Marshal(registration)
				if err != nil {
					logger.Println("Cannot marshal registration request", registration)
				}
				resp, err = http.Post(url, "application/json", bytes.NewReader(body))
				if err != nil {
					logger.Println("error registering instance", resp, err)
				}
				logger.Println("Registration response", resp)
			} else if (resp.StatusCode < 200) || (resp.StatusCode > 299) {
				logger.Println("non 2XX status in heartbeat response", resp)
			}
		case <-quit:
			return
		}
	}
}
