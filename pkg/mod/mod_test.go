package mod_test

import (
	"fmt"
	"testing"

	"github.com/saltfishpr/tools/pkg/mod"
)

func TestListVersions(t *testing.T) {
	got, err := mod.ListVersions("github.com/labstack/echo")
	if err != nil {
		t.Fatalf("ListVersions() error = %v", err)
	}
	fmt.Println(got)
}
