package mod

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

var goProxy = getGoProxy()

var defaultClient *http.Client = &http.Client{
	Timeout: 3 * time.Second,
}

func SetDefaultClient(c *http.Client) {
	defaultClient = c
}

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

func getGoProxy() string {
	proxy := os.Getenv("GOPROXY")
	if proxy == "" {
		return "https://proxy.golang.org"
	}
	if i := strings.Index(proxy, ","); i > 0 {
		return proxy[:i]
	}
	return proxy
}
