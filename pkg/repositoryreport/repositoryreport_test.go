package repositoryreport

import (
	"github.com/securityclippy/imagemanager/pkg/config"
	"testing"
	"time"
)

var testRR = &RepositoryReport{
	Registry: "testorg",
	Repository: "golang",
}

var testCfg = &config.Config{
		DeprecationThresholdDays: 180,
		DeprecationWarningDays: 14,
		DeletionThresholdDays: 200,
		DeletionWarningDays: 7,
}

func TestRepositoryReport_DaysToDeprecationDate(t *testing.T) {
	var tests = map[string]struct{
		Input time.Time
		Want int
		Got int
	}{
		"180 since last update": {
			Input: time.Now(),
			Want: 0,
		},
		"180 days to deprecation": {
			Input: time.Now().Add(time.Duration(time.Hour * 24 * 180)),
			Want: 179,

		},
		"20 days to deprecation": {
			Input: time.Now().Add(time.Duration(time.Hour * 24 * 21)),
			Want: 20,

		},
		"negative 180 days to deprecation": {
			Input: time.Now().Add(-time.Duration(time.Hour * 24 * 180)),
			Want: -180,

		},
		"not set": {
			Input:time.Time{},
			Want: 365,
		},

	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testRR.DeprecationDate = test.Input
			test.Got = testRR.DaysToDeprecationDate()

			if test.Got != test.Want {
				t.Errorf("Want: %d, got: %d", test.Want, test.Got)
				t.Log(testRR.JsonString())
			}
		})

	}

}

func TestRepositoryReport_DaysSinceDeprecationMark(t *testing.T) {

	var tests = map[string]struct{
		Input time.Time
		Want int
		Got int
	}{
		"dep mark not set": {
			Input: time.Time{},
			Want: 0,
		},
		"2 days since deprecation mark": {
			Input: time.Now().Add(-time.Duration(time.Hour * 24 * 2)),
			Want: 2,

		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testRR.DeprecationMarkedOn = test.Input
			test.Got = testRR.DaysSinceDeprecationMark()

			if test.Got != test.Want {
				t.Errorf("Want: %d, got: %d", test.Want, test.Got)
				t.Log(testRR.JsonString())
			}
		})

	}
}

func TestRepositoryReport_DaysToDeletionDate(t *testing.T) {

	var tests = map[string]struct{
		Input time.Time
		Want int
		Got int
	}{
		"180 since last update": {
			Input: time.Now(),
			Want: 0,
		},
		"180 days to deprecation": {
			Input: time.Now().Add(time.Duration(time.Hour * 24 * 180)),
			Want: 179,

		},
		"20 days to deprecation": {
			Input: time.Now().Add(time.Duration(time.Hour * 24 * 21)),
			Want: 20,

		},
		"negative 180 days to deprecation": {
			Input: time.Now().Add(-time.Duration(time.Hour * 24 * 180)),
			Want: -180,

		},
		"not set": {
			Input:time.Time{},
			Want: 365,
		},

	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testRR.DeletionDate = test.Input
			test.Got = testRR.DaysToDeletionDate()

			if test.Got != test.Want {
				t.Errorf("Want: %d, got: %d", test.Want, test.Got)
				t.Log(testRR.JsonString())
			}
		})

	}
}