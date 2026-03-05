package utils

import (
	"html"
	"html/template"
	"regexp"
	"strings"
)

var hashtagRE = regexp.MustCompile(`#([0-9A-Za-z]+)`)

func ParseBodyForBacklinks(s string) []uint32 {
	matches := hashtagRE.FindAllStringSubmatch(s, -1)
	res := make([]uint32, 0)
	for _, m := range matches {
		n, err := AToID(m[1])
		if err != nil {
			continue
		}
		res = append(res, n)
	}
	return res
}

func RenderTextBody(s string) template.HTML {
	var out strings.Builder
	last := 0
	matches := hashtagRE.FindAllStringSubmatchIndex(s, -1)
	for _, m := range matches {
		start, end := m[0], m[1]
		capStart, capEnd := m[2], m[3]
		out.WriteString(html.EscapeString(s[last:start]))
		capture := s[capStart:capEnd]
		out.WriteString(`<a href="/p/`)
		out.WriteString(capture)
		out.WriteString(`">#`)
		out.WriteString(capture)
		out.WriteString(`</a>`)
		last = end
	}

	out.WriteString(html.EscapeString(s[last:]))
	return template.HTML(out.String())
}

func RenderImageBody(cid string, alt *string) template.HTML {
	var out strings.Builder
	isgif := strings.HasSuffix(cid, ".gif")
	out.WriteString(`<div class="image-wrapper`)
	if !isgif {
		out.WriteString(` thumb`)
	}
	out.WriteString(`"`)
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
