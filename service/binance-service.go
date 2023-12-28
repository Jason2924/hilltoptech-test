package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	bc "github.com/binance/binance-connector-go"
	"github.com/gorilla/websocket"
)

var (
	bncConn *websocket.Conn
	bncOnce sync.Once
)

type BinanceService interface {
	Connect() (*websocket.Conn, error)
	GetPrice(from, to string) (float64, error)
}

type binanceService struct {
	WsUrl string
}

func NewBinanceService(wsUrl string) BinanceService {
	return &binanceService{
		WsUrl: wsUrl,
	}
}

func (serv *binanceService) Connect() (*websocket.Conn, error) {
	var erro error
	bncOnce.Do(func() {
		var err error
		bncConn, _, err = websocket.DefaultDialer.Dial(serv.WsUrl, nil)
		if err != nil {
			erro = err
			return
		}
	})
	return bncConn, erro
}

func (serv *binanceService) GetPrice(from, to string) (float64, error) {
	if bncConn == nil {
		var err error
		bncConn, err = serv.Connect()
		if err != nil {
			return 0, fmt.Errorf("Error connect websocket:", err)
		}
		lower := strings.ToLower(from) + strings.ToLower(to)
		if err := bncConn.WriteMessage(websocket.TextMessage, []byte(`{"method":"SUBSCRIBE","params":["`+lower+`@trade"],"id": 1}`)); err != nil {
			return 0, fmt.Errorf("Error writing message:", err)
		}
	}
	_, message, err := bncConn.ReadMessage()
	if err != nil {
		return 0, fmt.Errorf("Error reading message:", err)
	}
	stream := &bc.WsAggTradeEvent{}
	json.Unmarshal(message, stream)
	if stream.Price != "" {
		price, err := strconv.ParseFloat(stream.Price, 64)
		if err != nil {
			return 0, fmt.Errorf("Error parse flooat:", err)
		}
		return price, nil
	}
	return 0, nil
}
