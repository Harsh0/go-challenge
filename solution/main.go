package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"
)

/*Response json response struct */
type Response struct {
	Numbers []int `json:"numbers"`
}

/*HTTPResponse body and error struct */
type HTTPResponse struct {
	res Response
	err error
}

var testServerURL = "localhost:8090"
var maxResponseTime = 500 * time.Millisecond

func main() {
	listenAddr := flag.String("http.addr", ":8080", "http listen address")
	flag.Parse()

	http.HandleFunc("/numbers", numbers)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

func numbers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("---------new request came-----------")
	t := time.NewTimer(maxResponseTime)
	response := []int{}
	done := false
	//start timer in go routine for edge cases
	//if time exceed 500 milisecond send response without wait
	//where response pointer will be shared so that if response received from even one url
	//that can be sent
	go func(w http.ResponseWriter, response *[]int) {
		//receive timer after its expiration and fire send response function
		_ = <-t.C
		done = true
		SendReposne(w, response)
	}(w, &response)

	urls := r.URL.Query()["u"]

	urlsUpdated := []string{}
	//check whether url are syntactically valid
	//and update url host to run on local test server
	for _, u := range urls {
		if u, ok := IsSyntacticallyValid(u); ok {
			urlsUpdated = append(urlsUpdated, u)
		}
	}
	//send updated url to aggregate result to query test server to aggregate
	// result and assigned to response varibale whose pointer is passed
	AggregateResult(urlsUpdated, &response)
	//stop timer if result received in time to prevent multiple response sent to client
	t.Stop()
	if !done {
		SendReposne(w, &response)
	}
}

/*IsSyntacticallyValid this function will check whether url are syntactically valid
and change their host with testserverurl  which is localhost:8090*/
func IsSyntacticallyValid(urlToCheck string) (string, bool) {
	u, err := url.Parse(urlToCheck)
	if u.Host == "" {
		//host is not there
		return "", false
	}
	u.Host = testServerURL
	if err != nil {
		return "", false
	}
	return u.String(), true
}

/*SendReposne this function will sort the array and remove duplicate element
and send the resposne to client */
func SendReposne(w http.ResponseWriter, response *[]int) {
	sort.Ints(*response)
	var prev int
	res := []int{}
	for i, num := range *response {
		if i == 0 {
			prev = num
			res = append(res, num)
			continue
		}
		if prev == num {
			continue
		} else {
			res = append(res, num)
			prev = num
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"numbers": res})
}

/*AggregateResult this function fire multiple go routine to parallel query to server
for getting result and assign to response variable whose pointer is passed as argument */
func AggregateResult(urls []string, response *[]int) {
	httpChan := make(chan *Response)
	for _, u := range urls {
		//fire go routine for every request
		go GetRequest(u, httpChan)
	}
	chanLength := len(urls)
	//wait for response to come from channel
	for chanLength > 0 {
		chanLength--
		res := <-httpChan
		//add data to response
		*response = append(*response, res.Numbers...)
	}
}

/*GetRequest this function will make 'GET' request to server url and
send response through channel variable passed as argument of type HTTPResponse*/
func GetRequest(url string, httpChan chan *Response) {
	//init client for max timeout of max Response time
	var client = &http.Client{Timeout: maxResponseTime}
	r, err := client.Get(url)
	//close request after all ops
	//create new instance of Response type
	var target = new(Response)
	if err != nil {
		//timeout error or network error
		fmt.Println(err.Error())
		httpChan <- target
		return
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusServiceUnavailable {
		//pass empty target
		httpChan <- target
		fmt.Println("service unavailable error")
		return
	}
	if r.StatusCode != http.StatusOK {
		//pass empty target
		httpChan <- target
		fmt.Println("unknown error")
		return
	}
	err = json.NewDecoder(r.Body).Decode(target)
	httpChan <- target
}
