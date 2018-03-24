package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"golang.org/x/net/html"
)

var explored map[string]struct{}

var frontier chan []string

var wg = sync.WaitGroup{}

var mu = sync.RWMutex{}

func init() {
	explored = make(map[string]struct{})
	frontier = make(chan []string, NumberOfCrawlers)
	_ = os.Mkdir(DefaultImagePath, 0755)
}

/*
Run is the starting of the crawler
*/
func Run(entryPoints []string) {
	wg.Add(1)
	go func() {
		frontier <- entryPoints
		wg.Wait()
		close(frontier)
	}()
	for fetchedUrls := range frontier {
		for _, url := range fetchedUrls {
			wg.Add(1)
			mu.RLock()
			_, ok := explored[url]
			mu.RUnlock()
			if !ok {
				fmt.Println(url)
				go handleUrl(url)
			}
		}
	}
}

func handleUrl(url string) {
	mu.Lock()
	explored[url] = struct{}{}
	mu.Unlock()
	defer wg.Done()
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()
	rootNode, err := html.Parse(res.Body)
	if err != nil {
		log.Println(err)
		return
	}
	visitNode := func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			links := make([]string, 0)
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					url, err := res.Request.URL.Parse(attr.Val)
					if err != nil {
						log.Println(err)
						continue
					}
					links = append(links, url.String())
				}
			}
			frontier <- links
		}
	}
	saveImage := func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "img" {
			for _, attr := range node.Attr {
				if attr.Key == "src" {
					url, err := res.Request.URL.Parse(attr.Val)
					if err != nil {
						fmt.Println(err)
						continue
					}
					go handleImage(url.String())
				}
			}
		}
	}
	traverse(rootNode, visitNode, saveImage)
}

func traverse(node *html.Node, pre, post func(*html.Node)) {
	if pre != nil {
		pre(node)
	}
	for current := node.FirstChild; current != nil; current = current.NextSibling {
		traverse(current, pre, post)
	}
	if post != nil {
		post(node)
	}
}

func handleImage(url string) {
	fmt.Println("Fetching image", url)
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()
	file, err := os.Create(DefaultImagePath + path.Base(url))
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = io.Copy(file, res.Body)
	if err != nil {
		fmt.Println(err)
	}
}
