package listener

type AuthListener interface {
	OnWaitingConfirmation()
	OnAuthorized()
	OnError(err error)
}
