package gateway

import (
	"log"

	"github.com/it-chain/bifrost"
	"github.com/it-chain/bifrost/client"
	"github.com/it-chain/heimdall/key"
	"github.com/it-chain/midgard"
)

type ConnectionCommandHandler struct {
	store     *bifrost.ConnectionStore
	priKey    key.PriKey
	pubKey    key.PubKey
	publisher midgard.Publisher
}

func NewConnectionCommandHandler(store *bifrost.ConnectionStore, priKey key.PriKey, pubKey key.PubKey, publisher midgard.Publisher) *ConnectionCommandHandler {
	return &ConnectionCommandHandler{
		publisher: publisher, //grpc 인터페이스에서 이벤트를 발생시키기 위해 필요하다.
		store:     store,
		pubKey:    pubKey,
		priKey:    priKey,
	}
}

// 새로운 connection 이 생성되면 처리하는 함수이다.
func (c ConnectionCommandHandler) HandleConnectionCreate(command ConnectionCreateCommand) {

	log.Println(command)

	if command.Address == "" {
		return
	}

	clientOpt := client.ClientOpts{
		Ip:     command.Address,
		PriKey: c.priKey,
		PubKey: c.pubKey,
	}

	grpcOpt := client.GrpcOpts{
		TlsEnabled: false,
		Creds:      nil,
	}

	connection, err := client.Dial(command.Address, clientOpt, grpcOpt)

	if err != nil {

		c.publisher.Publish("Event", "Error", ErrorCreatedEvent{
			Err:   err.Error(),
			Event: "Connection fail to create",
		})

		return
	}

	err = c.publisher.Publish("Event", "Connection", ConnectionCreatedEvent{
		Address: connection.GetIP(),
		EventModel: midgard.EventModel{
			ID: connection.GetID(),
		},
	})

	if err != nil {
		log.Println(err.Error())
		return
	}

	connection.Handle(NewRequestHandler(c.publisher))
	c.store.AddConnection(connection)

	go func() {
		defer connection.Close()

		if err := connection.Start(); err != nil {
			connection.Close()
		}
		log.Printf("connection are deleted")
	}()
}

type MessageCommandHandler struct {
	store     *bifrost.ConnectionStore
	publisher midgard.Publisher
}

func NewMessageCommandHandler(store *bifrost.ConnectionStore, publisher midgard.Publisher) *MessageCommandHandler {
	return &MessageCommandHandler{
		store:     store,
		publisher: publisher,
	}
}

func (m MessageCommandHandler) HandleMessageDeliver(command MessageDeliverCommand) {

	for _, recipient := range command.Recipients {
		connection := m.store.GetConnection(bifrost.ConnID(recipient))

		if connection != nil {
			connection.Send(command.Body, command.Protocol, nil, nil)
		}
	}
}