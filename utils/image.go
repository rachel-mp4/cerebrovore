package utils

import (
	"html"
	"html/template"
	"strings"
)

func RenderAvatarPFP(cidp *string, alt *string, isPixelArt *bool) template.HTML {
	if cidp == nil {
		return template.HTML("pls report bug, nil avatarpfp cid")
	}
	thumb := "yes"
	if isPixelArt != nil && !*isPixelArt {
		thumb = "jpg"
	}
	cid := *cidp
	var out strings.Builder
	isgif := strings.HasSuffix(cid, ".gif")
	out.WriteString(`<div class="avatar image-wrapper thumb"`)
	if !isgif {
		out.WriteString(` data-thumb="/blob?cid=`)
		out.WriteString(html.EscapeString(cid))
		out.WriteString(`&thumb=`)
		out.WriteString(thumb)
		out.WriteString(`" data-full="/blob?cid=`)
		out.WriteString(html.EscapeString(cid))
		out.WriteString(`"`)
	}
	out.WriteString(`><img class="bg-img" src="/blob?cid=`)
	out.WriteString(html.EscapeString(cid))
	if !isgif {
		out.WriteString(`&thumb=`)
		out.WriteString(thumb)
	}
	if alt != nil {
		out.WriteString(`" alt="`)
		out.WriteString(html.EscapeString(*alt))
		out.WriteString(`" title="`)
		out.WriteString(html.EscapeString(*alt))
	}
	out.WriteString(`" /><img class="fg-img" src="/blob?cid=`)
	out.WriteString(html.EscapeString(cid))
	if !isgif {
		out.WriteString(`&thumb=`)
		out.WriteString(thumb)
	}
	if alt != nil {
		out.WriteString(`" alt="`)
		out.WriteString(html.EscapeString(*alt))
		out.WriteString(`" title="`)
		out.WriteString(html.EscapeString(*alt))
	}
	out.WriteString(`" /></div>`)
	return template.HTML(out.String())
}
