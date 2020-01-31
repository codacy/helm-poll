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

var STATUSES = [5]string{"UNKNOWN", "DEPLOYED", "DELETED", "SUPERSEDED", "FAILED"}

const (
	NUM_MOCKED_RELEASES = "NUM_MOCKED_RELEASES"
	DEPLOYED_STATUS     = "DEPLOYED"
	INSTALLING_STATUS   = "INSTALLING"
	RELEASE_STATUSES    = "RELEASE_STATUSES"
)

type Runner interface {
	Run(string, ...string) string
}

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

type Release struct {
	Name       string `json:Name`
	Revision   int    `json:Revision`
	Updated    string `json:Updated`
	Status     string `json:Status`
	Chart      string `json:Chart`
	AppVersion string `json:AppVersion`
	Namespace  string `json:Namespace`
}

type ReleaseList struct {
	Next     string    `json:Next`
	Releases []Release `json:Releases`
}

func (release Release) isAvailableStatus() bool {
	for _, status := range STATUSES {
		if release.Status == status {
			return true
		}
	}
	return false
}

func GetRelease(runner Runner, releaseName string) Release {
	out := runner.Run("helm", "list", "--output", "json")
	decoder := json.NewDecoder(strings.NewReader(out))
	var decodedJson ReleaseList
	err := decoder.Decode(&decodedJson)
	if err != nil {
		log.Fatal(err)
	}
	for _, release := range decodedJson.Releases {
		if release.Name == releaseName {
			return release
		}
	}
	return Release{}
}

func PollRelease(runner Runner, releaseName string, timeout int, interval int) Release {
	for i := 1; i <= int(timeout/interval); i++ {
		release := GetRelease(runner, releaseName)
		if (release != Release{}) && release.isAvailableStatus() {
			return release
		}
		fmt.Sprintf("%s is %s... waiting...", release.Name, release.Status)
		time.Sleep(time.Duration(interval) * time.Second)
	}
	fmt.Sprintf("%s took to long to become available... exiting...", releaseName)

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
	release := PollRelease(RealRunner{}, *optRelease, *optTimeout, *optInterval)
	marshal, err := json.Marshal(release)
	if err != nil {
		log.Println(err)
	}
	fmt.Print(string(marshal))
}
