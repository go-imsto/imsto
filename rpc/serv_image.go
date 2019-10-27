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
	app, err := storage.LoadApp(in.ApiKey)
	if err != nil {
		reportError(err, nil)
		return nil, err
	}

	entry, err := storage.Fetch(storage.FetchInput{
		URI:     in.Uri,
		Roof:    in.Roof,
		Referer: in.Referer,
		AppID:   int(app.Id),
		UserID:  int(in.UserID),
	})
	if err != nil {
		reportError(err, nil)
		return nil, err
	}

	return ri.loadImageOutput(entry, in.SizeOp)
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

	err = <-entry.Store(in.Roof)
	if err != nil {
		reportError(err, nil)
		return nil, err
	}

	return ri.loadImageOutput(entry, in.SizeOp)
}

func (ri *rpcImage) loadImageOutput(entry *storage.Entry, sizeOp string) (*pb.ImageOutput, error) {

	spath := "orig/" + entry.Path
	if sizeOp != "" {
		spath = sizeOp + "/" + entry.Path
		_, oerr := storage.LoadPath(storage.ViewName + "/" + spath)
		if oerr != nil {
			reportError(oerr, nil)
			return nil, oerr
		}
	}

	return &pb.ImageOutput{
		Path: entry.Path,
		Uri:  "/" + storage.ViewName + "/" + spath,
		Host: config.Current.StageHost,
		ID:   uint64(entry.Id),
		Meta: &pb.ImageMeta{
			Width:   int32(entry.Meta.Width),
			Height:  int32(entry.Meta.Height),
			Quality: int32(entry.Meta.Quality),
			Size:    int32(entry.Size),
			Ext:     entry.Meta.Ext,
			Mime:    entry.Meta.Mime,
		},
	}, nil
}
