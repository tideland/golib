// Tideland Go Library - RSS Feed
//
// Copyright (C) 2012-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package rss

//--------------------
// IMPORTS
//--------------------

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/feed/utils"
)

//--------------------
// CONST
//--------------------

// Version of the RSS.
const (
	Version = "2.0"
)

const (
	rssDate   = "Mon, 02 Jan 2006 15:04 MST"
	rssDateV1 = "Mon, 02 Jan 2006 15:04:05 MST"
	rssDateV2 = "02 Jan 2006 15:04 MST"
	rssDateV3 = "Mon, 02 Jan 2006 15:04 -0700"
	rssDateV4 = "02 Jan 2006 15:04 -0700"
)

//--------------------
// MODEL
//--------------------

// RSS is the root element of the document.
type RSS struct {
	XMLName string  `xml:"rss"`
	Version string  `xml:"version,attr"`
	Channel Channel `xml:"channel"`
}

// Validate checks if the RSS document is valid.
func (r *RSS) Validate() error {
	if r.Version != Version {
		return errors.New(ErrValidation, errorMessages, "invalid RSS document version: %q", r.Version)
	}
	return r.Channel.Validate()
}

// Channel is the one channel element of the RSS document.
type Channel struct {
	Title          string      `xml:"title"`
	Description    string      `xml:"description"`
	Link           string      `xml:"link"`
	Categories     []*Category `xml:"category,omitempty"`
	Cloud          *Cloud      `xml:"cloud,omitempty"`
	Copyright      string      `xml:"copyright,omitempty"`
	Docs           string      `xml:"docs,omitempty"`
	Generator      string      `xml:"generator,omitempty"`
	Image          *Image      `xml:"image,omitempty"`
	Language       string      `xml:"language,omitempty"`
	LastBuildDate  string      `xml:"lastBuildDate,omitempty"`
	ManagingEditor string      `xml:"managingEditor,omitempty"`
	PubDate        string      `xml:"pubDate,omitempty"`
	Rating         string      `xml:"rating,omitempty"`
	SkipDays       *SkipDays   `xml:"skipDays,omitempty"`
	SkipHours      *SkipHours  `xml:"skipHours,omitempty"`
	TextInput      string      `xml:"textInput,omitempty"`
	TTL            int         `xml:"ttl,omitempty"`
	WebMaster      string      `xml:"webMaster,omitempty"`
	Items          []*Item     `xml:"item,omitempty"`
}

// Validate checks if the cannel is valid.
func (c Channel) Validate() error {
	if c.Title == "" {
		return errors.New(ErrValidation, errorMessages, "channel title must not be empty")
	}
	if c.Description == "" {
		return errors.New(ErrValidation, errorMessages, "channel description must not be empty")
	}
	if c.Link == "" {
		return errors.New(ErrValidation, errorMessages, "channel link must not be empty")
	}
	if _, err := url.Parse(c.Link); err != nil {
		return errors.Annotate(err, ErrParsing, errorMessages, "channel link")
	}
	for _, category := range c.Categories {
		if err := category.Validate(); err != nil {
			return err
		}
	}
	if c.Cloud != nil {
		if err := c.Cloud.Validate(); err != nil {
			return err
		}
	}
	if c.Docs != "" && c.Docs != "http://blogs.law.harvard.edu/tech/rss" {
		return errors.New(ErrValidation, errorMessages, "docs %q is not valid", c.Docs)
	}
	if c.Image != nil {
		if err := c.Image.Validate(); err != nil {
			return err
		}
	}
	if c.Language != "" {
		// TODO(mue) Language has to be validated.
	}
	if c.LastBuildDate != "" {
		if _, err := ParseTime(c.LastBuildDate); err != nil {
			return errors.Annotate(err, ErrParsing, errorMessages, "channel last build date")
		}
	}
	if c.PubDate != "" {
		if _, err := ParseTime(c.PubDate); err != nil {
			return errors.Annotate(err, ErrParsing, errorMessages, "channel publication date")
		}
	}
	if c.SkipDays != nil {
		if err := c.SkipDays.Validate(); err != nil {
			return err
		}
	}
	if c.SkipHours != nil {
		if err := c.SkipHours.Validate(); err != nil {
			return err
		}
	}
	if c.TTL < 0 {
		return errors.New(ErrValidation, errorMessages, "channel ttl is below zero")
	}
	for _, item := range c.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Category identifies a category or tag to which the feed belongs.
type Category struct {
	Category string `xml:",chardata"`
	Domain   string `xml:"domain,attr,omitempty"`
}

// Validate checks if the category is valid.
func (c *Category) Validate() error {
	if c.Category == "" {
		return errors.New(ErrValidation, errorMessages, "channel category must not be empty")
	}
	return nil
}

// Cloud indicates that updates to the feed can be monitored using a web service
// that implements the RssCloud application programming interface.
type Cloud struct {
	Domain            string `xml:"domain,attr"`
	Port              int    `xml:"port,attr,omitempty"`
	Path              string `xml:"path,attr"`
	RegisterProcedure string `xml:"registerProcedure,attr"`
	Protocol          string `xml:"protocol,attr"`
}

// Validate checks if the cloud is valid.
func (c *Cloud) Validate() error {
	if c.Domain == "" {
		return errors.New(ErrValidation, errorMessages, "cloud domain must not be empty")
	}
	if c.Path == "" || c.Path[0] != '/' {
		return errors.New(ErrValidation, errorMessages, "cloud path %q must not be empty and has to start with a slash", c.Path)
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New(ErrValidation, errorMessages, "cloud port %d is out of range", c.Port)
	}
	return nil
}

// Image supplies a graphical logo for the feed .
type Image struct {
	Link        string `xml:"link"`
	Title       string `xml:"title"`
	URL         string `xml:"url"`
	Description string `xml:"description,omitempty"`
	Height      int    `xml:"height,omitempty"`
	Width       int    `xml:"width,omitempty"`
}

// Validate checks if the image is valid.
func (i *Image) Validate() error {
	if _, err := url.Parse(i.Link); err != nil {
		return errors.New(ErrValidation, errorMessages, "image link is not parsable", i.Link)
	}
	if i.Title == "" {
		return errors.New(ErrValidation, errorMessages, "image title must not be empty")
	}
	if _, err := url.Parse(i.URL); err != nil {
		return errors.Annotate(err, ErrParsing, errorMessages, "image title")
	}
	if i.Height < 1 || i.Height > 400 {
		return errors.New(ErrValidation, errorMessages, "image height %d is out of range from 1 to 400", i.Height)
	}
	if i.Width < 1 || i.Width > 144 {
		return errors.New(ErrValidation, errorMessages, "image width %d is out of range from 1 to 144", i.Width)
	}
	return nil
}

// SkipDays identifies days of the week during which the feed is not updated.
type SkipDays struct {
	Days []string `xml:"day"`
}

// Validate checks if the skip days are valid.
func (s *SkipDays) Validate() error {
	skipDays := map[string]bool{
		"Monday":    true,
		"Tuesday":   true,
		"Wednesday": true,
		"Thursday":  true,
		"Friday":    true,
		"Saturday":  true,
		"Sunday":    true,
	}
	for _, day := range s.Days {
		if !skipDays[day] {
			return errors.New(ErrValidation, errorMessages, "skip day %q is invalid", day)
		}
	}
	return nil
}

// SkipHours identifies the hours of the day during which the feed is not updated.
type SkipHours struct {
	Hours []int `xml:"hour"`
}

// Validate checks if the skip hours are valid.
func (s *SkipHours) Validate() error {
	for _, hour := range s.Hours {
		if hour < 0 || hour > 23 {
			return errors.New(ErrValidation, errorMessages, "skip hour %d is out of range from 0 to 23", hour)
		}
	}
	return nil
}

// TextInput defines a form to submit a text query to the feed's publisher over
// the Common Gateway Interface (CGI).
type TextInput struct {
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Name        string `xml:"name"`
	Title       string `xml:"title"`
}

// Validate checks if the text input is valid.
func (t *TextInput) Validate() error {
	if t.Description == "" {
		return errors.New(ErrValidation, errorMessages, "text input description must not be empty")
	}
	if _, err := url.Parse(t.Link); err != nil {
		return errors.Annotate(err, ErrParsing, errorMessages, "text input link")
	}
	if t.Name == "" {
		return errors.New(ErrValidation, errorMessages, "text input name must not be empty")
	}
	if t.Title == "" {
		return errors.New(ErrValidation, errorMessages, "text input title must not be empty")
	}
	return nil
}

// Item represents distinct content published in the feed such as a news article,
// weblog entry or some other form of discrete update. It must contain either a
// title or description.
type Item struct {
	Title       string      `xml:"title,omitempty"`
	Description string      `xml:"description,omitempty"`
	Author      string      `xml:"author,omitempty"`
	Categories  []*Category `xml:"category,omitempty"`
	Comments    string      `xml:"comments,omitempty"`
	Enclosure   *Enclosure  `xml:"enclosure,omitempty"`
	GUID        *GUID       `xml:"guid,omitempty"`
	Link        string      `xml:"link,omitempty"`
	PubDate     string      `xml:"pubDate,omitempty"`
	Source      *Source     `xml:"source,omitempty"`
}

// Validate checks if the item is valid.
func (i *Item) Validate() error {
	if i.Title == "" {
		if i.Description == "" {
			return errors.New(ErrValidation, errorMessages, "item title or description must not be empty")
		}
	}
	if i.Comments != "" {
		if _, err := url.Parse(i.Comments); err != nil {
			return errors.Annotate(err, ErrParsing, errorMessages, "item comments")
		}
	}
	if i.Enclosure != nil {
		if err := i.Enclosure.Validate(); err != nil {
			return err
		}
	}
	if i.GUID != nil {
		if err := i.GUID.Validate(); err != nil {
			return err
		}
	}
	if i.Link != "" {
		if _, err := url.Parse(i.Link); err != nil {
			return errors.Annotate(err, ErrParsing, errorMessages, "item link")
		}
	}
	if i.PubDate != "" {
		if _, err := ParseTime(i.PubDate); err != nil {
			return errors.Annotate(err, ErrParsing, errorMessages, "item publcation date")
		}
	}
	if i.Source != nil {
		if err := i.Source.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Enclosure associates a media object such as an audio or video file with the item.
type Enclosure struct {
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
	URL    string `xml:"url,attr"`
}

// Validate checks if the enclosure is valid.
func (e *Enclosure) Validate() error {
	if e.Length < 1 {
		return errors.New(ErrValidation, errorMessages, "item enclosure length %d is too small", e.Length)
	}
	if e.Type == "" {
		return errors.New(ErrValidation, errorMessages, "item enclosure type must not be empty")
	}
	if _, err := url.Parse(e.URL); err != nil {
		return errors.Annotate(err, ErrParsing, errorMessages, "item enclosure url")
	}
	return nil
}

// GUID provides a string that uniquely identifies the item.
type GUID struct {
	GUID        string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr,omitempty"`
}

// Validate checks if the GUID is valid.
func (g *GUID) Validate() error {
	if g.IsPermaLink {
		if _, err := url.Parse(g.GUID); err != nil {
			return errors.Annotate(err, ErrParsing, errorMessages, "item GUID")
		}
	}
	return nil
}

// Source indicates the fact that the item has been republished from another RSS feed.
type Source struct {
	Source string `xml:",chardata"`
	URL    string `xml:"url,attr"`
}

// Validate checks if the source is valid.
func (s *Source) Validate() error {
	if s.Source == "" {
		return errors.New(ErrValidation, errorMessages, "item source must not be empty")
	}
	if _, err := url.Parse(s.URL); err != nil {
		return errors.Annotate(err, ErrParsing, errorMessages, "item source URL")
	}
	return nil
}

//--------------------
// FUNCTIONS
//--------------------

// ParseTime analyzes the RSS date/time string and returns it as Go time.
func ParseTime(s string) (t time.Time, err error) {
	formats := []string{rssDate, rssDateV1, rssDateV2, rssDateV3, rssDateV4, time.RFC822, time.RFC822Z}
	for _, format := range formats {
		t, err = time.Parse(format, s)
		if err == nil {
			return
		}
	}
	return
}

// ComposeTime takes a Go time and converts it into a valid RSS time string.
func ComposeTime(t time.Time) string {
	return t.Format(rssDate)
}

// Encode writes the RSS document to the writer.
func Encode(w io.Writer, rss *RSS) error {
	enc := xml.NewEncoder(w)
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	return enc.Encode(rss)
}

// Decode reads the RSS document from the reader.
func Decode(r io.Reader) (*RSS, error) {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = utils.CharsetReader
	rss := &RSS{}
	if err := dec.Decode(rss); err != nil {
		return nil, err
	}
	return rss, nil
}

// Get retrieves an RSS document from the given URL.
func Get(u *url.URL) (*RSS, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return Decode(resp.Body)
}

// EOF
