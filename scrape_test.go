package scrape

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"time"

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
		Depth: 1,
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
	}, LimitRule{
		Parallelism: 1,
		DomainGlob:  "",
		Delay:       time.Duration(1) * time.Second,
		RandomDelay: time.Duration(5) * time.Second,
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
