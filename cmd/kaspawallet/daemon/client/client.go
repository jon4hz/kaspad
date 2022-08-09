package client

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Connect connects to the kaspawalletd server, and returns the client instance
func Connect(address string, tlsCert ...string) (pb.KaspawalletdClient, func(), error) {
	// Connection is local, so 1 second timeout is sufficient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	creds := insecure.NewCredentials()
	var err error
	if len(tlsCert) > 0 && tlsCert[0] != "" {
		creds, err = credentials.NewClientTLSFromFile(tlsCert[0], "")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load tls certificate: %w", err)
		}
	}

	conn, err := grpc.DialContext(ctx, address, grpc.WithBlock(), grpc.WithTransportCredentials(creds))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, errors.New("kaspawallet daemon is not running, start it with `kaspawallet start-daemon`")
		}
		return nil, nil, err
	}

	return pb.NewKaspawalletdClient(conn), func() {
		conn.Close()
	}, nil
}
