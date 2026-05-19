package service

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/mini-proxy/mod-proxy/config"
	"github.com/semanggilab/lib-go-fhir/entity"
	"github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/semanggilab/lib-go-fhir/service"
	"github.com/webcore-go/webcore/app/core"
)

type AuthDbService struct {
	Context     *core.AppContext
	Config      *config.ModuleConfig
	FhirService *service.FhirTransactionService
	Token       *string
}

func NewAuthDbService(wctx *core.AppContext, cfg *config.ModuleConfig, fhirService *service.FhirTransactionService) *AuthDbService {
	return &AuthDbService{
		Context:     wctx,
		Config:      cfg,
		FhirService: fhirService,
	}
}

func (r *AuthDbService) ValidateToken(ctx *fiber.Ctx, token string, env string) error {
	clientCredential := r.getLocalClientCredentials(ctx)
	if clientCredential == nil {
		var err error
		clientCredential, err = r.FhirService.GetClientCredentials(token, env)
		if err != nil {
			return err
		}
	}

	// Check if the access token has expired
	if clientCredential.ExpiredAt != nil {
		if time.Now().After(*clientCredential.ExpiredAt) {
			return fmt.Errorf("access token has expired at %s", clientCredential.ExpiredAt)
		}
	}

	r.Token = &token
	r.setLocalClientCredentials(ctx, clientCredential)

	return nil
}

func (r *AuthDbService) GetOrganizationID(ctx *fiber.Ctx, entry *types.BaseResource, env string) string {
	clientCredential := r.getLocalClientCredentials(ctx)
	if clientCredential == nil {
		return ""
	}

	if clientCredential.OrganizationID != nil && *clientCredential.OrganizationID != "" {
		return *clientCredential.OrganizationID
	}

	return r.scanEntry(ctx, clientCredential, entry, env)
}

func (r *AuthDbService) scanEntry(ctx *fiber.Ctx, clientCredential *entity.ClientCredential, entry *types.BaseResource, env string) string {
	orgId := r.FhirService.ScanOrganization(clientCredential, entry, env)
	if orgId != "" {
		clientCredential.OrganizationID = &orgId
		r.FhirService.Repository.SaveClientOrganization(*clientCredential, env)

		// update local
		r.setLocalClientCredentials(ctx, clientCredential)
	}

	return orgId
}

func (r *AuthDbService) getLocalClientCredentials(ctx *fiber.Ctx) *entity.ClientCredential {
	cCred := ctx.Locals("client_credential")
	if cCred == nil {
		return nil
	}

	val, ok := cCred.(*entity.ClientCredential)
	if !ok {
		return nil
	}

	return val
}

func (r *AuthDbService) setLocalClientCredentials(ctx *fiber.Ctx, clientCredential *entity.ClientCredential) {
	ctx.Locals("client_credential", clientCredential)
}
