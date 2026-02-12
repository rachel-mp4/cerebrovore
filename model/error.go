package model

type ModelError string

const ErrThreadDNE = ModelError("requested thread does not exist")
const ErrServerDNE = ModelError("requested thread's server does not exist")

func (te ModelError) Error() string {
	return string(te)
}
