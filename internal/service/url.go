package service

import (
	"context"
	"net/url"
	"strings"

	"github.com/gbvillarinho/base-project/config"
	"github.com/gbvillarinho/base-project/internal/domain"
)

type URLService interface {
	GenerateVerificationURL(ctx context.Context, token string, flow domain.VerificationFlow) string
}

type urlService struct {
	feURL string
	beURL string
}

func NewURLService(appConfig *config.Config) URLService {
	return &urlService{
		feURL: appConfig.URL.APPBaseURL,
		beURL: appConfig.URL.APIBaseURL,
	}
}

func (s *urlService) GenerateVerificationURL(ctx context.Context, token string, flow domain.VerificationFlow) string {
	paths := map[domain.VerificationFlow]string{
		domain.VerificationEmailFlow: "/auth/verify-email",
		domain.ResetPasswordFlow:     "/auth/reset-password",
		domain.ChangeEmailFlow:       "/auth/change-email",
		domain.MagicLinkLoginFlow:    "/auth/magic-link/verify",
	}

	path, ok := paths[flow]
	if !ok {
		path = "/auth/verify"
	}

	queryParams := map[string]string{
		"token": token,
	}

	return buildURL(s.feURL, path, queryParams)
}

func buildURL(base, path string, queryParams map[string]string) string {
	u, err := url.Parse(base)
	if err != nil {
		cleanedBase := strings.TrimSuffix(base, "/")
		cleanedPath := strings.TrimPrefix(path, "/")
		var urlStr strings.Builder
		urlStr.WriteString(cleanedBase + "/" + cleanedPath)

		if len(queryParams) > 0 {
			urlStr.WriteString("?")
			first := true
			for key, value := range queryParams {
				if !first {
					urlStr.WriteString("&")
				}
				urlStr.WriteString(key + "=" + url.QueryEscape(value))
				first = false
			}
		}
		return urlStr.String()
	}

	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(path, "/")

	if len(queryParams) > 0 {
		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}
