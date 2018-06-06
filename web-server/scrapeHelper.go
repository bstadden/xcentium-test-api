package main

import (
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func getResponse(u string) response {
	scrapeResp := response{}
	resp, err := http.Get(u)
	if err != nil {
		scrapeResp.Err = "Error connecting to the website."
	}

	scrapeResp.Images = findImages(resp)

	resp, err = http.Get(u)
	if err != nil {
		scrapeResp.Err = "Error connecting to the website."
	}
	scrapeResp.Words = findWords(resp)
	return scrapeResp
}

// Cycle through all img tags and return parsed data
func findImages(r *http.Response) []image {
	t := html.NewTokenizer(r.Body)
	defer r.Body.Close()

	imgs := []image{}
	for {
		htmlTok := t.Next()

		switch {
		case htmlTok == html.ErrorToken:
			return imgs // EOF return
		case htmlTok == html.StartTagToken || htmlTok == html.SelfClosingTagToken:
			tok := t.Token()
			if strings.ToLower(tok.Data) == "img" {
				img, ok := parseImgTag(tok.Attr)
				if ok {
					imgs = append(imgs, img)
				}
			}
		}
	}
}

func parseImgTag(attrs []html.Attribute) (image, bool) {
	img := image{}
	ok := false
	for _, attr := range attrs {
		switch strings.ToLower(attr.Key) {
		case "src":
			img.ImgURL = attr.Val
			ok = true // only URL is required
		case "alt":
			img.Desc = attr.Val
		}
	}
	if img.ImgURL != "" {
		img.Name = path.Base(img.ImgURL)
	}

	return img, ok
}

func findWords(r *http.Response) words {
	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		// handle
	}
	defer r.Body.Close()

	doc.Find("script").Remove()
	doc.Find("style").Remove()
	bodyText := doc.Find("body").First().Text()

	wordList := extractWords(bodyText)
	words := getTopWords(wordList)

	return words
}

func extractWords(s string) []string {
	regex := regexp.MustCompile(`[a-zA-Z]+`)
	return regex.FindAllString(strings.ToLower(s), -1)
}

func getTopWords(allWords []string) words {
	counts := make(map[string]int)
	totalCount := 0
	for _, w := range allWords {
		// workaround - go regex does not have exceptions
		// was unable to remove newline chars
		if w != "n" {
			counts[w]++
			totalCount++
		}
	}

	allWordsList := wordList{}
	for k, v := range counts {
		allWordsList = append(allWordsList, word{Word: k, Count: v})
	}
	sort.Sort(allWordsList)
	return words{TotalCount: totalCount, TopWords: allWordsList[:10]}
}
