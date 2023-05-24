package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	awshttp "github.com/shellingford330/aws-signed-http-client/http"
	"golang.org/x/sync/errgroup"
)

func main() {
	var eg errgroup.Group
	for i := 0; i < 100; i++ {
		eg.Go(func() error {
			ctx := context.Background()

			url := "https://localhost:3000"
			method := "GET"
			payload := strings.NewReader(`{ "id": 33907 }`)

			client, err := awshttp.NewClient(ctx, awshttp.ServiceNameAPIGateway)
			if err != nil {
				return err
			}

			req, err := http.NewRequest(method, url, payload)
			if err != nil {
				return err
			}
			req.Header.Add("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			_, err = ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		fmt.Printf("error :%v\n", err)
	}
	fmt.Println("Done!!!")
}
