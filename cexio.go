// Golang API for cex.io trading
// Author: julien@hautefeuille.eu
// Contributor : Aleksandr Shepelev (https://github.com/AleksandrShepelev)
// Date: 09/03/2014
// Update: 12/04/2018
// Version: 0.4

package cexio

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type CexKey struct {
	Username   string
	Api_key    string
	Api_secret string
}

func (cexapi *CexKey) Nonce() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func (cexapi *CexKey) ToHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func (cexapi *CexKey) Signature() (string, string) {
	nonce := cexapi.Nonce()
	message := nonce + cexapi.Username + cexapi.Api_key
	signature := cexapi.ToHmac256(message, cexapi.Api_secret)
	return signature, nonce
}

func (cexapi *CexKey) GetMethod(u string) []byte {
	res, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func (cexapi *CexKey) PostMethod(u string, v url.Values) []byte {
	res, err := http.PostForm(u, v)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func (cexapi *CexKey) ApiCall(method string, id string, param map[string]string, private bool, opt string) []byte {
	var data []byte
	u := "https://cex.io/api/" + method + "/"
	w := "https://cex.io/api/ghash.io/" + method
	if len(opt) != 0 {
		u = u + opt + "/"
	}
	if private {
		// Post method for private method
		signature, nonce := cexapi.Signature()
		v := url.Values{}
		v.Set("key", cexapi.Api_key)
		v.Add("signature", signature)
		v.Add("nonce", nonce)
		// Place order param
		if len(param) != 0 {
			if param["ordertype"] == "market" {
				v.Add("order_type", param["ordertype"])
				v.Add("type", param["type"])
				v.Add("amount", param["amount"])
			} else if param["ordertype"] == "limit" {
				v.Add("order_type", param["ordertype"])
				v.Add("type", param["type"])
				v.Add("amount", param["amount"])
				v.Add("price", param["price"])
			}
		}
		// Cancel order id
		if len(id) != 0 {
			v.Add("id", id)
		}
		v.Encode()
		if method == "workers" || method == "hashrate" {
			// Ghash.io post method
			data = cexapi.PostMethod(w, v) // url ghash.io , param
		} else {
			// Cex.io post method
			data = cexapi.PostMethod(u, v) // url cex.io, param
		}
	} else {
		// Get method for public method
		data = cexapi.GetMethod(u)
	}
	return data
}

// Public functions
func (cexapi *CexKey) Ticker(opt string) []byte {
	return cexapi.ApiCall("ticker", "", map[string]string{}, false, opt)
}

func (cexapi *CexKey) OrderBook(opt string) []byte {
	return cexapi.ApiCall("order_book", "", map[string]string{}, false, opt)
}

func (cexapi *CexKey) TradeHistory(opt string) []byte {
	return cexapi.ApiCall("trade_history", "", map[string]string{}, false, opt)
}

// Private functions
func (cexapi *CexKey) Balance() []byte {
	return cexapi.ApiCall("balance", "", map[string]string{}, true, "")
}

func (cexapi *CexKey) OpenOrders(opt string) []byte {
	return cexapi.ApiCall("open_orders", "", map[string]string{}, true, opt)
}

// Orders functions
func (cexapi *CexKey) PlaceLimitOrder(ordertype string, amount string, price string, opt string) []byte {
	var param = map[string]string{
		"ordertype": "limit",
		"type":      ordertype,
		"amount":    amount,
		"price":     price}
	return cexapi.ApiCall("place_order", "", param, true, opt)
}

func (cexapi *CexKey) PlaceMarketOrder(ordertype string, amount string, opt string) []byte {
	var param = map[string]string{
		"ordertype": "market",
		"type":      ordertype,
		"amount":    amount}
	return cexapi.ApiCall("place_order", "", param, true, opt)
}

func (cexapi *CexKey) CancelOrder(id string) []byte {
	return cexapi.ApiCall("cancel_order", id, map[string]string{}, true, "")
}

// Workers functions
func (cexapi *CexKey) Hashrate() []byte {
	return cexapi.ApiCall("hashrate", "", map[string]string{}, true, "")
}

func (cexapi *CexKey) Workers() []byte {
	return cexapi.ApiCall("workers", "", map[string]string{}, true, "")
}
