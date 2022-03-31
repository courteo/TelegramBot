package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Тут писать SearchServer
type compare func(user1 User, user2 User, orderField string, orderBy int) bool

type Data struct {
	Id        int    `xml:"id"`
	Age       int    `xml:"age"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type DataSet struct {
	Dt []Data `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != "123321213321" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		file, _ := json.Marshal(SearchErrorResponse{Error: "limit must me int"})
		_, err := w.Write(file)
		if err != nil {
			return
		}
		//w.Write()
		return
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		file, _ := json.Marshal(SearchErrorResponse{Error: "offset must me int"})
		_, err := w.Write(file)
		if err != nil {
			return
		}

		return
	}

	query := r.URL.Query().Get("query")

	orderField := r.URL.Query().Get("order_field")
	if orderField == "" {
		orderField = "Name"
	} else if orderField != "Name" && orderField != "ID" && orderField != "Age" {
		w.WriteHeader(http.StatusBadRequest)

		var er string
		if orderField == "Neme" {
			er = ErrorBadOrderField
		}

		file, _ := json.Marshal(SearchErrorResponse{Error: er})
		_, err := w.Write(file)
		if err != nil {
			return
		}
		return
	}

	orderByStr := r.URL.Query().Get("order_by")
	orderBy, err := strconv.Atoi(orderByStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		file, _ := json.Marshal(SearchErrorResponse{Error: "orderBy must me int"})
		_, err := w.Write(file)
		if err != nil {
			return
		}

		return
	}

	if orderBy != 0 && orderBy != 1 && orderBy != -1 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ForDataSet := new(DataSet)
	dataset, _ := os.Open("dataset.xml")

	data, _ := ioutil.ReadAll(dataset)

	err = xml.Unmarshal(data, &ForDataSet)
	if err != nil {
		return
	}

	response := SearchResponse{}
	for _, temp := range (*ForDataSet).Dt {
		response.Users = append(response.Users, User{
			ID:     temp.Id,
			Name:   temp.FirstName + temp.LastName,
			Age:    temp.Age,
			About:  temp.About,
			Gender: temp.Gender,
		})
	}

	if orderBy != 0 {
		//sort.Slice(response.Users, Compare)
		response.Users = mergeSort(response.Users, Compare, orderField, orderBy)
	}

	if len(response.Users) > offset {
		var users []User
		var res []User

		if limit+offset > len(response.Users) {
			users = response.Users[offset : len(response.Users)-1]
		} else {
			users = response.Users[offset : offset+limit+1]
		}

		if query != "" {
			for _, temp := range users {
				if strings.Contains(temp.Name, query) || strings.Contains(temp.About, query) {
					res = append(res, temp)
				}
			}
			response.Users = res
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, _ := json.Marshal(response.Users)

	if len(response.Users) == 0 {
		_, err := w.Write([]byte(`error`))
		if err != nil {
			return
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(file)

	}

}

func Compare(user1 User, user2 User, orderField string, orderBy int) bool {

	if orderBy == 1 {
		switch orderField {
		case "Name":
			return user1.Name < user2.Name
		case "ID":
			return user1.ID < user2.ID
		default:
			return user1.Age < user2.Age
		}

	} else {
		switch orderField {
		case "Name":
			return user1.Name > user2.Name
		case "ID":
			return user1.ID > user2.ID
		default:
			return user1.Age > user2.Age
		}

	}
}

func mergeSort(items []User, cmp compare, orderField string, orderBy int) []User {
	if len(items) < 2 {
		return items
	}
	first := mergeSort(items[:len(items)/2], cmp, orderField, orderBy)
	second := mergeSort(items[len(items)/2:], cmp, orderField, orderBy)
	return merge(first, second, cmp, orderField, orderBy)
}

func merge(a []User, b []User, cmp compare, orderField string, orderBy int) []User {
	final := []User{}
	i := 0
	j := 0
	for i < len(a) && j < len(b) {
		if cmp(a[i], b[j], orderField, orderBy) {
			final = append(final, a[i])
			i++
		} else {
			final = append(final, b[j])
			j++
		}
	}
	for ; i < len(a); i++ {
		final = append(final, a[i])
	}
	for ; j < len(b); j++ {
		final = append(final, b[j])
	}
	return final
}
