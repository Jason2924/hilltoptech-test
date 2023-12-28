package client

import (
	"hilltoptech-test/proto"

	"google.golang.org/grpc"
)

func NewGRPCClient(remoteAddr string) (proto.PriceFetcherClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := proto.NewPriceFetcherClient(conn)

	return c, nil
}
