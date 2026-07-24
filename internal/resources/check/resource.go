package check

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/systeampl/terraform-provider-systeam/internal/client"
)

var (
	_ resource.Resource                = &checkResource{}
	_ resource.ResourceWithConfigure   = &checkResource{}
	_ resource.ResourceWithImportState = &checkResource{}
)

type checkResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &checkResource{}
}

func (r *checkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check"
}

func (r *checkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a health check.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Check ID.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Check name.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Check type.",
				Validators: []validator.String{
					stringvalidator.OneOf("heartbeat", "uptime", "icmp", "tcp", "udp", "dns", "ftp", "mail", "database", "api_scenario", "oidc"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.Int64Attribute{
				Required:    true,
				Description: "Project ID that owns this check.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Check description.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the check is active.",
			},
			"interval": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Description: "Check interval in seconds.",
			},
			"grace_period": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Description: "Grace period in seconds before marking as down.",
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Description: "Timeout in seconds.",
			},
			"url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "URL to check (uptime checks).",
			},
			"host": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Host to check (icmp, tcp, udp, dns, ftp checks).",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Port number (tcp, udp, ftp checks).",
			},
			"ping_count": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
				Description: "Number of ICMP pings to send.",
			},
			"ssl_verify": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Verify SSL certificate.",
			},
			"alert_after_failures": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "Number of consecutive failures before alerting.",
			},
			"geo_monitoring_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enable geo-distributed monitoring.",
			},
			"assigned_agent_id": schema.Int64Attribute{
				Optional:    true,
				Description: "ID of the private agent assigned to this check.",
			},
			"escalation_policy_id": schema.Int64Attribute{
				Optional:    true,
				Description: "ID of the escalation policy.",
			},
			"traceroute_on_timeout": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Run traceroute automatically on timeout.",
			},
			"dns_server": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "DNS server to query (dns checks).",
			},
			"dns_record_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "DNS record type to query (dns checks). Server default: A.",
			},
			"dns_expected_value": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Expected DNS record value (dns checks).",
			},
			"http_method": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "HTTP method (uptime checks). Server default: GET.",
			},
			"auth_method": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Authentication method (uptime checks). Server default: NONE.",
			},
			"schedule_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Schedule type (heartbeat checks). Server default: interval.",
			},
			"cron_expression": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Cron expression (heartbeat checks).",
			},
			"cron_timezone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cron timezone (heartbeat checks). Server default: UTC.",
			},
			"mail_domain": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Mail domain to check (mail checks).",
			},
			"runbook_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Runbook URL surfaced on incidents for this check.",
			},
			// Database checks (type = "database").
			"db_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Database engine (database checks): postgresql, mysql, mongodb, redis.",
			},
			"db_host": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Database host (database checks).",
			},
			"db_port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Database port (database checks).",
			},
			"db_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Database name (database checks).",
			},
			"db_username": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Database username (database checks).",
			},
			"db_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Database password (database checks). Write-only: encrypted at rest and never returned by the API, so it is not refreshed into state.",
			},
			"db_ssl_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Use SSL/TLS for the database connection (database checks).",
			},
			"db_query": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Query to execute (database checks).",
			},
			"db_expected_result": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Expected query result for validation (database checks).",
			},
			// HTTP auth (auth_password / auth_bearer_token are write-only secrets).
			"auth_username":                   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "HTTP basic-auth username."},
			"auth_password":                   schema.StringAttribute{Optional: true, Sensitive: true, Description: "HTTP basic-auth password. Write-only: encrypted server-side, never read back."},
			"auth_bearer_token":               schema.StringAttribute{Optional: true, Sensitive: true, Description: "HTTP bearer token. Write-only: encrypted server-side, never read back."},
			"http_body":                       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Request body (POST/PUT/PATCH)."},
			"http_body_type":                  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("none"), Description: "Request body type: none, json, form, raw."},
			"http_follow_redirects":           schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Follow HTTP redirects."},
			"content_match_enabled":           schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Enable response content matching."},
			"content_match_text":              schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Text or regex to match in the response."},
			"content_match_type":              schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("contains"), Description: "contains, not_contains, regex."},
			"content_match_case_sensitive":    schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Case-sensitive content match."},
			"http_form_login_enabled":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Authenticate via an HTML form before checking."},
			"http_form_login_url":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "URL to POST the login form to."},
			"http_form_login_success_text":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Text confirming successful login."},
			"http_form_check_after_login_url": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "URL to check after logging in."},
			// FTP/SFTP (ftp_password is a write-only secret).
			"ftp_username": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "FTP username (empty for anonymous)."},
			"ftp_password": schema.StringAttribute{Optional: true, Sensitive: true, Description: "FTP password. Write-only: encrypted server-side, never read back."},
			"ftp_protocol": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("FTP"), Description: "FTP, FTPS or SFTP."},
			"ftp_path":     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Path to list/verify."},
			"ftp_passive":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Use passive mode."},
			// DNS advanced.
			"dns_soa_alert_on_change":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Alert when the SOA serial changes."},
			"dns_hijack_alert_enabled":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Enable DNS-hijacking alerts."},
			"dns_hijack_alert_channel_ids": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Channel IDs for hijack alerts."},
			"dns_txt_monitoring_enabled":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Enable TXT (SPF/DKIM/DMARC) monitoring."},
			"dns_dkim_selector":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "DKIM selector to check."},
			"dns_multi_record_enabled":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Enable multi-record DNS monitoring."},
			// Mail server.
			"mail_smtp_enabled":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Check SMTP."},
			"mail_smtp_port":         schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(25), Description: "SMTP port (25/465/587)."},
			"mail_smtp_starttls":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Check STARTTLS support."},
			"mail_smtp_open_relay":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Check for an open relay."},
			"mail_imap_enabled":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Check IMAP."},
			"mail_imap_port":         schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(993), Description: "IMAP port (143/993)."},
			"mail_imap_ssl":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Use IMAP SSL."},
			"mail_pop3_enabled":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Check POP3."},
			"mail_pop3_port":         schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(995), Description: "POP3 port (110/995)."},
			"mail_pop3_ssl":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Use POP3 SSL."},
			"mail_check_spf":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Check SPF record."},
			"mail_check_dkim":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Check DKIM."},
			"mail_dkim_selectors":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Comma-separated DKIM selectors."},
			"mail_check_dmarc":       schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Check DMARC."},
			"mail_check_ptr":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Check reverse DNS (PTR)."},
			"mail_check_blacklist":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Check RBL blacklists."},
			"mail_blacklist_servers": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Comma-separated RBL servers."},
			// List fields.
			"expected_status_codes": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Accepted HTTP status codes (uptime checks). Defaults to [200] server-side.",
			},
			"dns_expected_ips": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Expected IPs for DNS-hijacking detection (dns checks).",
			},
			"assigned_agent_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "IDs of agents assigned to this check (multi/geo monitoring).",
			},
			// JSON-object fields — use jsonencode(...). Compared semantically, so
			// whitespace and key order don't produce spurious diffs.
			"http_headers": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Sensitive:   true, // commonly carries Authorization / API keys
				Description: "Custom HTTP request headers as a JSON object, e.g. jsonencode({\"X-Api-Key\"=\"...\"}).",
			},
			"http_form_login_data": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Sensitive:   true, // typically contains {username, password}
				Description: "Form fields for HTTP form login as a JSON object, e.g. jsonencode({username=\"u\",password=\"p\"}).",
			},
			"api_scenario_steps": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Description: "Multi-step API scenario definition as JSON (api_scenario checks).",
			},
			"api_scenario_secrets": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Sensitive:   true,
				Description: "Secrets for api_scenario steps as a JSON object. Write-only: encrypted server-side, never read back.",
			},
			"oidc_config": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Description: "OIDC identity-provider check configuration as JSON (oidc checks).",
			},
			"dns_records_config": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Description: "Per-record config for multi-record DNS monitoring as JSON.",
			},
			"check_source_critical": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Description: "Per-source criticality config as JSON.",
			},
			"response_assertions": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Description: "Response assertions as a JSON array (http/uptime checks).",
			},
		},
	}
}

func (r *checkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *checkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CheckModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CheckCreateRequest{
		Name:                       plan.Name.ValueString(),
		Type:                       plan.Type.ValueString(),
		Description:                plan.Description.ValueString(),
		IsActive:                   plan.IsActive.ValueBool(),
		ProjectID:                  int(plan.ProjectID.ValueInt64()),
		Interval:                   int(plan.Interval.ValueInt64()),
		GracePeriod:                int(plan.GracePeriod.ValueInt64()),
		Timeout:                    int(plan.Timeout.ValueInt64()),
		URL:                        plan.URL.ValueString(),
		Host:                       plan.Host.ValueString(),
		Port:                       int(plan.Port.ValueInt64()),
		PingCount:                  int(plan.PingCount.ValueInt64()),
		SSLVerify:                  plan.SSLVerify.ValueBool(),
		AlertAfterFailures:         int(plan.AlertAfterFailures.ValueInt64()),
		GeoMonitoringEnabled:       plan.GeoMonitoringEnabled.ValueBool(),
		TracerouteOnTimeout:        plan.TracerouteOnTimeout.ValueBool(),
		DNSServer:                  plan.DNSServer.ValueString(),
		DNSRecordType:              plan.DNSRecordType.ValueString(),
		DNSExpectedValue:           plan.DNSExpectedValue.ValueString(),
		HTTPMethod:                 plan.HTTPMethod.ValueString(),
		AuthMethod:                 plan.AuthMethod.ValueString(),
		ScheduleType:               plan.ScheduleType.ValueString(),
		CronExpression:             plan.CronExpression.ValueString(),
		CronTimezone:               plan.CronTimezone.ValueString(),
		MailDomain:                 plan.MailDomain.ValueString(),
		RunbookURL:                 plan.RunbookURL.ValueString(),
		DBType:                     plan.DBType.ValueString(),
		DBHost:                     plan.DBHost.ValueString(),
		DBPort:                     int(plan.DBPort.ValueInt64()),
		DBName:                     plan.DBName.ValueString(),
		DBUsername:                 plan.DBUsername.ValueString(),
		DBPassword:                 plan.DBPassword.ValueString(),
		DBSSLEnabled:               plan.DBSSLEnabled.ValueBool(),
		DBQuery:                    plan.DBQuery.ValueString(),
		DBExpectedResult:           plan.DBExpectedResult.ValueString(),
		AuthUsername:               plan.AuthUsername.ValueString(),
		AuthPassword:               plan.AuthPassword.ValueString(),
		AuthBearerToken:            plan.AuthBearerToken.ValueString(),
		HTTPBody:                   plan.HTTPBody.ValueString(),
		HTTPBodyType:               plan.HTTPBodyType.ValueString(),
		HTTPFollowRedirects:        plan.HTTPFollowRedirects.ValueBool(),
		ContentMatchEnabled:        plan.ContentMatchEnabled.ValueBool(),
		ContentMatchText:           plan.ContentMatchText.ValueString(),
		ContentMatchType:           plan.ContentMatchType.ValueString(),
		ContentMatchCaseSensitive:  plan.ContentMatchCaseSensitive.ValueBool(),
		HTTPFormLoginEnabled:       plan.HTTPFormLoginEnabled.ValueBool(),
		HTTPFormLoginURL:           plan.HTTPFormLoginURL.ValueString(),
		HTTPFormLoginSuccessText:   plan.HTTPFormLoginSuccessText.ValueString(),
		HTTPFormCheckAfterLoginURL: plan.HTTPFormCheckAfterLoginURL.ValueString(),
		FTPUsername:                plan.FTPUsername.ValueString(),
		FTPPassword:                plan.FTPPassword.ValueString(),
		FTPProtocol:                plan.FTPProtocol.ValueString(),
		FTPPath:                    plan.FTPPath.ValueString(),
		FTPPassive:                 plan.FTPPassive.ValueBool(),
		DNSSOAAlertOnChange:        plan.DNSSOAAlertOnChange.ValueBool(),
		DNSHijackAlertEnabled:      plan.DNSHijackAlertEnabled.ValueBool(),
		DNSHijackAlertChannelIDs:   plan.DNSHijackAlertChannelIDs.ValueString(),
		DNSTXTMonitoringEnabled:    plan.DNSTXTMonitoringEnabled.ValueBool(),
		DNSDKIMSelector:            plan.DNSDKIMSelector.ValueString(),
		DNSMultiRecordEnabled:      plan.DNSMultiRecordEnabled.ValueBool(),
		MailSMTPEnabled:            plan.MailSMTPEnabled.ValueBool(),
		MailSMTPPort:               int(plan.MailSMTPPort.ValueInt64()),
		MailSMTPStartTLS:           plan.MailSMTPStartTLS.ValueBool(),
		MailSMTPOpenRelay:          plan.MailSMTPOpenRelay.ValueBool(),
		MailIMAPEnabled:            plan.MailIMAPEnabled.ValueBool(),
		MailIMAPPort:               int(plan.MailIMAPPort.ValueInt64()),
		MailIMAPSSL:                plan.MailIMAPSSL.ValueBool(),
		MailPOP3Enabled:            plan.MailPOP3Enabled.ValueBool(),
		MailPOP3Port:               int(plan.MailPOP3Port.ValueInt64()),
		MailPOP3SSL:                plan.MailPOP3SSL.ValueBool(),
		MailCheckSPF:               plan.MailCheckSPF.ValueBool(),
		MailCheckDKIM:              plan.MailCheckDKIM.ValueBool(),
		MailDKIMSelectors:          plan.MailDKIMSelectors.ValueString(),
		MailCheckDMARC:             plan.MailCheckDMARC.ValueBool(),
		MailCheckPTR:               plan.MailCheckPTR.ValueBool(),
		MailCheckBlacklist:         plan.MailCheckBlacklist.ValueBool(),
		MailBlacklistServers:       plan.MailBlacklistServers.ValueString(),
	}

	if !plan.AssignedAgentID.IsNull() && !plan.AssignedAgentID.IsUnknown() {
		v := int(plan.AssignedAgentID.ValueInt64())
		createReq.AssignedAgentID = &v
	}
	if !plan.EscalationPolicyID.IsNull() && !plan.EscalationPolicyID.IsUnknown() {
		v := int(plan.EscalationPolicyID.ValueInt64())
		createReq.EscalationPolicyID = &v
	}

	statusCodes, d := listToInts(ctx, plan.ExpectedStatusCodes)
	resp.Diagnostics.Append(d...)
	createReq.ExpectedStatusCodes = statusCodes
	ips, d := listToStrs(ctx, plan.DNSExpectedIPs)
	resp.Diagnostics.Append(d...)
	createReq.DNSExpectedIPs = ips
	agentIDs, d := listToInts(ctx, plan.AssignedAgentIDs)
	resp.Diagnostics.Append(d...)
	createReq.AssignedAgentIDs = agentIDs
	if !plan.HTTPHeaders.IsNull() && !plan.HTTPHeaders.IsUnknown() {
		createReq.HTTPHeaders = json.RawMessage(plan.HTTPHeaders.ValueString())
	}
	if !plan.HTTPFormLoginData.IsNull() && !plan.HTTPFormLoginData.IsUnknown() {
		createReq.HTTPFormLoginData = json.RawMessage(plan.HTTPFormLoginData.ValueString())
	}
	createReq.APIScenarioSteps = normToRaw(plan.APIScenarioSteps)
	createReq.APIScenarioSecrets = normToRaw(plan.APIScenarioSecrets)
	createReq.OIDCConfig = normToRaw(plan.OIDCConfig)
	createReq.DNSRecordsConfig = normToRaw(plan.DNSRecordsConfig)
	createReq.CheckSourceCritical = normToRaw(plan.CheckSourceCritical)
	createReq.ResponseAssertions = normToRaw(plan.ResponseAssertions)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.client.CreateCheck(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating check", err.Error())
		return
	}

	resp.Diagnostics.Append(mapCheckToState(ctx, check, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *checkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CheckModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.client.GetCheck(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading check", err.Error())
		return
	}

	resp.Diagnostics.Append(mapCheckToState(ctx, check, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *checkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CheckModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Prior state — used to detect write-only secrets being removed so we can
	// clear them server-side (sending "" wipes the encrypted value).
	var prior CheckModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	isActive := plan.IsActive.ValueBool()
	interval := int(plan.Interval.ValueInt64())
	gracePeriod := int(plan.GracePeriod.ValueInt64())
	timeout := int(plan.Timeout.ValueInt64())
	url := plan.URL.ValueString()
	host := plan.Host.ValueString()
	port := int(plan.Port.ValueInt64())
	pingCount := int(plan.PingCount.ValueInt64())
	sslVerify := plan.SSLVerify.ValueBool()
	alertAfter := int(plan.AlertAfterFailures.ValueInt64())
	geoEnabled := plan.GeoMonitoringEnabled.ValueBool()
	traceroute := plan.TracerouteOnTimeout.ValueBool()
	dnsServer := plan.DNSServer.ValueString()
	dnsRecordType := plan.DNSRecordType.ValueString()
	dnsExpected := plan.DNSExpectedValue.ValueString()
	httpMethod := plan.HTTPMethod.ValueString()
	authMethod := plan.AuthMethod.ValueString()
	scheduleType := plan.ScheduleType.ValueString()
	cronExpr := plan.CronExpression.ValueString()
	cronTz := plan.CronTimezone.ValueString()
	mailDomain := plan.MailDomain.ValueString()
	runbookURL := plan.RunbookURL.ValueString()
	dbType := plan.DBType.ValueString()
	dbHost := plan.DBHost.ValueString()
	dbPort := int(plan.DBPort.ValueInt64())
	dbName := plan.DBName.ValueString()
	dbUsername := plan.DBUsername.ValueString()
	dbSSLEnabled := plan.DBSSLEnabled.ValueBool()
	dbQuery := plan.DBQuery.ValueString()
	dbExpectedResult := plan.DBExpectedResult.ValueString()

	updateReq := client.CheckUpdateRequest{
		Name:                       &name,
		Description:                &desc,
		IsActive:                   &isActive,
		Interval:                   &interval,
		GracePeriod:                &gracePeriod,
		Timeout:                    &timeout,
		URL:                        &url,
		Host:                       &host,
		Port:                       &port,
		PingCount:                  &pingCount,
		SSLVerify:                  &sslVerify,
		AlertAfterFailures:         &alertAfter,
		GeoMonitoringEnabled:       &geoEnabled,
		TracerouteOnTimeout:        &traceroute,
		DNSServer:                  &dnsServer,
		DNSRecordType:              &dnsRecordType,
		DNSExpectedValue:           &dnsExpected,
		HTTPMethod:                 &httpMethod,
		AuthMethod:                 &authMethod,
		ScheduleType:               &scheduleType,
		CronExpression:             &cronExpr,
		CronTimezone:               &cronTz,
		MailDomain:                 &mailDomain,
		RunbookURL:                 &runbookURL,
		DBType:                     &dbType,
		DBHost:                     &dbHost,
		DBPort:                     &dbPort,
		DBName:                     &dbName,
		DBUsername:                 &dbUsername,
		DBSSLEnabled:               &dbSSLEnabled,
		DBQuery:                    &dbQuery,
		DBExpectedResult:           &dbExpectedResult,
		AuthUsername:               strPtr(plan.AuthUsername.ValueString()),
		HTTPBody:                   strPtr(plan.HTTPBody.ValueString()),
		HTTPBodyType:               strPtr(plan.HTTPBodyType.ValueString()),
		HTTPFollowRedirects:        boolPtr(plan.HTTPFollowRedirects.ValueBool()),
		ContentMatchEnabled:        boolPtr(plan.ContentMatchEnabled.ValueBool()),
		ContentMatchText:           strPtr(plan.ContentMatchText.ValueString()),
		ContentMatchType:           strPtr(plan.ContentMatchType.ValueString()),
		ContentMatchCaseSensitive:  boolPtr(plan.ContentMatchCaseSensitive.ValueBool()),
		HTTPFormLoginEnabled:       boolPtr(plan.HTTPFormLoginEnabled.ValueBool()),
		HTTPFormLoginURL:           strPtr(plan.HTTPFormLoginURL.ValueString()),
		HTTPFormLoginSuccessText:   strPtr(plan.HTTPFormLoginSuccessText.ValueString()),
		HTTPFormCheckAfterLoginURL: strPtr(plan.HTTPFormCheckAfterLoginURL.ValueString()),
		FTPUsername:                strPtr(plan.FTPUsername.ValueString()),
		FTPProtocol:                strPtr(plan.FTPProtocol.ValueString()),
		FTPPath:                    strPtr(plan.FTPPath.ValueString()),
		FTPPassive:                 boolPtr(plan.FTPPassive.ValueBool()),
		DNSSOAAlertOnChange:        boolPtr(plan.DNSSOAAlertOnChange.ValueBool()),
		DNSHijackAlertEnabled:      boolPtr(plan.DNSHijackAlertEnabled.ValueBool()),
		DNSHijackAlertChannelIDs:   strPtr(plan.DNSHijackAlertChannelIDs.ValueString()),
		DNSTXTMonitoringEnabled:    boolPtr(plan.DNSTXTMonitoringEnabled.ValueBool()),
		DNSDKIMSelector:            strPtr(plan.DNSDKIMSelector.ValueString()),
		DNSMultiRecordEnabled:      boolPtr(plan.DNSMultiRecordEnabled.ValueBool()),
		MailSMTPEnabled:            boolPtr(plan.MailSMTPEnabled.ValueBool()),
		MailSMTPPort:               intPtr(int(plan.MailSMTPPort.ValueInt64())),
		MailSMTPStartTLS:           boolPtr(plan.MailSMTPStartTLS.ValueBool()),
		MailSMTPOpenRelay:          boolPtr(plan.MailSMTPOpenRelay.ValueBool()),
		MailIMAPEnabled:            boolPtr(plan.MailIMAPEnabled.ValueBool()),
		MailIMAPPort:               intPtr(int(plan.MailIMAPPort.ValueInt64())),
		MailIMAPSSL:                boolPtr(plan.MailIMAPSSL.ValueBool()),
		MailPOP3Enabled:            boolPtr(plan.MailPOP3Enabled.ValueBool()),
		MailPOP3Port:               intPtr(int(plan.MailPOP3Port.ValueInt64())),
		MailPOP3SSL:                boolPtr(plan.MailPOP3SSL.ValueBool()),
		MailCheckSPF:               boolPtr(plan.MailCheckSPF.ValueBool()),
		MailCheckDKIM:              boolPtr(plan.MailCheckDKIM.ValueBool()),
		MailDKIMSelectors:          strPtr(plan.MailDKIMSelectors.ValueString()),
		MailCheckDMARC:             boolPtr(plan.MailCheckDMARC.ValueBool()),
		MailCheckPTR:               boolPtr(plan.MailCheckPTR.ValueBool()),
		MailCheckBlacklist:         boolPtr(plan.MailCheckBlacklist.ValueBool()),
		MailBlacklistServers:       strPtr(plan.MailBlacklistServers.ValueString()),
	}

	// Write-only secrets: send the new value when set; send "" to CLEAR it when it
	// was set before but removed now; leave nil (untouched) when it was never set.
	updateReq.AuthPassword = secretUpdate(plan.AuthPassword.ValueString(), prior.AuthPassword.ValueString())
	updateReq.AuthBearerToken = secretUpdate(plan.AuthBearerToken.ValueString(), prior.AuthBearerToken.ValueString())
	updateReq.FTPPassword = secretUpdate(plan.FTPPassword.ValueString(), prior.FTPPassword.ValueString())
	updateReq.DBPassword = secretUpdate(plan.DBPassword.ValueString(), prior.DBPassword.ValueString())

	if !plan.AssignedAgentID.IsNull() && !plan.AssignedAgentID.IsUnknown() {
		v := int(plan.AssignedAgentID.ValueInt64())
		updateReq.AssignedAgentID = &v
	}
	if !plan.EscalationPolicyID.IsNull() && !plan.EscalationPolicyID.IsUnknown() {
		v := int(plan.EscalationPolicyID.ValueInt64())
		updateReq.EscalationPolicyID = &v
	}

	scodes, d := listToInts(ctx, plan.ExpectedStatusCodes)
	resp.Diagnostics.Append(d...)
	if scodes != nil {
		updateReq.ExpectedStatusCodes = &scodes
	}
	uips, d := listToStrs(ctx, plan.DNSExpectedIPs)
	resp.Diagnostics.Append(d...)
	if uips != nil {
		updateReq.DNSExpectedIPs = &uips
	}
	uagents, d := listToInts(ctx, plan.AssignedAgentIDs)
	resp.Diagnostics.Append(d...)
	if uagents != nil {
		updateReq.AssignedAgentIDs = &uagents
	}
	if !plan.HTTPHeaders.IsNull() && !plan.HTTPHeaders.IsUnknown() {
		updateReq.HTTPHeaders = json.RawMessage(plan.HTTPHeaders.ValueString())
	}
	if !plan.HTTPFormLoginData.IsNull() && !plan.HTTPFormLoginData.IsUnknown() {
		updateReq.HTTPFormLoginData = json.RawMessage(plan.HTTPFormLoginData.ValueString())
	}
	updateReq.APIScenarioSteps = normToRaw(plan.APIScenarioSteps)
	updateReq.OIDCConfig = normToRaw(plan.OIDCConfig)
	updateReq.DNSRecordsConfig = normToRaw(plan.DNSRecordsConfig)
	updateReq.CheckSourceCritical = normToRaw(plan.CheckSourceCritical)
	updateReq.ResponseAssertions = normToRaw(plan.ResponseAssertions)
	updateReq.APIScenarioSecrets = secretRawUpdate(plan.APIScenarioSecrets, prior.APIScenarioSecrets)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.client.UpdateCheck(ctx, int(plan.ID.ValueInt64()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating check", err.Error())
		return
	}

	resp.Diagnostics.Append(mapCheckToState(ctx, check, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *checkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CheckModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCheck(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Error deleting check", err.Error())
	}
}

func (r *checkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected integer check ID, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func mapCheckToState(ctx context.Context, check *client.Check, state *CheckModel) diag.Diagnostics {
	var diags diag.Diagnostics
	state.ID = types.Int64Value(int64(check.ID))
	state.Name = types.StringValue(check.Name)
	state.Type = types.StringValue(check.Type)
	state.Description = types.StringValue(check.Description)
	state.IsActive = types.BoolValue(check.IsActive)
	state.ProjectID = types.Int64Value(int64(check.ProjectID))
	state.Interval = types.Int64Value(int64(check.Interval))
	state.GracePeriod = types.Int64Value(int64(check.GracePeriod))
	state.Timeout = types.Int64Value(int64(check.Timeout))
	state.URL = types.StringValue(check.URL)
	state.Host = types.StringValue(check.Host)
	state.Port = types.Int64Value(int64(check.Port))
	state.PingCount = types.Int64Value(int64(check.PingCount))
	state.SSLVerify = types.BoolValue(check.SSLVerify)
	state.AlertAfterFailures = types.Int64Value(int64(check.AlertAfterFailures))
	state.GeoMonitoringEnabled = types.BoolValue(check.GeoMonitoringEnabled)
	state.TracerouteOnTimeout = types.BoolValue(check.TracerouteOnTimeout)
	state.DNSServer = types.StringValue(check.DNSServer)
	state.DNSRecordType = types.StringValue(check.DNSRecordType)
	state.DNSExpectedValue = types.StringValue(check.DNSExpectedValue)
	state.HTTPMethod = types.StringValue(check.HTTPMethod)
	state.AuthMethod = types.StringValue(check.AuthMethod)
	state.ScheduleType = types.StringValue(check.ScheduleType)
	state.CronExpression = types.StringValue(check.CronExpression)
	state.CronTimezone = types.StringValue(check.CronTimezone)
	state.MailDomain = types.StringValue(check.MailDomain)
	state.RunbookURL = types.StringValue(check.RunbookURL)
	state.DBType = types.StringValue(check.DBType)
	state.DBHost = types.StringValue(check.DBHost)
	state.DBPort = types.Int64Value(int64(check.DBPort))
	state.DBName = types.StringValue(check.DBName)
	state.DBUsername = types.StringValue(check.DBUsername)
	state.DBSSLEnabled = types.BoolValue(check.DBSSLEnabled)
	state.DBQuery = types.StringValue(check.DBQuery)
	state.DBExpectedResult = types.StringValue(check.DBExpectedResult)
	// Secrets (db_password, auth_password, auth_bearer_token, ftp_password) are
	// intentionally NOT refreshed — write-only, not returned by the API; keep
	// whatever is already in state/plan.
	state.AuthUsername = types.StringValue(check.AuthUsername)
	state.HTTPBody = types.StringValue(check.HTTPBody)
	state.HTTPBodyType = types.StringValue(check.HTTPBodyType)
	state.HTTPFollowRedirects = types.BoolValue(check.HTTPFollowRedirects)
	state.ContentMatchEnabled = types.BoolValue(check.ContentMatchEnabled)
	state.ContentMatchText = types.StringValue(check.ContentMatchText)
	state.ContentMatchType = types.StringValue(check.ContentMatchType)
	state.ContentMatchCaseSensitive = types.BoolValue(check.ContentMatchCaseSensitive)
	state.HTTPFormLoginEnabled = types.BoolValue(check.HTTPFormLoginEnabled)
	state.HTTPFormLoginURL = types.StringValue(check.HTTPFormLoginURL)
	state.HTTPFormLoginSuccessText = types.StringValue(check.HTTPFormLoginSuccessText)
	state.HTTPFormCheckAfterLoginURL = types.StringValue(check.HTTPFormCheckAfterLoginURL)
	state.FTPUsername = types.StringValue(check.FTPUsername)
	state.FTPProtocol = types.StringValue(check.FTPProtocol)
	state.FTPPath = types.StringValue(check.FTPPath)
	state.FTPPassive = types.BoolValue(check.FTPPassive)
	state.DNSSOAAlertOnChange = types.BoolValue(check.DNSSOAAlertOnChange)
	state.DNSHijackAlertEnabled = types.BoolValue(check.DNSHijackAlertEnabled)
	state.DNSHijackAlertChannelIDs = types.StringValue(check.DNSHijackAlertChannelIDs)
	state.DNSTXTMonitoringEnabled = types.BoolValue(check.DNSTXTMonitoringEnabled)
	state.DNSDKIMSelector = types.StringValue(check.DNSDKIMSelector)
	state.DNSMultiRecordEnabled = types.BoolValue(check.DNSMultiRecordEnabled)
	state.MailSMTPEnabled = types.BoolValue(check.MailSMTPEnabled)
	state.MailSMTPPort = types.Int64Value(int64(check.MailSMTPPort))
	state.MailSMTPStartTLS = types.BoolValue(check.MailSMTPStartTLS)
	state.MailSMTPOpenRelay = types.BoolValue(check.MailSMTPOpenRelay)
	state.MailIMAPEnabled = types.BoolValue(check.MailIMAPEnabled)
	state.MailIMAPPort = types.Int64Value(int64(check.MailIMAPPort))
	state.MailIMAPSSL = types.BoolValue(check.MailIMAPSSL)
	state.MailPOP3Enabled = types.BoolValue(check.MailPOP3Enabled)
	state.MailPOP3Port = types.Int64Value(int64(check.MailPOP3Port))
	state.MailPOP3SSL = types.BoolValue(check.MailPOP3SSL)
	state.MailCheckSPF = types.BoolValue(check.MailCheckSPF)
	state.MailCheckDKIM = types.BoolValue(check.MailCheckDKIM)
	state.MailDKIMSelectors = types.StringValue(check.MailDKIMSelectors)
	state.MailCheckDMARC = types.BoolValue(check.MailCheckDMARC)
	state.MailCheckPTR = types.BoolValue(check.MailCheckPTR)
	state.MailCheckBlacklist = types.BoolValue(check.MailCheckBlacklist)
	state.MailBlacklistServers = types.StringValue(check.MailBlacklistServers)

	if check.AssignedAgentID != nil {
		state.AssignedAgentID = types.Int64Value(int64(*check.AssignedAgentID))
	} else {
		state.AssignedAgentID = types.Int64Null()
	}
	if check.EscalationPolicyID != nil {
		state.EscalationPolicyID = types.Int64Value(int64(*check.EscalationPolicyID))
	} else {
		state.EscalationPolicyID = types.Int64Null()
	}

	escList, d := types.ListValueFrom(ctx, types.Int64Type, intsToInt64s(check.ExpectedStatusCodes))
	diags.Append(d...)
	state.ExpectedStatusCodes = escList
	ipList, d := types.ListValueFrom(ctx, types.StringType, check.DNSExpectedIPs)
	diags.Append(d...)
	state.DNSExpectedIPs = ipList
	agentList, d := types.ListValueFrom(ctx, types.Int64Type, intsToInt64s(check.AssignedAgentIDs))
	diags.Append(d...)
	state.AssignedAgentIDs = agentList

	state.HTTPHeaders = rawToNormalized(check.HTTPHeaders)
	state.HTTPFormLoginData = rawToNormalized(check.HTTPFormLoginData)
	state.APIScenarioSteps = rawToNormalized(check.APIScenarioSteps)
	state.OIDCConfig = rawToNormalized(check.OIDCConfig)
	state.DNSRecordsConfig = rawToNormalized(check.DNSRecordsConfig)
	state.CheckSourceCritical = rawToNormalized(check.CheckSourceCritical)
	state.ResponseAssertions = rawToNormalized(check.ResponseAssertions)
	// api_scenario_secrets is write-only — never refreshed from the API.

	return diags
}

func rawToNormalized(raw json.RawMessage) jsontypes.Normalized {
	s := string(raw)
	if len(raw) == 0 || s == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(s)
}

// normToRaw converts a plan/config jsontypes.Normalized into a request payload
// (nil when null/unknown so omitempty drops it).
func normToRaw(n jsontypes.Normalized) json.RawMessage {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	return json.RawMessage(n.ValueString())
}

// secretRawUpdate is the write-only-secret analogue for a JSON-object secret:
// send the new value if set, send an empty object {} to CLEAR it when it was set
// before and removed now, or nil (omit — untouched) when it was never set.
// NOTE: clear must be {} not null — the backend's update is `if value is not None`,
// so a JSON null would be skipped and the encrypted secret would linger.
func secretRawUpdate(plan, prior jsontypes.Normalized) json.RawMessage {
	if !plan.IsNull() && !plan.IsUnknown() {
		return json.RawMessage(plan.ValueString())
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		return json.RawMessage("{}")
	}
	return nil
}

// secretUpdate decides what to send for a write-only secret on update:
// new value if set, "" to clear it when it was previously set but now removed,
// or nil (omit — leave untouched) when it was never set.
func secretUpdate(planVal, priorVal string) *string {
	if planVal != "" {
		return &planVal
	}
	if priorVal != "" {
		empty := ""
		return &empty
	}
	return nil
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func intPtr(i int) *int       { return &i }

func intsToInt64s(in []int) []int64 {
	out := make([]int64, len(in))
	for i, v := range in {
		out[i] = int64(v)
	}
	return out
}

// listToInts / listToStrs convert a plan/config types.List into a Go slice.
func listToInts(ctx context.Context, l types.List) ([]int, diag.Diagnostics) {
	if l.IsNull() || l.IsUnknown() {
		return nil, nil
	}
	var v64 []int64
	diags := l.ElementsAs(ctx, &v64, false)
	out := make([]int, len(v64))
	for i, x := range v64 {
		out[i] = int(x)
	}
	return out, diags
}

func listToStrs(ctx context.Context, l types.List) ([]string, diag.Diagnostics) {
	if l.IsNull() || l.IsUnknown() {
		return nil, nil
	}
	var v []string
	diags := l.ElementsAs(ctx, &v, false)
	return v, diags
}
