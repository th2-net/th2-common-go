package message

type MessageGroupBatchSubscriber interface {
	Start() error
	//isClose()
	AddListener(listener *MessageListener)
	//close()
}
