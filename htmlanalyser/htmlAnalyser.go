package htmlanalyser

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// LinksCount holds the frequency of occurences of both external and internal links in a
// given html document
type LinksCount struct {
	Internal, External int
}

// Add increments the corresponding internal frequency count
func (lc *LinksCount) Add(s string) {
	switch s {
	case "internal":
		lc.Internal++
	case "external":
		lc.External++
	}
}

// HeadingsCount holds the frequency of occurences of each heading tag by level in a given
// html document
type HeadingsCount struct {
	H1, H2, H3, H4, H5, H6 int
}

// Add accepts a heading tag name and increments the corresponding internal frequency count
func (hc *HeadingsCount) Add(h string) {
	switch h {
	case "h1":
		hc.H1++
	case "h2":
		hc.H2++
	case "h3":
		hc.H3++
	case "h4":
		hc.H4++
	case "h5":
		hc.H5++
	case "h6":
		hc.H6++
	}
}

// HTMLPageAnalyser encapsulates all the required by the client
type HTMLPageAnalyser struct {
	logger            *log.Logger
	URL               string `json:"url"`
	document          string
	HTMLVersion       string         `json:"htmlVersion"`
	PageTitle         string         `json:"pageTitle"`
	HeadingsByLevel   *HeadingsCount `json:"headings"`
	LinksByType       *LinksCount    `json:"linksByType"`
	InaccessibleLinks int            `json:"inaccessibleLinks"`
	LoginForm         bool           `json:"loginForm"`
}

// New returns the address of a fresh instance of HTMLPageAnalyser
func New(doc, url string, logger *log.Logger) *HTMLPageAnalyser {
	return &HTMLPageAnalyser{
		logger:   logger,
		URL:      url,
		document: doc,
	}
}

// Analyse runs the full suite of analysis functions over the given html document
func (hpa *HTMLPageAnalyser) Analyse() error {
	// We make one initial pass over the document to check if it is parseable.
	// If this check fails we know that each of the various analysis methods will also fail,
	// so we should fail fast and let the caller know we received unparseable HTML.
	// if this first pass succeeds we continue with running the analysis.
	// This has the added benefit of simplifying the analysis functions code
	// since we can avoid defensive programming, safe in the knowledge that parsing won't fail
	hpa.logger.Print("Running test pass on html document")
	z := html.NewTokenizer(strings.NewReader(hpa.document))
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			// case when we encounter a TokenType of Error
			err := z.Err()
			if err == io.EOF {
				break
			}
			hpa.logger.Print("Test pass on html document FAILED")
			return err
		}
	}
	hpa.logger.Print("Test pass on html document SUCCEEDED")

	// test pass has succeeded, so now we run the full suite of analysis
	hpa.HTMLVersion = hpa.getHTMLDocType()
	hpa.PageTitle = hpa.getPageTitle()
	hpa.HeadingsByLevel = hpa.getHeadingsCountByLevel()
	hpa.LinksByType = hpa.getLinksCount()
	hpa.InaccessibleLinks = hpa.getCountOfInaccessibleLinks()
	hpa.LoginForm = hpa.hasLoginForm()

	return nil
}

// DetermineHTMLDocType searches for a doctype tag and parses out the version
func (hpa *HTMLPageAnalyser) getHTMLDocType() string {
	z := html.NewTokenizer(strings.NewReader(hpa.document))
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			// we know this error will be io.EOF due to success of test pass
			return "Document contains no doctype element"
		} else if tokenType == html.DoctypeToken {
			// case when we encounter a DocType Token Type
			docTypeToken := z.Token()
			docTypeTokenValue := strings.ToLower(docTypeToken.Data)

			switch {
			case docTypeTokenValue == "html":
				return "HTML 5.0"
			case strings.Contains(docTypeTokenValue, "-//w3c//dtd html"):
				// we have a legacy doc type < 5
				i := strings.Index(docTypeTokenValue, "-//w3c//dtd html")
				start := i + len("-//w3c//dtd html ")
				var stop int
				if _, err := strconv.Atoi(string(docTypeTokenValue[start+3])); err != nil {
					stop = start + 3
				} else {
					stop = start + 4
				}
				return "HTML " + docTypeTokenValue[start:stop]
			case strings.Contains(docTypeTokenValue, "-//w3c//dtd xhtml"):
				// we have an xhtml doc type
				i := strings.Index(docTypeTokenValue, "-//w3c//dtd xhtml")
				start := i + len("-//w3c//dtd xhtml ")
				var stop int
				if _, err := strconv.Atoi(string(docTypeTokenValue[start+3])); err != nil {
					stop = start + 3
				} else {
					stop = start + 4
				}
				return "XHTML " + docTypeTokenValue[start:stop]
			}
		}
	}
}

// getPageTitle searches for and returns the value of the first title element it finds
func (hpa *HTMLPageAnalyser) getPageTitle() string {
	z := html.NewTokenizer(strings.NewReader(hpa.document))
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			return "Document contains no title element"
		} else if tokenType == html.StartTagToken {
			Token := z.Token()
			if strings.HasPrefix(Token.Data, "title") {
				z.Next()
				return z.Token().Data
			}
		}
	}
}

// getHeadingsCountByLevel returns a HeadingsCount representing the frequency of occurences of each heading
// tag by level found in the given html document
func (hpa *HTMLPageAnalyser) getHeadingsCountByLevel() *HeadingsCount {
	var headingsCount HeadingsCount
	z := html.NewTokenizer(strings.NewReader(hpa.document))
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			break
		} else if tokenType == html.StartTagToken {
			if token := z.Token(); len(token.Data) > 1 {
				tag := token.Data[:2]
				headingsCount.Add(tag)
			}
		}
	}
	return &headingsCount
}

// getLinksCount returns aHeadingsCount containing the frequency of occurences internal/external
// hyperlinks found in the given html document
func (hpa *HTMLPageAnalyser) getLinksCount() *LinksCount {
	var LinksCount LinksCount
	z := html.NewTokenizer(strings.NewReader(hpa.document))
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			break
		} else if tokenType == html.StartTagToken {
			token := z.Token()
			if token.Data == "a" {
				hrefVal := ""
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						hrefVal = attr.Val
						break
					}
				}
				if hrefVal == "" {
					continue
				}
				if hpa.isLinkExternal(hrefVal) {
					LinksCount.Add("external")
				} else {
					LinksCount.Add("internal")
				}
			}
		}
	}
	return &LinksCount
}

// isLinkExternal naively determines whether a link is to an external domain
// by checking an href value to see if it references a local path
func (hpa *HTMLPageAnalyser) isLinkExternal(href string) bool {
	if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "#") || strings.HasPrefix(href, hpa.URL) {
		return false
	}
	return true
}

// hasLoginForm makes a best effort attempt to determine whether the given html document
// contains a login form. We naively assume that a page contains a login form if it is found
// to contain input elements of type 'password' and 'submit'.
func (hpa *HTMLPageAnalyser) hasLoginForm() bool {
	z := html.NewTokenizer(strings.NewReader(hpa.document))
	hasSubmit := false
	hasPassword := false
	for {
		if hasSubmit && hasPassword {
			return true
		}
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			break
		} else if tokenType == html.StartTagToken {
			token := z.Token()
			if strings.HasPrefix(token.Data, "input") {
				typeVal := ""
				for _, attr := range token.Attr {
					if attr.Key == "type" {
						typeVal = attr.Val
						break
					}
				}
				if typeVal == "submit" {
					hasSubmit = true
				} else if typeVal == "password" {
					hasPassword = true
				}
			}
		}
	}
	return false
}

func (hpa *HTMLPageAnalyser) getCountOfInaccessibleLinks() int {
	z := html.NewTokenizer(strings.NewReader(hpa.document))
	inaccessibleLinks := 0
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			break
		} else if tokenType == html.StartTagToken {
			token := z.Token()
			if token.Data == "a" {
				hrefVal := ""
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						hrefVal = attr.Val
						break
					}
				}
				if hrefVal == "" || strings.HasPrefix(hrefVal, "#") {
					// we know hrefs that begin with # are for the current page which must be accessible
					continue
				}
				var url string
				if hpa.isLinkExternal(hrefVal) {
					url = hrefVal
				} else {
					url = hpa.URL + hrefVal
				}
				resp, err := http.Get(url)
				if err != nil {
					hpa.logger.Printf("Whilst testing accessibility of links, GET request to: %s failed", url)
					continue
				}
				if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
					inaccessibleLinks++
				}
			}
		}
	}
	return inaccessibleLinks
}
