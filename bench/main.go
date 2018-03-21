package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/workpool"
	"github.com/logrusorgru/aurora"
)

const (
	instances          = 30
	initialRequests    = instances * 5
	afterScaleRequests = 10000
	maxInFlight        = 500
)

func scaleUp() {
	cmd := exec.Command("cf", "scale", "-i", strconv.Itoa(instances), "test-server")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("cf scale -i %d command failed\n", instances)
		os.Exit(1)
	}

	fmt.Print("waiting for app to scale up")

	defer fmt.Println("\n")

	for {
		fmt.Print(".")
		cmd = exec.Command("bash", "-c", "cf app test-server | grep running | wc -l")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("cf app command failed")
			os.Exit(1)
		}
		i, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err != nil {
			fmt.Printf("cannot parse int: %s\n", string(out))
			os.Exit(1)
		}
		if i == instances {
			return
		}
	}
}

func scaleDown() {
	cmd := exec.Command("cf", "scale", "-i", "5", "test-server")
	err := cmd.Run()
	if err != nil {
		fmt.Println("cf scale command failed")
		os.Exit(1)
	}
}

func stateColor(state string) func(interface{}) aurora.Value {
	color := aurora.Red
	if strings.HasPrefix(state, "404") {
		color = aurora.Brown
	} else if strings.HasPrefix(state, "200") {
		color = aurora.Green
	}
	return color
}

func main() {
	// avoid dns lookup which can slow down the requests
	url := "http://10.244.0.34"

	testStart := time.Now()

	scaleUp()

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1000,
		},
	}

	initialRespChan := make(chan string, initialRequests)
	initialReqDur := make(chan time.Duration, initialRequests)
	for i := 0; i < initialRequests; i++ {
		go func(c chan string) {
			s := time.Now()
			req, err := http.NewRequest("GET", url+"?wait=7s", nil)
			if err != nil {
				panic(err)
			}
			req.Host = "test-server.bosh-lite.com"
			resp, err := client.Do(req)
			initialReqDur <- time.Since(s)
			if err != nil {
				c <- err.Error()
			} else {
				resp.Body.Close()
				c <- resp.Status
			}
		}(initialRespChan)
	}

	go func() {
		for i := 0; i < initialRequests; i++ {
			resp := <-initialRespChan
			if resp != "200 OK" {
				fmt.Printf("got '%s' from one of the initial requests\n", resp)
				os.Exit(1)
			}
		}
		// something that can be spotted in the ocean of dots printed below
		fmt.Print("D")
	}()

	afterScaleRespChan := make(chan string, afterScaleRequests)
	afterScaleReqDurationChan := make(chan time.Duration, afterScaleRequests)
	works := []func(){}
	for i := 0; i < afterScaleRequests; i++ {
		works = append(works, func() {
			s := time.Now()
			req, err := http.NewRequest("GET", url+"?wait=1us", nil)
			if err != nil {
				panic(err)
			}
			req.Host = "test-server.bosh-lite.com"
			resp, err := client.Do(req)
			afterScaleReqDurationChan <- time.Since(s)
			if err != nil {
				afterScaleRespChan <- err.Error()
			} else {
				resp.Body.Close()
				afterScaleRespChan <- resp.Status
			}
		})
	}

	throttler, err := workpool.NewThrottler(maxInFlight, works)
	if err != nil {
		panic(err)
	}

	// make sure all initial requests have been routed on the gorouter
	time.Sleep(time.Second)

	go throttler.Work()

	// allow a few requests to be served
	time.Sleep(time.Second)
	scaleDown()

	results := map[string]int{}
	var minDur, maxDur time.Duration
	for i := 0; i < afterScaleRequests; i++ {
		state := <-afterScaleRespChan
		color := stateColor(state)
		fmt.Print(color("."))
		results[state]++
		dur := <-afterScaleReqDurationChan
		if dur < minDur || minDur == 0 {
			minDur = dur
		}
		if dur > maxDur || maxDur == 0 {
			maxDur = dur
		}
	}
	fmt.Println("")

	var minInitialDur, maxInitialDur time.Duration
	for i := 0; i < initialRequests; i++ {
		dur := <-initialReqDur
		if dur < minInitialDur || minInitialDur == 0 {
			minInitialDur = dur
		}
		if dur > maxInitialDur || maxInitialDur == 0 {
			maxInitialDur = dur
		}
	}

	for k, v := range results {
		fmt.Printf("Count: %d\t\tResp: '%s'\n", v, k)
	}
	fmt.Println("")
	fmt.Printf("initial requests dur\t\tmax: %s\tmin: %s\n", maxInitialDur, minInitialDur)
	fmt.Printf("after scale requests dur\tmax: %s\tmin: %s\n", maxDur, minDur)
	fmt.Println("")
	fmt.Printf("started:\t%s\nended:\t%s\nduration:\t%s\n", testStart, time.Now(), time.Since(testStart))
}
