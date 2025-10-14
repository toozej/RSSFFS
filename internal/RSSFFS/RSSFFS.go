package RSSFFS

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/toozej/RSSFFS/pkg/config"
	"golang.org/x/net/html"
)

// Category struct to unmarshal the JSON response
type Category struct {
	Title  string `json:"title"`
	UserID int    `json:"user_id"`
	ID     int    `json:"id"`
}

var (
	apiEndpoint string
	apiKey      string
)

var commonPatterns = []string{"/index.xml", "/feed", "/rss", "/atom.xml"}

const maxRedirects = 10
const timeoutSeconds = 10

// getAllDomainsFromPage retrieves all unique domain names from a webpage
func getAllDomainsFromPage(pageURL string) (map[string]bool, error) {
	client := &http.Client{
		Timeout: time.Second * timeoutSeconds,
	}

	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tokenizer := html.NewTokenizer(resp.Body)
	domains := make(map[string]bool)

	// Parse HTML and extract URLs
	for {
		tt := tokenizer.Next()
		switch {
		case tt == html.ErrorToken:
			return domains, nil
		case tt == html.StartTagToken:
			t := tokenizer.Token()
			if t.Data == "a" {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						u, err := url.Parse(attr.Val)
						if err == nil && u.Host != "" {
							domain := u.Hostname()
							domains[domain] = true
						}
					}
				}
			}
		}
	}
}

// checkDomainsForRSS checks for RSS feeds on the given domains with concurrency
func checkDomainsForRSS(domains map[string]bool) []string {
	var wg sync.WaitGroup
	feedChan := make(chan string)
	feedMap := make(map[string]bool)
	mu := sync.Mutex{}

	for domain := range domains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			feed := findPreferredRSSFeed(domain)
			if feed != "" {
				mu.Lock()
				if !feedMap[domain] {
					feedMap[domain] = true
					feedChan <- feed
				}
				mu.Unlock()
			}
		}(domain)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(feedChan)
	}()

	var validFeeds []string
	for feed := range feedChan {
		validFeeds = append(validFeeds, feed)
	}

	return validFeeds
}

// findPreferredRSSFeed checks RSS patterns for a domain and returns the first valid one based on preference
func findPreferredRSSFeed(domain string) string {
	client := &http.Client{
		Timeout: time.Second * timeoutSeconds,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	for _, pattern := range commonPatterns {
		feedURL := "https://" + domain + pattern
		if checkRSSFeed(client, feedURL) {
			return feedURL
		}
	}
	return ""
}

// checkRSSFeed checks if the given URL is a valid RSS feed
func checkRSSFeed(client *http.Client, feedURL string) bool {
	resp, err := client.Get(feedURL)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	defer resp.Body.Close()

	// Check if the Content-Type header indicates an RSS feed
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "xml") || strings.Contains(contentType, "rss")
}

func Run(pageURL string, category string, debug bool, clearCategoryFeeds bool, conf config.Config) {
	// Use configuration passed from caller
	apiEndpoint, apiKey = conf.RSSReaderEndpoint, conf.RSSReaderAPIKey

	// Get categoryId of user-input category if it exists
	categoryId, err := getCategoryId(apiEndpoint, apiKey, category)
	if err != nil {
		log.Fatalf("Error getting categoryId from category %s: %s\n", category, err)
	}

	// delete all feeds within categoryId if user requested it
	if clearCategoryFeeds {
		feedIds, err := getCategoryFeeds(apiEndpoint, apiKey, categoryId)
		if err != nil {
			log.Fatalf("Error getting feeds in categoryId %d: %v\n", categoryId, err)
		}
		log.Info("Deleting feeds from categoryId: ", categoryId)
		for _, feedId := range feedIds {
			log.Debug("Deleting feedId ", feedId)
			err := deleteFeed(apiEndpoint, apiKey, feedId)
			if err != nil {
				log.Errorf("Error deleting feedId %d: %\n ", feedId, err)
			}
		}
	}

	// Get all unique domains from the page
	log.Info("Getting all unique domains from the URL: ", pageURL)
	domains, err := getAllDomainsFromPage(pageURL)
	if err != nil {
		log.Fatal("Error fetching page: ", err)
	}

	// Deduplicate valid RSS feeds
	validFeeds := checkDomainsForRSS(domains)

	// Subscribe to valid RSS feeds
	for _, feed := range validFeeds {
		if debug {
			log.Debug("Pretending to subscribe to feed: ", feed)
		} else {
			if err := subscribeToFeed(apiEndpoint, apiKey, categoryId, feed); err != nil {
				log.Errorf("Error subscribing to RSS feed %s: %v\n", feed, err)
			}
		}
	}
}
