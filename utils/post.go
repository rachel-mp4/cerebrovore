package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"net/http"

	"github.com/rachel-mp4/cerebrovore/clog"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var hashtagRE = regexp.MustCompile(`#([0-9A-Za-z]+)`)

func ParseBodyForBacklinks(s string) (backlinks []uint32, extras []uint64) {
	matches := hashtagRE.FindAllStringSubmatch(s, -1)
	blmap := make(map[string]bool)
	backlinks = make([]uint32, 0)
	extras = make([]uint64, 0)
	for _, m := range matches {
		b := m[1]
		added := blmap[b]
		if added {
			continue
		}
		blmap[b] = true
		bl, blerr, ex, exerr := AToEx(m[1])
		if blerr == nil {
			backlinks = append(backlinks, bl)
			continue
		}
		if exerr == nil {
			extras = append(extras, ex)
			continue
		}
	}
	return
}

var mentionRE = regexp.MustCompile(`@([0-9a-z]+)`)

func ParseBodyForMentions(s string) []string {
	matches := mentionRE.FindAllStringSubmatch(s, -1)
	menmap := make(map[string]bool)
	res := make([]string, 0)
	for _, m := range matches {
		men := m[1]
		added := menmap[men]
		if added {
			continue
		}
		menmap[men] = true
		res = append(res, m[1])
	}
	return res
}

type PlaySite int

const (
	Youtube = iota
	Soundcloud
)

func init() {
	if Youtube != 0 {
		panic("don't change me!")
	}
}

type PlayInput struct {
	Site     PlaySite      `json:"site"`
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Duration time.Duration `json:"duration"`
	Width    *int          `json:"width,omitempty"`
	Height   *int          `json:"height,omitempty"`
}

// ParseBodyForPlays finds each instance of #play {play_url} and then
// according to the play_url, makes appropriate api calls to produce
// a slice of PlayInput that we can inform the client of. i think this
// is ok, but depending on the desired ui, we can put off at least
// youtube api calls until the next PlayInput that hasn't looked up its
// until the next PlayInput that hasn't looked up its data is up next
// in queue. if we do it this way, i think we can cut down api calls
// a lot, however, ux is a bit
func ParseBodyForPlays(s string) (res []*PlayInput, unpause bool) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	res = make([]*PlayInput, 0)
	for scanner.Scan() {
		l := scanner.Text()
		literal, found := strings.CutPrefix(l, "#play ")
		if found {
			playurl, err := url.Parse(literal)
			if err != nil {
				clog.Warn("post parse: %s", err)
				continue
			}
			switch playurl.Host {
			case "youtube.com", "www.youtube.com", "m.youtube.com":
				id := playurl.Query().Get("v")
				pi, err := getDurationForYoutubeId(id)
				if err != nil {
					clog.Warn("post parse: %s", err)
					continue
				}
				res = append(res, pi)
			case "youtu.be":
				id := strings.TrimPrefix(playurl.Path, "/")
				pi, err := getDurationForYoutubeId(id)
				if err != nil {
					clog.Warn("post parse: %s", err)
					continue
				}
				res = append(res, pi)
				// case "soundcloud.com", "on.soundcloud.com", "www.soundcloud.com":
			}
		} else if l == "#play" {
			unpause = true
		}
	}
	return
}

func getDurationForYoutubeId(id string) (*PlayInput, error) {
	type YTResp struct {
		Items []struct {
			Snippet struct {
				Title string `json:"title"`
				// Description  string `json:"description"` // maybe include these if we want to intuit
				// ChannelTitle string `json:"channelTitle"` // if it's a square auto-generated song
				LiveBroadcastContent string `json:"liveBroadcastContent"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
			Status struct {
				PrivacyStatus string `json:"privacyStatus"`
				Embeddable    bool   `json:"embeddable"`
			} `json:"status"`
			Player struct {
				EmbedHeight string `json:"embedHeight"`
				EmbedWidth  string `json:"embedWidth"`
			}
		} `json:"items"`
	}
	//max dimensions are used to get the embed Height and Width, we use 576 x 324 because it's 16:9 and the max width is 3x the margin of the sidebar
	apiurl := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails,status,player&maxWidth=576&maxHeight=324&id=%s&key=%s", id, os.Getenv("YOUTUBE_API_KEY"))
	resp, err := http.DefaultClient.Get(apiurl)
	if err != nil {
		return nil, err
	}
	var ytresp YTResp
	err = json.NewDecoder(resp.Body).Decode(&ytresp)
	if err != nil {
		return nil, err
	}
	if len(ytresp.Items) == 0 {
		return nil, errors.New("no items")
	}
	ti := ytresp.Items[0]
	if ti.Status.PrivacyStatus == "private" || !ti.Status.Embeddable || ti.Snippet.LiveBroadcastContent == "live" {
		return nil, errors.New("not embeddable")
	}
	title := ti.Snippet.Title
	dstring := strings.ToLower(strings.TrimPrefix(ti.ContentDetails.Duration, "PT"))
	duration, err := time.ParseDuration(dstring)
	if err != nil {
		return nil, err
	}
	wstring := ti.Player.EmbedWidth
	hstring := ti.Player.EmbedHeight
	var width *int
	var height *int
	wnum, err := strconv.Atoi(wstring)
	if err == nil {
		width = &wnum
	}
	hnum, err := strconv.Atoi(hstring)
	if err == nil {
		height = &hnum
	}
	return &PlayInput{
		Site:     Youtube,
		ID:       id,
		Title:    title,
		Duration: duration,
		Width:    width,
		Height:   height,
	}, nil
}

func RenderTextBody(s string) template.HTML {
	var out strings.Builder
	last := 0
	matches := hashtagRE.FindAllStringSubmatchIndex(s, -1)
	for _, m := range matches {
		start, end := m[0], m[1]
		capStart, capEnd := m[2], m[3]
		out.WriteString(ExpandUrls(s[last:start]))
		capture := s[capStart:capEnd]
		out.WriteString(`<a href="/p/`)
		out.WriteString(capture)
		out.WriteString(`">#`)
		out.WriteString(capture)
		out.WriteString(`</a>`)
		last = end
	}

	out.WriteString(ExpandUrls(s[last:]))
	return template.HTML(out.String())
}

var urlRE = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

func ExpandUrls(s string) string {
	var out strings.Builder
	last := 0
	matches := urlRE.FindAllStringIndex(s, -1)
	for _, m := range matches {
		start, end := m[0], m[1]
		out.WriteString(html.EscapeString(s[last:start]))
		tryurl := s[start:end]
		yesurl, err := url.Parse(tryurl)
		if err != nil {
			clog.Warn("url parse error")
			out.WriteString(html.EscapeString(tryurl))
			last = end
			continue
		}
		q := yesurl.Query()
		q.Del("si")
		yesurl.RawQuery = q.Encode()
		yesurlstr := yesurl.String()
		out.WriteString(`<a href="`)
		out.WriteString(yesurlstr)
		out.WriteString(`" target="_blank" rel="noopener noreferrer">`)
		out.WriteString(yesurlstr)
		out.WriteString(`</a>`)
		last = end
	}
	out.WriteString(html.EscapeString(s[last:]))
	return out.String()
}

// for some reason i felt like this was stupid a few days ago...
// btw the reason why images are duplicated is because we have one
// with z-index negative something thats full opacity and the other
// with z-index positive thats like half opacity, the result is that
// it looks normal, but this way we can wedge vfx between the two that
// only half obscure it. maybe there's a better way with svg filter
// and blend modes, but this is dead simple at the cost of more verbose
// html
func RenderImageBody(cid string, alt *string) template.HTML {
	var out strings.Builder
	isgif := strings.HasSuffix(cid, ".gif")
	out.WriteString(`<div class="image-wrapper thumb"`)
	if !isgif {
		out.WriteString(` data-thumb="/blob?cid=`)
		out.WriteString(html.EscapeString(cid))
		out.WriteString(`&thumb=yes" data-full="/blob?cid=`)
		out.WriteString(html.EscapeString(cid))
		out.WriteString(`"`)
	}
	out.WriteString(`><img class="bg-img" src="/blob?cid=`)
	out.WriteString(html.EscapeString(cid))
	if !isgif {
		out.WriteString(`&thumb=yes`)
	}
	if alt != nil {
		out.WriteString(`" alt="`)
		out.WriteString(html.EscapeString(*alt))
	}
	out.WriteString(`" /><img class="fg-img" src="/blob?cid=`)
	out.WriteString(html.EscapeString(cid))
	if !isgif {
		out.WriteString(`&thumb=yes`)
	}
	if alt != nil {
		out.WriteString(`" alt="`)
		out.WriteString(html.EscapeString(*alt))
	}
	out.WriteString(`" /></div>`)
	return template.HTML(out.String())
}

func RenderTextBodyNew(s string) template.HTML {
	rdr := Render(Parse(s))
	return template.HTML(rdr)
}

type line struct {
	isRightQuote bool
	isUpQuote    bool
	isLeftQuote  bool
	isDownQuote  bool
	tokens       []token
}

type token struct {
	t       state
	text    string
	hashtag string
	mention string
	link    *url.URL
	linkstr string
}

type state = int

const (
	hashtag state = iota
	mention
	normal
	link
	italic
	bold
	code
)

func Render(ll []line) string {
	out := strings.Builder{}
	isbold := false
	isitalic := false
	iscode := false
	for i, l := range ll {
		if i != 0 {
			out.WriteRune('\n')
		}
		if l.isUpQuote {
			out.WriteString(`<span class="up quote">`)
		}
		if l.isLeftQuote {
			out.WriteString(`<span class="left quote">`)
		}
		if l.isRightQuote {
			out.WriteString(`<span class="right quote">`)
		}
		if l.isDownQuote {
			out.WriteString(`<span class="down quote">`)
		}
		out.WriteString(ibcstart(isitalic, isbold, iscode))
		for _, t := range l.tokens {
			switch t.t {
			case hashtag:
				out.WriteString(ibcend(isitalic, isbold, iscode))
				out.WriteString(`<a href="/p/`)
				out.WriteString(t.hashtag)
				out.WriteString(`">`)
				out.WriteString(ibcstart(isitalic, isbold, iscode))
				out.WriteRune('#')
				out.WriteString(t.hashtag)
				out.WriteString(ibcend(isitalic, isbold, iscode))
				out.WriteString(`</a>`)
				out.WriteString(ibcstart(isitalic, isbold, iscode))
			case mention:
				out.WriteString(ibcend(isitalic, isbold, iscode))
				out.WriteString(`<a href="/profile/`)
				out.WriteString(t.mention)
				out.WriteString(`">`)
				out.WriteString(ibcstart(isitalic, isbold, iscode))
				out.WriteRune('@')
				out.WriteString(t.mention)
				out.WriteString(ibcend(isitalic, isbold, iscode))
				out.WriteString(`</a>`)
				out.WriteString(ibcstart(isitalic, isbold, iscode))
			case link:
				out.WriteString(ibcend(isitalic, isbold, iscode))
				out.WriteString(`<a target="_blank" ref="noopener noreferrer" href="`)
				out.WriteString(t.link.String())
				out.WriteString(`">`)
				out.WriteString(ibcstart(isitalic, isbold, iscode))
				out.WriteString(t.linkstr)
				out.WriteString(ibcend(isitalic, isbold, iscode))
				out.WriteString(`</a>`)
				out.WriteString(ibcstart(isitalic, isbold, iscode))
			case normal:
				out.WriteString(html.EscapeString(t.text))
			case italic:
				if isitalic {
					out.WriteString(`*</em>`)
				} else {
					out.WriteString(`<em>*`)
				}
				isitalic = !isitalic
			case bold:
				if isitalic {
					out.WriteString(`</em>`)
				}
				if isbold {
					out.WriteString(`**</b>`)
				} else {
					out.WriteString(`<b>**`)
				}
				if isitalic {
					out.WriteString(`<em>`)
				}
				isbold = !isbold
			case code:
				if isitalic {
					out.WriteString(`</em>`)
				}
				if isbold {
					out.WriteString(`</b>`)
				}
				if iscode {
					out.WriteString("`</code>")
				} else {
					out.WriteString("<code>`")
				}
				if isbold {
					out.WriteString(`<b>`)
				}
				if isitalic {
					out.WriteString(`<em>`)
				}
				iscode = !iscode
			}
		}
		out.WriteString(ibcend(isitalic, isbold, iscode))
		if l.isLeftQuote || l.isRightQuote || l.isUpQuote || l.isDownQuote {
			out.WriteString(`</span>`)
		}
	}
	return out.String()
}

// ibcstart writes the starting tags for italic, bold, and code
// the basic idea is that italic, bold and code should be three
// different independent toggles that you can turn on and off
// which means that we need to turn them off whenever we start
// a link, and then turn them back on again
func ibcstart(isitalic, isbold, iscode bool) string {
	out := strings.Builder{}
	if iscode {
		out.WriteString(`<code>`)
	}
	if isbold {
		out.WriteString(`<b>`)
	}
	if isitalic {
		out.WriteString(`<em>`)
	}
	return out.String()
}

// ibcend writes the ending tags for italic, bold, and code
func ibcend(isitalic, isbold, iscode bool) string {
	out := strings.Builder{}
	if isitalic {
		out.WriteString(`</em>`)
	}
	if isbold {
		out.WriteString(`</b>`)
	}
	if iscode {
		out.WriteString(`</code>`)
	}
	return out.String()
}

func Parse(str string) []line {
	s := bufio.NewScanner(strings.NewReader(str))
	var res []line

	bigword := make([]rune, 0)
	word := make([]rune, 0)
lineloop:
	for s.Scan() {
		l := s.Text()
		p := line{}
		if l == "" {
			res = append(res, p)
			continue
		}
		switch l[0] {
		case '>':
			p.isRightQuote = true
		case '<':
			p.isLeftQuote = true
		case '^':
			p.isUpQuote = true
		case 'v', 'V':
			p.isDownQuote = true
		}
		curstate := normal
		rr := []rune(l)
		bigword = bigword[:0]
		word = word[:0]

		skipme := false
		for idx, char := range rr {
			if skipme {
				skipme = false
				continue
			} else {
				if char == '*' || char == '`' {
					switch curstate {
					case hashtag:
						p.tokens = append(p.tokens, token{t: hashtag, hashtag: string(word)})
						word = word[:0]
					case mention:
						p.tokens = append(p.tokens, token{t: mention, mention: string(word)})
						word = word[:0]
					case normal:
						if len(word) != 0 {
							myword := string(word)
							u, err := validURL(myword)
							if err != nil {
								bigword = append(bigword, word...)
							} else {
								p.tokens = append(p.tokens, token{t: normal, text: string(bigword)}, token{t: link, linkstr: clean(word), link: u})
								bigword = bigword[:0]
							}
							word = word[:0]
						}
						if len(bigword) != 0 {
							p.tokens = append(p.tokens, token{t: normal, text: string(bigword)})
							bigword = bigword[:0]
						}
					}
					curstate = normal
					if char == '*' {
						if idx+1 != len(rr) && rr[idx+1] == '*' {
							p.tokens = append(p.tokens, token{t: bold})
							skipme = true
						} else {
							p.tokens = append(p.tokens, token{t: italic})
						}
					} else {
						p.tokens = append(p.tokens, token{t: code})
					}
					if idx+1 == len(rr) {
						res = append(res, p)
					}
					continue
				}
			}
			switch curstate {
			case hashtag:
				if isAlphanumeric(char) {
					word = append(word, char)
					continue
				}
				p.tokens = append(p.tokens, token{t: hashtag, hashtag: string(word)})
				if idx+1 == len(rr) {
					p.tokens = append(p.tokens, token{t: normal, text: string([]rune{char})})
					res = append(res, p)
					continue lineloop
				}
				word = word[:0]
				switch char {
				case '#':
					if isAlphanumericNonzero(rr[idx+1]) {
						curstate = hashtag // state stays the same
					} else {
						curstate = normal
						bigword = append(bigword, char)
					}
				case '@':
					if isAlphanumeric(rr[idx+1]) {
						curstate = mention
					} else {
						curstate = normal
						bigword = append(bigword, char)
					}
				default:
					curstate = normal
					if isOkInUrl(char) {
						word = append(word, char)
					} else {
						bigword = append(bigword, char)
					}
				}

			case mention:
				if isAlphanumeric(char) {
					word = append(word, char)
					continue
				}
				p.tokens = append(p.tokens, token{t: mention, mention: string(word)})
				if idx+1 == len(rr) {
					p.tokens = append(p.tokens, token{t: normal, text: string([]rune{char})})
					res = append(res, p)
					continue lineloop
				}
				word = word[:0]
				switch char {
				case '#':
					if isAlphanumericNonzero(rr[idx+1]) {
						curstate = hashtag
					} else {
						curstate = normal
						bigword = append(bigword, char)
					}
				case '@':
					if isAlphanumeric(rr[idx+1]) {
						curstate = mention
					} else {
						curstate = normal
						bigword = append(bigword, char)
					}
				default:
					curstate = normal
					if isOkInUrl(char) {
						word = append(word, char)
					} else {
						bigword = append(bigword, char)
					}
				}

			case normal:
				if isOkInUrl(char) {
					word = append(word, char)
					continue
				}
				if len(word) != 0 {
					myword := string(word)
					u, err := validURL(myword)
					if err != nil {
						bigword = append(bigword, word...)
					} else {
						p.tokens = append(p.tokens, token{t: normal, text: string(bigword)}, token{t: link, linkstr: myword, link: u})
						bigword = bigword[:0]
					}
					word = word[:0]
				}
				if idx+1 == len(rr) {
					bigword = append(bigword, char)
					p.tokens = append(p.tokens, token{t: normal, text: string(bigword)})
					res = append(res, p)
					continue lineloop
				}
				switch char {
				case '#':
					if isAlphanumericNonzero(rr[idx+1]) {
						p.tokens = append(p.tokens, token{t: normal, text: string(bigword)})
						bigword = bigword[:0]
						curstate = hashtag
					} else {
						bigword = append(bigword, char)
					}
				case '@':
					if isAlphanumeric(rr[idx+1]) {
						p.tokens = append(p.tokens, token{t: normal, text: string(bigword)})
						bigword = bigword[:0]
						curstate = mention
					} else {
						bigword = append(bigword, char)
					}
				default:
					bigword = append(bigword, char)
				}
			}
		}

		switch curstate {
		case hashtag:
			p.tokens = append(p.tokens, token{t: hashtag, hashtag: string(word)})
			res = append(res, p)
		case mention:
			p.tokens = append(p.tokens, token{t: mention, mention: string(word)})
			res = append(res, p)
		case normal:
			if len(word) != 0 {
				myword := string(word)
				u, err := validURL(myword)
				if err != nil {
					bigword = append(bigword, word...)
				} else {
					p.tokens = append(p.tokens, token{t: normal, text: string(bigword)}, token{t: link, linkstr: clean(word), link: u})
					res = append(res, p)
					continue
				}
			}
			if len(bigword) != 0 {
				p.tokens = append(p.tokens, token{t: normal, text: string(bigword)})
				res = append(res, p)
			}
		}
	}
	return res
}

func validURL(s string) (*url.URL, error) {
	if !(strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")) {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	var last string
	hn := u.Hostname()
	for w := range strings.SplitSeq(hn, ".") {
		if w == "" {
			return nil, errors.New("not a valid url")
		}
		last = w
	}
	if last == hn {
		return nil, errors.New("url doesn't contain period")
	}
	if !isTLD(last) {
		return nil, errors.New("url doesn't end with valid tld")
	}
	q := u.Query()
	q.Del("si")
	u.RawQuery = q.Encode()
	return u, nil
}

func clean(rr []rune) string {
	out := strings.Builder{}
	strip := false
	for idx, r := range rr {
		switch r {
		case '?', '&':
			if strip {
				strip = false
				out.WriteRune(r)
				continue
			}
			if idx+4 < len(rr) {
				if rr[idx+1] == 's' && rr[idx+2] == 'i' && rr[idx+3] == '=' {
					strip = true
					continue
				}
			}
			out.WriteRune(r)
		default:
			if strip {
				continue
			}
			out.WriteRune(r)
		}
	}
	return out.String()
}

func isAlphanumeric(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		return true
	default:
		return false
	}
}

func isAlphanumericNonzero(r rune) bool {
	switch r {
	case '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		return true
	default:
		return false
	}
}

func isOkInUrl(r rune) bool {
	if isAlphanumeric(r) {
		return true
	}
	switch r {
	case ':', '/', '?', '[', ']', '$', '-', '_', '.', '+', '!', '\'', '(', ')', ',', '%', ';', '=', '&':
		return true
	default:
		return false
	}
}

var tlds = map[string]struct{}{
	"AAA":                      {},
	"AARP":                     {},
	"ABB":                      {},
	"ABBOTT":                   {},
	"ABBVIE":                   {},
	"ABC":                      {},
	"ABLE":                     {},
	"ABOGADO":                  {},
	"ABUDHABI":                 {},
	"AC":                       {},
	"ACADEMY":                  {},
	"ACCENTURE":                {},
	"ACCOUNTANT":               {},
	"ACCOUNTANTS":              {},
	"ACO":                      {},
	"ACTOR":                    {},
	"AD":                       {},
	"ADS":                      {},
	"ADULT":                    {},
	"AE":                       {},
	"AEG":                      {},
	"AERO":                     {},
	"AETNA":                    {},
	"AF":                       {},
	"AFL":                      {},
	"AFRICA":                   {},
	"AG":                       {},
	"AGAKHAN":                  {},
	"AGENCY":                   {},
	"AI":                       {},
	"AIG":                      {},
	"AIRBUS":                   {},
	"AIRFORCE":                 {},
	"AIRTEL":                   {},
	"AKDN":                     {},
	"AL":                       {},
	"ALIBABA":                  {},
	"ALIPAY":                   {},
	"ALLFINANZ":                {},
	"ALLSTATE":                 {},
	"ALLY":                     {},
	"ALSACE":                   {},
	"ALSTOM":                   {},
	"AM":                       {},
	"AMAZON":                   {},
	"AMERICANEXPRESS":          {},
	"AMERICANFAMILY":           {},
	"AMEX":                     {},
	"AMFAM":                    {},
	"AMICA":                    {},
	"AMSTERDAM":                {},
	"ANALYTICS":                {},
	"ANDROID":                  {},
	"ANQUAN":                   {},
	"ANZ":                      {},
	"AO":                       {},
	"AOL":                      {},
	"APARTMENTS":               {},
	"APP":                      {},
	"APPLE":                    {},
	"AQ":                       {},
	"AQUARELLE":                {},
	"AR":                       {},
	"ARAB":                     {},
	"ARAMCO":                   {},
	"ARCHI":                    {},
	"ARMY":                     {},
	"ARPA":                     {},
	"ART":                      {},
	"ARTE":                     {},
	"AS":                       {},
	"ASDA":                     {},
	"ASIA":                     {},
	"ASSOCIATES":               {},
	"AT":                       {},
	"ATHLETA":                  {},
	"ATTORNEY":                 {},
	"AU":                       {},
	"AUCTION":                  {},
	"AUDI":                     {},
	"AUDIBLE":                  {},
	"AUDIO":                    {},
	"AUSPOST":                  {},
	"AUTHOR":                   {},
	"AUTO":                     {},
	"AUTOS":                    {},
	"AW":                       {},
	"AWS":                      {},
	"AX":                       {},
	"AXA":                      {},
	"AZ":                       {},
	"AZURE":                    {},
	"BA":                       {},
	"BABY":                     {},
	"BAIDU":                    {},
	"BANAMEX":                  {},
	"BAND":                     {},
	"BANK":                     {},
	"BAR":                      {},
	"BARCELONA":                {},
	"BARCLAYCARD":              {},
	"BARCLAYS":                 {},
	"BAREFOOT":                 {},
	"BARGAINS":                 {},
	"BASEBALL":                 {},
	"BASKETBALL":               {},
	"BAUHAUS":                  {},
	"BAYERN":                   {},
	"BB":                       {},
	"BBC":                      {},
	"BBT":                      {},
	"BBVA":                     {},
	"BCG":                      {},
	"BCN":                      {},
	"BD":                       {},
	"BE":                       {},
	"BEATS":                    {},
	"BEAUTY":                   {},
	"BEER":                     {},
	"BERLIN":                   {},
	"BEST":                     {},
	"BESTBUY":                  {},
	"BET":                      {},
	"BF":                       {},
	"BG":                       {},
	"BH":                       {},
	"BHARTI":                   {},
	"BI":                       {},
	"BIBLE":                    {},
	"BID":                      {},
	"BIKE":                     {},
	"BING":                     {},
	"BINGO":                    {},
	"BIO":                      {},
	"BIZ":                      {},
	"BJ":                       {},
	"BLACK":                    {},
	"BLACKFRIDAY":              {},
	"BLOCKBUSTER":              {},
	"BLOG":                     {},
	"BLOOMBERG":                {},
	"BLUE":                     {},
	"BM":                       {},
	"BMS":                      {},
	"BMW":                      {},
	"BN":                       {},
	"BNPPARIBAS":               {},
	"BO":                       {},
	"BOATS":                    {},
	"BOEHRINGER":               {},
	"BOFA":                     {},
	"BOM":                      {},
	"BOND":                     {},
	"BOO":                      {},
	"BOOK":                     {},
	"BOOKING":                  {},
	"BOSCH":                    {},
	"BOSTIK":                   {},
	"BOSTON":                   {},
	"BOT":                      {},
	"BOUTIQUE":                 {},
	"BOX":                      {},
	"BR":                       {},
	"BRADESCO":                 {},
	"BRIDGESTONE":              {},
	"BROADWAY":                 {},
	"BROKER":                   {},
	"BROTHER":                  {},
	"BRUSSELS":                 {},
	"BS":                       {},
	"BT":                       {},
	"BUILD":                    {},
	"BUILDERS":                 {},
	"BUSINESS":                 {},
	"BUY":                      {},
	"BUZZ":                     {},
	"BV":                       {},
	"BW":                       {},
	"BY":                       {},
	"BZ":                       {},
	"BZH":                      {},
	"CA":                       {},
	"CAB":                      {},
	"CAFE":                     {},
	"CAL":                      {},
	"CALL":                     {},
	"CALVINKLEIN":              {},
	"CAM":                      {},
	"CAMERA":                   {},
	"CAMP":                     {},
	"CANON":                    {},
	"CAPETOWN":                 {},
	"CAPITAL":                  {},
	"CAPITALONE":               {},
	"CAR":                      {},
	"CARAVAN":                  {},
	"CARDS":                    {},
	"CARE":                     {},
	"CAREER":                   {},
	"CAREERS":                  {},
	"CARS":                     {},
	"CASA":                     {},
	"CASE":                     {},
	"CASH":                     {},
	"CASINO":                   {},
	"CAT":                      {},
	"CATERING":                 {},
	"CATHOLIC":                 {},
	"CBA":                      {},
	"CBN":                      {},
	"CBRE":                     {},
	"CC":                       {},
	"CD":                       {},
	"CENTER":                   {},
	"CEO":                      {},
	"CERN":                     {},
	"CF":                       {},
	"CFA":                      {},
	"CFD":                      {},
	"CG":                       {},
	"CH":                       {},
	"CHANEL":                   {},
	"CHANNEL":                  {},
	"CHARITY":                  {},
	"CHASE":                    {},
	"CHAT":                     {},
	"CHEAP":                    {},
	"CHINTAI":                  {},
	"CHRISTMAS":                {},
	"CHROME":                   {},
	"CHURCH":                   {},
	"CI":                       {},
	"CIPRIANI":                 {},
	"CIRCLE":                   {},
	"CISCO":                    {},
	"CITADEL":                  {},
	"CITI":                     {},
	"CITIC":                    {},
	"CITY":                     {},
	"CK":                       {},
	"CL":                       {},
	"CLAIMS":                   {},
	"CLEANING":                 {},
	"CLICK":                    {},
	"CLINIC":                   {},
	"CLINIQUE":                 {},
	"CLOTHING":                 {},
	"CLOUD":                    {},
	"CLUB":                     {},
	"CLUBMED":                  {},
	"CM":                       {},
	"CN":                       {},
	"CO":                       {},
	"COACH":                    {},
	"CODES":                    {},
	"COFFEE":                   {},
	"COLLEGE":                  {},
	"COLOGNE":                  {},
	"COM":                      {},
	"COMMBANK":                 {},
	"COMMUNITY":                {},
	"COMPANY":                  {},
	"COMPARE":                  {},
	"COMPUTER":                 {},
	"COMSEC":                   {},
	"CONDOS":                   {},
	"CONSTRUCTION":             {},
	"CONSULTING":               {},
	"CONTACT":                  {},
	"CONTRACTORS":              {},
	"COOKING":                  {},
	"COOL":                     {},
	"COOP":                     {},
	"CORSICA":                  {},
	"COUNTRY":                  {},
	"COUPON":                   {},
	"COUPONS":                  {},
	"COURSES":                  {},
	"CPA":                      {},
	"CR":                       {},
	"CREDIT":                   {},
	"CREDITCARD":               {},
	"CREDITUNION":              {},
	"CRICKET":                  {},
	"CROWN":                    {},
	"CRS":                      {},
	"CRUISE":                   {},
	"CRUISES":                  {},
	"CU":                       {},
	"CUISINELLA":               {},
	"CV":                       {},
	"CW":                       {},
	"CX":                       {},
	"CY":                       {},
	"CYMRU":                    {},
	"CYOU":                     {},
	"CZ":                       {},
	"DAD":                      {},
	"DANCE":                    {},
	"DATA":                     {},
	"DATE":                     {},
	"DATING":                   {},
	"DATSUN":                   {},
	"DAY":                      {},
	"DCLK":                     {},
	"DDS":                      {},
	"DE":                       {},
	"DEAL":                     {},
	"DEALER":                   {},
	"DEALS":                    {},
	"DEGREE":                   {},
	"DELIVERY":                 {},
	"DELL":                     {},
	"DELOITTE":                 {},
	"DELTA":                    {},
	"DEMOCRAT":                 {},
	"DENTAL":                   {},
	"DENTIST":                  {},
	"DESI":                     {},
	"DESIGN":                   {},
	"DEV":                      {},
	"DHL":                      {},
	"DIAMONDS":                 {},
	"DIET":                     {},
	"DIGITAL":                  {},
	"DIRECT":                   {},
	"DIRECTORY":                {},
	"DISCOUNT":                 {},
	"DISCOVER":                 {},
	"DISH":                     {},
	"DIY":                      {},
	"DJ":                       {},
	"DK":                       {},
	"DM":                       {},
	"DNP":                      {},
	"DO":                       {},
	"DOCS":                     {},
	"DOCTOR":                   {},
	"DOG":                      {},
	"DOMAINS":                  {},
	"DOT":                      {},
	"DOWNLOAD":                 {},
	"DRIVE":                    {},
	"DTV":                      {},
	"DUBAI":                    {},
	"DUPONT":                   {},
	"DURBAN":                   {},
	"DVAG":                     {},
	"DVR":                      {},
	"DZ":                       {},
	"EARTH":                    {},
	"EAT":                      {},
	"EC":                       {},
	"ECO":                      {},
	"EDEKA":                    {},
	"EDU":                      {},
	"EDUCATION":                {},
	"EE":                       {},
	"EG":                       {},
	"EMAIL":                    {},
	"EMERCK":                   {},
	"ENERGY":                   {},
	"ENGINEER":                 {},
	"ENGINEERING":              {},
	"ENTERPRISES":              {},
	"EPSON":                    {},
	"EQUIPMENT":                {},
	"ER":                       {},
	"ERICSSON":                 {},
	"ERNI":                     {},
	"ES":                       {},
	"ESQ":                      {},
	"ESTATE":                   {},
	"ET":                       {},
	"EU":                       {},
	"EUROVISION":               {},
	"EUS":                      {},
	"EVENTS":                   {},
	"EXCHANGE":                 {},
	"EXPERT":                   {},
	"EXPOSED":                  {},
	"EXPRESS":                  {},
	"EXTRASPACE":               {},
	"FAGE":                     {},
	"FAIL":                     {},
	"FAIRWINDS":                {},
	"FAITH":                    {},
	"FAMILY":                   {},
	"FAN":                      {},
	"FANS":                     {},
	"FARM":                     {},
	"FARMERS":                  {},
	"FASHION":                  {},
	"FAST":                     {},
	"FEDEX":                    {},
	"FEEDBACK":                 {},
	"FERRARI":                  {},
	"FERRERO":                  {},
	"FI":                       {},
	"FIDELITY":                 {},
	"FIDO":                     {},
	"FILM":                     {},
	"FINAL":                    {},
	"FINANCE":                  {},
	"FINANCIAL":                {},
	"FIRE":                     {},
	"FIRESTONE":                {},
	"FIRMDALE":                 {},
	"FISH":                     {},
	"FISHING":                  {},
	"FIT":                      {},
	"FITNESS":                  {},
	"FJ":                       {},
	"FK":                       {},
	"FLICKR":                   {},
	"FLIGHTS":                  {},
	"FLIR":                     {},
	"FLORIST":                  {},
	"FLOWERS":                  {},
	"FLY":                      {},
	"FM":                       {},
	"FO":                       {},
	"FOO":                      {},
	"FOOD":                     {},
	"FOOTBALL":                 {},
	"FORD":                     {},
	"FOREX":                    {},
	"FORSALE":                  {},
	"FORUM":                    {},
	"FOUNDATION":               {},
	"FOX":                      {},
	"FR":                       {},
	"FREE":                     {},
	"FRESENIUS":                {},
	"FRL":                      {},
	"FROGANS":                  {},
	"FRONTIER":                 {},
	"FTR":                      {},
	"FUJITSU":                  {},
	"FUN":                      {},
	"FUND":                     {},
	"FURNITURE":                {},
	"FUTBOL":                   {},
	"FYI":                      {},
	"GA":                       {},
	"GAL":                      {},
	"GALLERY":                  {},
	"GALLO":                    {},
	"GALLUP":                   {},
	"GAME":                     {},
	"GAMES":                    {},
	"GAP":                      {},
	"GARDEN":                   {},
	"GAY":                      {},
	"GB":                       {},
	"GBIZ":                     {},
	"GD":                       {},
	"GDN":                      {},
	"GE":                       {},
	"GEA":                      {},
	"GENT":                     {},
	"GENTING":                  {},
	"GEORGE":                   {},
	"GF":                       {},
	"GG":                       {},
	"GGEE":                     {},
	"GH":                       {},
	"GI":                       {},
	"GIFT":                     {},
	"GIFTS":                    {},
	"GIVES":                    {},
	"GIVING":                   {},
	"GL":                       {},
	"GLASS":                    {},
	"GLE":                      {},
	"GLOBAL":                   {},
	"GLOBO":                    {},
	"GM":                       {},
	"GMAIL":                    {},
	"GMBH":                     {},
	"GMO":                      {},
	"GMX":                      {},
	"GN":                       {},
	"GODADDY":                  {},
	"GOLD":                     {},
	"GOLDPOINT":                {},
	"GOLF":                     {},
	"GOODYEAR":                 {},
	"GOOG":                     {},
	"GOOGLE":                   {},
	"GOP":                      {},
	"GOT":                      {},
	"GOV":                      {},
	"GP":                       {},
	"GQ":                       {},
	"GR":                       {},
	"GRAINGER":                 {},
	"GRAPHICS":                 {},
	"GRATIS":                   {},
	"GREEN":                    {},
	"GRIPE":                    {},
	"GROCERY":                  {},
	"GROUP":                    {},
	"GS":                       {},
	"GT":                       {},
	"GU":                       {},
	"GUCCI":                    {},
	"GUGE":                     {},
	"GUIDE":                    {},
	"GUITARS":                  {},
	"GURU":                     {},
	"GW":                       {},
	"GY":                       {},
	"HAIR":                     {},
	"HAMBURG":                  {},
	"HANGOUT":                  {},
	"HAUS":                     {},
	"HBO":                      {},
	"HDFC":                     {},
	"HDFCBANK":                 {},
	"HEALTH":                   {},
	"HEALTHCARE":               {},
	"HELP":                     {},
	"HELSINKI":                 {},
	"HERE":                     {},
	"HERMES":                   {},
	"HIPHOP":                   {},
	"HISAMITSU":                {},
	"HITACHI":                  {},
	"HIV":                      {},
	"HK":                       {},
	"HKT":                      {},
	"HM":                       {},
	"HN":                       {},
	"HOCKEY":                   {},
	"HOLDINGS":                 {},
	"HOLIDAY":                  {},
	"HOMEDEPOT":                {},
	"HOMEGOODS":                {},
	"HOMES":                    {},
	"HOMESENSE":                {},
	"HONDA":                    {},
	"HORSE":                    {},
	"HOSPITAL":                 {},
	"HOST":                     {},
	"HOSTING":                  {},
	"HOT":                      {},
	"HOTELS":                   {},
	"HOTMAIL":                  {},
	"HOUSE":                    {},
	"HOW":                      {},
	"HR":                       {},
	"HSBC":                     {},
	"HT":                       {},
	"HU":                       {},
	"HUGHES":                   {},
	"HYATT":                    {},
	"HYUNDAI":                  {},
	"IBM":                      {},
	"ICBC":                     {},
	"ICE":                      {},
	"ICU":                      {},
	"ID":                       {},
	"IE":                       {},
	"IEEE":                     {},
	"IFM":                      {},
	"IKANO":                    {},
	"IL":                       {},
	"IM":                       {},
	"IMAMAT":                   {},
	"IMDB":                     {},
	"IMMO":                     {},
	"IMMOBILIEN":               {},
	"IN":                       {},
	"INC":                      {},
	"INDUSTRIES":               {},
	"INFINITI":                 {},
	"INFO":                     {},
	"ING":                      {},
	"INK":                      {},
	"INSTITUTE":                {},
	"INSURANCE":                {},
	"INSURE":                   {},
	"INT":                      {},
	"INTERNATIONAL":            {},
	"INTUIT":                   {},
	"INVESTMENTS":              {},
	"IO":                       {},
	"IPIRANGA":                 {},
	"IQ":                       {},
	"IR":                       {},
	"IRISH":                    {},
	"IS":                       {},
	"ISMAILI":                  {},
	"IST":                      {},
	"ISTANBUL":                 {},
	"IT":                       {},
	"ITAU":                     {},
	"ITV":                      {},
	"JAGUAR":                   {},
	"JAVA":                     {},
	"JCB":                      {},
	"JE":                       {},
	"JEEP":                     {},
	"JETZT":                    {},
	"JEWELRY":                  {},
	"JIO":                      {},
	"JLL":                      {},
	"JM":                       {},
	"JMP":                      {},
	"JNJ":                      {},
	"JO":                       {},
	"JOBS":                     {},
	"JOBURG":                   {},
	"JOT":                      {},
	"JOY":                      {},
	"JP":                       {},
	"JPMORGAN":                 {},
	"JPRS":                     {},
	"JUEGOS":                   {},
	"JUNIPER":                  {},
	"KAUFEN":                   {},
	"KDDI":                     {},
	"KE":                       {},
	"KERRYHOTELS":              {},
	"KERRYPROPERTIES":          {},
	"KFH":                      {},
	"KG":                       {},
	"KH":                       {},
	"KI":                       {},
	"KIA":                      {},
	"KIDS":                     {},
	"KIM":                      {},
	"KINDLE":                   {},
	"KITCHEN":                  {},
	"KIWI":                     {},
	"KM":                       {},
	"KN":                       {},
	"KOELN":                    {},
	"KOMATSU":                  {},
	"KOSHER":                   {},
	"KP":                       {},
	"KPMG":                     {},
	"KPN":                      {},
	"KR":                       {},
	"KRD":                      {},
	"KRED":                     {},
	"KUOKGROUP":                {},
	"KW":                       {},
	"KY":                       {},
	"KYOTO":                    {},
	"KZ":                       {},
	"LA":                       {},
	"LACAIXA":                  {},
	"LAMBORGHINI":              {},
	"LAMER":                    {},
	"LAND":                     {},
	"LANDROVER":                {},
	"LANXESS":                  {},
	"LASALLE":                  {},
	"LAT":                      {},
	"LATINO":                   {},
	"LATROBE":                  {},
	"LAW":                      {},
	"LAWYER":                   {},
	"LB":                       {},
	"LC":                       {},
	"LDS":                      {},
	"LEASE":                    {},
	"LECLERC":                  {},
	"LEFRAK":                   {},
	"LEGAL":                    {},
	"LEGO":                     {},
	"LEXUS":                    {},
	"LGBT":                     {},
	"LI":                       {},
	"LIDL":                     {},
	"LIFE":                     {},
	"LIFEINSURANCE":            {},
	"LIFESTYLE":                {},
	"LIGHTING":                 {},
	"LIKE":                     {},
	"LILLY":                    {},
	"LIMITED":                  {},
	"LIMO":                     {},
	"LINCOLN":                  {},
	"LINK":                     {},
	"LIVE":                     {},
	"LIVING":                   {},
	"LK":                       {},
	"LLC":                      {},
	"LLP":                      {},
	"LOAN":                     {},
	"LOANS":                    {},
	"LOCKER":                   {},
	"LOCUS":                    {},
	"LOL":                      {},
	"LONDON":                   {},
	"LOTTE":                    {},
	"LOTTO":                    {},
	"LOVE":                     {},
	"LPL":                      {},
	"LPLFINANCIAL":             {},
	"LR":                       {},
	"LS":                       {},
	"LT":                       {},
	"LTD":                      {},
	"LTDA":                     {},
	"LU":                       {},
	"LUNDBECK":                 {},
	"LUXE":                     {},
	"LUXURY":                   {},
	"LV":                       {},
	"LY":                       {},
	"MA":                       {},
	"MADRID":                   {},
	"MAIF":                     {},
	"MAISON":                   {},
	"MAKEUP":                   {},
	"MAN":                      {},
	"MANAGEMENT":               {},
	"MANGO":                    {},
	"MAP":                      {},
	"MARKET":                   {},
	"MARKETING":                {},
	"MARKETS":                  {},
	"MARRIOTT":                 {},
	"MARSHALLS":                {},
	"MATTEL":                   {},
	"MBA":                      {},
	"MC":                       {},
	"MCKINSEY":                 {},
	"MD":                       {},
	"ME":                       {},
	"MED":                      {},
	"MEDIA":                    {},
	"MEET":                     {},
	"MELBOURNE":                {},
	"MEME":                     {},
	"MEMORIAL":                 {},
	"MEN":                      {},
	"MENU":                     {},
	"MERCK":                    {},
	"MERCKMSD":                 {},
	"MG":                       {},
	"MH":                       {},
	"MIAMI":                    {},
	"MICROSOFT":                {},
	"MIL":                      {},
	"MINI":                     {},
	"MINT":                     {},
	"MIT":                      {},
	"MITSUBISHI":               {},
	"MK":                       {},
	"ML":                       {},
	"MLB":                      {},
	"MLS":                      {},
	"MM":                       {},
	"MMA":                      {},
	"MN":                       {},
	"MO":                       {},
	"MOBI":                     {},
	"MOBILE":                   {},
	"MODA":                     {},
	"MOE":                      {},
	"MOI":                      {},
	"MOM":                      {},
	"MONASH":                   {},
	"MONEY":                    {},
	"MONSTER":                  {},
	"MORMON":                   {},
	"MORTGAGE":                 {},
	"MOSCOW":                   {},
	"MOTO":                     {},
	"MOTORCYCLES":              {},
	"MOV":                      {},
	"MOVIE":                    {},
	"MP":                       {},
	"MQ":                       {},
	"MR":                       {},
	"MS":                       {},
	"MSD":                      {},
	"MT":                       {},
	"MTN":                      {},
	"MTR":                      {},
	"MU":                       {},
	"MUSEUM":                   {},
	"MUSIC":                    {},
	"MV":                       {},
	"MW":                       {},
	"MX":                       {},
	"MY":                       {},
	"MZ":                       {},
	"NA":                       {},
	"NAB":                      {},
	"NAGOYA":                   {},
	"NAME":                     {},
	"NAVY":                     {},
	"NBA":                      {},
	"NC":                       {},
	"NE":                       {},
	"NEC":                      {},
	"NET":                      {},
	"NETBANK":                  {},
	"NETFLIX":                  {},
	"NETWORK":                  {},
	"NEUSTAR":                  {},
	"NEW":                      {},
	"NEWS":                     {},
	"NEXT":                     {},
	"NEXTDIRECT":               {},
	"NEXUS":                    {},
	"NF":                       {},
	"NFL":                      {},
	"NG":                       {},
	"NGO":                      {},
	"NHK":                      {},
	"NI":                       {},
	"NICO":                     {},
	"NIKE":                     {},
	"NIKON":                    {},
	"NINJA":                    {},
	"NISSAN":                   {},
	"NISSAY":                   {},
	"NL":                       {},
	"NO":                       {},
	"NOKIA":                    {},
	"NORTON":                   {},
	"NOW":                      {},
	"NOWRUZ":                   {},
	"NOWTV":                    {},
	"NP":                       {},
	"NR":                       {},
	"NRA":                      {},
	"NRW":                      {},
	"NTT":                      {},
	"NU":                       {},
	"NYC":                      {},
	"NZ":                       {},
	"OBI":                      {},
	"OBSERVER":                 {},
	"OFFICE":                   {},
	"OKINAWA":                  {},
	"OLAYAN":                   {},
	"OLAYANGROUP":              {},
	"OLLO":                     {},
	"OM":                       {},
	"OMEGA":                    {},
	"ONE":                      {},
	"ONG":                      {},
	"ONL":                      {},
	"ONLINE":                   {},
	"OOO":                      {},
	"OPEN":                     {},
	"ORACLE":                   {},
	"ORANGE":                   {},
	"ORG":                      {},
	"ORGANIC":                  {},
	"ORIGINS":                  {},
	"OSAKA":                    {},
	"OTSUKA":                   {},
	"OTT":                      {},
	"OVH":                      {},
	"PA":                       {},
	"PAGE":                     {},
	"PANASONIC":                {},
	"PARIS":                    {},
	"PARS":                     {},
	"PARTNERS":                 {},
	"PARTS":                    {},
	"PARTY":                    {},
	"PAY":                      {},
	"PCCW":                     {},
	"PE":                       {},
	"PET":                      {},
	"PF":                       {},
	"PFIZER":                   {},
	"PG":                       {},
	"PH":                       {},
	"PHARMACY":                 {},
	"PHD":                      {},
	"PHILIPS":                  {},
	"PHONE":                    {},
	"PHOTO":                    {},
	"PHOTOGRAPHY":              {},
	"PHOTOS":                   {},
	"PHYSIO":                   {},
	"PICS":                     {},
	"PICTET":                   {},
	"PICTURES":                 {},
	"PID":                      {},
	"PIN":                      {},
	"PING":                     {},
	"PINK":                     {},
	"PIONEER":                  {},
	"PIZZA":                    {},
	"PK":                       {},
	"PL":                       {},
	"PLACE":                    {},
	"PLAY":                     {},
	"PLAYSTATION":              {},
	"PLUMBING":                 {},
	"PLUS":                     {},
	"PM":                       {},
	"PN":                       {},
	"PNC":                      {},
	"POHL":                     {},
	"POKER":                    {},
	"POLITIE":                  {},
	"PORN":                     {},
	"POST":                     {},
	"PR":                       {},
	"PRAXI":                    {},
	"PRESS":                    {},
	"PRIME":                    {},
	"PRO":                      {},
	"PROD":                     {},
	"PRODUCTIONS":              {},
	"PROF":                     {},
	"PROGRESSIVE":              {},
	"PROMO":                    {},
	"PROPERTIES":               {},
	"PROPERTY":                 {},
	"PROTECTION":               {},
	"PRU":                      {},
	"PRUDENTIAL":               {},
	"PS":                       {},
	"PT":                       {},
	"PUB":                      {},
	"PW":                       {},
	"PWC":                      {},
	"PY":                       {},
	"QA":                       {},
	"QPON":                     {},
	"QUEBEC":                   {},
	"QUEST":                    {},
	"RACING":                   {},
	"RADIO":                    {},
	"RE":                       {},
	"READ":                     {},
	"REALESTATE":               {},
	"REALTOR":                  {},
	"REALTY":                   {},
	"RECIPES":                  {},
	"RED":                      {},
	"REDUMBRELLA":              {},
	"REHAB":                    {},
	"REISE":                    {},
	"REISEN":                   {},
	"REIT":                     {},
	"RELIANCE":                 {},
	"REN":                      {},
	"RENT":                     {},
	"RENTALS":                  {},
	"REPAIR":                   {},
	"REPORT":                   {},
	"REPUBLICAN":               {},
	"REST":                     {},
	"RESTAURANT":               {},
	"REVIEW":                   {},
	"REVIEWS":                  {},
	"REXROTH":                  {},
	"RICH":                     {},
	"RICHARDLI":                {},
	"RICOH":                    {},
	"RIL":                      {},
	"RIO":                      {},
	"RIP":                      {},
	"RO":                       {},
	"ROCKS":                    {},
	"RODEO":                    {},
	"ROGERS":                   {},
	"ROOM":                     {},
	"RS":                       {},
	"RSVP":                     {},
	"RU":                       {},
	"RUGBY":                    {},
	"RUHR":                     {},
	"RUN":                      {},
	"RW":                       {},
	"RWE":                      {},
	"RYUKYU":                   {},
	"SA":                       {},
	"SAARLAND":                 {},
	"SAFE":                     {},
	"SAFETY":                   {},
	"SAKURA":                   {},
	"SALE":                     {},
	"SALON":                    {},
	"SAMSCLUB":                 {},
	"SAMSUNG":                  {},
	"SANDVIK":                  {},
	"SANDVIKCOROMANT":          {},
	"SANOFI":                   {},
	"SAP":                      {},
	"SARL":                     {},
	"SAS":                      {},
	"SAVE":                     {},
	"SAXO":                     {},
	"SB":                       {},
	"SBI":                      {},
	"SBS":                      {},
	"SC":                       {},
	"SCB":                      {},
	"SCHAEFFLER":               {},
	"SCHMIDT":                  {},
	"SCHOLARSHIPS":             {},
	"SCHOOL":                   {},
	"SCHULE":                   {},
	"SCHWARZ":                  {},
	"SCIENCE":                  {},
	"SCOT":                     {},
	"SD":                       {},
	"SE":                       {},
	"SEARCH":                   {},
	"SEAT":                     {},
	"SECURE":                   {},
	"SECURITY":                 {},
	"SEEK":                     {},
	"SELECT":                   {},
	"SENER":                    {},
	"SERVICES":                 {},
	"SEVEN":                    {},
	"SEW":                      {},
	"SEX":                      {},
	"SEXY":                     {},
	"SFR":                      {},
	"SG":                       {},
	"SH":                       {},
	"SHANGRILA":                {},
	"SHARP":                    {},
	"SHELL":                    {},
	"SHIA":                     {},
	"SHIKSHA":                  {},
	"SHOES":                    {},
	"SHOP":                     {},
	"SHOPPING":                 {},
	"SHOUJI":                   {},
	"SHOW":                     {},
	"SI":                       {},
	"SILK":                     {},
	"SINA":                     {},
	"SINGLES":                  {},
	"SITE":                     {},
	"SJ":                       {},
	"SK":                       {},
	"SKI":                      {},
	"SKIN":                     {},
	"SKY":                      {},
	"SKYPE":                    {},
	"SL":                       {},
	"SLING":                    {},
	"SM":                       {},
	"SMART":                    {},
	"SMILE":                    {},
	"SN":                       {},
	"SNCF":                     {},
	"SO":                       {},
	"SOCCER":                   {},
	"SOCIAL":                   {},
	"SOFTBANK":                 {},
	"SOFTWARE":                 {},
	"SOHU":                     {},
	"SOLAR":                    {},
	"SOLUTIONS":                {},
	"SONG":                     {},
	"SONY":                     {},
	"SOY":                      {},
	"SPA":                      {},
	"SPACE":                    {},
	"SPORT":                    {},
	"SPOT":                     {},
	"SR":                       {},
	"SRL":                      {},
	"SS":                       {},
	"ST":                       {},
	"STADA":                    {},
	"STAPLES":                  {},
	"STAR":                     {},
	"STATEBANK":                {},
	"STATEFARM":                {},
	"STC":                      {},
	"STCGROUP":                 {},
	"STOCKHOLM":                {},
	"STORAGE":                  {},
	"STORE":                    {},
	"STREAM":                   {},
	"STUDIO":                   {},
	"STUDY":                    {},
	"STYLE":                    {},
	"SU":                       {},
	"SUCKS":                    {},
	"SUPPLIES":                 {},
	"SUPPLY":                   {},
	"SUPPORT":                  {},
	"SURF":                     {},
	"SURGERY":                  {},
	"SUZUKI":                   {},
	"SV":                       {},
	"SWATCH":                   {},
	"SWISS":                    {},
	"SX":                       {},
	"SY":                       {},
	"SYDNEY":                   {},
	"SYSTEMS":                  {},
	"SZ":                       {},
	"TAB":                      {},
	"TAIPEI":                   {},
	"TALK":                     {},
	"TAOBAO":                   {},
	"TARGET":                   {},
	"TATAMOTORS":               {},
	"TATAR":                    {},
	"TATTOO":                   {},
	"TAX":                      {},
	"TAXI":                     {},
	"TC":                       {},
	"TCI":                      {},
	"TD":                       {},
	"TDK":                      {},
	"TEAM":                     {},
	"TECH":                     {},
	"TECHNOLOGY":               {},
	"TEL":                      {},
	"TEMASEK":                  {},
	"TENNIS":                   {},
	"TEVA":                     {},
	"TF":                       {},
	"TG":                       {},
	"TH":                       {},
	"THD":                      {},
	"THEATER":                  {},
	"THEATRE":                  {},
	"TIAA":                     {},
	"TICKETS":                  {},
	"TIENDA":                   {},
	"TIPS":                     {},
	"TIRES":                    {},
	"TIROL":                    {},
	"TJ":                       {},
	"TJMAXX":                   {},
	"TJX":                      {},
	"TK":                       {},
	"TKMAXX":                   {},
	"TL":                       {},
	"TM":                       {},
	"TMALL":                    {},
	"TN":                       {},
	"TO":                       {},
	"TODAY":                    {},
	"TOKYO":                    {},
	"TOOLS":                    {},
	"TOP":                      {},
	"TORAY":                    {},
	"TOSHIBA":                  {},
	"TOTAL":                    {},
	"TOURS":                    {},
	"TOWN":                     {},
	"TOYOTA":                   {},
	"TOYS":                     {},
	"TR":                       {},
	"TRADE":                    {},
	"TRADING":                  {},
	"TRAINING":                 {},
	"TRAVEL":                   {},
	"TRAVELERS":                {},
	"TRAVELERSINSURANCE":       {},
	"TRUST":                    {},
	"TRV":                      {},
	"TT":                       {},
	"TUBE":                     {},
	"TUI":                      {},
	"TUNES":                    {},
	"TUSHU":                    {},
	"TV":                       {},
	"TVS":                      {},
	"TW":                       {},
	"TZ":                       {},
	"UA":                       {},
	"UBANK":                    {},
	"UBS":                      {},
	"UG":                       {},
	"UK":                       {},
	"UNICOM":                   {},
	"UNIVERSITY":               {},
	"UNO":                      {},
	"UOL":                      {},
	"UPS":                      {},
	"US":                       {},
	"UY":                       {},
	"UZ":                       {},
	"VA":                       {},
	"VACATIONS":                {},
	"VANA":                     {},
	"VANGUARD":                 {},
	"VC":                       {},
	"VE":                       {},
	"VEGAS":                    {},
	"VENTURES":                 {},
	"VERISIGN":                 {},
	"VERSICHERUNG":             {},
	"VET":                      {},
	"VG":                       {},
	"VI":                       {},
	"VIAJES":                   {},
	"VIDEO":                    {},
	"VIG":                      {},
	"VIKING":                   {},
	"VILLAS":                   {},
	"VIN":                      {},
	"VIP":                      {},
	"VIRGIN":                   {},
	"VISA":                     {},
	"VISION":                   {},
	"VIVA":                     {},
	"VIVO":                     {},
	"VLAANDEREN":               {},
	"VN":                       {},
	"VODKA":                    {},
	"VOLVO":                    {},
	"VOTE":                     {},
	"VOTING":                   {},
	"VOTO":                     {},
	"VOYAGE":                   {},
	"VU":                       {},
	"WALES":                    {},
	"WALMART":                  {},
	"WALTER":                   {},
	"WANG":                     {},
	"WANGGOU":                  {},
	"WATCH":                    {},
	"WATCHES":                  {},
	"WEATHER":                  {},
	"WEATHERCHANNEL":           {},
	"WEBCAM":                   {},
	"WEBER":                    {},
	"WEBSITE":                  {},
	"WED":                      {},
	"WEDDING":                  {},
	"WEIBO":                    {},
	"WEIR":                     {},
	"WF":                       {},
	"WHOSWHO":                  {},
	"WIEN":                     {},
	"WIKI":                     {},
	"WILLIAMHILL":              {},
	"WIN":                      {},
	"WINDOWS":                  {},
	"WINE":                     {},
	"WINNERS":                  {},
	"WME":                      {},
	"WOODSIDE":                 {},
	"WORK":                     {},
	"WORKS":                    {},
	"WORLD":                    {},
	"WOW":                      {},
	"WS":                       {},
	"WTC":                      {},
	"WTF":                      {},
	"XBOX":                     {},
	"XEROX":                    {},
	"XIHUAN":                   {},
	"XIN":                      {},
	"XN--11B4C3D":              {},
	"XN--1CK2E1B":              {},
	"XN--1QQW23A":              {},
	"XN--2SCRJ9C":              {},
	"XN--30RR7Y":               {},
	"XN--3BST00M":              {},
	"XN--3DS443G":              {},
	"XN--3E0B707E":             {},
	"XN--3HCRJ9C":              {},
	"XN--3PXU8K":               {},
	"XN--42C2D9A":              {},
	"XN--45BR5CYL":             {},
	"XN--45BRJ9C":              {},
	"XN--45Q11C":               {},
	"XN--4DBRK0CE":             {},
	"XN--4GBRIM":               {},
	"XN--54B7FTA0CC":           {},
	"XN--55QW42G":              {},
	"XN--55QX5D":               {},
	"XN--5SU34J936BGSG":        {},
	"XN--5TZM5G":               {},
	"XN--6FRZ82G":              {},
	"XN--6QQ986B3XL":           {},
	"XN--80ADXHKS":             {},
	"XN--80AO21A":              {},
	"XN--80AQECDR1A":           {},
	"XN--80ASEHDB":             {},
	"XN--80ASWG":               {},
	"XN--8Y0A063A":             {},
	"XN--90A3AC":               {},
	"XN--90AE":                 {},
	"XN--90AIS":                {},
	"XN--9DBQ2A":               {},
	"XN--9ET52U":               {},
	"XN--9KRT00A":              {},
	"XN--B4W605FERD":           {},
	"XN--BCK1B9A5DRE4C":        {},
	"XN--C1AVG":                {},
	"XN--C2BR7G":               {},
	"XN--CCK2B3B":              {},
	"XN--CCKWCXETD":            {},
	"XN--CG4BKI":               {},
	"XN--CLCHC0EA0B2G2A9GCD":   {},
	"XN--CZR694B":              {},
	"XN--CZRS0T":               {},
	"XN--CZRU2D":               {},
	"XN--D1ACJ3B":              {},
	"XN--D1ALF":                {},
	"XN--E1A4C":                {},
	"XN--ECKVDTC9D":            {},
	"XN--EFVY88H":              {},
	"XN--FCT429K":              {},
	"XN--FHBEI":                {},
	"XN--FIQ228C5HS":           {},
	"XN--FIQ64B":               {},
	"XN--FIQS8S":               {},
	"XN--FIQZ9S":               {},
	"XN--FJQ720A":              {},
	"XN--FLW351E":              {},
	"XN--FPCRJ9C3D":            {},
	"XN--FZC2C9E2C":            {},
	"XN--FZYS8D69UVGM":         {},
	"XN--G2XX48C":              {},
	"XN--GCKR3F0F":             {},
	"XN--GECRJ9C":              {},
	"XN--GK3AT1E":              {},
	"XN--H2BREG3EVE":           {},
	"XN--H2BRJ9C":              {},
	"XN--H2BRJ9C8C":            {},
	"XN--HXT814E":              {},
	"XN--I1B6B1A6A2E":          {},
	"XN--IMR513N":              {},
	"XN--IO0A7I":               {},
	"XN--J1AEF":                {},
	"XN--J1AMH":                {},
	"XN--J6W193G":              {},
	"XN--JLQ480N2RG":           {},
	"XN--JVR189M":              {},
	"XN--KCRX77D1X4A":          {},
	"XN--KPRW13D":              {},
	"XN--KPRY57D":              {},
	"XN--KPUT3I":               {},
	"XN--L1ACC":                {},
	"XN--LGBBAT1AD8J":          {},
	"XN--MGB9AWBF":             {},
	"XN--MGBA3A3EJT":           {},
	"XN--MGBA3A4F16A":          {},
	"XN--MGBA7C0BBN0A":         {},
	"XN--MGBAAM7A8H":           {},
	"XN--MGBAB2BD":             {},
	"XN--MGBAH1A3HJKRD":        {},
	"XN--MGBAI9AZGQP6J":        {},
	"XN--MGBAYH7GPA":           {},
	"XN--MGBBH1A":              {},
	"XN--MGBBH1A71E":           {},
	"XN--MGBC0A9AZCG":          {},
	"XN--MGBCA7DZDO":           {},
	"XN--MGBCPQ6GPA1A":         {},
	"XN--MGBERP4A5D4AR":        {},
	"XN--MGBGU82A":             {},
	"XN--MGBI4ECEXP":           {},
	"XN--MGBPL2FH":             {},
	"XN--MGBT3DHD":             {},
	"XN--MGBTX2B":              {},
	"XN--MGBX4CD0AB":           {},
	"XN--MIX891F":              {},
	"XN--MK1BU44C":             {},
	"XN--MXTQ1M":               {},
	"XN--NGBC5AZD":             {},
	"XN--NGBE9E0A":             {},
	"XN--NGBRX":                {},
	"XN--NODE":                 {},
	"XN--NQV7F":                {},
	"XN--NQV7FS00EMA":          {},
	"XN--NYQY26A":              {},
	"XN--O3CW4H":               {},
	"XN--OGBPF8FL":             {},
	"XN--OTU796D":              {},
	"XN--P1ACF":                {},
	"XN--P1AI":                 {},
	"XN--PGBS0DH":              {},
	"XN--PSSY2U":               {},
	"XN--Q7CE6A":               {},
	"XN--Q9JYB4C":              {},
	"XN--QCKA1PMC":             {},
	"XN--QXA6A":                {},
	"XN--QXAM":                 {},
	"XN--RHQV96G":              {},
	"XN--ROVU88B":              {},
	"XN--RVC1E0AM3E":           {},
	"XN--S9BRJ9C":              {},
	"XN--SES554G":              {},
	"XN--T60B56A":              {},
	"XN--TCKWE":                {},
	"XN--TIQ49XQYJ":            {},
	"XN--UNUP4Y":               {},
	"XN--VERMGENSBERATER-CTB":  {},
	"XN--VERMGENSBERATUNG-PWB": {},
	"XN--VHQUV":                {},
	"XN--VUQ861B":              {},
	"XN--W4R85EL8FHU5DNRA":     {},
	"XN--W4RS40L":              {},
	"XN--WGBH1C":               {},
	"XN--WGBL6A":               {},
	"XN--XHQ521B":              {},
	"XN--XKC2AL3HYE2A":         {},
	"XN--XKC2DL3A5EE0H":        {},
	"XN--Y9A3AQ":               {},
	"XN--YFRO4I67O":            {},
	"XN--YGBI2AMMX":            {},
	"XN--ZFR164B":              {},
	"XXX":                      {},
	"XYZ":                      {},
	"YACHTS":                   {},
	"YAHOO":                    {},
	"YAMAXUN":                  {},
	"YANDEX":                   {},
	"YE":                       {},
	"YODOBASHI":                {},
	"YOGA":                     {},
	"YOKOHAMA":                 {},
	"YOU":                      {},
	"YOUTUBE":                  {},
	"YT":                       {},
	"YUN":                      {},
	"ZA":                       {},
	"ZAPPOS":                   {},
	"ZARA":                     {},
	"ZERO":                     {},
	"ZIP":                      {},
	"ZM":                       {},
	"ZONE":                     {},
	"ZUERICH":                  {},
	"ZW":                       {},
}

func isTLD(s string) bool {
	_, ok := tlds[strings.ToUpper(s)]
	return ok
}
