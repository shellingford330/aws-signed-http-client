package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

const (
	// EmptyStringSHA256 is the hex encoded sha256 value of an empty string
	EmptyStringSHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

type ServiceName string

const (
	ServiceNameLambda ServiceName = "lambda"
	ServiceNameAPIGateway ServiceName = "execute-api"
)

type Transport struct {
	serviceName ServiceName
	config      aws.Config
}

func NewClient(ctx context.Context, serviceName ServiceName) (*http.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &Transport{
			serviceName: serviceName,
			config:      cfg,
		},
	}, nil
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := sign(req, t.config, t.serviceName)
	if err != nil {
		return nil, err
	}

	return http.DefaultTransport.RoundTrip(req)
}

func sign(req *http.Request, cfg aws.Config, serviceName ServiceName) error {
	// Calculate payload hash
	payloadHash := EmptyStringSHA256
	if req.Body != nil {
		body, err := req.GetBody()
		if err != nil {
			return err
		}
		b, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		payloadHash = hex.EncodeToString(sum[:])
	}

	// Sign request with AWS Signature Version 4 (SigV4)
	credentials, err := cfg.Credentials.Retrieve(req.Context())
	if err != nil {
		return err
	}
	signer := v4.NewSigner()
	return signer.SignHTTP(req.Context(), credentials, req, payloadHash, string(serviceName), cfg.Region, time.Now())
}
