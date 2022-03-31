package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

// тут писать код тестов

func TestRequest(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	searchClient := SearchClient{URL: ts.URL, AccessToken: "123321213321"}

	requests := GetRequests()
	response := GetResponse()

	for j, i := range requests {
		values, err := searchClient.FindUsers(i)
		if err != nil {
			fmt.Errorf("wrong request")
		}
		if !reflect.DeepEqual(values, response[j]) {
			fmt.Errorf("wrong answer")
		}
	}

	ts.Close()
}

func TestBadType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	params := GetParams("qwe", "eas", "q", "s", "swq")
	searcherReq, err := http.NewRequest("GET", ts.URL+"?"+params.Encode(), nil) //nolint:errcheck
	if err != nil {
		fmt.Errorf("wrong request")
	}
	searcherReq.Header.Add("AccessToken", "123321213321")

	params1 := GetParams("12", "2", "qWqqqqq", "Name", "1q2")
	searcherReq1, err := http.NewRequest("GET", ts.URL+"?"+params1.Encode(), nil) //nolint:errcheck
	if err != nil {
		fmt.Errorf("wrong request")
	}
	searcherReq1.Header.Add("AccessToken", "123321213321")

	params2 := GetParams("12", "we", "q", "Name", "1q2")
	searcherReq2, err := http.NewRequest("GET", ts.URL+"?"+params2.Encode(), nil) //nolint:errcheck
	if err != nil {
		fmt.Errorf("wrong request")
	}
	searcherReq2.Header.Add("AccessToken", "123321213321")

	searchReq := []http.Request{*searcherReq, *searcherReq1, *searcherReq2}
	Erorrs := []string{"limit must be int", "orderBy must be int", "offset must be int"}

	for index, i := range searchReq {
		resp, do := ts.Client().Do(&i)
		if do != nil {
			fmt.Errorf(Erorrs[index])
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Errorf("wrong request")
		}

		errResp := SearchErrorResponse{}
		json.Unmarshal(body, &errResp)
		if errResp.Error != Erorrs[index] {
			fmt.Errorf("wrong answer")
		}
	}

	ts.Close()
}

func TestBadRequest(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	badRequests := GetBadRequest()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 100)
	}))
	searchClient := []SearchClient{
		SearchClient{URL: ts.URL, AccessToken: "123321213321"},
		SearchClient{URL: "qwe.co]m/courteo", AccessToken: "123321213321"},
		SearchClient{URL: ts.URL, AccessToken: "12"},
		SearchClient{URL: ts2.URL, AccessToken: "12"},
	}

	for i := range searchClient {
		if i == 0 {
			for _, i := range badRequests {
				value, err := searchClient[0].FindUsers(i)
				if err != nil {
					fmt.Errorf("wrong request")
				}
				if value != nil {
					fmt.Errorf("wrong answer")
				}
			}
		} else {
			value, err := searchClient[i].FindUsers(badRequests[0])
			if err != nil {
				fmt.Errorf("wrong request")
			}
			if value != nil {
				fmt.Errorf("wrong answer")
			}
		}
	}

	ts.Close()
}

func GetParams(limit, offset, query, orderField, orderBy string) url.Values {
	params1 := url.Values{}
	params1.Add("limit", limit)
	params1.Add("offset", offset)
	params1.Add("query", query)
	params1.Add("order_field", orderField)
	params1.Add("order_by", orderBy)
	return params1
}

func GetResponse() []SearchResponse {
	searchResponse2 := SearchResponse{
		Users: []User{
			User{
				ID:     26,
				Name:   "SimsCotton",
				Age:    39,
				About:  "Ex cupidatat est velit consequat ad. Tempor non cillum labore non voluptate. Et proident culpa labore deserunt ut aliquip commodo laborum nostrud. Anim minim occaecat est est minim.\n.",
				Gender: "male",
			},
			User{
				ID:     6,
				Name:   "JenningsMays",
				Age:    39,
				About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	searchResponse3 := SearchResponse{
		Users: []User{
			User{
				ID:     26,
				Name:   "SimsCotton",
				Age:    39,
				About:  "Ex cupidatat est velit consequat ad. Tempor non cillum labore non voluptate. Et proident culpa labore deserunt ut aliquip commodo laborum nostrud. Anim minim occaecat est est minim.\n.",
				Gender: "male",
			},
			User{
				ID:     9,
				Name:   "RoseCarney",
				Age:    36,
				About:  "Voluptate ipsum ad consequat elit ipsum tempor irure consectetur amet. Et veniam sunt in sunt ipsum non elit ullamco est est eu. Exercitation ipsum do deserunt do eu adipisicing id deserunt duis nulla ullamco eu. Ad duis voluptate amet quis commodo nostrud occaecat minim occaecat commodo. Irure sint incididunt est cupidatat laborum in duis enim nulla duis ut in ut. Cupidatat ex incididunt do ullamco do laboris eiusmod quis nostrud excepteur quis ea.\n",
				Gender: "female",
			},
		},
		NextPage: false,
	}
	searchResponse4 := SearchResponse{
		Users: []User{
			User{
				ID:     32,
				Name:   "ChristyKnapp",
				Age:    40,
				About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
				Gender: "female",
			},
			User{
				ID:     31,
				Name:   "PalmerScott",
				Age:    37,
				About:  "Elit fugiat commodo laborum quis eu consequat. In velit magna sit fugiat non proident ipsum tempor eu. Consectetur exercitation labore eiusmod occaecat adipisicing irure consequat fugiat ullamco aliquip nostrud anim irure enim. Duis do amet cillum eiusmod eu sunt. Minim minim sunt sit sit enim velit sint tempor enim sint aliquip voluptate reprehenderit officia. Voluptate magna sit consequat adipisicing ut eu qui.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	return []SearchResponse{searchResponse2, searchResponse3, searchResponse4}
}

func GetRequests() []SearchRequest {

	request6 := SearchRequest{Limit: 2,
		Offset:     2,
		Query:      "E",
		OrderField: "Age",
		OrderBy:    -1}

	request7 := SearchRequest{Limit: 2,
		Offset:     2,
		Query:      "E",
		OrderField: "Name",
		OrderBy:    -1}

	request8 := SearchRequest{Limit: 2,
		Offset:     2,
		Query:      "E",
		OrderField: "ID",
		OrderBy:    -1}

	res := []SearchRequest{request6, request7, request8}
	return res
}

func GetBadRequest() []SearchRequest {
	request1 := SearchRequest{Limit: 21,
		Offset:     10,
		Query:      "q",
		OrderField: "Name",
		OrderBy:    0}

	request2 := SearchRequest{Limit: -1,
		Offset:     2,
		Query:      "Enim",
		OrderField: "Name",
		OrderBy:    0}

	request3 := SearchRequest{Limit: 7,
		Offset:     -1,
		Query:      "Enim",
		OrderField: "Name",
		OrderBy:    0}

	request4 := SearchRequest{Limit: 26,
		Offset:     78,
		Query:      "Enim",
		OrderField: "Name",
		OrderBy:    0}

	request5 := SearchRequest{Limit: 27,
		Offset:     26,
		Query:      "Bo",
		OrderField: "",
		OrderBy:    0}

	request9 := SearchRequest{Limit: 10,
		Offset:     2,
		Query:      "E",
		OrderField: "Age",
		OrderBy:    1}

	request10 := SearchRequest{Limit: 27,
		Offset:     0,
		Query:      "ayta",
		OrderField: "Name",
		OrderBy:    -1}

	request11 := SearchRequest{Limit: 1,
		Offset:     2,
		Query:      "Erqereqrrqerewerqwerq",
		OrderField: "ID",
		OrderBy:    1}

	request12 := SearchRequest{Limit: 0,
		Offset:     2,
		Query:      "Bo",
		OrderField: "Name",
		OrderBy:    1}

	request13 := SearchRequest{Limit: 0,
		Offset:     2,
		Query:      "Bo",
		OrderField: "Name",
		OrderBy:    -21}

	request16 := SearchRequest{Limit: 127, // unknown request
		Offset:     26,
		Query:      "q",
		OrderField: "qqwe",
		OrderBy:    0}
	request17 := SearchRequest{Limit: 127, // invalid orderField
		Offset:     26,
		Query:      "q",
		OrderField: "Neme",
		OrderBy:    0}
	res := []SearchRequest{
		request1,
		request2,
		request3,
		request4,
		request5,
		request9,
		request10,
		request11,
		request12,
		request13,
		request16,
		request17,
	}
	return res
}
