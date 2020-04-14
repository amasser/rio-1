package examples

import (
	"context"
	"fmt"
	"github.com/susamn/rio"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/graphql", SampleHandler)
	http.ListenAndServe(":7070", nil)
}

func backEndCall1(id string) (name string) {
	time.Sleep(time.Duration(10) * time.Second)
	return "RIO"
}

func backEndCall2(name, locationId string) (streetAddress string) {
	time.Sleep(time.Duration(10) * time.Second)
	return "Route 66"
}

func GetNameById(id string) rio.Callback {
	return func(bconn *rio.BridgeConnection) *rio.FutureTaskResponse {
		response := backEndCall1(id)
		return &rio.FutureTaskResponse{
			Data:         response,
			ResponseCode: 200,
		}

	}
}

func GetStreetAddressByNameAndLocationId(name, locationId string) rio.Callback {
	return func(bconn *rio.BridgeConnection) *rio.FutureTaskResponse {
		var innerName string

		if bconn != nil {
			innerName = bconn.Data[0].(string)
		} else {
			innerName = name
		}

		if innerName != "" && locationId != "" {
			response := backEndCall2(innerName, locationId)
			return &rio.FutureTaskResponse{
				Data:         response,
				ResponseCode: 200,
			}
		} else {
			return rio.EMPTY_CALLBACK_RESPONSE
		}

	}
}

// Bridges
func Call1ToCall2(response interface{}) *rio.BridgeConnection {
	bridge := make([]interface{}, 1)
	typedResponse := response.(string)
	bridge[0] = typedResponse
	return &rio.BridgeConnection{
		Data:  bridge,
		Error: nil,
	}
}

func SampleHandler(w http.ResponseWriter, r *http.Request) {
	// Create the load balancer, this should be created only once.
	balancer := rio.GetBalancer(10, 2) // 10 threads

	// Setup the callbacks
	callback1 := GetNameById("Some Name")
	callback2 := GetStreetAddressByNameAndLocationId(rio.EMPTY_ARG_PLACEHOLDER, "Some Location ID")

	// Set up the pipeline
	request := rio.BuildRequests(context.Background(),
		rio.NewFutureTask(callback1).WithMilliSecondTimeout(10).WithRetry(3), 2).
		FollowedBy(Call1ToCall2, rio.NewFutureTask(callback2).WithMilliSecondTimeout(20))

	// Post job
	balancer.PostJob(request)

	// Wait for response
	<-request.CompletedChannel

	// Responses
	response1, err := request.GetResponse(0)
	if err == nil {
		// Do something with the response
		fmt.Println(response1)
	}
	response2, err := request.GetResponse(1)
	if err == nil {
		// Do something with the response
		fmt.Println(response2)
	}

}
