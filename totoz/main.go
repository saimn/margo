// See https://linuxfr.org/users/n_e/journaux/recherche-totoz-en-javascript

package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Totozes struct {
	XMLName xml.Name `xml:"totozes"`
	Totozes []Totoz  `xml:"totoz"`
}

type Totoz struct {
	XMLName  xml.Name `xml:"totoz"`
	Name     string   `xml:"name"`
	UserName string   `xml:"username"`
	Nsfw     string   `xml:"nsfw"`
	Tags     []string `xml:"tag"`
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: totoz name")
	}
	ch := make(chan string)

	go func(query string) {
		resp, err := http.Get("https://totoz.eu/search.xml?terms=" + url.QueryEscape(query))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		var totozes Totozes
		xml.Unmarshal(body, &totozes)

		for i := 0; i < len(totozes.Totozes); i++ {
			ch <- totozes.Totozes[i].Name
		}
		close(ch)
	}(os.Args[1])

	for i := range ch {
		fmt.Println(i)
	}
}
