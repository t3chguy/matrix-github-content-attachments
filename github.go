package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/palantir/go-githubapp/githubapp"
	"regexp"
)

type ContentReferenceHandler struct {
	githubapp.ClientCreator
	MatrixClient *Client

	RoomRegexes []*regexp.Regexp
}

func (h *ContentReferenceHandler) Handles() []string {
	return []string{"content_reference"}
}

type Request struct {
	ContentReference struct {
		ID        int    `json:"id"`
		NodeID    string `json:"node_id"`
		Reference string `json:"reference"`
	} `json:"content_reference"`
	Installation *github.Installation `json:"installation"`
}

type response struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *ContentReferenceHandler) GetResponse(ctx context.Context, url string) *response {
	var roomID *string = nil
	for _, regex := range h.RoomRegexes {
		if r := regex.FindStringSubmatch(url); r != nil {
			roomID = &r[1]
			break
		}
	}

	if roomID == nil {
		return nil
	}

	roomInfo := h.MatrixClient.GetCachedRoomInfo(ctx, *roomID)
	if roomInfo == nil {
		return nil
	}

	return &response{
		roomInfo.Name,
		roomInfo.Topic,
	}
}

func (h *ContentReferenceHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var req Request
	if err := json.Unmarshal(payload, &req); err != nil {
		return err
	}

	installationID := req.Installation.GetID()

	client, err := h.NewInstallationClient(installationID)
	if err != nil {
		return err
	}

	u := fmt.Sprintf("content_references/%v/attachments", req.ContentReference.ID)
	pl := h.GetResponse(ctx, req.ContentReference.Reference)
	if pl == nil {
		return errors.New("unable to handle URL")
	}

	r, err := client.NewRequest("POST", u, pl)
	if err != nil {
		return err
	}

	r.Header.Set("Accept", "application/vnd.github.corsair-preview+json")
	_, err = client.Do(ctx, r, nil)
	return err
}
