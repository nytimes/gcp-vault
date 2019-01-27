package kitexample

import (
	"context"
	"net/http"

	"github.com/NYTimes/gizmo/server/kit"
)

func (s *service) getTopStories(ctx context.Context, _ interface{}) (interface{}, error) {
	stories, err := s.client.GetTopStories(ctx, "science")
	if err != nil {
		kit.LogErrorMsg(ctx, err, "unable to get storiess")
		return nil, kit.NewJSONStatusResponse("unable to get stories",
			http.StatusInternalServerError)
	}
	return stories, nil
}
