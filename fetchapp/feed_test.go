package fetchapp_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/spotlightpa/inky-fetch/fetchapp"
)

func assertErr(t *testing.T, msg string, err error) {
	t.Helper()
	if err != nil {
		t.Errorf(msg+" %v", err)
	}
}

func assertStrings(t *testing.T, ss ...string) {
	t.Helper()
	s0 := ss[0]
	for _, s := range ss[1:] {
		if s != s0 {
			t.Errorf("%s != %s", s0, s)
		}
	}
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/" + name)
	if err != nil {
		assertErr(t, "could not open test data", err)
	}
	return string(data)
}

func TestGetSpotlightLinks(t *testing.T) {
	var testcases = map[string]struct {
		src   string
		links string
	}{
		"case": {
			src:   "arc.xml",
			links: "https://www.inquirer.com/news/pennsylvania/spl/nurse-licensing-board-pennsylvania-delays-complaints-20190911.html",
		},
	}

	for name, test := range testcases {
		t.Run(name, func(t *testing.T) {
			data := readFile(t, test.src)
			u, err := fetchapp.GetSpotlightLinks(strings.NewReader(data))
			assertErr(t, "could not process RSS", err)
			for i, s := range strings.Split(test.links, ",") {
				assertStrings(t, u[i].String(), s)
			}
		})
	}
}
