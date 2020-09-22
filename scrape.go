package scrape

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/gocolly/colly/v2/queue"
)

// CrawlerParams : Specify behavior for Collector object
type CrawlerParams struct {
	Depth   int
	Agent   string
	Before  func(*colly.Request)
	Success func(*colly.Response)
	Failed  func(error, *colly.Response)
	Done    func(*colly.Response)
	Proxies []*url.URL
}

type LimitRule = colly.LimitRule

// Crawler : creates a Collector with specified parameters
func Crawler(args CrawlerParams, limits ...LimitRule) *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(args.Depth),
		colly.Debugger(&debug.LogDebugger{}),
	)

	if limits != nil {
		for _, rule := range limits {
			c.Limit(&rule)
		}
	}

	if args.Agent != "" {
		c.UserAgent = args.Agent
	} else {
		extensions.RandomUserAgent(c)
	}

	if args.Proxies != nil {
		c.SetProxyFunc(func(_ *http.Request) (*url.URL, error) {
			return args.Proxies[rand.Intn(len(args.Proxies))], nil
		})
	}

	c.OnError(func(res *colly.Response, err error) {
		log.Println("Something went wrong:", err)
		if args.Failed != nil {
			args.Failed(err, res)
		}
	})

	c.OnRequest(func(req *colly.Request) {
		req.Headers.Set("X-Requested-With", "XMLHttpRequest")
		req.Headers.Set("Referrer", req.URL.String())
		fmt.Println("Visiting", req.URL)
		if args.Before != nil {
			args.Before(req)
		}
	})

	// Warning: this extension works only if you use Request.Visit from callbacks instead of Collector.Visit.
	extensions.Referer(c)
	c.OnResponse(func(res *colly.Response) {
		// fmt.Println("Visited", res.Request.URL)
		if args.Success != nil {
			args.Success(res)
		}
	})

	c.OnScraped(func(res *colly.Response) {
		// fmt.Println("Finished", res.Request.URL)
		if args.Done != nil {
			args.Done(res)
		}
	})

	return c
}

// RequestQueue : Run the urls in a queue using a Collector
func RequestQueue(threads int, c *colly.Collector, urls []string) {
	q, _ := queue.New(
		threads, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	for _, url := range urls {
		q.AddURL(url)
	}

	q.Run(c)

}
