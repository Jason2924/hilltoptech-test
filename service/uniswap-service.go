package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/netbound/dex-feed/db"
	uniswapv3 "github.com/netbound/dex-feed/exchange/uniswap_v3"
	"github.com/netbound/dex-feed/token"
)

var (
	uniConn *ethclient.Client
	uniOnce sync.Once
)

type UniswapService interface {
	Connect() (*ethclient.Client, error)
	GetPrice(ctx context.Context, from, to string) (float64, error)
}

type uniswapService struct {
	WsUrl string
}

func NewUniswapService(wsUrl string) UniswapService {
	key := os.Getenv("INFURA_API_KEY")
	wsUrl += "/" + key
	return &uniswapService{
		WsUrl: wsUrl,
	}
}

func (serv *uniswapService) Connect() (*ethclient.Client, error) {
	var erro error
	uniOnce.Do(func() {
		var err error
		uniConn, err = ethclient.Dial(serv.WsUrl)
		if err != nil {
			erro = fmt.Errorf("Error connecting to WebSocket:", err)
			return
		}
	})
	return uniConn, erro
}

func (serv *uniswapService) GetPrice(ctx context.Context, from, to string) (float64, error) {
	if uniConn == nil {
		var err error
		uniConn, err = serv.Connect()
		if err != nil {
			return 0, fmt.Errorf("Error connect websocket:", err)
		}
	}
	fromHex := common.HexToAddress(getTokenByCurrency(from))
	toHex := common.HexToAddress(getTokenByCurrency(to))
	factory, fee := getTokenAndFee(from, to)
	tokenManager := token.NewTokenDB(uniConn, db.Opts{})
	uni := uniswapv3.New(uniConn, tokenManager, common.HexToAddress(factory), db.Opts{})
	price, err := uni.GetPrice(ctx, fromHex, toHex, fee)
	if err != nil {
		return 0, fmt.Errorf("Get price error:", err)
	}
	return price, nil
}

func getTokenByCurrency(curr string) string {
	switch strings.ToLower(curr) {
	case "eth":
		return "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	case "usdt":
		return "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	default:
		return ""
	}
}

func getTokenAndFee(from, to string) (string, int64) {
	if strings.ToLower(from) == "eth" && strings.ToLower(to) == "usdt" {
		factory := "0x1F98431c8aD98523631AE4a59f267346ea31F984"
		fee := int64(500)
		return factory, fee
	}
	return "", 0
}
