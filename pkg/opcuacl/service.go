package opcuacl

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gopcua/opcua"
	uatest "github.com/gopcua/opcua/tests/python"
	"github.com/gopcua/opcua/ua"
	"github.com/mrscorpio/uahelper/configs"
)

func NewCl(ctx context.Context, cfg *configs.Config, dontconnect bool) ([]*opcua.Client, error) {
	if dontconnect {
		return nil, nil
	}
	srvs := strings.Split(cfg.Endpoint, ";")
	cl := make([]*opcua.Client, len(srvs))
	var err error
	cl[0], err = opcua.NewClient(srvs[0], opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err != nil {
		return nil, err
	}
	if err := cl[0].Connect(ctx); err != nil {
		return nil, err
	}

	if len(srvs) > 1 {

		c, k, err := uatest.GenerateCert("debianvm", 2048, 666*24*time.Hour)
		if err != nil {
			return cl, err
		}
		os.WriteFile("cert.pem", c, 0644)
		os.WriteFile("key.der", k, 0644)

		ck, err := tls.LoadX509KeyPair("cert.pem", "key.der")

		if err != nil {
			return cl, fmt.Errorf("generator:%s", err)
		}

		eps, err := opcua.GetEndpoints(ctx, srvs[1])
		if err != nil {
			return cl, fmt.Errorf("OPC GetEndpoints: %w", err)
		}

		ep, err := opcua.SelectEndpoint(eps, ua.SecurityPolicyURIBasic256, ua.MessageSecurityModeSign)
		if err != nil {
			return cl, fmt.Errorf("OPC SelectEndpoints: %w", err)
		}

		cl[1], err = opcua.NewClient(srvs[1],
			opcua.SecurityMode(ua.MessageSecurityModeSign),
			opcua.SecurityPolicy(ua.SecurityPolicyURIBasic256),
			opcua.AuthUsername("scada", "xog52o3j8"),
			opcua.PrivateKey(ck.PrivateKey.(*rsa.PrivateKey)),
			opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeUserName),
			opcua.Certificate(ck.Certificate[0]),
		)

		if err != nil {
			return cl, err
		}
		if err := cl[1].Connect(ctx); err != nil {
			return cl, err
		}
	}

	return cl, nil
}
