package MQcommon

type Monitor interface {
	Unsubscribe() error
}

type Delivery struct {
	Redelivered bool
}

type Confirmation interface {
	Confirm() error
	Reject() error
}

type CloseListener interface {
	OnClose() error
}
