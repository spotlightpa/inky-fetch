package fetchapp

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/mmcdole/gofeed"
)

func GetSpotlightLinks(r io.Reader) (urls []*url.URL, err error) {
	fp := gofeed.NewParser()
	feed, err := fp.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("could not extract Spotlight links: %w", err)
	}
	for _, item := range feed.Items {
		link, err := url.Parse(item.Link)
		if err != nil {
			return nil, fmt.Errorf("could not extract Spotlight links: %w", err)
		}
		if strings.Contains(link.Path, "/spl/") {
			urls = append(urls, link)
		}
	}
	return
}
