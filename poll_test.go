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
	releases int
	status   string
}

func (t TestRunner) Run(command string, args ...string) string {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cmd := exec.Command(os.Args[0], cs...)
	numberOfReleases := fmt.Sprintf("%s=%d", NUM_MOCKED_RELEASES, t.releases)
	status := t.mockReleaseStatus()
	releaseStatuses := fmt.Sprintf("%s=%s", RELEASE_STATUSES, status)
	cmd.Env = []string{numberOfReleases, releaseStatuses}
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
	status := DEPLOYED_STATUS
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
	status := os.Getenv(RELEASE_STATUSES)
	if os.Getenv(NUM_MOCKED_RELEASES) == "0" {
		fmt.Println(`{"Next":"","Releases":[]}`)
	}
	if os.Getenv(NUM_MOCKED_RELEASES) == "1" {
		fmt.Println(fmt.Sprintf(`{"Next":"","Releases":[{"Name":"fakerelease","Revision":45,"Updated":"Wed Jan 29 08:56:03 2020","Status":"%s","Chart":"codacy-0.5.0-NIGHTLY.29-01-2020","AppVersion":"0.5.0-NIGHTLY.29-01-2020","Namespace":"codacy-nightly"}]}`, status))
	}
	if os.Getenv(NUM_MOCKED_RELEASES) == "2" {
		fmt.Println(fmt.Sprintf(`{"Next":"","Releases":[{"Name":"fakerelease","Revision":45,"Updated":"Wed Jan 29 08:56:03 2020","Status":"%s","Chart":"codacy-0.5.0-NIGHTLY.29-01-2020","AppVersion":"0.5.0-NIGHTLY.29-01-2020","Namespace":"codacy-nightly"},{"Name":"kubernetes-dashboard","Revision":1,"Updated":"Wed Dec 11 16:07:45 2019","Status":"%s","Chart":"kubernetes-dashboard-1.10.1","AppVersion":"1.10.1","Namespace":"kube-system"}]}`, status, status))
	}
}

func TestWhenReleaseExistsGetReleaseReturnsRelease(t *testing.T) {
	runner := TestRunner{1, ""}
	expectedReleaseName := "fakerelease"
	out := GetRelease(runner, expectedReleaseName)
	assert.Equal(t, expectedReleaseName, out.Name)
}

func TestWhenPollingForNonExistingReleaseReturnsEmptyRelease(t *testing.T) {
	runner := TestRunner{0, ""}
	out := PollRelease(runner, "fakerelease", 10, 10)
	assert.Equal(t, out, Release{})
}

func TestIfReleaseAvailableWhenPollingForExistingReleaseReturnsRelease(t *testing.T) {
	runner := TestRunner{1, ""}
	expectedReleaseName := "fakerelease"
	out := PollRelease(runner, expectedReleaseName, 10, 10)
	assert.Equal(t, expectedReleaseName, out.Name)
}

func TestIfReleaseNotAvailableWhenPollingTimesoutForExistingReleaseReturnsEmptyRelease(t *testing.T) {
	runner := TestRunner{1, INSTALLING_STATUS}
	out := PollRelease(runner, "fakerelease", 10, 10)
	assert.Equal(t, Release{}, out)
}

func TestIfReleaseBecomesAvailableWhenPollingReturnsRelease(t *testing.T) {
	for _, n := range STATUSES {
		mockStatusCount = 0
		runner := TestRunner{1, fmt.Sprintf("aRandomNotFinalState;%s", n)}
		out := PollRelease(runner, "fakerelease", 10, 5)
		assert.True(t, out.isAvailableStatus())
	}
}
