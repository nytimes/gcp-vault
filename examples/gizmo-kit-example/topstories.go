package kitexample

import (
	"context"
	"net/http"

	"github.com/NYTimes/gizmo/server/kit"
)

func (s *service) getTopScienceStories(ctx context.Context, _ interface{}) (interface{}, error) {
	stories, err := s.client.GetTopStories(ctx, "science")
	if err != nil {
		kit.LogErrorMsg(ctx, err, "unable to get stories")
		return nil, kit.NewJSONStatusResponse("unable to get stories",
			http.StatusInternalServerError)
	}
	return stories, nil
}
