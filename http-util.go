package http_util

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/axgle/mahonia"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

//GetWebConFromUrl simply get web content
//from net
func GetWebConFromUrl(url string) (string, error) {
	response, err := doRequest(url, nil, "GET", nil, 10*time.Second, "")
	if err != nil {
		return "", err
	}
	return ReadContentFromResponse(response, "")
}

// get http.Response from url
func GetWebResponseFromUrl(url string) (*http.Response, error) {
	return doRequest(url, nil, "GET", nil, 10*time.Second, "")

}

func GetWebResponseFromUrlWithHeader(url string, headerMap map[string]string) (*http.Response, error) {
	return doRequest(url, headerMap, "GET", nil, 60*time.Second, "")

}

//GetWebConFromUrlWithAllArgs get web content
//with some args
func GetWebConFromUrlWithAllArgs(url string, headerMap map[string]string, method string, postData []byte, timeOut time.Duration) (string, error) {
	response, err := doRequest(url, headerMap, method, postData, timeOut, "")
	if err != nil {
		return "", err
	}
	return ReadContentFromResponse(response, "")
}

//GetWebConFromUrlWithHeader get web con from target url
//param headerMap is some header info
func GetWebConFromUrlWithHeader(url string, headerMap map[string]string) (string, error) {
	response, err := doRequest(url, headerMap, "GET", nil, 10*time.Second, "")
	if err != nil {
		return "", err
	}
	return ReadContentFromResponse(response, "")
}

func GetContentFromResponse(response *http.Response) (string, error) {
	defer response.Body.Close()
	if response.StatusCode >= 300 || response.StatusCode < 200 {
		return "", errors.New(fmt.Sprintf("状态码为: %d", response.StatusCode))
	}
	char, data := detectContentCharset(response.Body)
	if data == nil {
		return "", errors.New("数据为空")
	}

	dec := mahonia.NewDecoder(char)
	preRd := dec.NewReader(data)
	preBytes, err := ioutil.ReadAll(preRd)
	if err != nil {
		return "", err
	}
	return string(preBytes), err
}

func SendRequest(targetUrl string, headerMap map[string]string, method string, postData []byte, timeOut time.Duration) (*http.Response, error) {
	return doRequest(targetUrl, headerMap, method, postData, timeOut, "")
}

func SendRequestWithProxy(targetUrl string, headerMap map[string]string, method string, postData []byte, timeOut time.Duration, proxy string) (*http.Response, error) {
	return doRequest(targetUrl, headerMap, method, postData, timeOut, proxy)
}

func doRequest(targetUrl string, headerMap map[string]string, method string, postData []byte, timeOut time.Duration, proxy string) (*http.Response, error) {

	//timeOut = time.Duration(timeOut * time.Millisecond)
	//urli := url.URL{}
	//urlproxy, _ := urli.Parse("https://127.0.0.1:9743")
	//https认证
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		//Proxy:http.ProxyURL(urlproxy),

	}
	if proxy != "" {
		urli := url.URL{}
		urlProxy, _ := urli.Parse(proxy)
		tr.Proxy = http.ProxyURL(urlProxy)
	}
	client := http.Client{
		Timeout:   timeOut,
		Transport: tr,
	}

	client.Jar, _ = cookiejar.New(nil)

	method = strings.ToUpper(method)
	var req *http.Request
	var err error
	if postData != nil && (method == "POST" || method == "PUT") {
		req, err = http.NewRequest(method, targetUrl, bytes.NewReader(postData))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	} else {
		req, err = http.NewRequest(method, targetUrl, nil)
		if err != nil {
			return nil, err
		}

	}

	for key, value := range headerMap {
		req.Header.Set(key, value)
	}
	res, err := client.Do(req)
	client.CloseIdleConnections()
	return res, err
}

func URLEncode(keyword string) string {
	return url.QueryEscape(keyword)
}

// 这里他娘的Peek是针对Reader的 针对不了io.Reader 所以 io.Reader其实是进行了位移的
func detectContentCharset(body io.Reader) (string, *bufio.Reader) {
	r := bufio.NewReader(body)
	if data, err := r.Peek(1024); err == nil {

		if _, name, _ := charset.DetermineEncoding(data, ""); name != "" {
			if name == "windows-1252" {
				datastr := strings.ToLower(string(data))
				if strings.Contains(datastr, "charset=utf-8") && !strings.Contains(datastr, `charset=gbk`) {
					return "utf-8", r
				} else {
					return "GBK", r
				}
			}
			return name, r
		}
	}
	return "utf-8", r
}
func ReadContentFromResponse(response *http.Response, charset string) (string, error) {
	defer response.Body.Close()
	var err error
	var htmlbytes []byte
	contentEncoding, ok := response.Header["Content-Encoding"]
	if ok && contentEncoding[0] == "gzip" {
		gzreader, err := gzip.NewReader(response.Body)
		if err != nil {
			return "", err
		}
		for {
			buf := make([]byte, 1024)
			n, err := gzreader.Read(buf)
			if err != nil && err != io.EOF {
				gzreader.Close()
				return "", err
			}
			if n == 0 {
				break
			}
			htmlbytes = append(htmlbytes, buf[:n]...)
		}
		gzreader.Close()
		//htmlbytes,err=ioutil.ReadAll(gzreader)
		//println(string(htmlbytes))
	} else {
		htmlbytes, err = ioutil.ReadAll(response.Body)
	}
	//response.Body = reader

	if response.StatusCode >= 300 || response.StatusCode < 200 {
		return "", errors.New(fmt.Sprintf("状态码为: %d", response.StatusCode))
	}
	hreader := bytes.NewReader(htmlbytes)

	char, data := detectContentCharset(hreader)
	if charset != "" {
		char = charset
	}
	if char == "windows-1252" {
		char = "GBK"
	}
	if data == nil {
		return "", errors.New("数据为空")
	}

	dec := mahonia.NewDecoder(char)
	if dec == nil {
		dec = mahonia.NewDecoder("utf-8")
	}
	preRd := dec.NewReader(data)
	preBytes, err := ioutil.ReadAll(preRd)
	reBytes, err := ioutil.ReadAll(hreader)
	if err != nil {
		return "", err
	}
	return string(append(preBytes, reBytes...)), err
}
