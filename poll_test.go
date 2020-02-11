package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type TestRunner struct {
	releaseName string
	status   string
}

func (t TestRunner) Run(command string, args ...string) string {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cmd := exec.Command(os.Args[0], cs...)
	status := t.mockReleaseStatus()
	releasename := fmt.Sprintf("%s=%s", mockReleaseName, t.releaseName)
	releaseStatuses := fmt.Sprintf("%s=%s", releaseStates, status)
	cmd.Env = []string{releasename, releaseStatuses}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String()
}

var mockStatusCount = 0

func (t TestRunner) mockReleaseStatus() string {
	status := deployedState
	if t.status != "" {
		statuses := strings.Split(t.status, ";")
		if mockStatusCount > len(statuses)-1 {
			status = statuses[len(statuses)-1]
		} else {
			status = statuses[mockStatusCount]
			mockStatusCount++
		}
	}
	return status
}

func TestHelperProcess(t *testing.T) {
	status := os.Getenv(releaseStates)
	if os.Getenv(mockReleaseName) == "fakerelease" {
		fmt.Println(fmt.Sprintf(`[{"revision":46,"updated":"Tue Feb 11 12:14:53 2020","status":"%s","chart":"codacy-0.1.1","appVersion":"1.0","description":"Preparing upgrade"}]`, status))
	} else {
		fmt.Println(fmt.Sprintf(`[{"revision":0,"updated":"","status":"","chart":"","appVersion":"","description":""}]`))
	}


}

func TestWhenReleaseExistsGetReleaseReturnsRelease(t *testing.T) {
	expectedReleaseName := "fakerelease"
	runner := TestRunner{expectedReleaseName,""}
	out := getRelease(runner, expectedReleaseName)
	assert.NotEqual(t, out, Release{})
}

func TestWhenPollingForNonExistingReleaseReturnsEmptyRelease(t *testing.T) {
	expectedReleaseName := "nonexistingfakerelease"
	runner := TestRunner{expectedReleaseName,""}
	out := pollRelease(runner, expectedReleaseName, 10, 10)
	assert.Equal(t, out, Release{})
}

func TestIfReleaseAvailableWhenPollingForExistingReleaseReturnsRelease(t *testing.T) {
	expectedReleaseName := "fakerelease"
	runner := TestRunner{expectedReleaseName,""}
	out := pollRelease(runner, expectedReleaseName, 10, 10)
	assert.NotEqual(t, out, Release{})
}

func TestIfReleaseNotAvailableWhenPollingTimesoutForExistingReleaseReturnsEmptyRelease(t *testing.T) {
	runner := TestRunner{ "fakerelease", installingState}
	out := pollRelease(runner, "fakerelease", 10, 10)
	assert.Equal(t, Release{}, out)
}

func TestIfReleaseBecomesAvailableWhenPollingReturnsRelease(t *testing.T) {
	for _, n := range statuses {
		mockStatusCount = 0
		runner := TestRunner{ "fakerelease", fmt.Sprintf("aRandomNotFinalState;%s", n)}
		out := pollRelease(runner, "fakerelease", 10, 5)
		assert.True(t, out.isAvailableStatus())
	}
}
