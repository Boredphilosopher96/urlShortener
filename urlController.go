package main

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/base62"
	"net/http"
	"strings"
	"sync"
)

const baseUrl string = "https://bit.ly/"

var (
	urlMap   = map[string]string{}
	mapMutex = sync.RWMutex{}
)

func getUrlFromMap(shortUrl string) (string, bool) {
	mapMutex.RLock()
	ogUrl, exists := urlMap[shortUrl]
	mapMutex.RUnlock()
	return ogUrl, exists
}

func getOriginalUrl(c *gin.Context) {
	url, exists := c.GetPostForm("url")
	if !exists {
		fmt.Println("This url does not exist")
		c.JSON(http.StatusNotFound, "URL not found")
	}
	splitUrl := strings.Split(url, "?")

	if originalUrl, exists := getUrlFromMap(splitUrl[0]); exists {
		if len(splitUrl) == 2 {
			originalUrl = originalUrl + "?" + splitUrl[1]
		}
		c.JSON(http.StatusOK, originalUrl)
	} else {
		c.JSON(http.StatusNotFound, "URL not found")
	}
}

func home(c *gin.Context) {
	c.JSON(http.StatusOK, "This is the home page")
}

func getMD5Hash(text string) [16]byte {
	return md5.Sum([]byte(text))
}

func setUrlValue(shortUrl string, originalUrl string) {
	mapMutex.Lock()
	urlMap[shortUrl] = originalUrl
	mapMutex.Unlock()
}

func getShortUrl(c *gin.Context) {
	url, exists := c.GetPostForm("url")

	if !exists {
		fmt.Println("No query result")
		return
	}

	if len(strings.Split(url, "?")) > 2 {
		fmt.Println("Seems like you an invalid url")
		c.JSON(http.StatusBadRequest, "Invalid URL")
	}

	finalUrl := createShortUrl(url)

	c.JSON(http.StatusOK, map[string]interface{}{
		"shortUrl": finalUrl,
	})
}

func createShortUrl(url string) string {
	splitStrings := strings.Split(url, "?")

	hashed := getMD5Hash(splitStrings[0])
	fmt.Println("Hashed URL ", hashed)

	encoded := base62.EncodeToString(hashed[:])
	fmt.Println("Encoded string ", encoded)

	finalUrl := baseUrl + encoded[:7]
	fmt.Println("This is the final url ", finalUrl)

	go setUrlValue(finalUrl, splitStrings[0])

	return finalUrl
}

func main() {
	r := gin.Default()
	r.GET("/", home)
	r.POST("/shorten", getShortUrl)
	r.POST("/original", getOriginalUrl)

	err := r.Run()
	if err != nil {
		fmt.Print("Something errored out")
		return
	}
}
