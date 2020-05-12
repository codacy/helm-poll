package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pborman/getopt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var statuses = [5]string{"unknown", "deployed", "deleted", "superseded", "failed"}

const (
	mockReleaseName = "mockReleaseName"
	deployedState   = "DEPLOYED"
	installingState = "INSTALLING"
	releaseStates   = "releaseStates"
)

// Runner is an interface to wrap around a exec.Command.Run
type Runner interface {
	Run(string, ...string) string
}

// RealRunner is a concrete implementation of Runner
type RealRunner struct{}

func (r RealRunner) Run(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String()
}

// Release is an Helm release
type Release struct {
	Revision   int    `json:revision`
	Updated    string `json:updated`
	Status     string `json:status`
	Chart      string `json:chart`
	AppVersion string `json:appVersion`
	Description  string `json:description`
}

// ReleaseList is a collection of Helm releases
type ReleaseList []Release

func (release Release) isAvailableStatus() bool {
	for _, status := range statuses {
		if release.Status == status {
			return true
		}
	}
	return false
}

func getRelease(runner Runner, releaseName string) Release {
	out := runner.Run("helm", "history", releaseName, "--max=1", "--output", "json")
	decoder := json.NewDecoder(strings.NewReader(out))
	var decodedJSON ReleaseList
	err := decoder.Decode(&decodedJSON)
	if err != nil {
		log.Fatal(err)
	}

	if decodedJSON[0].Revision == 0 {
		return Release{}
	}

	return decodedJSON[0]
}

func pollRelease(runner Runner, releaseName string, timeout int, interval int) Release {
	for i := 1; i <= int(timeout/interval); i++ {
		release := getRelease(runner, releaseName)
		if (release != Release{}) && release.isAvailableStatus() {
			return release
		}
		fmt.Println(fmt.Sprintf("%s is %s... waiting...", releaseName, release.Status))
		time.Sleep(time.Duration(interval) * time.Second)
	}
	fmt.Println(fmt.Sprintf("%s took to long to become available... exiting...", releaseName))

	return Release{}
}

func parseArgs() (*string, *int, *int) {
	optRelease := getopt.StringLong("release", 'r', "", "Release name to poll for.")
	optTimeout := getopt.IntLong("timeout", 't', 300, "The timeout in seconds (default: 300)")
	optInterval := getopt.IntLong("interval", 'i', 5, "The polling interval in seconds (default: 5)")
	optHelp := getopt.BoolLong("help", 0, "Help")

	getopt.Parse()
	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}

	if *optRelease == "" {
		fmt.Println("You must specify a release name to poll for!")
		getopt.Usage()
		os.Exit(0)
	}
	return optRelease, optTimeout, optInterval
}

func main() {
	optRelease, optTimeout, optInterval := parseArgs()
	release := pollRelease(RealRunner{}, *optRelease, *optTimeout, *optInterval)
	marshal, err := json.Marshal(release)
	if err != nil {
		log.Println(err)
	}
	fmt.Print(string(marshal))
}
