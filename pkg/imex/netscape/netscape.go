// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package netscape implements Netscape Bookmark File.
package netscape

import (
	"bufio"
	"fmt"
	sthtml "html"
	"io"
	"strings"
	"time"

	"golang.org/x/net/html"

	"git.sr.ht/~bouncepaw/betula/pkg/ticks"
)

type (
	Bookmark struct {
		URL      string
		Title    string
		Tags     []string
		Added    time.Time
		Modified time.Time
	}
	Folder struct {
		Title         string
		Added         time.Time
		Modified      time.Time
		IsReadingList bool
		Items         []Item
	}
	Item interface {
		isItem()
	}
)

func (Bookmark) isItem() {}

func (*Folder) isItem() {}

// Probe reports whether r contains a Netscape Bookmark File by checking its
// DOCTYPE declaration. It seeks r back to the start before returning, so the
// caller can pass the same reader to Read.
func Probe(r io.ReadSeeker) (bool, error) {
	const doctype = "<!DOCTYPE NETSCAPE-Bookmark-file-1>"
	buf := make([]byte, 64)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	return strings.HasPrefix(strings.TrimSpace(string(buf[:n])), doctype), nil
}

// Read parses a Netscape Bookmark file from r and returns the root folder.
// The root folder's Title is taken from the <H1> element.
func Read(r io.Reader) (*Folder, error) {
	type parseState int
	const (
		stateNormal parseState = iota
		stateRootTitle
		stateFolderTitle
		stateBookmarkTitle
	)

	var (
		root          = &Folder{}
		stack         = []*Folder{root}
		pendingFolder *Folder
		curBookmark   *Bookmark
		textBuf       strings.Builder
		state         = stateNormal
		tokenizer     = html.NewTokenizer(r)
	)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return root, nil
			}
			return nil, tokenizer.Err()

		case html.StartTagToken:
			tok := tokenizer.Token()
			tag := strings.ToLower(tok.Data)
			switch tag {
			case "h1":
				state = stateRootTitle
				textBuf.Reset()
			case "h3":
				pendingFolder = &Folder{}
				for _, a := range tok.Attr {
					switch strings.ToLower(a.Key) {
					case "add_date":
						pendingFolder.Added = ticks.ParseUnixTimestamp(a.Val)
					case "last_modified":
						pendingFolder.Modified = ticks.ParseUnixTimestamp(a.Val)
					case "id":
						pendingFolder.IsReadingList = a.Val == "com.apple.ReadingList"
					}
				}
				state = stateFolderTitle
				textBuf.Reset()
			case "a":
				curBookmark = &Bookmark{}
				for _, a := range tok.Attr {
					switch strings.ToLower(a.Key) {
					case "href":
						curBookmark.URL = a.Val
					case "add_date":
						curBookmark.Added = ticks.ParseUnixTimestamp(a.Val)
					case "last_modified":
						curBookmark.Modified = ticks.ParseUnixTimestamp(a.Val)
					case "tags":
						for t := range strings.SplitSeq(a.Val, ",") {
							t = strings.TrimSpace(t)
							if t != "" {
								curBookmark.Tags = append(curBookmark.Tags, t)
							}
						}
					}
				}
				state = stateBookmarkTitle
				textBuf.Reset()
			case "dl":
				if pendingFolder != nil {
					top := stack[len(stack)-1]
					top.Items = append(top.Items, pendingFolder)
					stack = append(stack, pendingFolder)
					pendingFolder = nil
				}
			}

		case html.SelfClosingTagToken, html.CommentToken, html.DoctypeToken:
			// nothing to do

		case html.TextToken:
			if state != stateNormal {
				textBuf.Write(tokenizer.Text())
			}

		case html.EndTagToken:
			tok := tokenizer.Token()
			switch strings.ToLower(tok.Data) {
			case "h1":
				root.Title = strings.TrimSpace(textBuf.String())
				state = stateNormal
			case "h3":
				if pendingFolder != nil {
					pendingFolder.Title = strings.TrimSpace(textBuf.String())
				}
				state = stateNormal
			case "a":
				if curBookmark != nil {
					curBookmark.Title = strings.TrimSpace(textBuf.String())
					top := stack[len(stack)-1]
					top.Items = append(top.Items, *curBookmark)
					curBookmark = nil
				}
				state = stateNormal
			case "dl":
				if len(stack) > 1 {
					stack = stack[:len(stack)-1]
				}
			}
		}
	}
}

// Write writes the folder as a Netscape Bookmark file to w.
func (f *Folder) Write(w io.Writer) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>%s</H1>

<DL><p>
`, sthtml.EscapeString(f.Title))
	f.writeItems(bw, 1)
	fmt.Fprint(bw, "</DL>")
	return bw.Flush()
}

func (b Bookmark) writeItem(w *bufio.Writer, indent string) {
	tagsAttr := ""
	if len(b.Tags) > 0 {
		tagsAttr = fmt.Sprintf(` TAGS="%s"`, sthtml.EscapeString(strings.Join(b.Tags, ",")))
	}
	fmt.Fprintf(
		w,
		`%s<DT><A HREF="%s" ADD_DATE="%s" LAST_MODIFIED="%s"%s>%s</A>
`,
		indent,
		sthtml.EscapeString(b.URL),
		ticks.FormatUnixTimestamp(b.Added),
		ticks.FormatUnixTimestamp(b.Modified),
		tagsAttr,
		sthtml.EscapeString(b.Title),
	)
}

func (f *Folder) writeItems(w *bufio.Writer, depth int) {
	indent := strings.Repeat("    ", depth)
	for _, item := range f.Items {
		switch v := item.(type) {
		case Bookmark:
			v.writeItem(w, indent)
		case *Folder:
			fmt.Fprintf(w,
				`%s<DT><H3 ADD_DATE="%s" LAST_MODIFIED="%s">%s</H3>
%s<DL><p>
`,
				indent,
				ticks.FormatUnixTimestamp(v.Added),
				ticks.FormatUnixTimestamp(v.Modified),
				sthtml.EscapeString(v.Title),
				indent)
			v.writeItems(w, depth+1)
			fmt.Fprintf(w, "%s</DL><p>\n", indent)
		}
	}
}
