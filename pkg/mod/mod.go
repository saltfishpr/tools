package mod

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"

	"github.com/charmbracelet/log"

	"github.com/saltfishpr/tools/pkg/util"
)

var defaultClient *http.Client = &http.Client{
	Timeout: 3 * time.Second,
}

func SetDefaultClient(c *http.Client) {
	defaultClient = c
}

// ListVersions lists all versions of a module. Sorts the versions in descending order by semver.
func ListVersions(modulePath string) ([]string, error) {
	url := fmt.Sprintf("%s/%s/@v/list", goProxy, modulePath)
	resp, err := defaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	versions := strings.Split(strings.TrimSpace(string(body)), "\n")
	semver.Sort(versions)
	slices.Reverse(versions)
	return versions, nil
}

// GetLatestVersion returns the go.mod file of a module.
func GetModFile(modulePath, version string) (*modfile.File, error) {
	url := fmt.Sprintf("%s/%s/@v/%s.mod", goProxy, modulePath, version)
	resp, err := defaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, err
	}

	return f, nil
}

var goProxy = getGoProxy()

func getGoProxy() string {
	proxy, err := util.GetGoProxy()
	if err != nil {
		log.Errorf("get go proxy from env error: %v", err)
	}
	log.Infof("using go proxy: %s", proxy)
	return proxy
}
