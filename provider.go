// cloudflareplus/provider.go
package cloudflareplus

import (
	"context"
	"net/http"
	"os"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_tokens": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of Cloudflare API tokens to use for round-robin requests.",
			},
			"account_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Cloudflare account ID.",
			},
		},
		ConfigureContextFunc: providerConfigure,
		ResourcesMap: map[string]*schema.Resource{
			"cloudflareplus_ruleset": resourceCloudflareplusRuleset(),
			"cloudflare_ruleset":     resourceCloudflareRuleset(),
		},
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	tokensRaw := d.Get("api_tokens").([]interface{})
	var tokens []string
	for _, t := range tokensRaw {
		tokens = append(tokens, t.(string))
	}

	if len(tokens) == 0 {
		if env := os.Getenv("CF_API_TOKENS"); env != "" {
			tokens = strings.Split(env, ",")
		}
	}

	if len(tokens) == 0 {
		return nil, diag.Errorf("At least one API token must be provided via 'api_tokens' or CF_API_TOKENS env variable")
	}

	rotator := &TokenRotator{Tokens: tokens}
	client := &http.Client{Transport: rotator}

	api, err := cloudflare.NewWithHTTPClient(tokens[0], "", client)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return &cloudflareplusClient{
		client:    api,
		tokenPool: rotator,
		accountID: d.Get("account_id").(string),
	}, diags
}

type cloudflareplusClient struct {
	client    *cloudflare.API
	tokenPool *TokenRotator
	accountID string
}
