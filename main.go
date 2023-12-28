package main

import (
	"context"
	"flag"
	"fmt"
	"hilltoptech-test/client"
	"hilltoptech-test/logger"
	"hilltoptech-test/proto"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	grpdAddr := flag.String("grpc", ":8000", "listen grpc transprt")

	grpcClient, err := client.NewGRPCClient(":8000")
	if err != nil {
		log.Fatal(err)
	}

	file, err := logger.SetLoggerFile("logger/test.log")
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			wgroup := sync.WaitGroup{}
			wgroup.Add(2)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			bncPrice, uniPrice := float64(0), float64(0)
			go func() {
				price, err := grpcClient.FetchPrice(ctx, &proto.PriceRequest{Platform: "binance", From: "ETH", To: "USDT"})
				if err != nil {
					log.Fatal(err)
				}
				bncPrice = price.Price
				wgroup.Done()
			}()
			go func() {
				price, err := grpcClient.FetchPrice(ctx, &proto.PriceRequest{Platform: "uniswap", From: "ETH", To: "USDT"})
				if err != nil {
					log.Fatal(err)
				}
				uniPrice = price.Price
				wgroup.Done()
			}()
			select {
			case <-ctx.Done():
				return
			default:
				wgroup.Wait()
				if bncPrice > 0 && uniPrice > 0 {
					if bncPrice > uniPrice {
						log.Printf("Binance has benefit orver Uniswap: %.4f > %.4f\n", bncPrice, uniPrice)
					} else if bncPrice < uniPrice {
						log.Printf("Uniswap has benefit orver Binance: %.4f < %.4f\n", bncPrice, uniPrice)
					}
				}
			}
		}
	}()

	go RunGRPCServer(*grpdAddr)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Shutting down...")
	defer func() {
		close(sigChan)
		file.Close()
	}()
}
