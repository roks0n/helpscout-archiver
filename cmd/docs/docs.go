package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/roks0n/helpscout-archiver/internal/docs"
)

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}
}

func saveImages(a *docs.Article) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(a.Text))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		fmt.Printf("Downloading image %s\n", src)

		resp, err := http.Get(src)
		if err != nil {
			fmt.Println("Error downloading image:", err)
			return
		}
		defer resp.Body.Close()

		var filename = strings.Split(src, "/")[len(strings.Split(src, "/"))-1]
		var newSrc = "data/docs/" + a.Slug + "/" + filename

		img, err := os.Create(newSrc)
		if err != nil {
			fmt.Println("Error creating image file:", err)
			return
		}

		_, err = io.Copy(img, resp.Body)
		if err != nil {
			fmt.Println("Error saving image:", err)
			return
		}

		s.SetAttr("src", filename)
		fmt.Println("Image downloaded")
	})

	// Update the article's text with the modified HTML
	html, err := doc.Html()
	if err != nil {
		fmt.Println("Error generating HTML:", err)
		return
	}
	a.Text = html
}

func archiveArticle(a *docs.Article) {
	fmt.Printf("Archiving article %s\n", a.Name)

	createDirIfNotExist("data/docs/" + a.Slug + "/")

	saveImages(a)

	f, err := os.Create("data/docs/" + a.Slug + "/index.html")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer f.Close()
	_, err = f.WriteString(a.Text)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("Article archived")
}

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Usage: go run main.go <argument>")
		return
	}
	
	token := args[1]
	docs := docs.NewDocs(token)

	cids := docs.CollectionIDS(1)

	var aids []int
	for _, cid := range cids {
		res := docs.ArticleNumbers(cid, 1)
		aids = append(aids, res...)
	}

	for _, aid := range aids {
		a := docs.Article(aid)
		archiveArticle(a)
	}
}
