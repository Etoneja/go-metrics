package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func retryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	backoffSchedule := common.DefaultBackoffSchedule
	attemptNum := 0

	for _, backoff := range backoffSchedule {
		attemptNum++
		attemptString := fmt.Sprintf("[%d/%d]", attemptNum, len(backoffSchedule)+1)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			log.Printf("%s gRPC request succeeded", attemptString)
			return nil
		}

		if shouldRetry(err) {
			log.Printf("%s gRPC request failed, retrying: %v", attemptString, err)
			time.Sleep(backoff)
			continue
		}

		log.Printf("%s gRPC request failed: %v", attemptString, err)
		return err
	}

	return fmt.Errorf("all gRPC attempts failed")
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.DeadlineExceeded, codes.Unavailable, codes.ResourceExhausted, codes.Internal, codes.Unknown:
		return true
	default:
		return false
	}
}
