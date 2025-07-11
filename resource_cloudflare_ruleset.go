// cloudflareplus/resource_cloudflare_ruleset.go
package cloudflareplus

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudflareRuleset() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRulesetCreate,
		ReadContext:   resourceRulesetRead,
		UpdateContext: resourceRulesetUpdate,
		DeleteContext: resourceRulesetDelete,
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			client := meta.(*cloudflareplusClient)

			rules, ok := d.Get("rules").([]interface{})
			if !ok || len(rules) == 0 {
				return nil
			}

			for i, r := range rules {
				rule, ok := r.(map[string]interface{})
				if !ok {
					continue
				}
				expr, ok := rule["expression"].(string)
				if ok && expr != "" {
					if err := ValidateExpression(expr, client.accountID, client.tokenPool); err != nil {
						return fmt.Errorf("rule %d: %w", i, err)
					}
				}
			}
			return nil
		},
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Zone ID for which to create the ruleset.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the ruleset.",
			},
			"phase": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The phase at which the ruleset is evaluated.",
			},
			"rules": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression":        {Type: schema.TypeString, Required: true},
						"action":            {Type: schema.TypeString, Required: true},
						"description":       {Type: schema.TypeString, Optional: true},
						"enabled":           {Type: schema.TypeBool, Optional: true, Default: true},
						"action_parameters": {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
						"logging": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {Type: schema.TypeBool, Required: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceRulesetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflareplusClient)

	zoneID := d.Get("zone_id").(string)
	name := d.Get("name").(string)
	phase := d.Get("phase").(string)

	ruleInputs := d.Get("rules").([]interface{})
	rules := make([]cloudflare.RulesetRule, 0, len(ruleInputs))
	for _, raw := range ruleInputs {
		rule := raw.(map[string]interface{})
		desc := ""
		if v, ok := rule["description"]; ok {
			desc = v.(string)
		}

		params := make(map[string]interface{})
		if ap, ok := rule["action_parameters"].(map[string]interface{}); ok {
			params = ap
		}

		var logging *cloudflare.Logging
		if lRaw, ok := rule["logging"].([]interface{}); ok && len(lRaw) > 0 {
			l := lRaw[0].(map[string]interface{})
			logging = &cloudflare.Logging{Enabled: l["enabled"].(bool)}
		}

		rules = append(rules, cloudflare.RulesetRule{
			Expression:       rule["expression"].(string),
			Action:           rule["action"].(string),
			Description:      desc,
			Enabled:          rule["enabled"].(bool),
			ActionParameters: params,
			Logging:          logging,
		})
	}

	ruleset := cloudflare.Ruleset{
		Name:  name,
		Phase: phase,
		Kind:  "zone",
		Rules: rules,
	}

	resp, err := client.client.CreateRuleset(ctx, zoneID, ruleset)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)
	return nil
}

func resourceRulesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflareplusClient)
	zoneID := d.Get("zone_id").(string)
	id := d.Id()

	ruleset, err := client.client.GetRuleset(ctx, zoneID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", ruleset.Name)
	d.Set("phase", ruleset.Phase)

	var rules []map[string]interface{}
	for _, r := range ruleset.Rules {
		rule := map[string]interface{}{
			"expression":        r.Expression,
			"action":            r.Action,
			"enabled":           r.Enabled,
			"description":       r.Description,
			"action_parameters": r.ActionParameters,
		}
		if r.Logging != nil {
			rule["logging"] = []interface{}{map[string]interface{}{"enabled": r.Logging.Enabled}}
		}
		rules = append(rules, rule)
	}
	d.Set("rules", rules)
	return nil
}

func resourceRulesetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflareplusClient)
	zoneID := d.Get("zone_id").(string)
	id := d.Id()

	ruleInputs := d.Get("rules").([]interface{})
	rules := make([]cloudflare.RulesetRule, 0, len(ruleInputs))
	for _, raw := range ruleInputs {
		rule := raw.(map[string]interface{})
		desc := ""
		if v, ok := rule["description"]; ok {
			desc = v.(string)
		}
		params := make(map[string]interface{})
		if ap, ok := rule["action_parameters"].(map[string]interface{}); ok {
			params = ap
		}
		var logging *cloudflare.Logging
		if lRaw, ok := rule["logging"].([]interface{}); ok && len(lRaw) > 0 {
			l := lRaw[0].(map[string]interface{})
			logging = &cloudflare.Logging{Enabled: l["enabled"].(bool)}
		}
		rules = append(rules, cloudflare.RulesetRule{
			Expression:       rule["expression"].(string),
			Action:           rule["action"].(string),
			Description:      desc,
			Enabled:          rule["enabled"].(bool),
			ActionParameters: params,
			Logging:          logging,
		})
	}

	ruleset := cloudflare.Ruleset{
		ID:    id,
		Name:  d.Get("name").(string),
		Phase: d.Get("phase").(string),
		Kind:  "zone",
		Rules: rules,
	}

	_, err := client.client.UpdateRuleset(ctx, zoneID, ruleset)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRulesetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflareplusClient)
	zoneID := d.Get("zone_id").(string)
	id := d.Id()

	err := client.client.DeleteRuleset(ctx, zoneID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
