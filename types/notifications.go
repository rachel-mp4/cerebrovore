package types

type Notification struct {
	Id       int
	Username string
	// if the notification is a reply notification, this is the relevant post
	ReplyId *uint32
	Reply   *Post // post

	// if the notification is a mention notification, this is the relevant post
	MentionId   *uint32
	Mentioner   *string // username
	MentionAnon *bool
	MentionNick *string

	// if the notification is a thread watch notification, this is the relevant thread
	ThreadId *uint32
	Topic    *string // topic of the thread

	// if the notification is a poke notification, this is who poked
	Poker       *string
	PokeMessage *string

	// if the notification is a mod notification, this is the reason for notifying user
	Reason *string

	GetId *uint32
	Value *int

	IsLastRead bool
}
