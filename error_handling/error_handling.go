package main

import (
	"fmt"
	"net/http"
)

// Zen: Separate the concerns of error handling from a producer goroutine. Instead,
// couple the potential result with a potential error so that the calling goroutine
// (which has more context about the running program) is able to take informed decisions

// this function fetches http responses but is unable to pass/communicate an error event
func errorHandlingThatIsNotAbleToPropagateValues() {
	checkStatus := func(done <-chan interface{}, urls ...string) <-chan *http.Response {
		responses := make(chan *http.Response)
		go func() {
			defer close(responses)
			//make request
			for _, url := range urls {
				resp, err := http.Get(url)
				if err != nil {
					// there's nothing we can do here other than to log the error
					fmt.Printf("[Hopeful] Error processing request %v: %v\n", url, err)
					continue
				}
				select {
				case <-done:
					return
				case responses <- resp:
				}
			}
		}()
		return responses
	}

	done := make(chan interface{})
	defer close(done)

	urls := []string{
		"https://www.google.com",
		"https://badhost",
	}
	for resp := range checkStatus(done, urls...) {
		fmt.Printf("Response for %v: %d\n", resp.Request.URL, resp.StatusCode)
	}
}

type Result struct {
	Error    error
	Response *http.Response
	Url      string
}

// this function fetches http responses and is able to pass/communicate an error event
func errorHandlingThatIsAbleToPropagateValues() {
	checkStatus := func(done <-chan interface{}, urls ...string) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)
			//make request
			for _, url := range urls {
				resp, err := http.Get(url)
				result := Result{Error: err, Response: resp, Url: url}
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()
		return results
	}
	done := make(chan interface{})
	defer close(done)

	urls := []string{
		"https://www.google.com",
		"https://badhost",
	}
	for res := range checkStatus(done, urls...) {
		if res.Error != nil {
			// we can take an informed decision about the error now
			fmt.Printf("[Informed] Error processing request %v: %v\n", res.Url, res.Error)
			continue
		}
		fmt.Printf("Response for %v: %d\n", res.Url, res.Response.StatusCode)
	}
}

func main() {
	errorHandlingThatIsNotAbleToPropagateValues()
	errorHandlingThatIsAbleToPropagateValues()
}
