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
	mockNamespace = "mockNamespace"
	deployedState   = "deployed"
	installingState = "installing"
	releaseStates   = "releaseStates"
)

// Runner is an interface to wrap around a exec.Command.Run
type Runner interface {
	Run(string, ...string) string
}

// RealRunner is a concrete implementation of Runner
type RealRunner struct{
	isDebugEnabled bool
}

func (r RealRunner) LogDebug(message string) {
	if r.isDebugEnabled {
		fmt.Println(message)
	}
}

func (r RealRunner) Run(command string, args ...string) string {
	arguments := args
	if r.isDebugEnabled {
		r.LogDebug(fmt.Sprintf("enabling helm with --debug"))
		arguments = append(args, "--debug")
	}
	r.LogDebug(fmt.Sprintf("%s %s", command, arguments))
	cmd := exec.Command(command, arguments...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Release not found in namespace.")
		out.WriteString(fmt.Sprintf(`[{"revision":0,"updated":"","status":"","chart":"","appVersion":"","description":""}]`))
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
		if strings.ToLower(release.Status) == status {
			return true
		}
	}
	return false
}

func getRelease(runner Runner, releaseName string, namespace string) Release {
	out := runner.Run("helm", "history", releaseName, "-n", namespace, "--max=1", "--output", "json")
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

func pollRelease(runner Runner, releaseName string, namespace string, timeout int, interval int) Release {
	for i := 1; i <= int(timeout/interval); i++ {
		release := getRelease(runner, releaseName, namespace)
		if (release != Release{}) && release.isAvailableStatus() {
			return release
		}
		fmt.Println(fmt.Sprintf("%s is %s... waiting...", releaseName, strings.ToLower(release.Status)))
		time.Sleep(time.Duration(interval) * time.Second)
	}
	fmt.Println(fmt.Sprintf("%s took to long to become available... exiting...", releaseName))

	return Release{}
}

func parseArgs() (*string, *string, *int, *int, *bool) {
	optRelease := getopt.StringLong("release", 'r', "", "Release name to poll for.")
	optNamespace := getopt.StringLong("namespace", 'n', "default", "Namespace where the release is installed. (default: \"default\")")
	optTimeout := getopt.IntLong("timeout", 't', 300, "The timeout in seconds (default: 300)")
	optInterval := getopt.IntLong("interval", 'i', 5, "The polling interval in seconds (default: 5)")
	optDebug := getopt.BoolLong("debug", 0, "Run with debug messages on")
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

	return optRelease, optNamespace, optTimeout, optInterval, optDebug
}

func main() {
	optRelease, optNamespace, optTimeout, optInterval, optDebug := parseArgs()
	runner := RealRunner{}
	runner.isDebugEnabled = *optDebug
	release := pollRelease(runner, *optRelease, *optNamespace, *optTimeout, *optInterval)
	marshal, err := json.Marshal(release)
	if err != nil {
		log.Println(err)
	}
	fmt.Print(string(marshal))
}
