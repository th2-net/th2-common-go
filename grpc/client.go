package main

import (
	"context"
	ac "exactpro/th2/th2-common-go/grpc/proto" //act proto
	cg "exactpro/th2/th2-common-go/proto"      //common proto
	"github.com/google/uuid"
	"log"
	"time"
)

func _() {
	grpcRouter := CommonGrpcRouter{Config: GrpcConfig{}}
	cp := &ConfigProviderFromFile{DirectoryPath: "../resources"}
	cfgErr := cp.GetConfig(grpcJsonFileName, &grpcRouter.Config)
	if cfgErr != nil {
		log.Fatalf(cfgErr.Error())
	}
	con, conErr := grpcRouter.GetConnection("actAttr")
	if conErr != nil {
		log.Fatalf(conErr.Error())
	}
	//defer con.Close() //doesnt have that func
	c := ac.NewActClient(con)

	// getting data ready for placing order

	eventID := cg.EventID{Id: uuid.New().String()}

	tradingPartyFields := map[string]*cg.Value{
		"NoPartyIDs": {Kind: &cg.Value_ListValue{ListValue: &cg.ListValue{
			Values: []*cg.Value{{Kind: &cg.Value_MessageValue{MessageValue: &cg.Message{
				Metadata: &cg.MessageMetadata{MessageType: "TradingParty_NoPartyIDs"},
				Fields: map[string]*cg.Value{
					"PartyID":       {Kind: &cg.Value_SimpleValue{SimpleValue: "Trader1"}},
					"PartyIDSource": {Kind: &cg.Value_SimpleValue{SimpleValue: "D"}},
					"PartyRole":     {Kind: &cg.Value_SimpleValue{SimpleValue: "76"}},
				},
			}}},
				{Kind: &cg.Value_MessageValue{MessageValue: &cg.Message{
					Metadata: &cg.MessageMetadata{MessageType: "TradingParty_NoPartyIDs"},
					Fields: map[string]*cg.Value{
						"PartyID":       {Kind: &cg.Value_SimpleValue{SimpleValue: "0"}},
						"PartyIDSource": {Kind: &cg.Value_SimpleValue{SimpleValue: "D"}},
						"PartyRole":     {Kind: &cg.Value_SimpleValue{SimpleValue: "3"}},
					},
				},
				}},
			}}},
		},
	}

	fields := map[string]*cg.Value{
		"Side":             {Kind: &cg.Value_SimpleValue{SimpleValue: "1"}},
		"SecurityID":       {Kind: &cg.Value_SimpleValue{SimpleValue: "INSTR1"}},
		"SecurityIDSource": {Kind: &cg.Value_SimpleValue{SimpleValue: "8"}},
		"OrdType":          {Kind: &cg.Value_SimpleValue{SimpleValue: "2"}},
		"AccountType":      {Kind: &cg.Value_SimpleValue{SimpleValue: "1"}},
		"OrderCapacity":    {Kind: &cg.Value_SimpleValue{SimpleValue: "A"}},
		"OrderQty":         {Kind: &cg.Value_SimpleValue{SimpleValue: "100"}},
		"Price":            {Kind: &cg.Value_SimpleValue{SimpleValue: "10"}},
		"ClOrdID":          {Kind: &cg.Value_SimpleValue{SimpleValue: "123"}}, //random in py
		"SecondaryClOrdID": {Kind: &cg.Value_SimpleValue{SimpleValue: "2"}},   //random in py
		"TransactTime":     {Kind: &cg.Value_SimpleValue{SimpleValue: time.Now().Format(time.RFC3339)}},
		"TradingParty":     {Kind: &cg.Value_MessageValue{MessageValue: &cg.Message{Fields: tradingPartyFields}}},
	}

	msg := cg.Message{Metadata: &cg.MessageMetadata{
		MessageType: "NewOrderSingle",
		Id:          &cg.MessageID{ConnectionId: &cg.ConnectionID{SessionAlias: "demo-conn1"}},
	},
		Fields: fields,
	}

	request := ac.PlaceMessageRequest{
		Message:       &msg,
		ParentEventId: &eventID,
		Description:   "User places an order.",
	}

	resp, err := c.PlaceOrderFIX(context.Background(), &request)
	if err != nil {
		log.Fatalf("could not send order: %v", err)
	}
	log.Printf("order sent. response: %s", resp.String())
}
