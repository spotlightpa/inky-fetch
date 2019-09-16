package fetchapp

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/peterbourgon/ff"
	"github.com/spotlightpa/inky-fetch/internal/cachedata"
	"github.com/spotlightpa/inky-fetch/internal/feed"
	"github.com/spotlightpa/inky-fetch/internal/slack"
)

func CLI(args []string) error {
	fl := flag.NewFlagSet("app", flag.ContinueOnError)
	feed := flagext.FileOrURL("https://www.inquirer.com/arcio/rss/", nil)
	fl.Var(feed, "feed", "source file or URL")
	cacheloc := cachedata.Default("inky-feed")
	fl.Var(&cacheloc, "cache-location", "file `path` to save seen URLs")
	verbose := fl.Bool("verbose", false, "log debug output")
	slackURL := fl.String("slack-web-hook", "", "web hook to post Slack messages")
	interval := fl.Duration("interval", 0, "poll interval (if 0, only runs once)")
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `inky-fetch - Fetches Spotlight PA stories from the Philadelphia Inquirer

Usage:

	inky-fetch [options]

Options:
`)
		fl.PrintDefaults()
	}
	if err := ff.Parse(fl, args, ff.WithEnvVarPrefix("INKY_FETCH")); err != nil {
		return err
	}

	return appExec(feed, *slackURL, *verbose, cacheloc, *interval)
}

func appExec(feed io.ReadCloser, slackURL string, verbose bool, loc cachedata.Loc, interval time.Duration) error {
	l := nooplogger
	if verbose {
		l = log.New(os.Stderr, "inky-fetch ", log.LstdFlags).Printf
	}

	var sc *slack.Client
	if slackURL != "" {
		sc = slack.New(slackURL, l, nil)
	}
	a := app{feed, sc, l, loc}
	if interval == 0 {
		if err := a.exec(); err != nil {
			fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
			return err
		}
	} else {
		a.loop(interval)
	}
	return nil
}

type logger = func(format string, v ...interface{})

func nooplogger(format string, v ...interface{}) {}

type app struct {
	feed io.ReadCloser
	sc   *slack.Client
	log  logger
	loc  cachedata.Loc
}

func (a *app) exec() (err error) {
	a.log("starting")
	defer func() { a.log("done") }()

	urls, err := feed.GetSpotlightLinks(a.feed)
	if err != nil {
		return err
	}

	urls, err = a.dedupe(urls)
	if err != nil {
		return err
	}

	a.log("no error fetching %d urls", len(urls))
	if a.sc != nil {
		err = a.postToSlack(urls)
	} else {
		a.logToTerm(urls)
	}

	return err
}

func (a *app) dedupe(urls []*url.URL) ([]*url.URL, error) {
	a.log("deduping %d urls", len(urls))

	var seen map[string]bool
	err := a.loc.Read(&seen)
	if err != nil {
		return nil, err
	}
	if seen == nil {
		seen = make(map[string]bool)
	}

	filteredURLs := urls[:0]
	for _, u := range urls {
		if !seen[u.String()] {
			filteredURLs = append(filteredURLs, u)
			seen[u.String()] = true
		}
	}

	if err = a.loc.Write(&seen); err != nil {
		return nil, err
	}

	a.log("found %d new urls", len(filteredURLs))
	return filteredURLs, nil
}

func (a *app) postToSlack(urls []*url.URL) (err error) {
	a.log("posting to slack")
	if len(urls) > 0 {
		err = a.sc.Post(a.messageFrom(urls))
	}
	return
}

func (a *app) messageFrom(urls []*url.URL) slack.Message {
	attachments := make([]slack.Attachment, len(urls))
	for i := range urls {
		u := urls[i].String()
		attachments[i] = slack.Attachment{
			Fallback:  u,
			Color:     "#005eb8",
			Title:     u,
			TitleLink: u,
		}
	}
	return slack.Message{
		Text:        "Found Spotlight Inquirer Page",
		Attachments: attachments,
	}
}

func (a *app) logToTerm(urls []*url.URL) {
	fmt.Printf("found %d url(s):\n", len(urls))
	for _, u := range urls {
		fmt.Printf("- %v\n", u)
	}
}

func (a *app) loop(interval time.Duration) {
	for {
		wait := time.After(interval)
		if err := a.exec(); err != nil {
			fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		}
		<-wait
	}
}
