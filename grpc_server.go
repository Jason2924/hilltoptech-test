package main

import (
	"context"
	"fmt"
	"hilltoptech-test/proto"
	"hilltoptech-test/service"
	"net"

	"google.golang.org/grpc"
)

func RunGRPCServer(listenAddr string) error {
	// get binance service here
	binanceService := service.NewBinanceService("wss://stream.binance.com:9443/ws")
	uniswapService := service.NewUniswapService("wss://mainnet.infura.io/ws/v3")
	grpcServer := NewGRPCPriceFetcherServer(binanceService, uniswapService)
	listen, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	options := []grpc.ServerOption{}
	server := grpc.NewServer(options...)
	proto.RegisterPriceFetcherServer(server, grpcServer)
	return server.Serve(listen)
}

type GRPCPriceFetcherServer struct {
	binanceService service.BinanceService
	uniswapService service.UniswapService
	proto.UnimplementedPriceFetcherServer
}

func NewGRPCPriceFetcherServer(bncServ service.BinanceService, uniServ service.UniswapService) *GRPCPriceFetcherServer {
	return &GRPCPriceFetcherServer{
		binanceService: bncServ,
		uniswapService: uniServ,
	}
}

func (sver *GRPCPriceFetcherServer) FetchPrice(ctx context.Context, reqt *proto.PriceRequest) (*proto.PriceResponse, error) {
	if reqt.Platform == "binance" {
		price, err := sver.binanceService.GetPrice(reqt.From, reqt.To)
		if err != nil {
			return nil, err
		}
		resp := &proto.PriceResponse{
			Price: price,
		}
		return resp, nil
	} else if reqt.Platform == "uniswap" {
		price, err := sver.uniswapService.GetPrice(ctx, reqt.From, reqt.To)
		if err != nil {
			return nil, err
		}
		resp := &proto.PriceResponse{
			Price: price,
		}
		return resp, nil
	}
	return nil, fmt.Errorf("platform not exist")
}
