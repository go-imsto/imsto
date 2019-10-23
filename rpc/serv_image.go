package rpc

import (
	"bytes"
	"context"

	pb "github.com/go-imsto/imsto-client/impb"
	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage"
)

var (
	_ pb.ImageSvcServer = (*rpcImage)(nil)
)

type rpcImage struct{}

func (ri *rpcImage) Fetch(ctx context.Context, in *pb.FetchInput) (*pb.ImageOutput, error) {
	// TODO:
	return nil, nil
}

func (ri *rpcImage) Store(ctx context.Context, in *pb.ImageInput) (*pb.ImageOutput, error) {
	app, err := storage.LoadApp(in.ApiKey)
	if err != nil {
		reportError(err, nil)
		return nil, err
	}
	entry, err := storage.NewEntryReader(bytes.NewReader(in.Image), in.Name)
	if err != nil {
		reportError(err, nil)
		return nil, err
	}
	entry.AppId = app.Id
	entry.Author = storage.Author(in.UserID)

	err = entry.Store(in.Roof)
	if err != nil {
		reportError(err, nil)
		return nil, err
	}
	return &pb.ImageOutput{
		Path: entry.Path,
		Uri:  config.GetValue(in.Roof, "stage_host"),
		ID:   uint64(entry.Id),
	}, err
}
