package squaresvc

import (
	"context"

	"github.com/pkg/errors"
)

func (svc *squareSvc) IsAgentExists(ctx context.Context, agentID string) (exists bool, err error) {
	exists, err = svc.agentConfRepo.ExistsByID(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "[squareSvc.IsAgentExists]:svc.agentConfRepo.ExistsByID(ctx, %s)", agentID)
		return
	}

	return
}
