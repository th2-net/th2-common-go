package MessageConverter

type MessageStruct struct {
	Metadata      metaData               `json:"metadata,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
	ParentEventID string                 `json:"parentEventID,omitempty"`
}
type metaData struct {
	Id          id                `json:"id,omitempty"`
	MessageType string            `json:"messageType,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
	Protocol    string            `json:"protocol,omitempty"`
	Timestamp   timeStamp         `json:"timestamp,omitempty"`
}

type timeStamp struct {
	Nanos   string `json:"nanos,omitempty"`
	Seconds string `json:"seconds,omitempty"`
}
type connectionID struct {
	SessionAlias string `json:"sessionAlias,omitempty"`
}

type id struct {
	ConnectionID connectionID `json:"connectionId,omitempty"`
	Direction    string       `json:"direction,omitempty"`
	Sequence     string       `json:"sequence,omitempty"`
	SubSequence  []uint32     `json:"subsequence,omitempty"`
}
