package opcuacl

import (
	"context"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/mrscorpio/uahelper/configs"
)

func NewCl(ctx context.Context, cfg *configs.Config, dontconnect bool) (*opcua.Client, error) {
	if dontconnect {
		return nil, nil
	}
	cl, err := opcua.NewClient(cfg.Endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err != nil {
		return nil, err
	}
	if err := cl.Connect(ctx); err != nil {
		return nil, err
	}
	return cl, nil
}
