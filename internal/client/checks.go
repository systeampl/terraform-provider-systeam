package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Check struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	Type                 string `json:"type"`
	Description          string `json:"description"`
	IsActive             bool   `json:"is_active"`
	ProjectID            int    `json:"project_id"`
	Interval             int    `json:"interval"`
	GracePeriod          int    `json:"grace_period"`
	Timeout              int    `json:"timeout"`
	URL                  string `json:"url"`
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	PingCount            int    `json:"ping_count"`
	SSLVerify            bool   `json:"ssl_verify"`
	AlertAfterFailures   int    `json:"alert_after_failures"`
	GeoMonitoringEnabled bool   `json:"geo_monitoring_enabled"`
	AssignedAgentID      *int   `json:"assigned_agent_id"`
	EscalationPolicyID   *int   `json:"escalation_policy_id"`
	TracerouteOnTimeout  bool   `json:"traceroute_on_timeout"`
	DNSServer            string `json:"dns_server"`
	DNSRecordType        string `json:"dns_record_type"`
	DNSExpectedValue     string `json:"dns_expected_value"`
	HTTPMethod           string `json:"http_method"`
	AuthMethod           string `json:"auth_method"`
	ScheduleType         string `json:"schedule_type"`
	CronExpression       string `json:"cron_expression"`
	CronTimezone         string `json:"cron_timezone"`
	MailDomain           string `json:"mail_domain"`
	RunbookURL           string `json:"runbook_url"`
	// Database checks. db_password is write-only — the API never returns it,
	// so it is intentionally absent from this response struct.
	DBType           string `json:"db_type"`
	DBHost           string `json:"db_host"`
	DBPort           int    `json:"db_port"`
	DBName           string `json:"db_name"`
	DBUsername       string `json:"db_username"`
	DBSSLEnabled     bool   `json:"db_ssl_enabled"`
	DBQuery          string `json:"db_query"`
	DBExpectedResult string `json:"db_expected_result"`
	// HTTP auth/request (auth_password & auth_bearer_token are write-only — omitted here)
	AuthUsername               string `json:"auth_username"`
	HTTPBody                   string `json:"http_body"`
	HTTPBodyType               string `json:"http_body_type"`
	HTTPFollowRedirects        bool   `json:"http_follow_redirects"`
	ContentMatchEnabled        bool   `json:"content_match_enabled"`
	ContentMatchText           string `json:"content_match_text"`
	ContentMatchType           string `json:"content_match_type"`
	ContentMatchCaseSensitive  bool   `json:"content_match_case_sensitive"`
	HTTPFormLoginEnabled       bool   `json:"http_form_login_enabled"`
	HTTPFormLoginURL           string `json:"http_form_login_url"`
	HTTPFormLoginSuccessText   string `json:"http_form_login_success_text"`
	HTTPFormCheckAfterLoginURL string `json:"http_form_check_after_login_url"`
	// FTP (ftp_password write-only — omitted here)
	FTPUsername string `json:"ftp_username"`
	FTPProtocol string `json:"ftp_protocol"`
	FTPPath     string `json:"ftp_path"`
	FTPPassive  bool   `json:"ftp_passive"`
	// DNS advanced
	DNSSOAAlertOnChange      bool   `json:"dns_soa_alert_on_change"`
	DNSHijackAlertEnabled    bool   `json:"dns_hijack_alert_enabled"`
	DNSHijackAlertChannelIDs string `json:"dns_hijack_alert_channel_ids"`
	DNSTXTMonitoringEnabled  bool   `json:"dns_txt_monitoring_enabled"`
	DNSDKIMSelector          string `json:"dns_dkim_selector"`
	DNSMultiRecordEnabled    bool   `json:"dns_multi_record_enabled"`
	// Mail server
	MailSMTPEnabled      bool            `json:"mail_smtp_enabled"`
	MailSMTPPort         int             `json:"mail_smtp_port"`
	MailSMTPStartTLS     bool            `json:"mail_smtp_starttls"`
	MailSMTPOpenRelay    bool            `json:"mail_smtp_open_relay"`
	MailIMAPEnabled      bool            `json:"mail_imap_enabled"`
	MailIMAPPort         int             `json:"mail_imap_port"`
	MailIMAPSSL          bool            `json:"mail_imap_ssl"`
	MailPOP3Enabled      bool            `json:"mail_pop3_enabled"`
	MailPOP3Port         int             `json:"mail_pop3_port"`
	MailPOP3SSL          bool            `json:"mail_pop3_ssl"`
	MailCheckSPF         bool            `json:"mail_check_spf"`
	MailCheckDKIM        bool            `json:"mail_check_dkim"`
	MailDKIMSelectors    string          `json:"mail_dkim_selectors"`
	MailCheckDMARC       bool            `json:"mail_check_dmarc"`
	MailCheckPTR         bool            `json:"mail_check_ptr"`
	MailCheckBlacklist   bool            `json:"mail_check_blacklist"`
	MailBlacklistServers string          `json:"mail_blacklist_servers"`
	ExpectedStatusCodes  []int           `json:"expected_status_codes"`
	DNSExpectedIPs       []string        `json:"dns_expected_ips"`
	AssignedAgentIDs     []int           `json:"assigned_agent_ids"`
	HTTPHeaders          json.RawMessage `json:"http_headers"`
	HTTPFormLoginData    json.RawMessage `json:"http_form_login_data"`
	APIScenarioSteps     json.RawMessage `json:"api_scenario_steps"`
	OIDCConfig           json.RawMessage `json:"oidc_config"`
	DNSRecordsConfig     json.RawMessage `json:"dns_records_config"`
	CheckSourceCritical  json.RawMessage `json:"check_source_critical"`
	ResponseAssertions   json.RawMessage `json:"response_assertions"`
}

type CheckCreateRequest struct {
	Name                       string          `json:"name"`
	Type                       string          `json:"type"`
	Description                string          `json:"description,omitempty"`
	IsActive                   bool            `json:"is_active"`
	ProjectID                  int             `json:"project_id"`
	Interval                   int             `json:"interval,omitempty"`
	GracePeriod                int             `json:"grace_period,omitempty"`
	Timeout                    int             `json:"timeout,omitempty"`
	URL                        string          `json:"url,omitempty"`
	Host                       string          `json:"host,omitempty"`
	Port                       int             `json:"port,omitempty"`
	PingCount                  int             `json:"ping_count,omitempty"`
	SSLVerify                  bool            `json:"ssl_verify"`
	AlertAfterFailures         int             `json:"alert_after_failures,omitempty"`
	GeoMonitoringEnabled       bool            `json:"geo_monitoring_enabled"`
	AssignedAgentID            *int            `json:"assigned_agent_id,omitempty"`
	EscalationPolicyID         *int            `json:"escalation_policy_id,omitempty"`
	TracerouteOnTimeout        bool            `json:"traceroute_on_timeout"`
	DNSServer                  string          `json:"dns_server,omitempty"`
	DNSRecordType              string          `json:"dns_record_type,omitempty"`
	DNSExpectedValue           string          `json:"dns_expected_value,omitempty"`
	HTTPMethod                 string          `json:"http_method,omitempty"`
	AuthMethod                 string          `json:"auth_method,omitempty"`
	ScheduleType               string          `json:"schedule_type,omitempty"`
	CronExpression             string          `json:"cron_expression,omitempty"`
	CronTimezone               string          `json:"cron_timezone,omitempty"`
	MailDomain                 string          `json:"mail_domain,omitempty"`
	RunbookURL                 string          `json:"runbook_url,omitempty"`
	DBType                     string          `json:"db_type,omitempty"`
	DBHost                     string          `json:"db_host,omitempty"`
	DBPort                     int             `json:"db_port,omitempty"`
	DBName                     string          `json:"db_name,omitempty"`
	DBUsername                 string          `json:"db_username,omitempty"`
	DBPassword                 string          `json:"db_password,omitempty"`
	DBSSLEnabled               bool            `json:"db_ssl_enabled"`
	DBQuery                    string          `json:"db_query,omitempty"`
	DBExpectedResult           string          `json:"db_expected_result,omitempty"`
	AuthUsername               string          `json:"auth_username,omitempty"`
	AuthPassword               string          `json:"auth_password,omitempty"`
	AuthBearerToken            string          `json:"auth_bearer_token,omitempty"`
	HTTPBody                   string          `json:"http_body,omitempty"`
	HTTPBodyType               string          `json:"http_body_type,omitempty"`
	HTTPFollowRedirects        bool            `json:"http_follow_redirects"`
	ContentMatchEnabled        bool            `json:"content_match_enabled"`
	ContentMatchText           string          `json:"content_match_text,omitempty"`
	ContentMatchType           string          `json:"content_match_type,omitempty"`
	ContentMatchCaseSensitive  bool            `json:"content_match_case_sensitive"`
	HTTPFormLoginEnabled       bool            `json:"http_form_login_enabled"`
	HTTPFormLoginURL           string          `json:"http_form_login_url,omitempty"`
	HTTPFormLoginSuccessText   string          `json:"http_form_login_success_text,omitempty"`
	HTTPFormCheckAfterLoginURL string          `json:"http_form_check_after_login_url,omitempty"`
	FTPUsername                string          `json:"ftp_username,omitempty"`
	FTPPassword                string          `json:"ftp_password,omitempty"`
	FTPProtocol                string          `json:"ftp_protocol,omitempty"`
	FTPPath                    string          `json:"ftp_path,omitempty"`
	FTPPassive                 bool            `json:"ftp_passive"`
	DNSSOAAlertOnChange        bool            `json:"dns_soa_alert_on_change"`
	DNSHijackAlertEnabled      bool            `json:"dns_hijack_alert_enabled"`
	DNSHijackAlertChannelIDs   string          `json:"dns_hijack_alert_channel_ids,omitempty"`
	DNSTXTMonitoringEnabled    bool            `json:"dns_txt_monitoring_enabled"`
	DNSDKIMSelector            string          `json:"dns_dkim_selector,omitempty"`
	DNSMultiRecordEnabled      bool            `json:"dns_multi_record_enabled"`
	MailSMTPEnabled            bool            `json:"mail_smtp_enabled"`
	MailSMTPPort               int             `json:"mail_smtp_port,omitempty"`
	MailSMTPStartTLS           bool            `json:"mail_smtp_starttls"`
	MailSMTPOpenRelay          bool            `json:"mail_smtp_open_relay"`
	MailIMAPEnabled            bool            `json:"mail_imap_enabled"`
	MailIMAPPort               int             `json:"mail_imap_port,omitempty"`
	MailIMAPSSL                bool            `json:"mail_imap_ssl"`
	MailPOP3Enabled            bool            `json:"mail_pop3_enabled"`
	MailPOP3Port               int             `json:"mail_pop3_port,omitempty"`
	MailPOP3SSL                bool            `json:"mail_pop3_ssl"`
	MailCheckSPF               bool            `json:"mail_check_spf"`
	MailCheckDKIM              bool            `json:"mail_check_dkim"`
	MailDKIMSelectors          string          `json:"mail_dkim_selectors,omitempty"`
	MailCheckDMARC             bool            `json:"mail_check_dmarc"`
	MailCheckPTR               bool            `json:"mail_check_ptr"`
	MailCheckBlacklist         bool            `json:"mail_check_blacklist"`
	MailBlacklistServers       string          `json:"mail_blacklist_servers,omitempty"`
	ExpectedStatusCodes        []int           `json:"expected_status_codes,omitempty"`
	DNSExpectedIPs             []string        `json:"dns_expected_ips,omitempty"`
	AssignedAgentIDs           []int           `json:"assigned_agent_ids,omitempty"`
	HTTPHeaders                json.RawMessage `json:"http_headers,omitempty"`
	HTTPFormLoginData          json.RawMessage `json:"http_form_login_data,omitempty"`
	APIScenarioSteps           json.RawMessage `json:"api_scenario_steps,omitempty"`
	APIScenarioSecrets         json.RawMessage `json:"api_scenario_secrets,omitempty"`
	OIDCConfig                 json.RawMessage `json:"oidc_config,omitempty"`
	DNSRecordsConfig           json.RawMessage `json:"dns_records_config,omitempty"`
	CheckSourceCritical        json.RawMessage `json:"check_source_critical,omitempty"`
	ResponseAssertions         json.RawMessage `json:"response_assertions,omitempty"`
}

type CheckUpdateRequest struct {
	Name                       *string         `json:"name,omitempty"`
	Description                *string         `json:"description,omitempty"`
	IsActive                   *bool           `json:"is_active,omitempty"`
	Interval                   *int            `json:"interval,omitempty"`
	GracePeriod                *int            `json:"grace_period,omitempty"`
	Timeout                    *int            `json:"timeout,omitempty"`
	URL                        *string         `json:"url,omitempty"`
	Host                       *string         `json:"host,omitempty"`
	Port                       *int            `json:"port,omitempty"`
	PingCount                  *int            `json:"ping_count,omitempty"`
	SSLVerify                  *bool           `json:"ssl_verify,omitempty"`
	AlertAfterFailures         *int            `json:"alert_after_failures,omitempty"`
	GeoMonitoringEnabled       *bool           `json:"geo_monitoring_enabled,omitempty"`
	AssignedAgentID            *int            `json:"assigned_agent_id,omitempty"`
	EscalationPolicyID         *int            `json:"escalation_policy_id,omitempty"`
	TracerouteOnTimeout        *bool           `json:"traceroute_on_timeout,omitempty"`
	DNSServer                  *string         `json:"dns_server,omitempty"`
	DNSRecordType              *string         `json:"dns_record_type,omitempty"`
	DNSExpectedValue           *string         `json:"dns_expected_value,omitempty"`
	HTTPMethod                 *string         `json:"http_method,omitempty"`
	AuthMethod                 *string         `json:"auth_method,omitempty"`
	ScheduleType               *string         `json:"schedule_type,omitempty"`
	CronExpression             *string         `json:"cron_expression,omitempty"`
	CronTimezone               *string         `json:"cron_timezone,omitempty"`
	MailDomain                 *string         `json:"mail_domain,omitempty"`
	RunbookURL                 *string         `json:"runbook_url,omitempty"`
	DBType                     *string         `json:"db_type,omitempty"`
	DBHost                     *string         `json:"db_host,omitempty"`
	DBPort                     *int            `json:"db_port,omitempty"`
	DBName                     *string         `json:"db_name,omitempty"`
	DBUsername                 *string         `json:"db_username,omitempty"`
	DBPassword                 *string         `json:"db_password,omitempty"`
	DBSSLEnabled               *bool           `json:"db_ssl_enabled,omitempty"`
	DBQuery                    *string         `json:"db_query,omitempty"`
	DBExpectedResult           *string         `json:"db_expected_result,omitempty"`
	AuthUsername               *string         `json:"auth_username,omitempty"`
	AuthPassword               *string         `json:"auth_password,omitempty"`
	AuthBearerToken            *string         `json:"auth_bearer_token,omitempty"`
	HTTPBody                   *string         `json:"http_body,omitempty"`
	HTTPBodyType               *string         `json:"http_body_type,omitempty"`
	HTTPFollowRedirects        *bool           `json:"http_follow_redirects,omitempty"`
	ContentMatchEnabled        *bool           `json:"content_match_enabled,omitempty"`
	ContentMatchText           *string         `json:"content_match_text,omitempty"`
	ContentMatchType           *string         `json:"content_match_type,omitempty"`
	ContentMatchCaseSensitive  *bool           `json:"content_match_case_sensitive,omitempty"`
	HTTPFormLoginEnabled       *bool           `json:"http_form_login_enabled,omitempty"`
	HTTPFormLoginURL           *string         `json:"http_form_login_url,omitempty"`
	HTTPFormLoginSuccessText   *string         `json:"http_form_login_success_text,omitempty"`
	HTTPFormCheckAfterLoginURL *string         `json:"http_form_check_after_login_url,omitempty"`
	FTPUsername                *string         `json:"ftp_username,omitempty"`
	FTPPassword                *string         `json:"ftp_password,omitempty"`
	FTPProtocol                *string         `json:"ftp_protocol,omitempty"`
	FTPPath                    *string         `json:"ftp_path,omitempty"`
	FTPPassive                 *bool           `json:"ftp_passive,omitempty"`
	DNSSOAAlertOnChange        *bool           `json:"dns_soa_alert_on_change,omitempty"`
	DNSHijackAlertEnabled      *bool           `json:"dns_hijack_alert_enabled,omitempty"`
	DNSHijackAlertChannelIDs   *string         `json:"dns_hijack_alert_channel_ids,omitempty"`
	DNSTXTMonitoringEnabled    *bool           `json:"dns_txt_monitoring_enabled,omitempty"`
	DNSDKIMSelector            *string         `json:"dns_dkim_selector,omitempty"`
	DNSMultiRecordEnabled      *bool           `json:"dns_multi_record_enabled,omitempty"`
	MailSMTPEnabled            *bool           `json:"mail_smtp_enabled,omitempty"`
	MailSMTPPort               *int            `json:"mail_smtp_port,omitempty"`
	MailSMTPStartTLS           *bool           `json:"mail_smtp_starttls,omitempty"`
	MailSMTPOpenRelay          *bool           `json:"mail_smtp_open_relay,omitempty"`
	MailIMAPEnabled            *bool           `json:"mail_imap_enabled,omitempty"`
	MailIMAPPort               *int            `json:"mail_imap_port,omitempty"`
	MailIMAPSSL                *bool           `json:"mail_imap_ssl,omitempty"`
	MailPOP3Enabled            *bool           `json:"mail_pop3_enabled,omitempty"`
	MailPOP3Port               *int            `json:"mail_pop3_port,omitempty"`
	MailPOP3SSL                *bool           `json:"mail_pop3_ssl,omitempty"`
	MailCheckSPF               *bool           `json:"mail_check_spf,omitempty"`
	MailCheckDKIM              *bool           `json:"mail_check_dkim,omitempty"`
	MailDKIMSelectors          *string         `json:"mail_dkim_selectors,omitempty"`
	MailCheckDMARC             *bool           `json:"mail_check_dmarc,omitempty"`
	MailCheckPTR               *bool           `json:"mail_check_ptr,omitempty"`
	MailCheckBlacklist         *bool           `json:"mail_check_blacklist,omitempty"`
	MailBlacklistServers       *string         `json:"mail_blacklist_servers,omitempty"`
	ExpectedStatusCodes        *[]int          `json:"expected_status_codes,omitempty"`
	DNSExpectedIPs             *[]string       `json:"dns_expected_ips,omitempty"`
	AssignedAgentIDs           *[]int          `json:"assigned_agent_ids,omitempty"`
	HTTPHeaders                json.RawMessage `json:"http_headers,omitempty"`
	HTTPFormLoginData          json.RawMessage `json:"http_form_login_data,omitempty"`
	APIScenarioSteps           json.RawMessage `json:"api_scenario_steps,omitempty"`
	APIScenarioSecrets         json.RawMessage `json:"api_scenario_secrets,omitempty"`
	OIDCConfig                 json.RawMessage `json:"oidc_config,omitempty"`
	DNSRecordsConfig           json.RawMessage `json:"dns_records_config,omitempty"`
	CheckSourceCritical        json.RawMessage `json:"check_source_critical,omitempty"`
	ResponseAssertions         json.RawMessage `json:"response_assertions,omitempty"`
}

func (c *Client) CreateCheck(ctx context.Context, req CheckCreateRequest) (*Check, error) {
	var check Check
	err := c.DoRequest(ctx, http.MethodPost, "/api/checks", req, &check)
	if err != nil {
		return nil, err
	}
	return &check, nil
}

func (c *Client) GetCheck(ctx context.Context, checkID int) (*Check, error) {
	var check Check
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/checks/%d", checkID), nil, &check)
	if err != nil {
		return nil, err
	}
	return &check, nil
}

func (c *Client) UpdateCheck(ctx context.Context, checkID int, req CheckUpdateRequest) (*Check, error) {
	var check Check
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/checks/%d", checkID), req, &check)
	if err != nil {
		return nil, err
	}
	return &check, nil
}

func (c *Client) DeleteCheck(ctx context.Context, checkID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/checks/%d", checkID), nil, nil)
}
