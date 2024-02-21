package service

import (
	elementsrpc "github.com/rddl-network/elements-rpc"
	"github.com/rddl-network/elements-rpc/types"
)

type IElementsClient interface {
	GetTransaction(url string, params []string) (transactionResult types.GetTransactionResult, err error)
	DeriveAddresses(url string, params []string) (addresses types.DeriveAddressesResult, err error)
	GetNewAddress(url string, params []string) (address string, err error)
	GetAddressInfo(url string, params []string) (info types.GetAddressInfoResult, err error)
}

type ElementsClient struct{}

func NewElementsClient() *ElementsClient {
	return &ElementsClient{}
}

func (ec *ElementsClient) GetTransaction(url string, params []string) (transactionResult types.GetTransactionResult, err error) {
	return elementsrpc.GetTransaction(url, params)
}

func (ec *ElementsClient) DeriveAddresses(url string, params []string) (addresses types.DeriveAddressesResult, err error) {
	return elementsrpc.DeriveAddresses(url, params)
}

func (ec *ElementsClient) GetNewAddress(url string, params []string) (address string, err error) {
	return elementsrpc.GetNewAddress(url, params)
}

func (ec *ElementsClient) GetAddressInfo(url string, params []string) (info types.GetAddressInfoResult, err error) {
	return elementsrpc.GetAddressInfo(url, params)
}
