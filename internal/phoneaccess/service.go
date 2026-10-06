package phoneaccess

import (
	"context"
	"encoding/json"
	"fmt"
)

type Flow string

const (
	Signup Flow = "signup"
	Login  Flow = "login"
)

type AccessService struct {
	client *Client
}

func NewAccessService(client *Client) *AccessService {
	return &AccessService{client: client}
}

func (s *AccessService) RequestCode(ctx context.Context, flow Flow, phone, locale string) (json.RawMessage, error) {
	if err := validateFlow(flow); err != nil {
		return nil, err
	}
	return s.client.SendCode(ctx, phone)
}

func (s *AccessService) VerifyCode(ctx context.Context, flow Flow, phone, code string) (json.RawMessage, error) {
	if err := validateFlow(flow); err != nil {
		return nil, err
	}
	return s.client.VerifyPhone(ctx, phone, code, flow == Login)
}

func validateFlow(flow Flow) error {
	if flow != Signup && flow != Login {
		return fmt.Errorf("unknown phone access flow %q", flow)
	}
	return nil
}
