package MessageConverter

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"os"
)

func MessageToJson(message *p_buff.Message, pathToJson string) {
	data, err := protojson.Marshal(message)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	fail := os.WriteFile(pathToJson, data, 0644)
	if fail != nil {
		log.Fatal("Error Writing to file): ", fail)
	}
}

func JsonToMessage(path string) *p_buff.Message {
	/*Arg: path to json file of message
	returns protobuf message
	!! Warning !! Doesn't work with timestamp with format :
									"timestamp": { "seconds": "1668081620",
												  "nanos": "71501000" }
	*/
	message := p_buff.Message{}
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	fail := protojson.Unmarshal(content, &message)
	if fail != nil {
		log.Fatal("Error during Unmarshal(): ", fail)
	}
	return &message
}
