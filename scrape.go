package scrape

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/gocolly/colly/v2/queue"
)

// CrawlerParams : Specify behavior for Collector object
type CrawlerParams struct {
	threads, throttle, randomDelay, depth int
	agent                                 string
	domainGlob                            string
	before                                func(*colly.Request)
	success                               func(*colly.Response)
	failed                                func(error, *colly.Response)
	done                                  func(*colly.Response)
	proxies                               []*url.URL
}

// Crawler : creates a Collector with specified parameters
func Crawler(args CrawlerParams) *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(args.depth),
		colly.Debugger(&debug.LogDebugger{}),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  args.domainGlob,
		Parallelism: args.threads,
		Delay:       time.Duration(args.throttle) * time.Second,
		RandomDelay: time.Duration(args.randomDelay) * time.Second,
	})

	if args.agent != "" {
		c.UserAgent = args.agent
	} else {
		extensions.RandomUserAgent(c)
	}

	if args.proxies != nil {
		c.SetProxyFunc(func(_ *http.Request) (*url.URL, error) {
			return args.proxies[rand.Intn(len(args.proxies))], nil
		})
	}

	c.OnError(func(res *colly.Response, err error) {
		log.Println("Something went wrong:", err)
		if args.failed != nil {
			args.failed(err, res)
		}
	})

	c.OnRequest(func(req *colly.Request) {
		req.Headers.Set("X-Requested-With", "XMLHttpRequest")
		req.Headers.Set("Referrer", req.URL.String())
		fmt.Println("Visiting", req.URL)
		if args.before != nil {
			args.before(req)
		}
	})

	// Warning: this extension works only if you use Request.Visit from callbacks instead of Collector.Visit.
	extensions.Referer(c)
	c.OnResponse(func(res *colly.Response) {
		// fmt.Println("Visited", res.Request.URL)
		if args.success != nil {
			args.success(res)
		}
	})

	c.OnScraped(func(res *colly.Response) {
		// fmt.Println("Finished", res.Request.URL)
		if args.done != nil {
			args.done(res)
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
