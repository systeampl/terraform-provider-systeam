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
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &checkResource{}
	_ resource.ResourceWithConfigure   = &checkResource{}
	_ resource.ResourceWithImportState = &checkResource{}
)

type checkResource struct {
	sdk *syschecks.Client
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
				Sensitive:   true, // typically contains {username, password}
				Description: "Form fields for HTTP form login as a JSON object, e.g. jsonencode({username=\"u\",password=\"p\"}). Write-only: holds credentials, never read back from the API.",
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
	sdk, ok := req.ProviderData.(*syschecks.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *syschecks.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.sdk = sdk
}

func (r *checkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CheckModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// NOTE: models.CheckCreate has NO escalation_policy_id field (unlike CheckUpdate /
	// CheckResponse), so it cannot be set at create time via the SDK. It is still
	// mapped on Update and Read.
	createReq := models.CheckCreate{
		Name:                       plan.Name.ValueString(),
		Type:                       strPtr(plan.Type.ValueString()),
		Description:                strPtr(plan.Description.ValueString()),
		IsActive:                   boolPtr(plan.IsActive.ValueBool()),
		ProjectId:                  int(plan.ProjectID.ValueInt64()),
		Interval:                   intPtr(int(plan.Interval.ValueInt64())),
		GracePeriod:                intPtr(int(plan.GracePeriod.ValueInt64())),
		Timeout:                    intPtr(int(plan.Timeout.ValueInt64())),
		Url:                        strPtr(plan.URL.ValueString()),
		Host:                       strPtr(plan.Host.ValueString()),
		Port:                       intPtr(int(plan.Port.ValueInt64())),
		PingCount:                  intPtr(int(plan.PingCount.ValueInt64())),
		SslVerify:                  boolPtr(plan.SSLVerify.ValueBool()),
		AlertAfterFailures:         intPtr(int(plan.AlertAfterFailures.ValueInt64())),
		GeoMonitoringEnabled:       boolPtr(plan.GeoMonitoringEnabled.ValueBool()),
		TracerouteOnTimeout:        boolPtr(plan.TracerouteOnTimeout.ValueBool()),
		DnsServer:                  strPtr(plan.DNSServer.ValueString()),
		DnsRecordType:              strPtr(plan.DNSRecordType.ValueString()),
		DnsExpectedValue:           strPtr(plan.DNSExpectedValue.ValueString()),
		HttpMethod:                 strPtr(plan.HTTPMethod.ValueString()),
		AuthMethod:                 authMethodPtr(plan.AuthMethod.ValueString()),
		ScheduleType:               scheduleTypePtr(plan.ScheduleType.ValueString()),
		CronExpression:             strPtr(plan.CronExpression.ValueString()),
		CronTimezone:               strPtr(plan.CronTimezone.ValueString()),
		MailDomain:                 strPtr(plan.MailDomain.ValueString()),
		RunbookUrl:                 strPtr(plan.RunbookURL.ValueString()),
		DbType:                     strPtr(plan.DBType.ValueString()),
		DbHost:                     strPtr(plan.DBHost.ValueString()),
		DbPort:                     intPtr(int(plan.DBPort.ValueInt64())),
		DbName:                     strPtr(plan.DBName.ValueString()),
		DbUsername:                 strPtr(plan.DBUsername.ValueString()),
		DbPassword:                 strPtr(plan.DBPassword.ValueString()),
		DbSslEnabled:               boolPtr(plan.DBSSLEnabled.ValueBool()),
		DbQuery:                    strPtr(plan.DBQuery.ValueString()),
		DbExpectedResult:           strPtr(plan.DBExpectedResult.ValueString()),
		AuthUsername:               strPtr(plan.AuthUsername.ValueString()),
		AuthPassword:               strPtr(plan.AuthPassword.ValueString()),
		AuthBearerToken:            strPtr(plan.AuthBearerToken.ValueString()),
		HttpBody:                   strPtr(plan.HTTPBody.ValueString()),
		HttpBodyType:               strPtr(plan.HTTPBodyType.ValueString()),
		HttpFollowRedirects:        boolPtr(plan.HTTPFollowRedirects.ValueBool()),
		ContentMatchEnabled:        boolPtr(plan.ContentMatchEnabled.ValueBool()),
		ContentMatchText:           strPtr(plan.ContentMatchText.ValueString()),
		ContentMatchType:           strPtr(plan.ContentMatchType.ValueString()),
		ContentMatchCaseSensitive:  boolPtr(plan.ContentMatchCaseSensitive.ValueBool()),
		HttpFormLoginEnabled:       boolPtr(plan.HTTPFormLoginEnabled.ValueBool()),
		HttpFormLoginUrl:           strPtr(plan.HTTPFormLoginURL.ValueString()),
		HttpFormLoginSuccessText:   strPtr(plan.HTTPFormLoginSuccessText.ValueString()),
		HttpFormCheckAfterLoginUrl: strPtr(plan.HTTPFormCheckAfterLoginURL.ValueString()),
		FtpUsername:                strPtr(plan.FTPUsername.ValueString()),
		FtpPassword:                strPtr(plan.FTPPassword.ValueString()),
		FtpProtocol:                strPtr(plan.FTPProtocol.ValueString()),
		FtpPath:                    strPtr(plan.FTPPath.ValueString()),
		FtpPassive:                 boolPtr(plan.FTPPassive.ValueBool()),
		DnsSoaAlertOnChange:        boolPtr(plan.DNSSOAAlertOnChange.ValueBool()),
		DnsHijackAlertEnabled:      boolPtr(plan.DNSHijackAlertEnabled.ValueBool()),
		DnsHijackAlertChannelIds:   strPtr(plan.DNSHijackAlertChannelIDs.ValueString()),
		DnsTxtMonitoringEnabled:    boolPtr(plan.DNSTXTMonitoringEnabled.ValueBool()),
		DnsDkimSelector:            strPtr(plan.DNSDKIMSelector.ValueString()),
		DnsMultiRecordEnabled:      boolPtr(plan.DNSMultiRecordEnabled.ValueBool()),
		MailSmtpEnabled:            boolPtr(plan.MailSMTPEnabled.ValueBool()),
		MailSmtpPort:               intPtr(int(plan.MailSMTPPort.ValueInt64())),
		MailSmtpStarttls:           boolPtr(plan.MailSMTPStartTLS.ValueBool()),
		MailSmtpOpenRelay:          boolPtr(plan.MailSMTPOpenRelay.ValueBool()),
		MailImapEnabled:            boolPtr(plan.MailIMAPEnabled.ValueBool()),
		MailImapPort:               intPtr(int(plan.MailIMAPPort.ValueInt64())),
		MailImapSsl:                boolPtr(plan.MailIMAPSSL.ValueBool()),
		MailPop3Enabled:            boolPtr(plan.MailPOP3Enabled.ValueBool()),
		MailPop3Port:               intPtr(int(plan.MailPOP3Port.ValueInt64())),
		MailPop3Ssl:                boolPtr(plan.MailPOP3SSL.ValueBool()),
		MailCheckSpf:               boolPtr(plan.MailCheckSPF.ValueBool()),
		MailCheckDkim:              boolPtr(plan.MailCheckDKIM.ValueBool()),
		MailDkimSelectors:          strPtr(plan.MailDKIMSelectors.ValueString()),
		MailCheckDmarc:             boolPtr(plan.MailCheckDMARC.ValueBool()),
		MailCheckPtr:               boolPtr(plan.MailCheckPTR.ValueBool()),
		MailCheckBlacklist:         boolPtr(plan.MailCheckBlacklist.ValueBool()),
		MailBlacklistServers:       strPtr(plan.MailBlacklistServers.ValueString()),
	}

	createReq.AssignedAgentId = sdkutil.IntPtr(plan.AssignedAgentID)

	statusCodes, d := listToInts(ctx, plan.ExpectedStatusCodes)
	resp.Diagnostics.Append(d...)
	if statusCodes != nil {
		createReq.ExpectedStatusCodes = &statusCodes
	}
	ips, d := listToStrs(ctx, plan.DNSExpectedIPs)
	resp.Diagnostics.Append(d...)
	if ips != nil {
		createReq.DnsExpectedIps = &ips
	}
	agentIDs, d := listToInts(ctx, plan.AssignedAgentIDs)
	resp.Diagnostics.Append(d...)
	if agentIDs != nil {
		createReq.AssignedAgentIds = &agentIDs
	}

	createReq.HttpHeaders, d = normToMap(plan.HTTPHeaders)
	resp.Diagnostics.Append(d...)
	createReq.HttpFormLoginData, d = normToMap(plan.HTTPFormLoginData)
	resp.Diagnostics.Append(d...)
	createReq.ApiScenarioSteps, d = normToMap(plan.APIScenarioSteps)
	resp.Diagnostics.Append(d...)
	createReq.ApiScenarioSecrets, d = normToMap(plan.APIScenarioSecrets)
	resp.Diagnostics.Append(d...)
	createReq.OidcConfig, d = normToMap(plan.OIDCConfig)
	resp.Diagnostics.Append(d...)
	createReq.DnsRecordsConfig, d = normToMap(plan.DNSRecordsConfig)
	resp.Diagnostics.Append(d...)
	createReq.CheckSourceCritical, d = normToMap(plan.CheckSourceCritical)
	resp.Diagnostics.Append(d...)
	createReq.ResponseAssertions, d = normToSlice(plan.ResponseAssertions)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.sdk.Checks.CreateNewCheck(ctx, createReq)
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

	check, err := r.sdk.Checks.GetCheckDetails(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if sdkutil.IsNotFound(err) {
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

	updateReq := models.CheckUpdate{
		Name:                       strPtr(plan.Name.ValueString()),
		Description:                strPtr(plan.Description.ValueString()),
		IsActive:                   boolPtr(plan.IsActive.ValueBool()),
		Interval:                   intPtr(int(plan.Interval.ValueInt64())),
		GracePeriod:                intPtr(int(plan.GracePeriod.ValueInt64())),
		Timeout:                    intPtr(int(plan.Timeout.ValueInt64())),
		Url:                        strPtr(plan.URL.ValueString()),
		Host:                       strPtr(plan.Host.ValueString()),
		Port:                       intPtr(int(plan.Port.ValueInt64())),
		PingCount:                  intPtr(int(plan.PingCount.ValueInt64())),
		SslVerify:                  boolPtr(plan.SSLVerify.ValueBool()),
		AlertAfterFailures:         intPtr(int(plan.AlertAfterFailures.ValueInt64())),
		GeoMonitoringEnabled:       boolPtr(plan.GeoMonitoringEnabled.ValueBool()),
		TracerouteOnTimeout:        boolPtr(plan.TracerouteOnTimeout.ValueBool()),
		DnsServer:                  strPtr(plan.DNSServer.ValueString()),
		DnsRecordType:              strPtr(plan.DNSRecordType.ValueString()),
		DnsExpectedValue:           strPtr(plan.DNSExpectedValue.ValueString()),
		HttpMethod:                 strPtr(plan.HTTPMethod.ValueString()),
		AuthMethod:                 authMethodPtr(plan.AuthMethod.ValueString()),
		ScheduleType:               scheduleTypePtr(plan.ScheduleType.ValueString()),
		CronExpression:             strPtr(plan.CronExpression.ValueString()),
		CronTimezone:               strPtr(plan.CronTimezone.ValueString()),
		MailDomain:                 strPtr(plan.MailDomain.ValueString()),
		RunbookUrl:                 strPtr(plan.RunbookURL.ValueString()),
		DbType:                     strPtr(plan.DBType.ValueString()),
		DbHost:                     strPtr(plan.DBHost.ValueString()),
		DbPort:                     intPtr(int(plan.DBPort.ValueInt64())),
		DbName:                     strPtr(plan.DBName.ValueString()),
		DbUsername:                 strPtr(plan.DBUsername.ValueString()),
		DbSslEnabled:               boolPtr(plan.DBSSLEnabled.ValueBool()),
		DbQuery:                    strPtr(plan.DBQuery.ValueString()),
		DbExpectedResult:           strPtr(plan.DBExpectedResult.ValueString()),
		AuthUsername:               strPtr(plan.AuthUsername.ValueString()),
		HttpBody:                   strPtr(plan.HTTPBody.ValueString()),
		HttpBodyType:               strPtr(plan.HTTPBodyType.ValueString()),
		HttpFollowRedirects:        boolPtr(plan.HTTPFollowRedirects.ValueBool()),
		ContentMatchEnabled:        boolPtr(plan.ContentMatchEnabled.ValueBool()),
		ContentMatchText:           strPtr(plan.ContentMatchText.ValueString()),
		ContentMatchType:           strPtr(plan.ContentMatchType.ValueString()),
		ContentMatchCaseSensitive:  boolPtr(plan.ContentMatchCaseSensitive.ValueBool()),
		HttpFormLoginEnabled:       boolPtr(plan.HTTPFormLoginEnabled.ValueBool()),
		HttpFormLoginUrl:           strPtr(plan.HTTPFormLoginURL.ValueString()),
		HttpFormLoginSuccessText:   strPtr(plan.HTTPFormLoginSuccessText.ValueString()),
		HttpFormCheckAfterLoginUrl: strPtr(plan.HTTPFormCheckAfterLoginURL.ValueString()),
		FtpUsername:                strPtr(plan.FTPUsername.ValueString()),
		FtpProtocol:                strPtr(plan.FTPProtocol.ValueString()),
		FtpPath:                    strPtr(plan.FTPPath.ValueString()),
		FtpPassive:                 boolPtr(plan.FTPPassive.ValueBool()),
		DnsSoaAlertOnChange:        boolPtr(plan.DNSSOAAlertOnChange.ValueBool()),
		DnsHijackAlertEnabled:      boolPtr(plan.DNSHijackAlertEnabled.ValueBool()),
		DnsHijackAlertChannelIds:   strPtr(plan.DNSHijackAlertChannelIDs.ValueString()),
		DnsTxtMonitoringEnabled:    boolPtr(plan.DNSTXTMonitoringEnabled.ValueBool()),
		DnsDkimSelector:            strPtr(plan.DNSDKIMSelector.ValueString()),
		DnsMultiRecordEnabled:      boolPtr(plan.DNSMultiRecordEnabled.ValueBool()),
		MailSmtpEnabled:            boolPtr(plan.MailSMTPEnabled.ValueBool()),
		MailSmtpPort:               intPtr(int(plan.MailSMTPPort.ValueInt64())),
		MailSmtpStarttls:           boolPtr(plan.MailSMTPStartTLS.ValueBool()),
		MailSmtpOpenRelay:          boolPtr(plan.MailSMTPOpenRelay.ValueBool()),
		MailImapEnabled:            boolPtr(plan.MailIMAPEnabled.ValueBool()),
		MailImapPort:               intPtr(int(plan.MailIMAPPort.ValueInt64())),
		MailImapSsl:                boolPtr(plan.MailIMAPSSL.ValueBool()),
		MailPop3Enabled:            boolPtr(plan.MailPOP3Enabled.ValueBool()),
		MailPop3Port:               intPtr(int(plan.MailPOP3Port.ValueInt64())),
		MailPop3Ssl:                boolPtr(plan.MailPOP3SSL.ValueBool()),
		MailCheckSpf:               boolPtr(plan.MailCheckSPF.ValueBool()),
		MailCheckDkim:              boolPtr(plan.MailCheckDKIM.ValueBool()),
		MailDkimSelectors:          strPtr(plan.MailDKIMSelectors.ValueString()),
		MailCheckDmarc:             boolPtr(plan.MailCheckDMARC.ValueBool()),
		MailCheckPtr:               boolPtr(plan.MailCheckPTR.ValueBool()),
		MailCheckBlacklist:         boolPtr(plan.MailCheckBlacklist.ValueBool()),
		MailBlacklistServers:       strPtr(plan.MailBlacklistServers.ValueString()),
	}

	// Write-only secrets: send the new value when set; send "" to CLEAR it when it
	// was set before but removed now; leave nil (untouched) when it was never set.
	updateReq.AuthPassword = secretUpdate(plan.AuthPassword.ValueString(), prior.AuthPassword.ValueString())
	updateReq.AuthBearerToken = secretUpdate(plan.AuthBearerToken.ValueString(), prior.AuthBearerToken.ValueString())
	updateReq.FtpPassword = secretUpdate(plan.FTPPassword.ValueString(), prior.FTPPassword.ValueString())
	updateReq.DbPassword = secretUpdate(plan.DBPassword.ValueString(), prior.DBPassword.ValueString())

	updateReq.AssignedAgentId = sdkutil.IntPtr(plan.AssignedAgentID)
	updateReq.EscalationPolicyId = sdkutil.IntPtr(plan.EscalationPolicyID)

	scodes, d := listToInts(ctx, plan.ExpectedStatusCodes)
	resp.Diagnostics.Append(d...)
	if scodes != nil {
		updateReq.ExpectedStatusCodes = &scodes
	}
	uips, d := listToStrs(ctx, plan.DNSExpectedIPs)
	resp.Diagnostics.Append(d...)
	if uips != nil {
		updateReq.DnsExpectedIps = &uips
	}
	uagents, d := listToInts(ctx, plan.AssignedAgentIDs)
	resp.Diagnostics.Append(d...)
	if uagents != nil {
		updateReq.AssignedAgentIds = &uagents
	}

	updateReq.HttpHeaders, d = normToMap(plan.HTTPHeaders)
	resp.Diagnostics.Append(d...)
	updateReq.HttpFormLoginData, d = normToMap(plan.HTTPFormLoginData)
	resp.Diagnostics.Append(d...)
	updateReq.ApiScenarioSteps, d = normToMap(plan.APIScenarioSteps)
	resp.Diagnostics.Append(d...)
	updateReq.OidcConfig, d = normToMap(plan.OIDCConfig)
	resp.Diagnostics.Append(d...)
	updateReq.DnsRecordsConfig, d = normToMap(plan.DNSRecordsConfig)
	resp.Diagnostics.Append(d...)
	updateReq.CheckSourceCritical, d = normToMap(plan.CheckSourceCritical)
	resp.Diagnostics.Append(d...)
	updateReq.ResponseAssertions, d = normToSlice(plan.ResponseAssertions)
	resp.Diagnostics.Append(d...)
	updateReq.ApiScenarioSecrets, d = secretMapUpdate(plan.APIScenarioSecrets, prior.APIScenarioSecrets)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.sdk.Checks.UpdateCheck(ctx, int(plan.ID.ValueInt64()), updateReq)
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

	if _, err := r.sdk.Checks.RemoveCheck(ctx, int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
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

func mapCheckToState(ctx context.Context, check *models.CheckResponse, state *CheckModel) diag.Diagnostics {
	var diags diag.Diagnostics
	state.ID = types.Int64Value(int64(check.Id))
	state.Name = types.StringValue(check.Name)
	state.Type = strVal(check.Type)
	state.Description = strVal(check.Description)
	state.IsActive = boolVal(check.IsActive)
	state.ProjectID = intVal(check.ProjectId)
	state.Interval = intVal(check.Interval)
	state.GracePeriod = intVal(check.GracePeriod)
	state.Timeout = intVal(check.Timeout)
	state.URL = strVal(check.Url)
	state.Host = strVal(check.Host)
	state.Port = intVal(check.Port)
	state.PingCount = intVal(check.PingCount)
	state.SSLVerify = boolVal(check.SslVerify)
	state.AlertAfterFailures = intVal(check.AlertAfterFailures)
	state.GeoMonitoringEnabled = boolVal(check.GeoMonitoringEnabled)
	state.TracerouteOnTimeout = boolVal(check.TracerouteOnTimeout)
	state.DNSServer = strVal(check.DnsServer)
	state.DNSRecordType = strVal(check.DnsRecordType)
	state.DNSExpectedValue = strVal(check.DnsExpectedValue)
	state.HTTPMethod = strVal(check.HttpMethod)
	state.AuthMethod = authMethodVal(check.AuthMethod)
	state.ScheduleType = scheduleTypeVal(check.ScheduleType)
	state.CronExpression = strVal(check.CronExpression)
	state.CronTimezone = strVal(check.CronTimezone)
	state.MailDomain = strVal(check.MailDomain)
	state.RunbookURL = strVal(check.RunbookUrl)
	state.DBType = strVal(check.DbType)
	state.DBHost = strVal(check.DbHost)
	state.DBPort = intVal(check.DbPort)
	state.DBName = strVal(check.DbName)
	state.DBUsername = strVal(check.DbUsername)
	state.DBSSLEnabled = boolVal(check.DbSslEnabled)
	state.DBQuery = strVal(check.DbQuery)
	state.DBExpectedResult = strVal(check.DbExpectedResult)
	// Secrets (db_password, auth_password, auth_bearer_token, ftp_password) are
	// intentionally NOT refreshed — write-only, not returned by the API; keep
	// whatever is already in state/plan.
	state.AuthUsername = strVal(check.AuthUsername)
	state.HTTPBody = strVal(check.HttpBody)
	state.HTTPBodyType = strVal(check.HttpBodyType)
	state.HTTPFollowRedirects = boolVal(check.HttpFollowRedirects)
	state.ContentMatchEnabled = boolVal(check.ContentMatchEnabled)
	state.ContentMatchText = strVal(check.ContentMatchText)
	state.ContentMatchType = strVal(check.ContentMatchType)
	state.ContentMatchCaseSensitive = boolVal(check.ContentMatchCaseSensitive)
	state.HTTPFormLoginEnabled = boolVal(check.HttpFormLoginEnabled)
	state.HTTPFormLoginURL = strVal(check.HttpFormLoginUrl)
	state.HTTPFormLoginSuccessText = strVal(check.HttpFormLoginSuccessText)
	state.HTTPFormCheckAfterLoginURL = strVal(check.HttpFormCheckAfterLoginUrl)
	state.FTPUsername = strVal(check.FtpUsername)
	state.FTPProtocol = strVal(check.FtpProtocol)
	state.FTPPath = strVal(check.FtpPath)
	state.FTPPassive = boolVal(check.FtpPassive)
	state.DNSSOAAlertOnChange = boolVal(check.DnsSoaAlertOnChange)
	state.DNSHijackAlertEnabled = boolVal(check.DnsHijackAlertEnabled)
	state.DNSHijackAlertChannelIDs = strVal(check.DnsHijackAlertChannelIds)
	state.DNSTXTMonitoringEnabled = boolVal(check.DnsTxtMonitoringEnabled)
	state.DNSDKIMSelector = strVal(check.DnsDkimSelector)
	state.DNSMultiRecordEnabled = boolVal(check.DnsMultiRecordEnabled)
	state.MailSMTPEnabled = boolVal(check.MailSmtpEnabled)
	state.MailSMTPPort = intVal(check.MailSmtpPort)
	state.MailSMTPStartTLS = boolVal(check.MailSmtpStarttls)
	state.MailSMTPOpenRelay = boolVal(check.MailSmtpOpenRelay)
	state.MailIMAPEnabled = boolVal(check.MailImapEnabled)
	state.MailIMAPPort = intVal(check.MailImapPort)
	state.MailIMAPSSL = boolVal(check.MailImapSsl)
	state.MailPOP3Enabled = boolVal(check.MailPop3Enabled)
	state.MailPOP3Port = intVal(check.MailPop3Port)
	state.MailPOP3SSL = boolVal(check.MailPop3Ssl)
	state.MailCheckSPF = boolVal(check.MailCheckSpf)
	state.MailCheckDKIM = boolVal(check.MailCheckDkim)
	state.MailDKIMSelectors = strVal(check.MailDkimSelectors)
	state.MailCheckDMARC = boolVal(check.MailCheckDmarc)
	state.MailCheckPTR = boolVal(check.MailCheckPtr)
	state.MailCheckBlacklist = boolVal(check.MailCheckBlacklist)
	state.MailBlacklistServers = strVal(check.MailBlacklistServers)

	state.AssignedAgentID = sdkutil.Int(check.AssignedAgentId)
	state.EscalationPolicyID = sdkutil.Int(check.EscalationPolicyId)

	escList, d := types.ListValueFrom(ctx, types.Int64Type, intsToInt64s(derefInts(check.ExpectedStatusCodes)))
	diags.Append(d...)
	state.ExpectedStatusCodes = escList
	ipList, d := types.ListValueFrom(ctx, types.StringType, derefStrs(check.DnsExpectedIps))
	diags.Append(d...)
	state.DNSExpectedIPs = ipList
	agentList, d := types.ListValueFrom(ctx, types.Int64Type, intsToInt64s(derefInts(check.AssignedAgentIds)))
	diags.Append(d...)
	state.AssignedAgentIDs = agentList

	state.HTTPHeaders = mapToNormalized(check.HttpHeaders)
	state.APIScenarioSteps = mapToNormalized(check.ApiScenarioSteps)
	state.OIDCConfig = mapToNormalized(check.OidcConfig)
	state.DNSRecordsConfig = mapToNormalized(check.DnsRecordsConfig)
	state.CheckSourceCritical = mapToNormalized(check.CheckSourceCritical)
	state.ResponseAssertions = sliceToNormalized(check.ResponseAssertions)
	// http_form_login_data and api_scenario_secrets are write-only secrets — the API no
	// longer returns them on read (they hold credentials), so they are never refreshed
	// from the API; the prior plan/state value is preserved to keep `plan` a no-op.

	return diags
}

// strVal / intVal / boolVal deref an SDK response pointer into a concrete framework
// value (zero value when nil), keeping Computed attributes known rather than null.
func strVal(p *string) types.String {
	if p == nil {
		return types.StringValue("")
	}
	return types.StringValue(*p)
}

func intVal(p *int) types.Int64 {
	if p == nil {
		return types.Int64Value(0)
	}
	return types.Int64Value(int64(*p))
}

func boolVal(p *bool) types.Bool {
	if p == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(*p)
}

func authMethodVal(p *models.AuthMethod) types.String {
	if p == nil {
		return types.StringValue("")
	}
	return types.StringValue(string(*p))
}

func scheduleTypeVal(p *models.ScheduleType) types.String {
	if p == nil {
		return types.StringValue("")
	}
	return types.StringValue(string(*p))
}

func authMethodPtr(s string) *models.AuthMethod {
	if s == "" {
		return nil
	}
	m := models.AuthMethod(s)
	return &m
}

func scheduleTypePtr(s string) *models.ScheduleType {
	if s == "" {
		return nil
	}
	t := models.ScheduleType(s)
	return &t
}

// normToMap converts a plan/config JSON-object Normalized into the SDK's
// *map[string]interface{} (nil when null/unknown so omitempty drops it).
func normToMap(n jsontypes.Normalized) (*map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if n.IsNull() || n.IsUnknown() {
		return nil, diags
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(n.ValueString()), &m); err != nil {
		diags.AddError("Invalid JSON object", err.Error())
		return nil, diags
	}
	return &m, diags
}

// normToSlice converts a plan/config JSON-array Normalized into the SDK's
// *[]interface{} (nil when null/unknown so omitempty drops it).
func normToSlice(n jsontypes.Normalized) (*[]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if n.IsNull() || n.IsUnknown() {
		return nil, diags
	}
	var s []interface{}
	if err := json.Unmarshal([]byte(n.ValueString()), &s); err != nil {
		diags.AddError("Invalid JSON array", err.Error())
		return nil, diags
	}
	return &s, diags
}

// mapToNormalized / sliceToNormalized marshal an SDK response blob back into a
// semantic-equality Normalized value (null when the pointer is nil).
func mapToNormalized(m *map[string]interface{}) jsontypes.Normalized {
	if m == nil {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(*m)
	if err != nil || string(b) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}

func sliceToNormalized(s *[]interface{}) jsontypes.Normalized {
	if s == nil {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(*s)
	if err != nil || string(b) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}

// secretMapUpdate is the write-only-secret analogue for a JSON-object secret:
// send the new value if set, send an empty object {} to CLEAR it when it was set
// before and removed now, or nil (omit — untouched) when it was never set.
// NOTE: clear must be {} not null — the backend's update is `if value is not None`,
// so a JSON null would be skipped and the encrypted secret would linger.
func secretMapUpdate(plan, prior jsontypes.Normalized) (*map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !plan.IsNull() && !plan.IsUnknown() {
		return normToMap(plan)
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		empty := map[string]interface{}{}
		return &empty, diags
	}
	return nil, diags
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

func derefInts(p *[]int) []int {
	if p == nil {
		return nil
	}
	return *p
}

func derefStrs(p *[]string) []string {
	if p == nil {
		return nil
	}
	return *p
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
