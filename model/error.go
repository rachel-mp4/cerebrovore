package model

type ModelError string

const ErrThreadDNE = ModelError("requested thread does not exist")
const ErrThreadDead = ModelError("requested thread is dead, so this feature is not applicable")
const ErrThreadFull = ModelError("requested thread has hit the bump limit")
const ErrServerDNE = ModelError("requested thread's server does not exist")
const skipped = ModelError("wormwatch entry skipped")
const paused = ModelError("wormwatch entry paused")
const deleted = ModelError("thread was deleted")

func (te ModelError) Error() string {
	return string(te)
}
