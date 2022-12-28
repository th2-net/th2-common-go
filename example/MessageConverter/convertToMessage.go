package MessageConverter

import (
	"fmt"
	p_buff "github.com/th2-net/th2-common-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"reflect"
	"strconv"
)

func MessageFromStruct(obj *MessageStruct) *p_buff.Message {
	/*Arg : obj with type msg(Metadata,Fields,parentEventID)
	But we can adjust this function to pass directly those 3 values (Metadata, Fields, parentEventID)
	instead of whole struct.
	*/
	fields := make(map[string]*p_buff.Value) //Creating empty map to add message values later
	metadata := p_buff.MessageMetadata{}     //Create empty MessageMetadata variable
	var eventID *p_buff.EventID              // Declare *p_buff.EventID variable for message parentEventID
	for k, v := range parseMap(obj.Fields) { //Parse raw map from json
		fields[k] = valueConverter(v) //converting map type values into message values
	}
	metadata = *metadataConverter(obj.Metadata) // converting metadata into MessageMetadata
	if obj.ParentEventID != "" {
		eventID = &p_buff.EventID{Id: obj.ParentEventID}
	}
	message := p_buff.Message{Fields: fields, Metadata: &metadata, ParentEventId: eventID} //Creating Message
	return &message
}

func valueConverter(value interface{}) *p_buff.Value {
	/*Args: value with tye interface{} - Takes any type of variable and
	converts it into Message Value based on its type
	returns: *p_buff.Value - Pointer to Message value
	*/
	switch t := value.(type) {
	case string:
		//log.Printf("%T  \n", t)
		return &p_buff.Value{Kind: &p_buff.Value_SimpleValue{SimpleValue: value.(string)}}
	case int:
		//log.Printf("%T type value: %v \n", t, value)
		return &p_buff.Value{Kind: &p_buff.Value_SimpleValue{SimpleValue: strconv.Itoa(value.(int))}}
	case float64:
		//log.Printf("%T type value: %v \n", t, value)
		return &p_buff.Value{Kind: &p_buff.Value_SimpleValue{SimpleValue: strconv.FormatFloat(value.(float64), 'E', -1, 64)}}
	case bool:
		//log.Printf("%T type value: %v \n", t, value)
		return &p_buff.Value{Kind: &p_buff.Value_SimpleValue{SimpleValue: strconv.FormatBool(value.(bool))}}
	case nil:
		//log.Printf("%T type value: %v \n", t, value)
		return &p_buff.Value{Kind: &p_buff.Value_NullValue{NullValue: p_buff.NullValue_NULL_VALUE}}
	case *p_buff.Value:
		//log.Printf("%T type value: %v \n", t, value)
		return value.(*p_buff.Value)
	case *p_buff.Message:
		//log.Printf("%T type value: %v \n", t, value)
		return &p_buff.Value{Kind: &p_buff.Value_MessageValue{MessageValue: value.(*p_buff.Message)}}
	case map[string]interface{}:
		//log.Printf("%T type value: %v \n", t, value)
		fields := make(map[string]*p_buff.Value)
		for i, j := range value.(map[string]interface{}) {
			fields[i] = valueConverter(j)
		}
		return &p_buff.Value{Kind: &p_buff.Value_MessageValue{MessageValue: &p_buff.Message{Fields: fields}}}
	case *p_buff.ListValue:
		log.Printf("%T type \n", t)
		return &p_buff.Value{Kind: &p_buff.Value_ListValue{ListValue: value.(*p_buff.ListValue)}}
	case []interface{}:
		//log.Printf("%T type value: %v \n", t, value)
		var values []*p_buff.Value
		for _, i := range value.([]interface{}) {
			values = append(values, valueConverter(i))
		}
		return &p_buff.Value{Kind: &p_buff.Value_ListValue{ListValue: &p_buff.ListValue{Values: values}}}
	default:
		//log.Fatalf("not implemented for %v type", reflect.TypeOf(value))
		return &p_buff.Value{}
	}
}

func metadataConverter(meta metaData) *p_buff.MessageMetadata {
	/*Args: meta with type Metadata structure
	Creates MessageMetadata from argument's fields
	returns Pointer to MessageMetadata
	*/
	metadata := p_buff.MessageMetadata{Id: &p_buff.MessageID{}}
	if meta.Id.Sequence != "" {
		metadata.Id.ConnectionId = &p_buff.ConnectionID{SessionAlias: meta.Id.ConnectionID.SessionAlias}
	}
	if meta.Id.Direction != "" {
		if meta.Id.Direction == "FIRST" {
			metadata.Id.Direction = p_buff.Direction_FIRST
		} else if meta.Id.Direction == "SECOND" {
			metadata.Id.Direction = p_buff.Direction_SECOND
		}
	}
	if meta.Id.Sequence != "" || len(meta.Id.SubSequence) != 0 {
		sequence, err := strconv.ParseInt(meta.Id.Sequence, 10, 64)
		if err != nil {
			fmt.Printf("Can't parse string into int64 : %v", err)
		} else {
			metadata.Id.Sequence = sequence
		}
		metadata.Id.Subsequence = meta.Id.SubSequence
	}
	if meta.MessageType != "" {
		metadata.MessageType = meta.MessageType
	}
	if meta.Protocol != "" {
		metadata.Protocol = meta.Protocol
	}
	if len(meta.Properties) != 0 {
		metadata.Properties = meta.Properties
	}
	if meta.Timestamp.Nanos != "" || meta.Timestamp.Seconds != "" {
		nano, err := strconv.ParseInt(meta.Timestamp.Nanos, 10, 32)
		if err != nil {
			log.Printf("Can't parse '%v' into int32: %v \n", meta.Timestamp.Nanos, err)
		}
		seconds, er := strconv.ParseInt(meta.Timestamp.Seconds, 10, 64)
		if er != nil {
			log.Printf("Can't parse '%v' into int64: %v \n", meta.Timestamp.Seconds, err)
		}
		metadata.Timestamp = &timestamppb.Timestamp{Nanos: int32(nano), Seconds: seconds}
	}
	return &metadata
}

func parseMap(obj map[string]interface{}) map[string]interface{} {
	testMap := make(map[string]interface{})
	for key, value := range obj {
		var val interface{}
		if reflect.ValueOf(value).Kind() == reflect.Map {
			val = parseMapValue(value.(map[string]interface{}))
			testMap[key] = val
		} else {
			testMap[key] = value
		}

	}
	return testMap
}

func parseMapValue(value map[string]interface{}) interface{} {
	for k, v := range value {
		switch k {
		case "simpleValue":
			return v
		case "messageValue":
			for _, j := range v.(map[string]interface{}) {
				return parseMap(j.(map[string]interface{}))
			}
		case "listValue":
			for _, j := range v.(map[string]interface{}) {
				var res []interface{}
				val := j.([]interface{})
				for i := 0; i < len(val); i++ {
					res = append(res, parseMapValue((val[i]).(map[string]interface{})))
				}
				return res
			}
		}
	}
	return value
}
