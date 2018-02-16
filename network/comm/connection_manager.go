package comm

import (
	pb "it-chain/network/protos"
)

type OnError func(error)

//comm은 peer 들간의 connection을 유지하고있다.
//comm을 통해 peer들과 통신한다.
type ConnectionManager interface{

	SendStream(data *pb.StreamMessage, errorCallBack OnError, connectionID string)

	Stop()

	Close(connectionID string)

	CreateStreamClientConn(connectionID string, ip string, handle ReceiveMessageHandle) error

	Size() int

	//Server on function
	Stream(stream pb.StreamService_StreamServer) (error)
}