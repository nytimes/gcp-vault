package gizmoexample

import "context"

func (s *service) getMySecret(ctx context.Context, r interface{}) (interface{}, error) {
	return map[string]string{
		"my-secret": s.mySecret,
	}, nil
}
