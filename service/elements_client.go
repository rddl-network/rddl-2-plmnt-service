package service

import (
	elementsrpc "github.com/rddl-network/elements-rpc"
	"github.com/rddl-network/elements-rpc/types"
)

type IElementsClient interface {
	GetNewAddress(url string, params []string) (address string, err error)
	ListReceivedByAddress(url string, params []string) (receivedTx []types.ListReceivedByAddressResult, err error)
}

type ElementsClient struct{}

func NewElementsClient() *ElementsClient {
	return &ElementsClient{}
}

func (ec *ElementsClient) GetNewAddress(url string, params []string) (address string, err error) {
	return elementsrpc.GetNewAddress(url, params)
}

func (ec *ElementsClient) ListReceivedByAddress(url string, params []string) (receivedTx []types.ListReceivedByAddressResult, err error) {
	return elementsrpc.ListReceivedByAddress(url, params)
}
