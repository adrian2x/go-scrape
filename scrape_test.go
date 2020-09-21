package scrape

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"

	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"
)

func Trace(err error) {
	fmt.Println(errors.Wrap(err, 1).ErrorStack())
}

func BasePath(myurl string) string {
	urlObj, _ := url.Parse(myurl)
	return path.Base(urlObj.Path)
}

func main() {
	// List of urls we want to scrape
	urls := os.Args[1:]

	// Create the crawler with config
	client := Crawler(CrawlerParams{
		Threads:    1,
		Depth:      1,
		Throttle:   5,
		DomainGlob: "",
		Success: func(res *colly.Response) {
			// Headers
			contentType := res.Headers.Get("Content-Type")
			parts := regexp.MustCompile("[/;\\s]+").Split(contentType, -1)
			// Save to file using the last path segment and MIME type
			filext := parts[1]
			filename := BasePath(res.Request.URL.String()) + "." + filext
			println("Saving ", filename)
			if err := res.Save("./" + filename); err != nil {
				Trace(err)
			}
		},
	})

	// Visit urls
	// for _, url := range urls {
	// 	client.Visit(url)
	// }

	// Process using a queue
	RequestQueue(1, client, urls)

	// Wait until done
	client.Wait()
}
