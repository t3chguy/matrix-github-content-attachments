package main

import (
	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog"
	"goji.io/pat"
	"os"
)

const AppName = "matrix-github-content-attachments"

func main() {
	config, err := ReadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	client, err := NewClient(config.Matrix.HomeServerURL, config.Matrix.UserID, config.Matrix.AccessToken)
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	server, err := baseapp.NewServer(
		config.Server,
		baseapp.DefaultParams(logger, AppName+".")...,
	)
	if err != nil {
		panic(err)
	}

	cc, err := githubapp.NewDefaultCachingClientCreator(
		config.Github,
		githubapp.WithClientUserAgent(AppName),
		githubapp.WithClientMiddleware(
			githubapp.ClientMetrics(server.Registry()),
		),
	)
	if err != nil {
		panic(err)
	}

	contentReferenceHandler := &ContentReferenceHandler{
		ClientCreator: cc,
		MatrixClient:  client,

		RoomRegexes: config.GetRoomRegexes(),
	}

	webhookHandler := githubapp.NewDefaultEventDispatcher(config.Github, contentReferenceHandler)
	server.Mux().Handle(pat.Post(githubapp.DefaultWebhookRoute), webhookHandler)

	panic(server.Start())
}
