package util

import (
	"io/ioutil"
	"log"
	"net/http"
)

func handleErr(err error) {
	log.Fatalln(err)
}

func DownloadPage(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		handleErr(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleErr(err)
	}
	return string(body)
}
