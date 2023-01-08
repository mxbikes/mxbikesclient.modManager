package client

import (
	"context"
	"fmt"

	protobuffer "github.com/mxbikes/protobuf/mod"
	"google.golang.org/grpc"
)

type ModServiceClient struct {
	Client protobuffer.ModServiceClient
}

func InitModServiceClient(url string) ModServiceClient {
	cc, err := grpc.Dial(url, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Could not connect:", err)
	}

	c := ModServiceClient{
		Client: protobuffer.NewModServiceClient(cc),
	}

	return c
}

func (c *ModServiceClient) GetModByID(modID string) (*protobuffer.GetModByIDResponse, error) {
	req := &protobuffer.GetModByIDRequest{
		ID: modID,
	}

	return c.Client.GetModByID(context.Background(), req)
}

func (c *ModServiceClient) SearchMod(searchText string, size int64, page int64) (*protobuffer.SearchModResponse, error) {
	req := &protobuffer.SearchModRequest{
		SearchText: searchText,
		Size:       size,
		Page:       page,
	}

	return c.Client.SearchMod(context.Background(), req)
}
