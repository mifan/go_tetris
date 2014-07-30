package main

import (
	"fmt"
	"time"

	"github.com/conformal/btcjson"
	"github.com/gogames/go_tetris/types"
)

var (
	// block hash
	semiBH = ""
	lastBH = ""
	// error
	errNewAddrNotString               = fmt.Errorf("生成了无效的比特币地址")
	errCantConvToTxid                 = fmt.Errorf("无法获取交易id")
	errCantValidate                   = fmt.Errorf("无法验证比特币地址")
	errCantConvToListSinceBlockResult = fmt.Errorf("无法转换成所需数据")
)

var btcRpcCmd = func(msg []byte) (btcjson.Reply, error) {
	return btcjson.RpcCommand(btcUser, btcPass, btcServer, msg)
}

func initBitcoin() {
	go recv()
	log.Info("initialize bitcoin wallet...")
}

func recv() {
	for {
		func() {
			var err error
			defer func() {
				if err != nil {
					log.Error("can not deposit bitcoin, get error: %v", err)
				}
			}()
			res, err := listSinceBlock(lastBH)
			if err != nil {
				return
			}
			for _, v := range res.Transactions {
				switch v.Category {
				case "receive":
					u := users.GetByNickname(v.Account)
					if u == nil {
						continue
					}
					amount := int(v.Amount * 1000)
					// insert into database & update
					if err = insertDeposit(v.TxID, v.Account, v.Address, amount); err != nil {
						return
					}
					// update user cache
					if err = u.Update(types.NewUpdateInt(types.UF_Balance, u.GetBalance()+amount)); err != nil {
						return
					}
				}
			}
			if res.LastBlock != semiBH {
				lastBH = semiBH
				semiBH = res.LastBlock
				log.Info("semiBH changed to %v", semiBH)
				log.Info("lastBH changed to %v", lastBH)
			}
		}()
		time.Sleep(5 * time.Second)
	}
}

// get new address
func getNewAddress(nickname string) (addr string, err error) {
	msg, err := btcjson.CreateMessage("getnewaddress", nickname)
	if err != nil {
		return
	}
	reply, err := btcRpcCmd(msg)
	if err != nil {
		return
	}

	if reply.Error != nil {
		return "", fmt.Errorf("reply.Error is not nil: %v", reply.Error.Error())
	}

	addr, ok := reply.Result.(string)
	if !ok {
		fmt.Println("the result is: ", reply.Result, " can not convert to address")
		err = errNewAddrNotString
	}
	return
}

// withdraw bitcoin
func sendBitcoin(address string, amount int) (txid string, err error) {
	msg, err := btcjson.CreateMessage("sendtoaddress", address, float64(amount)/1e3)
	if err != nil {
		return
	}
	reply, err := btcRpcCmd(msg)
	if err != nil {
		return
	}

	if reply.Error != nil {
		return "", fmt.Errorf("reply.Error is not nil: %v", reply.Error.Error())
	}

	txid, ok := reply.Result.(string)
	if !ok {
		fmt.Println("the result is: ", reply.Result, " can not convert to txid")
		err = errCantConvToTxid
	}
	return
}

// validate address
func validateAddress(addr string) (isValid bool, err error) {
	msg, err := btcjson.CreateMessage("validateaddress", addr)
	if err != nil {
		return
	}

	reply, err := btcRpcCmd(msg)
	if err != nil {
		return
	}

	if reply.Error != nil {
		return false, fmt.Errorf("reply error is not nil: %v", reply.Error.Error())
	}

	v, ok := reply.Result.(*btcjson.ValidateAddressResult)
	if !ok {
		fmt.Println("the result is: ", reply.Result, " can not convert to *btcjson.ValidateAddressResult")
		err = errCantValidate
	}
	return v.IsValid, nil
}

// listsinceblock
func listSinceBlock(blockHash string) (res *btcjson.ListSinceBlockResult, err error) {
	msg, err := btcjson.CreateMessage("listsinceblock", blockHash)
	if err != nil {
		return
	}

	reply, err := btcRpcCmd(msg)
	if err != nil {
		return
	}

	if reply.Error != nil {
		return nil, fmt.Errorf("listSinceBlock error: %v", reply.Error.Error())
	}

	res, ok := reply.Result.(*btcjson.ListSinceBlockResult)
	if !ok {
		fmt.Println("the result is: ", reply.Result, " can not convert to *btcjson.ListSinceBlockResult")
		err = errCantConvToListSinceBlockResult
	}
	return
}
