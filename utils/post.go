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
