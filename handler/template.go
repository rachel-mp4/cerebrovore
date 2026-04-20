package handler

import (
	"context"
	"html/template"
	"io"
	"time"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
)

func init() {
	homeT = homeTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/home.html",
	)}
	patchT = patchTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/patch-notes.html",
		"./tmpl/partial/patch.html",
		"./tmpl/partial/note.html",
	)}
	moderateT = moderateTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/moderate.html",
		"./tmpl/partial/post.html",
		"./tmpl/partial/ban.html",
	)}
	meT = meTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/me.html",
	)}
	adminT = adminTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/admin.html",
	)}
	banT = banTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/emptyl.html",
		"./tmpl/empty.html",
		"./tmpl/ban.html",
		"./tmpl/partial/post.html",
		"./tmpl/partial/ban.html",
	)}
	loginT = loginTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/emptyl.html",
		"./tmpl/empty.html",
		"./tmpl/login.html",
	)}
	accountT = accountTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/emptyl.html",
		"./tmpl/empty.html",
		"./tmpl/account.html",
	)}
	beepT = beepTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/beep.html",
	)}
	threadT = threadTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/wormwatch.html",
		"./tmpl/thread.html",
		"./tmpl/partial/post.html",
	)}
	profileT = profileTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/profile.html",
		"./tmpl/partial/post.html",
	)}
	editprofileT = editprofileTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/editprofile.html",
	)}
	newthreadT = newthreadTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/newthread.html",
	)}
	threadsT = threadsTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/index.html",
		"./tmpl/partial/post.html",
	)}
	catalogT = catalogTemplate{newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/catalog.html",
		"./tmpl/partial/post.html",
	)}
	bumpedT = bumpedTemplate{newTemplate(
		"./tmpl/bumped-threads.html",
		"./tmpl/partial/threadlink.html",
	)}
	codeT = codeTemplate{newTemplate(
		"./tmpl/code.html",
	)}
	codeerrT = codeerrTemplate{newTemplate(
		"./tmpl/codeerr.html",
	)}
}

type baseresp struct {
	justbaseresp
	CompiledAssets *CompiledAssets
	Threads        []types.Thread
	Username       string
	IsMod          bool
}

func (h *Handler) makebase(title string, c *Client, ctx context.Context) (baseresp, error) {
	tt, err := h.db.GetBumps(ctx)
	return baseresp{
		h.makejustbase(title, true),
		h.ca,
		tt,
		c.Username,
		c.IsMod,
	}, err
}

func (h *Handler) makejustbase(title string, ws bool) justbaseresp {
	return justbaseresp{
		title,
		h.crack,
		nil,
		"var(--primary)",
		ws,
	}
}

type justbaseresp struct {
	Title      string
	Crack      string
	ReplyCount *int
	Accent     string
	Websockets bool
}

func newTemplate(files ...string) *template.Template {
	return template.Must(
		template.New("").Funcs(
			template.FuncMap{
				"idtoa":            utils.IDToA,
				"intto36a":         utils.IntTo36A,
				"renderImageBody":  utils.RenderImageBody,
				"renderAvatarPFP":  utils.RenderAvatarPFP,
				"renderTextBody":   utils.RenderTextBody,
				"colorIsDark":      utils.ColorIsDark,
				"colorToA":         utils.ColorToAp,
				"maxReplies":       utils.MaxReplies,
				"maxBumps":         utils.MaxBumps,
				"formatTime":       utils.FormatTime,
				"remainingTime":    utils.RemainingTime,
				"timeSince":        utils.TimeSince,
				"ftime":            utils.FTime,
				"topicOrIdtoa":     types.TopicOrIdtoa,
				"percentRemaining": utils.PercentRemaining,
				"boolPtrIsTrue":    boolPtrIsTrue,
			}).ParseFiles(files...),
	)
}

func boolPtrIsTrue(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

// maybe this is not ideal, it's a lot of boilerplate, however i want to give
// some static typing to our templates. perhaps a bit overkill for most cases
// but i think it's especially nice for htmx heavy parts of application like
// moderation, and should help throw errors whenever i extend base template for
// the things which inherit baseresp but aren't created by makebase. this
// is my reasoning...
// in addition to exec, you can define other methods as well. clog.Tmpl()
// automatically logs any errors from templating, if they exist, so make sure
// to wrap any template executions in this

var homeT homeTemplate

func (t *homeTemplate) exec(w io.Writer, home homeresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", home))
}

type homeTemplate struct {
	template *template.Template
}

type homeresp struct {
	baseresp
	Version string
	Time    *time.Time
	Commit  string
	Link    string
}

var moderateT moderateTemplate

func (t *moderateTemplate) inspect(w io.Writer, post *types.Post) {
	post.ViewerIsMod = true
	clog.Tmpl(t.template.ExecuteTemplate(w, "check-post", inspectresp{post}))
}

func (t *moderateTemplate) exec(w io.Writer, moderate moderateresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", moderate))
}

func (t *moderateTemplate) reject(w io.Writer) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "rejected", nil))
}

func (t *moderateTemplate) accept(w io.Writer) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "accepted", nil))
}

func (t *moderateTemplate) report(w io.Writer) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "report-submitted", nil))
}

func (t *moderateTemplate) error(w io.Writer, msg string) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "errored", moderateerrorresp{msg}))
}

func (t *moderateTemplate) confirm(w io.Writer, action *types.Action, actioning string, nid string, username string, reason string, comment string, until string) {
	action.IsMod = true
	if action.Post != nil {
		action.Post.ViewerIsMod = true
	}
	confirm := moderateconfirmresp{
		Action:   *action,
		Type:     actioning,
		Id:       nid,
		Username: username,
		Reason:   reason,
		Comment:  comment,
		Until:    until,
	}
	clog.Tmpl(t.template.ExecuteTemplate(w, "confirm", confirm))
}

func (t *moderateTemplate) reports(w io.Writer, reports []types.Report, cursor *int) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "reports", moderatereportsresp{Reports: reports, Cursor: cursor}))
}

func (t *moderateTemplate) confirmed(w io.Writer, action *types.Action, actioned string) {
	action.IsMod = true
	if action.Post != nil {
		action.Post.ViewerIsMod = true
	}
	confirmed := moderateconfirmedresp{
		Action: *action,
		Type:   actioned,
	}
	clog.Tmpl(t.template.ExecuteTemplate(w, "confirmed", confirmed))
}

func (t *moderateTemplate) canceled(w io.Writer) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "emptyactionform", nil))
}

type inspectresp struct {
	Post *types.Post
}

type moderateTemplate struct {
	template *template.Template
}

type moderatereportsresp struct {
	Reports []types.Report
	Cursor  *int
}

type moderateconfirmresp struct {
	types.Action
	Type     string
	Id       string
	Username string
	Reason   string
	Comment  string
	Until    string
}

type moderateconfirmedresp struct {
	types.Action
	Type string
}

type moderateerrorresp struct {
	Message string
}

type moderateresp struct {
	baseresp
	Appeals []types.Action
	Reports []types.Report
	Cursor  *int
}

var adminT adminTemplate

func (t *adminTemplate) exec(w io.Writer, admin adminresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", admin))
}

func (t *adminTemplate) plusmodsuccess(w io.Writer, username string) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "mod", username))
}

type adminTemplate struct {
	template *template.Template
}

type adminresp struct {
	baseresp
	Moderators []string
}

var meT meTemplate

func (t *meTemplate) exec(w io.Writer, me meresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", me))
}

type meTemplate struct {
	template *template.Template
}

type meresp struct {
	baseresp
	Username     string
	RequiresCode bool
}

var loginT loginTemplate

func (t *loginTemplate) exec(w io.Writer, login loginresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", login))
}

type loginTemplate struct {
	template *template.Template
}

type loginresp struct {
	justbaseresp
	Link string
}

var accountT accountTemplate

func (t *accountTemplate) exec(w io.Writer, account accountresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", account))
}

type accountTemplate struct {
	template *template.Template
}

type accountresp struct {
	justbaseresp
	Invite       string
	RequiresCode bool
	Link         string
}

var beepT beepTemplate

func (t *beepTemplate) exec(w io.Writer, beep beepresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", beep))
}

type beepTemplate struct {
	template *template.Template
}

type beepresp struct {
	baseresp
}

var threadT threadTemplate

func (t *threadTemplate) exec(w io.Writer, thread threadresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", thread))
}

type threadTemplate struct {
	template *template.Template
}

type threadresp struct {
	baseresp
	Thread   *types.Thread
	Archived bool
	Watched  bool
}

var newthreadT newthreadTemplate

func (t *newthreadTemplate) exec(w io.Writer, newthread newthreadresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", newthread))
}

type newthreadTemplate struct {
	template *template.Template
}

type newthreadresp struct {
	baseresp
}

var bumpedT bumpedTemplate

func (t *bumpedTemplate) exec(w io.Writer, tbumped bumpedresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "bumped-threads", tbumped))
}

type bumpedTemplate struct {
	template *template.Template
}

type bumpedresp struct {
	Threads []types.Thread
}

var threadsT threadsTemplate

func (t *threadsTemplate) exec(w io.Writer, threads catalogthreadsresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", threads))
}

type threadsTemplate struct {
	template *template.Template
}

var catalogT catalogTemplate

func (t *catalogTemplate) exec(w io.Writer, catalog catalogthreadsresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", catalog))
}

type catalogTemplate struct {
	template *template.Template
}

type catalogthreadsresp struct {
	baseresp
	IsChrono     bool
	ChronoCursor *uint32
	BumpCursor   *time.Time
	ThreadThumbs []types.Thread
}

var patchT patchTemplate

func (t *patchTemplate) exec(w io.Writer, patch patchresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", patch))
}

type patchTemplate struct {
	template *template.Template
}

type patchresp struct {
	baseresp
	Patches []types.Patch
}

var codeT codeTemplate

func (t *codeTemplate) exec(w io.Writer, code coderesp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", code))
}

type codeTemplate struct {
	template *template.Template
}

type coderesp struct {
	Code string
}

var codeerrT codeerrTemplate

func (t *codeerrTemplate) exec(w io.Writer, codeerr codeerrresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", codeerr))
}

type codeerrTemplate struct {
	template *template.Template
}

type codeerrresp struct {
	Reason string
}

var banT banTemplate

func (t *banTemplate) exec(w io.Writer, ban banresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", ban))
}

func (t *banTemplate) appeal(w io.Writer) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "appeal-submitted", nil))
}

type banTemplate struct {
	template *template.Template
}

type banresp struct {
	justbaseresp
	IsMod bool
	Ban   types.Ban
	Post  *types.Post
}

var profileT profileTemplate

func (t *profileTemplate) exec(w io.Writer, profile profileresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", profile))
}

type profileTemplate struct {
	template *template.Template
}

type profileresp struct {
	baseresp
	Profile *types.Profile
}

var editprofileT editprofileTemplate

func (t *editprofileTemplate) exec(w io.Writer, editprofile editprofileresp) {
	clog.Tmpl(t.template.ExecuteTemplate(w, "base", editprofile))
}

type editprofileTemplate struct {
	template *template.Template
}

type editprofileresp struct {
	baseresp
	Profile *types.Profile
}
