package http_util

import (
	"fmt"
	"testing"
)

func TestGetContentFromResponse(t *testing.T) {
	wec, err := GetWebConFromUrl("http://www.baidu.com")
	if err != nil {
		panic(err)
	}
	fmt.Println(wec)
}
