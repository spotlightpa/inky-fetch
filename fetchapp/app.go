package fetchapp

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/carlmjohnson/flagext"
	"github.com/peterbourgon/ff"
	"github.com/spotlightpa/inky-fetch/internal/feed"
	"github.com/spotlightpa/inky-fetch/internal/slack"
)

func CLI(args []string) error {
	fl := flag.NewFlagSet("app", flag.ContinueOnError)
	feed := flagext.FileOrURL("https://www.inquirer.com/arcio/rss/", nil)
	fl.Var(feed, "feed", "source file or URL")
	verbose := fl.Bool("verbose", false, "log debug output")
	slackURL := fl.String("slack-web-hook", "", "web hook to post Slack messages")
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
	if *slackURL == "" {
		fmt.Fprintf(fl.Output(), "Slack Web Hook not set\n\n")
		fl.Usage()
		return flag.ErrHelp
	}

	return appExec(feed, *slackURL, *verbose)
}

func appExec(feed io.ReadCloser, slackURL string, verbose bool) error {
	l := nooplogger
	if verbose {
		l = log.New(os.Stderr, "inky-fetch", log.LstdFlags).Printf
	}
	sc := slack.New(slackURL, l, nil)
	a := app{feed, sc, l}
	if err := a.exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return err
	}
	return nil
}

type logger = func(format string, v ...interface{})

func nooplogger(format string, v ...interface{}) {}

type app struct {
	feed io.ReadCloser
	sc   *slack.Client
	log  logger
}

func (a *app) exec() (err error) {
	a.log("starting")
	defer func() { a.log("done") }()

	urls, err := feed.GetSpotlightLinks(a.feed)
	if err != nil {
		return err
	}

	a.log("no error fetching %d urls", len(urls))
	if len(urls) > 0 {
		a.sc.Post(a.messageFrom(urls))
	}

	return err
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
