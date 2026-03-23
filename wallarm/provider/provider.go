// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wallarm

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/wallarm/terraform-provider-wallarm/version"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Provider() *schema.Provider {

	provider := &schema.Provider{

		Schema: map[string]*schema.Schema{
			"api_host": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("WALLARM_API_HOST", "https://api.wallarm.com"),
				Description:  "The API host address of the Wallarm Cloud for operations",
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"api_uuid": {
				Deprecated:   "This field is deprecated. Please use the api_token field instead.",
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("WALLARM_API_UUID", nil),
				Description:  "The API UUID of the user for operations",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[0-9a-f\-]+`), "API key must only contain characters 0-9 and a-f (all lowercased)"),
				Sensitive:    true,
			},
			"api_secret": {
				Deprecated:   "This field is deprecated. Please use the api_token field instead.",
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("WALLARM_API_SECRET", nil),
				Description:  "The API Secret of the user for operations",
				ValidateFunc: validation.StringMatch(regexp.MustCompile("[A-Za-z0-9-_]{40}"), "API tokens must only contain characters a-z, A-Z, 0-9 and underscores"),
				Sensitive:    true,
			},
			"api_token": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("WALLARM_API_TOKEN", nil),
				Description:   "The API Token of the user for operations",
				ValidateFunc:  validation.StringMatch(regexp.MustCompile("^[A-Za-z0-9+/]{64}$"), "API tokens must be a 64-character Base64 string (containing characters a-z, A-Z, 0-9, + and /)."),
				Sensitive:     true,
				ConflictsWith: []string{"api_secret", "api_uuid"},
			},
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_CLIENT_ID", nil),
				Description: "The Client ID to perform changes on",
			},
			"retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_RETRIES", 12),
				Description: "Maximum number of retries to perform when an API request fails",
			},
			"min_backoff": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_MIN_BACKOFF", 1),
				Description: "Minimum backoff period in seconds after failed API calls",
			},

			"max_backoff": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_MAX_BACKOFF", 5),
				Description: "Maximum backoff period in seconds after failed API calls",
			},
			"api_client_logging": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_API_CLIENT_LOGGING", false),
				Description: "Whether to print logs from the API client (using the default log library logger)",
			},
			"ignore_existing": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_IGNORE_EXISTING_RESOURCES", false),
				Description: "Whether ignore or raise an exception when a resource exists.",
			},
			"hint_prefetch": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_HINT_PREFETCH", true),
				Description: "Enable bulk prefetching of hints during plan/refresh to reduce API calls. " +
					"When enabled, the first rule read triggers a bulk fetch of all hints for the client, " +
					"and subsequent reads are served from an in-memory cache. Defaults to true.",
			},
			"require_explicit_client_id": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLARM_REQUIRE_EXPLICIT_CLIENT_ID", false),
				Description: "When true, every resource must set client_id explicitly. " +
					"Prevents accidental cross-tenant operations for Global Administrator tokens managing multiple tenants.",
			},
		},
		ProviderMetaSchema: map[string]*schema.Schema{
			"module_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"wallarm_actions":         dataSourceWallarmActions(),
			"wallarm_node":            dataSourceWallarmNode(),
			"wallarm_security_issues": dataSourceWallarmSecurityIssues(),
			"wallarm_hits":            dataSourceWallarmHits(),
			"wallarm_ip_lists":        dataSourceWallarmIPLists(),
			"wallarm_applications":    dataSourceWallarmApplications(),
			"wallarm_rules":           dataSourceWallarmRules(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"wallarm_action":                         resourceWallarmAction(),
			"wallarm_tenant":                         resourceWallarmTenant(),
			"wallarm_rules_settings":                 resourceWallarmRulesSettings(),
			"wallarm_global_mode":                    resourceWallarmGlobalMode(),
			"wallarm_node":                           resourceWallarmNode(),
			"wallarm_application":                    resourceWallarmApp(),
			"wallarm_user":                           resourceWallarmUser(),
			"wallarm_api_spec":                       resourceWallarmAPISpec(),
			"wallarm_denylist":                       resourceWallarmDenylist(),
			"wallarm_allowlist":                      resourceWallarmAllowlist(),
			"wallarm_graylist":                       resourceWallarmGraylist(),
			"wallarm_integration_email":              resourceWallarmEmail(),
			"wallarm_integration_opsgenie":           resourceWallarmOpsGenie(),
			"wallarm_integration_slack":              resourceWallarmSlack(),
			"wallarm_integration_pagerduty":          resourceWallarmPagerDuty(),
			"wallarm_integration_sumologic":          resourceWallarmSumologic(),
			"wallarm_integration_data_dog":           resourceWallarmDataDog(),
			"wallarm_integration_insightconnect":     resourceWallarmInsightConnect(),
			"wallarm_integration_splunk":             resourceWallarmSplunk(),
			"wallarm_integration_webhook":            resourceWallarmWebhook(),
			"wallarm_integration_telegram":           resourceWallarmTelegram(),
			"wallarm_integration_teams":              resourceWallarmTeams(),
			"wallarm_trigger":                        resourceWallarmTrigger(),
			"wallarm_rule_binary_data":               resourceWallarmBinaryData(),
			"wallarm_rule_enum":                      resourceWallarmEnum(),
			"wallarm_rule_disable_attack_type":       resourceWallarmDisableAttackType(),
			"wallarm_rule_disable_stamp":             resourceWallarmDisableStamp(),
			"wallarm_rule_vpatch":                    resourceWallarmVpatch(),
			"wallarm_rule_mode":                      resourceWallarmMode(),
			"wallarm_rule_masking":                   resourceWallarmSensitiveData(),
			"wallarm_rule_parser_state":              resourceWallarmParserState(),
			"wallarm_rule_regex":                     resourceWallarmRegex(),
			"wallarm_rule_ignore_regex":              resourceWallarmIgnoreRegex(),
			"wallarm_rule_set_response_header":       resourceWallarmSetResponseHeader(),
			"wallarm_rule_brute":                     resourceWallarmBrute(),
			"wallarm_rule_bruteforce_counter":        resourceWallarmBruteForceCounter(),
			"wallarm_rule_dirbust_counter":           resourceWallarmDirbustCounter(),
			"wallarm_rule_bola":                      resourceWallarmBola(),
			"wallarm_rule_bola_counter":              resourceWallarmBolaCounter(),
			"wallarm_rule_rate_limit":                resourceWallarmRateLimit(),
			"wallarm_rule_rate_limit_enum":           resourceWallarmRateLimitEnum(),
			"wallarm_rule_uploads":                   resourceWallarmUploads(),
			"wallarm_rule_credential_stuffing_regex": resourceWallarmCredentialStuffingRegex(),
			"wallarm_rule_credential_stuffing_point": resourceWallarmCredentialStuffingPoint(),
			"wallarm_rule_overlimit_res_settings":    resourceWallarmOverlimitResSettings(),
			"wallarm_rule_forced_browsing":           resourceWallarmForcedBrowsing(),
			"wallarm_rule_graphql_detection":         resourceWallarmGraphqlDetection(),
			"wallarm_rule_file_upload_size_limit":    resourceWallarmFileUploadSizeLimit(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return ProviderConfigure(ctx, d, provider)
	}
	return provider
}

func ProviderConfigure(_ context.Context, d *schema.ResourceData, p *schema.Provider) (interface{}, diag.Diagnostics) {
	retryOpt := wallarm.UsingRetryPolicy(d.Get("retries").(int), d.Get("min_backoff").(int), d.Get("max_backoff").(int))
	options := []wallarm.Option{retryOpt}

	c := cleanhttp.DefaultPooledClient()
	if d.Get("api_client_logging").(bool) {
		c.Transport = newLoggingTransport(c.Transport)
	} else {
		c.Transport = logging.NewSubsystemLoggingHTTPTransport("Wallarm", c.Transport)
	}
	options = append(options, wallarm.HTTPClient(c))

	ua := p.UserAgent("terraform-provider-wallarm", version.ProviderVersion)

	options = append(options, wallarm.UserAgent(ua))

	authHeaders := make(http.Header)
	config := Config{}

	if v, ok := d.GetOk("api_token"); ok {
		authHeaders.Add("X-WallarmAPI-Token", v.(string))
	} else {
		if v, ok := d.GetOk("api_uuid"); ok {
			authHeaders.Add("X-WallarmAPI-UUID", v.(string))
		} else {
			return nil, diag.FromErr(fmt.Errorf("api_uuid is required when api_token is not set: %w", wallarm.ErrInvalidCredentials))
		}
		if v, ok := d.GetOk("api_secret"); ok {
			authHeaders.Add("X-WallarmAPI-Secret", v.(string))
		} else {
			return nil, diag.FromErr(fmt.Errorf("api_secret is required when api_token is not set: %w", wallarm.ErrInvalidCredentials))
		}
	}

	if v, ok := d.GetOk("api_host"); ok {
		options = append(options, wallarm.UsingBaseURL(v.(string)))
	}
	options = append(options, wallarm.Headers(authHeaders))
	config.Options = options

	client, err := config.Client()
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("could not create Wallarm client: %w", err))
	}

	var defaultClientID int
	if v, ok := d.GetOk("client_id"); ok {
		defaultClientID = v.(int)
	} else {
		u, err := client.UserDetails()
		if err != nil {
			return nil, diag.FromErr(fmt.Errorf("could not fetch user details: %w", err))
		}

		defaultClientID = u.Body.Clientid
	}

	// Wrap with caching layer if hint_prefetch is enabled (default: true)
	if d.Get("hint_prefetch").(bool) {
		log.Printf("[INFO] Wallarm hint prefetch enabled — rule reads will use bulk cache")
		client = NewCachedClient(client)
	}

	return &ProviderMeta{
		Client:                  client,
		DefaultClientID:         defaultClientID,
		RequireExplicitClientID: d.Get("require_explicit_client_id").(bool),
		IPListCache:             NewIPListCache(),
	}, nil
}
