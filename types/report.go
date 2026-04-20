package types

type Report struct {
	Id         int
	Reporter   string
	Reported   string
	PostId     *uint32
	Post       *Post
	ForProfile bool
	Reason     *string
	ReviewedBy *string
}
