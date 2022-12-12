package main

import (
	"flag"
	"fmt"

	"log"
	"os"
	"runtime"
	"runtime/pprof"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	p_buff "github.com/th2-net/th2-common-go/proto"
	factory "github.com/th2-net/th2-common-go/schema/factory"
	cfg "github.com/th2-net/th2-common-go/schema/message/configuration"
)

func main() {

	rabbit_cfg := flag.String("rabbitConfiguration", "", "Rabbit config file")
	rabbit_router_cfg := flag.String("messageRouterConfiguration", "", "Message router config file")

	cpuprofile := flag.String("cpu", "", "write cpu profile to `file`")
	memprofile := flag.String("mem", "", "write memory profile to `file`")

	flag.Parse()

	//Enable CPU profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	////

	var cf *factory.CommonFactory
	var err error

	fmt.Println("Run")

	if *rabbit_cfg == "" || *rabbit_router_cfg == "" {

		cf, err = factory.NewCommonFactory()
		if err != nil {
			panic(err)
		}

	} else {

		fmt.Println("main args: ", *rabbit_cfg)
		fmt.Println("main args: ", *rabbit_router_cfg)

		cf, err = factory.NewCommonFactoryFromArgs(*rabbit_cfg, *rabbit_router_cfg)
		if err != nil {
			panic(err)
		}
	}

	cf.Init()

	fmt.Println("Sending raw message")
	{
		raw_router, err := cf.GetMessageRouterRawBatch()
		raw_msg_batch := p_buff.RawMessageBatch{}

		raw_msg := p_buff.RawMessage{

			Metadata: &p_buff.RawMessageMetadata{

				Id: &p_buff.MessageID{

					ConnectionId: &p_buff.ConnectionID{
						SessionAlias: "session_alias",
					},

					Sequence: 0,
				},

				Timestamp: &timestamppb.Timestamp{
					Seconds: 1234,
					Nanos:   6578,
				},

				Properties: map[string]string{"requestId": "0", "requestRef": "1"},
			},

			Body: []byte("Message body"),
		}

		msg_id := raw_msg.Metadata.GetId()

		from_client := true
		if from_client {
			msg_id.Direction = p_buff.Direction_SECOND
		} else {
			msg_id.Direction = p_buff.Direction_FIRST
		}

		raw_msg_batch.Messages = append(raw_msg_batch.Messages, &raw_msg)

		//attr_first := cfg.QueueAttribute{"publish", "raw", "first", "store"}
		attr_second := cfg.QueueAttribute{"publish", "raw", "second", "store"}

		//err := (*raw_router).Send(&raw_msg_batch)
		err = (*raw_router).SendByQueueAttributes(&raw_msg_batch, attr_second)
		//err := (*raw_router).SendAll(&raw_msg_batch, attr_second)

		if err != nil {
			fmt.Println(err)
		}

	}

	//Enable memory profiling
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
	////
}
