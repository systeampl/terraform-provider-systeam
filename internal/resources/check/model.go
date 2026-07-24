package check

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CheckModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Description          types.String `tfsdk:"description"`
	IsActive             types.Bool   `tfsdk:"is_active"`
	ProjectID            types.Int64  `tfsdk:"project_id"`
	Interval             types.Int64  `tfsdk:"interval"`
	GracePeriod          types.Int64  `tfsdk:"grace_period"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	URL                  types.String `tfsdk:"url"`
	Host                 types.String `tfsdk:"host"`
	Port                 types.Int64  `tfsdk:"port"`
	PingCount            types.Int64  `tfsdk:"ping_count"`
	SSLVerify            types.Bool   `tfsdk:"ssl_verify"`
	AlertAfterFailures   types.Int64  `tfsdk:"alert_after_failures"`
	GeoMonitoringEnabled types.Bool   `tfsdk:"geo_monitoring_enabled"`
	AssignedAgentID      types.Int64  `tfsdk:"assigned_agent_id"`
	EscalationPolicyID   types.Int64  `tfsdk:"escalation_policy_id"`
	TracerouteOnTimeout  types.Bool   `tfsdk:"traceroute_on_timeout"`
	DNSServer            types.String `tfsdk:"dns_server"`
	DNSRecordType        types.String `tfsdk:"dns_record_type"`
	DNSExpectedValue     types.String `tfsdk:"dns_expected_value"`
	HTTPMethod           types.String `tfsdk:"http_method"`
	AuthMethod           types.String `tfsdk:"auth_method"`
	ScheduleType         types.String `tfsdk:"schedule_type"`
	CronExpression       types.String `tfsdk:"cron_expression"`
	CronTimezone         types.String `tfsdk:"cron_timezone"`
	MailDomain           types.String `tfsdk:"mail_domain"`
	RunbookURL           types.String `tfsdk:"runbook_url"`
	// Database checks (type = "database")
	DBType           types.String `tfsdk:"db_type"`
	DBHost           types.String `tfsdk:"db_host"`
	DBPort           types.Int64  `tfsdk:"db_port"`
	DBName           types.String `tfsdk:"db_name"`
	DBUsername       types.String `tfsdk:"db_username"`
	DBPassword       types.String `tfsdk:"db_password"` // write-only secret; never read back
	DBSSLEnabled     types.Bool   `tfsdk:"db_ssl_enabled"`
	DBQuery          types.String `tfsdk:"db_query"`
	DBExpectedResult types.String `tfsdk:"db_expected_result"`
	// HTTP auth (auth_password/auth_bearer_token are write-only secrets)
	AuthUsername    types.String `tfsdk:"auth_username"`
	AuthPassword    types.String `tfsdk:"auth_password"`
	AuthBearerToken types.String `tfsdk:"auth_bearer_token"`
	// HTTP request
	HTTPBody            types.String `tfsdk:"http_body"`
	HTTPBodyType        types.String `tfsdk:"http_body_type"`
	HTTPFollowRedirects types.Bool   `tfsdk:"http_follow_redirects"`
	// Content matching
	ContentMatchEnabled       types.Bool   `tfsdk:"content_match_enabled"`
	ContentMatchText          types.String `tfsdk:"content_match_text"`
	ContentMatchType          types.String `tfsdk:"content_match_type"`
	ContentMatchCaseSensitive types.Bool   `tfsdk:"content_match_case_sensitive"`
	// HTTP form login
	HTTPFormLoginEnabled       types.Bool   `tfsdk:"http_form_login_enabled"`
	HTTPFormLoginURL           types.String `tfsdk:"http_form_login_url"`
	HTTPFormLoginSuccessText   types.String `tfsdk:"http_form_login_success_text"`
	HTTPFormCheckAfterLoginURL types.String `tfsdk:"http_form_check_after_login_url"`
	// FTP/SFTP (ftp_password is a write-only secret)
	FTPUsername types.String `tfsdk:"ftp_username"`
	FTPPassword types.String `tfsdk:"ftp_password"`
	FTPProtocol types.String `tfsdk:"ftp_protocol"`
	FTPPath     types.String `tfsdk:"ftp_path"`
	FTPPassive  types.Bool   `tfsdk:"ftp_passive"`
	// DNS advanced
	DNSSOAAlertOnChange      types.Bool   `tfsdk:"dns_soa_alert_on_change"`
	DNSHijackAlertEnabled    types.Bool   `tfsdk:"dns_hijack_alert_enabled"`
	DNSHijackAlertChannelIDs types.String `tfsdk:"dns_hijack_alert_channel_ids"`
	DNSTXTMonitoringEnabled  types.Bool   `tfsdk:"dns_txt_monitoring_enabled"`
	DNSDKIMSelector          types.String `tfsdk:"dns_dkim_selector"`
	DNSMultiRecordEnabled    types.Bool   `tfsdk:"dns_multi_record_enabled"`
	// Mail server
	MailSMTPEnabled      types.Bool   `tfsdk:"mail_smtp_enabled"`
	MailSMTPPort         types.Int64  `tfsdk:"mail_smtp_port"`
	MailSMTPStartTLS     types.Bool   `tfsdk:"mail_smtp_starttls"`
	MailSMTPOpenRelay    types.Bool   `tfsdk:"mail_smtp_open_relay"`
	MailIMAPEnabled      types.Bool   `tfsdk:"mail_imap_enabled"`
	MailIMAPPort         types.Int64  `tfsdk:"mail_imap_port"`
	MailIMAPSSL          types.Bool   `tfsdk:"mail_imap_ssl"`
	MailPOP3Enabled      types.Bool   `tfsdk:"mail_pop3_enabled"`
	MailPOP3Port         types.Int64  `tfsdk:"mail_pop3_port"`
	MailPOP3SSL          types.Bool   `tfsdk:"mail_pop3_ssl"`
	MailCheckSPF         types.Bool   `tfsdk:"mail_check_spf"`
	MailCheckDKIM        types.Bool   `tfsdk:"mail_check_dkim"`
	MailDKIMSelectors    types.String `tfsdk:"mail_dkim_selectors"`
	MailCheckDMARC       types.Bool   `tfsdk:"mail_check_dmarc"`
	MailCheckPTR         types.Bool   `tfsdk:"mail_check_ptr"`
	MailCheckBlacklist   types.Bool   `tfsdk:"mail_check_blacklist"`
	MailBlacklistServers types.String `tfsdk:"mail_blacklist_servers"`
	// List fields
	ExpectedStatusCodes types.List `tfsdk:"expected_status_codes"` // list[int]
	DNSExpectedIPs      types.List `tfsdk:"dns_expected_ips"`      // list[string]
	AssignedAgentIDs    types.List `tfsdk:"assigned_agent_ids"`    // list[int] — multi/geo agents
	// JSON-object fields (semantic equality — whitespace/key-order independent)
	HTTPHeaders         jsontypes.Normalized `tfsdk:"http_headers"`          // {"X-Api-Key":"..."}
	HTTPFormLoginData   jsontypes.Normalized `tfsdk:"http_form_login_data"`  // {"username":"...","password":"..."}
	APIScenarioSteps    jsontypes.Normalized `tfsdk:"api_scenario_steps"`    // api_scenario type
	APIScenarioSecrets  jsontypes.Normalized `tfsdk:"api_scenario_secrets"`  // write-only secret
	OIDCConfig          jsontypes.Normalized `tfsdk:"oidc_config"`           // oidc type
	DNSRecordsConfig    jsontypes.Normalized `tfsdk:"dns_records_config"`    // multi-record DNS
	CheckSourceCritical jsontypes.Normalized `tfsdk:"check_source_critical"` // per-source criticality
	ResponseAssertions  jsontypes.Normalized `tfsdk:"response_assertions"`   // JSON array of assertions
}
