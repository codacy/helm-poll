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

const NUM_MOCKED_RELEASES = "NUM_MOCKED_RELEASES"
const DEPLOYED_STATUS = "DEPLOYED"
const INSTALLING_STATUS = "INSTALLING"
const RELEASE_STATUSES = "RELEASE_STATUSES"

type Runner interface {
	Run(string, ...string) (string, error)
}

type RealRunner struct{}

var runner Runner

func (r RealRunner) Run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String(), err
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

func GetRelease(releaseName string) Release {
	out, err := runner.Run("helm", "list", "--output", "json")
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(strings.NewReader(out))
	var json ReleaseList
	err = decoder.Decode(&json)
	if err != nil {
		log.Fatal(err)
	}
	for _, release := range json.Releases {
		if release.Name == releaseName {
			return release
		}
	}
	return Release{}
}

func PollRelease(releaseName string, timeout int, interval int) Release {
	release := GetRelease(releaseName)
	if (release != Release{}) {
		for i := 1; i <= int(timeout/interval); i++ {
			if release.isAvailableStatus() {
				return release
			} else {
				fmt.Sprintf("%s is %s... waiting...", release.Name, release.Status)
				time.Sleep(time.Duration(interval) * time.Second)
				release = GetRelease(releaseName)
			}
		}
		fmt.Sprintf("%s took to long to become available... exiting...", release.Name)
	}
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
	runner = RealRunner{}
	optRelease, optTimeout, optInterval := parseArgs()
	release := PollRelease(*optRelease, *optTimeout, *optInterval)
	marshal, err := json.Marshal(release)
	if err != nil {
		log.Println(err)
	}
	fmt.Print(string(marshal))
}
