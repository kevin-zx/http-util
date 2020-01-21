package http_util

import (
	"fmt"
	"testing"
)

func TestGetContentFromResponse(t *testing.T) {
	wec, err := GetWebConFromUrl("http://www.foodjx.com/st116052/erlist_306726.html")
	if err != nil {
		panic(err)
	}
	fmt.Println(wec)
}
