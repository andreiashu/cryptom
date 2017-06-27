package krakenapi

import (
	"encoding/base64"
	"net/url"
	"reflect"
	"testing"
	"net/http/httptest"
	"net/http"
	"fmt"
	"io/ioutil"
)

var publicAPI = New("", "")

func TestCreateSignature(t *testing.T) {
	expectedSig := "Uog0MyIKZmXZ4/VFOh0g1u2U+A0ohuK8oCh0HFUiHLE2Csm23CuPCDaPquh/hpnAg/pSQLeXyBELpJejgOftCQ=="
	urlPath := "/0/private/"
	secret, _ := base64.StdEncoding.DecodeString("SECRET")
	values := url.Values{
		"TestKey": {"TestValue"},
	}

	sig := createSignature(urlPath, values, secret)

	if sig != expectedSig {
		t.Errorf("Expected Signature to be %s, got: %s\n", expectedSig, sig)
	}
}

func TestTime(t *testing.T) {
	resp, err := publicAPI.Time()
	if err != nil {
		t.Errorf("Time() should not return an error, got %s", err)
	}

	if resp.Unixtime <= 0 {
		t.Errorf("Time() should return valid Unixtime, got %d", resp.Unixtime)
	}
}

func TestAssets(t *testing.T) {
	_, err := publicAPI.Assets()
	if err != nil {
		t.Errorf("Assets() should not return an error, got %s", err)
	}
}

func TestAssetPairs(t *testing.T) {
	resp, err := publicAPI.AssetPairs()
	if err != nil {
		t.Errorf("AssetPairs() should not return an error, got %s", err)
	}

	if resp.XXBTZEUR.Base+resp.XXBTZEUR.Quote != XXBTZEUR {
		t.Errorf("AssetPairs() should return valid response, got %+v", resp.XXBTZEUR)
	}
}

func TestTicker(t *testing.T) {
	resp, err := publicAPI.Ticker(XXBTZGBP, XXBTZEUR)
	if err != nil {
		t.Errorf("Ticker() should not return an error, got %s", err)
	}

	if resp.XXBTZEUR.OpeningPrice == 0 {
		t.Errorf("Ticker() should return valid OpeningPrice, got %+v", resp.XXBTZEUR.OpeningPrice)
	}
}

func TestQueryTime(t *testing.T) {
	result, err := publicAPI.Query("Time", map[string]string{})
	resultKind := reflect.TypeOf(result).Kind()

	if err != nil {
		t.Errorf("Query should not return an error, got %s", err)
	}
	if resultKind != reflect.Map {
		t.Errorf("Query `Time` should return a Map, got: %s", resultKind)
	}
}

func TestQueryTicker(t *testing.T) {
	result, err := publicAPI.Query("Ticker", map[string]string{
		"pair": "XXBTZEUR",
	})
	resultKind := reflect.TypeOf(result).Kind()

	if err != nil {
		t.Errorf("Query should not return an error, got %s", err)
	}

	if resultKind != reflect.Map {
		t.Errorf("Query `Ticker` should return a Map, got: %s", resultKind)
	}
}

func TestQueryTrades(t *testing.T) {
	result, err := publicAPI.Trades(XXBTZEUR, 1495777604391411290)

	if err != nil {
		t.Errorf("Trades should not return an error, got %s", err)
	}

	if result.Last == 0 {
		t.Errorf("Returned parameter `last` should always have a value...")
	}

	if len(result.Trades) > 0 {
		for _, trade := range result.Trades {
			if trade.Buy == trade.Sell {
				t.Errorf("Trade should be buy or sell")
			}
			if trade.Market == trade.Limit {
				t.Errorf("Trade type should be market or limit")
			}
		}
	}
}

func TestOpenOrdersTrailingStop(t *testing.T) {
	resp := loadFixture(t, "./fixtures/open_orders_trailing_limit.json")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(resp))
	}))
	defer ts.Close()

	mockedApi := NewWithClient("", "", http.DefaultClient, &Config{
		url: ts.URL,
		apiversion: "0",
		ua: "Kraken GO API Agent test mode",
	})
	orders, err := mockedApi.OpenOrders(map[string]string{})

	if err != nil {
		t.Errorf("Got err for OpenOrders fetch: %s", err)
	}

	if len(orders.Open) != 1 {
		t.Errorf("Expected 1 open order got %d", len(orders.Open))
	}

	for id, order := range orders.Open {
		if id != "OHHGPP-NNI55-B4POPF" {
			t.Errorf("Unexpected order id")
		}

		if order.OpenTime != 1498459388.1265 {
			t.Errorf("Unexpected OpenTime")
		}

		if order.Description.PrimaryPrice != "-5.0000%" {
			t.Errorf("Unexpected PrimaryPrice in Description object")
		}

		if order.Description.SecondaryPrice != "-0.10000" {
			t.Errorf("Unexpected SecondaryPrice in Description object")
		}
	}

	// uncomment for debugging
	//j, err := json.Marshal(orders)
	//if err != nil {
	//	t.Errorf("Got err while marshalling orders: %s", err)
	//}
	//fmt.Printf("OpenOrders %s\n", string(j))
}

func TestOpenOrdersLimit(t *testing.T) {
	resp := loadFixture(t, "./fixtures/open_orders_limit.json")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(resp))
	}))
	defer ts.Close()

	mockedApi := NewWithClient("", "", http.DefaultClient, &Config{
		url: ts.URL,
		apiversion: "0",
		ua: "Kraken GO API Agent test mode",
	})
	orders, err := mockedApi.OpenOrders(map[string]string{})

	if err != nil {
		t.Errorf("Got err for OpenOrders fetch: %s", err)
	}

	if len(orders.Open) != 1 {
		t.Errorf("Expected 1 open order got %d", len(orders.Open))
	}

	for id, order := range orders.Open {
		if id != "OXVBVJ-EJOZJ-UP5E23" {
			t.Errorf("Unexpected order id")
		}

		if order.OpenTime != 1498464661.5428 {
			t.Errorf("Unexpected OpenTime")
		}

		if order.Description.PrimaryPrice != "320.00000" {
			t.Errorf("Unexpected PrimaryPrice in Description object")
		}

		if order.Description.SecondaryPrice != "0" {
			t.Errorf("Unexpected SecondaryPrice in Description object")
		}
	}

	// uncomment for debugging
	//j, err := json.Marshal(orders)
	//if err != nil {
	//	t.Errorf("Got err while marshalling orders: %s", err)
	//}
	//fmt.Printf("OpenOrders %s\n", string(j))
}

func TestTradeBalance(t *testing.T) {
	resp := loadFixture(t, "./fixtures/trade_balance.1.json")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(resp))
	}))
	defer ts.Close()

	mockedApi := NewWithClient("", "", http.DefaultClient, &Config{
		url: ts.URL,
		apiversion: "0",
		ua: "Kraken GO API Agent test mode",
	})

	balance, err := mockedApi.TradeBalance(map[string]string{})
	if err != nil {
		t.Errorf("Got error %s", err)
	}

	expected := &TradeBalanceResponse{
		EquivalentBalance: 3371.9760,
		TradeBalance: 3371.9760,
		MarginAmount: 0,
		UnrealizedNetProfit: 0,
		Cost: 0,
		Valuation: 0,
		Equity: 3371.9760,
		FreeMargin: 3371.9760,
	}

	if *balance != *expected {
		t.Errorf("Expected \n%+v \ngot \n%+v\n", expected, balance)
	}
}

func loadFixture(t *testing.T, path string) []byte {
	fixture := "./fixtures/trade_balance.1.json"
	resp, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Could not open fixture file %s", fixture)
	}

	return resp
}

