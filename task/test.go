package task

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type testRunner struct {
	Runner
}

type pkg struct {
	Name          string  `json:"name,omitempty"`
	ExecutionTime float64 `json:"execution_time,omitempty"`
	Success       bool    `json:"success"`
	Tests         []test  `json:"tests"`
}

type test struct {
	Name          string  `json:"name,omitempty"`
	ExecutionTime float64 `json:"execution_time"`
	Passed        bool    `json:"passed"`
}

// TestRunner is a runner used for testing Go projects
func TestRunner() Runnable {
	return &testRunner{Runner{Label: testName, Parallel: true}}
}

// Execute ...
func (r *testRunner) Execute() {
	defer r.trackTime(time.Now())

	out, err := exec.Command("go", "test", "-v", "./...").CombinedOutput()
	if err != nil {
		r.Error = &RunnerError{
			RawOutput: string(out),
			Message:   err,
		}
		return
	}

	r.RawOutput = string(out)
	r.parseTestOutput()
}

func (r *testRunner) parseTestOutput() {
	pkgs, p, tests := []pkg{}, pkg{}, []test{}

	pkgReregex, _ := regexp.Compile(`(?i)^(ok|fail|\?)\s+([\w\.\/\-]+)(?:\s+([\d\.]+)s)?(?:\s+coverage:\s(\d+(\.\d+)?))?`)
	testRegex, _ := regexp.Compile(`(?i)(FAIL|PASS):\s([\w\d]+)\s\(([\d\.]+)s\)`)

	lines := strings.Split(r.RawOutput, "\n")
	for _, l := range lines {
		submatch := testRegex.FindStringSubmatch(l)
		if len(submatch) != 0 {
			ef, _ := strconv.ParseFloat(submatch[3], 64)
			tests = append(tests, test{
				submatch[2],
				ef,
				submatch[1] == "PASS",
			})
			continue
		}

		submatch = pkgReregex.FindStringSubmatch(l)
		if len(submatch) != 0 {
			switch submatch[1] {
			case "FAIL":
				ef, _ := strconv.ParseFloat(submatch[3], 64)
				p = pkg{
					Name:          submatch[2],
					Success:       false,
					Tests:         tests,
					ExecutionTime: ef,
				}
			case "ok":
				ef, _ := strconv.ParseFloat(submatch[3], 64)
				p = pkg{
					Name:          submatch[2],
					Success:       true,
					Tests:         tests,
					ExecutionTime: ef,
				}
			default:
				continue
			}

			pkgs = append(pkgs, p)
			tests = []test{}
		}
	}

	r.Data = pkgs
}