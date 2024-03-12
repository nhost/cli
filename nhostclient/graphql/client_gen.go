// Code generated by github.com/Yamashou/gqlgenc, DO NOT EDIT.

package graphql

import (
	"context"
	"net/http"

	"github.com/Yamashou/gqlgenc/clientv2"
)

type Client struct {
	Client *clientv2.Client
}

func NewClient(cli *http.Client, baseURL string, options *clientv2.Options, interceptors ...clientv2.RequestInterceptor) *Client {
	return &Client{Client: clientv2.NewClient(cli, baseURL, options, interceptors...)}
}

type QueryRoot struct {
	Announcements                  []*Announcements                "json:\"announcements\" graphql:\"announcements\""
	AnnouncementsByPk              *Announcements                  "json:\"announcements_by_pk,omitempty\" graphql:\"announcements_by_pk\""
	App                            *Apps                           "json:\"app,omitempty\" graphql:\"app\""
	AppSecrets                     []*ConfigEnvironmentVariable    "json:\"appSecrets\" graphql:\"appSecrets\""
	AppStateHistories              []*AppStateHistory              "json:\"appStateHistories\" graphql:\"appStateHistories\""
	AppStateHistory                *AppStateHistory                "json:\"appStateHistory,omitempty\" graphql:\"appStateHistory\""
	Apps                           []*Apps                         "json:\"apps\" graphql:\"apps\""
	AuthRefreshToken               *AuthRefreshTokens              "json:\"authRefreshToken,omitempty\" graphql:\"authRefreshToken\""
	AuthRefreshTokens              []*AuthRefreshTokens            "json:\"authRefreshTokens\" graphql:\"authRefreshTokens\""
	Backup                         *Backups                        "json:\"backup,omitempty\" graphql:\"backup\""
	Backups                        []*Backups                      "json:\"backups\" graphql:\"backups\""
	BillingGetNextInvoice          *InvoiceSummary                 "json:\"billingGetNextInvoice,omitempty\" graphql:\"billingGetNextInvoice\""
	CliToken                       *CliTokens                      "json:\"cliToken,omitempty\" graphql:\"cliToken\""
	CliTokens                      []*CliTokens                    "json:\"cliTokens\" graphql:\"cliTokens\""
	Config                         *ConfigConfig                   "json:\"config,omitempty\" graphql:\"config\""
	ConfigRawJSON                  string                          "json:\"configRawJSON\" graphql:\"configRawJSON\""
	Continents                     []*Continents                   "json:\"continents\" graphql:\"continents\""
	ContinentsByPk                 *Continents                     "json:\"continents_by_pk,omitempty\" graphql:\"continents_by_pk\""
	Countries                      []*Countries                    "json:\"countries\" graphql:\"countries\""
	CountriesByPk                  *Countries                      "json:\"countries_by_pk,omitempty\" graphql:\"countries_by_pk\""
	Deployment                     *Deployments                    "json:\"deployment,omitempty\" graphql:\"deployment\""
	DeploymentLog                  *DeploymentLogs                 "json:\"deploymentLog,omitempty\" graphql:\"deploymentLog\""
	DeploymentLogs                 []*DeploymentLogs               "json:\"deploymentLogs\" graphql:\"deploymentLogs\""
	Deployments                    []*Deployments                  "json:\"deployments\" graphql:\"deployments\""
	DNSLookupCname                 string                          "json:\"dnsLookupCNAME\" graphql:\"dnsLookupCNAME\""
	FeatureFlag                    *FeatureFlags                   "json:\"featureFlag,omitempty\" graphql:\"featureFlag\""
	FeatureFlags                   []*FeatureFlags                 "json:\"featureFlags\" graphql:\"featureFlags\""
	GetBackupPresignedURL          BackupPresignedURL              "json:\"getBackupPresignedURL\" graphql:\"getBackupPresignedURL\""
	GetCPUSecondsUsage             Metrics                         "json:\"getCPUSecondsUsage\" graphql:\"getCPUSecondsUsage\""
	GetEgressVolume                Metrics                         "json:\"getEgressVolume\" graphql:\"getEgressVolume\""
	GetFunctionsDuration           Metrics                         "json:\"getFunctionsDuration\" graphql:\"getFunctionsDuration\""
	GetFunctionsInvocations        Metrics                         "json:\"getFunctionsInvocations\" graphql:\"getFunctionsInvocations\""
	GetLogsVolume                  Metrics                         "json:\"getLogsVolume\" graphql:\"getLogsVolume\""
	GetPostgresVolumeCapacity      Metrics                         "json:\"getPostgresVolumeCapacity\" graphql:\"getPostgresVolumeCapacity\""
	GetPostgresVolumeUsage         Metrics                         "json:\"getPostgresVolumeUsage\" graphql:\"getPostgresVolumeUsage\""
	GetServiceLabelValues          []string                        "json:\"getServiceLabelValues\" graphql:\"getServiceLabelValues\""
	GetTotalRequests               Metrics                         "json:\"getTotalRequests\" graphql:\"getTotalRequests\""
	GithubAppInstallation          *GithubAppInstallations         "json:\"githubAppInstallation,omitempty\" graphql:\"githubAppInstallation\""
	GithubAppInstallations         []*GithubAppInstallations       "json:\"githubAppInstallations\" graphql:\"githubAppInstallations\""
	GithubRepositories             []*GithubRepositories           "json:\"githubRepositories\" graphql:\"githubRepositories\""
	GithubRepository               *GithubRepositories             "json:\"githubRepository,omitempty\" graphql:\"githubRepository\""
	Logs                           []*Log                          "json:\"logs\" graphql:\"logs\""
	PaymentMethod                  *PaymentMethods                 "json:\"paymentMethod,omitempty\" graphql:\"paymentMethod\""
	PaymentMethods                 []*PaymentMethods               "json:\"paymentMethods\" graphql:\"paymentMethods\""
	Plan                           *Plans                          "json:\"plan,omitempty\" graphql:\"plan\""
	Plans                          []*Plans                        "json:\"plans\" graphql:\"plans\""
	Regions                        []*Regions                      "json:\"regions\" graphql:\"regions\""
	RegionsByPk                    *Regions                        "json:\"regions_by_pk,omitempty\" graphql:\"regions_by_pk\""
	RunService                     *RunService                     "json:\"runService,omitempty\" graphql:\"runService\""
	RunServiceConfig               *ConfigRunServiceConfig         "json:\"runServiceConfig,omitempty\" graphql:\"runServiceConfig\""
	RunServiceConfigRawJSON        string                          "json:\"runServiceConfigRawJSON\" graphql:\"runServiceConfigRawJSON\""
	RunServiceConfigs              []*ConfigRunServiceConfigWithID "json:\"runServiceConfigs\" graphql:\"runServiceConfigs\""
	RunServices                    []*RunService                   "json:\"runServices\" graphql:\"runServices\""
	RunServicesAggregate           RunServiceAggregate             "json:\"runServicesAggregate\" graphql:\"runServicesAggregate\""
	SelectRegionsAllowedWorkspace  *RegionsAllowedWorkspace        "json:\"selectRegionsAllowedWorkspace,omitempty\" graphql:\"selectRegionsAllowedWorkspace\""
	SelectRegionsAllowedWorkspaces []*RegionsAllowedWorkspace      "json:\"selectRegionsAllowedWorkspaces\" graphql:\"selectRegionsAllowedWorkspaces\""
	SoftwareVersion                *SoftwareVersions               "json:\"softwareVersion,omitempty\" graphql:\"softwareVersion\""
	SoftwareVersions               []*SoftwareVersions             "json:\"softwareVersions\" graphql:\"softwareVersions\""
	SystemConfig                   *ConfigSystemConfig             "json:\"systemConfig,omitempty\" graphql:\"systemConfig\""
	User                           *Users                          "json:\"user,omitempty\" graphql:\"user\""
	Users                          []*Users                        "json:\"users\" graphql:\"users\""
	UsersUsage                     *UsersUsage                     "json:\"usersUsage,omitempty\" graphql:\"usersUsage\""
	UsersUsages                    []*UsersUsage                   "json:\"usersUsages\" graphql:\"usersUsages\""
	Workspace                      *Workspaces                     "json:\"workspace,omitempty\" graphql:\"workspace\""
	WorkspaceMember                *WorkspaceMembers               "json:\"workspaceMember,omitempty\" graphql:\"workspaceMember\""
	WorkspaceMemberInvite          *WorkspaceMemberInvites         "json:\"workspaceMemberInvite,omitempty\" graphql:\"workspaceMemberInvite\""
	WorkspaceMemberInvites         []*WorkspaceMemberInvites       "json:\"workspaceMemberInvites\" graphql:\"workspaceMemberInvites\""
	WorkspaceMembers               []*WorkspaceMembers             "json:\"workspaceMembers\" graphql:\"workspaceMembers\""
	Workspaces                     []*Workspaces                   "json:\"workspaces\" graphql:\"workspaces\""
}
type MutationRoot struct {
	BackupApplicationDatabase        BackupResult                              "json:\"backupApplicationDatabase\" graphql:\"backupApplicationDatabase\""
	DeleteApp                        *Apps                                     "json:\"deleteApp,omitempty\" graphql:\"deleteApp\""
	DeleteApps                       *AppsMutationResponse                     "json:\"deleteApps,omitempty\" graphql:\"deleteApps\""
	DeleteAuthRefreshToken           *AuthRefreshTokens                        "json:\"deleteAuthRefreshToken,omitempty\" graphql:\"deleteAuthRefreshToken\""
	DeleteAuthRefreshTokens          *AuthRefreshTokensMutationResponse        "json:\"deleteAuthRefreshTokens,omitempty\" graphql:\"deleteAuthRefreshTokens\""
	DeleteCliToken                   *CliTokens                                "json:\"deleteCliToken,omitempty\" graphql:\"deleteCliToken\""
	DeleteCliTokens                  *CliTokensMutationResponse                "json:\"deleteCliTokens,omitempty\" graphql:\"deleteCliTokens\""
	DeletePaymentMethod              *PaymentMethods                           "json:\"deletePaymentMethod,omitempty\" graphql:\"deletePaymentMethod\""
	DeletePaymentMethods             *PaymentMethodsMutationResponse           "json:\"deletePaymentMethods,omitempty\" graphql:\"deletePaymentMethods\""
	DeleteRunService                 *RunService                               "json:\"deleteRunService,omitempty\" graphql:\"deleteRunService\""
	DeleteRunServiceConfig           *ConfigRunServiceConfig                   "json:\"deleteRunServiceConfig,omitempty\" graphql:\"deleteRunServiceConfig\""
	DeleteRunServices                *RunServiceMutationResponse               "json:\"deleteRunServices,omitempty\" graphql:\"deleteRunServices\""
	DeleteSecret                     *ConfigEnvironmentVariable                "json:\"deleteSecret,omitempty\" graphql:\"deleteSecret\""
	DeleteUser                       *Users                                    "json:\"deleteUser,omitempty\" graphql:\"deleteUser\""
	DeleteUsers                      *UsersMutationResponse                    "json:\"deleteUsers,omitempty\" graphql:\"deleteUsers\""
	DeleteWorkspace                  *Workspaces                               "json:\"deleteWorkspace,omitempty\" graphql:\"deleteWorkspace\""
	DeleteWorkspaceMember            *WorkspaceMembers                         "json:\"deleteWorkspaceMember,omitempty\" graphql:\"deleteWorkspaceMember\""
	DeleteWorkspaceMemberInvite      *WorkspaceMemberInvites                   "json:\"deleteWorkspaceMemberInvite,omitempty\" graphql:\"deleteWorkspaceMemberInvite\""
	DeleteWorkspaceMemberInvites     *WorkspaceMemberInvitesMutationResponse   "json:\"deleteWorkspaceMemberInvites,omitempty\" graphql:\"deleteWorkspaceMemberInvites\""
	DeleteWorkspaceMembers           *WorkspaceMembersMutationResponse         "json:\"deleteWorkspaceMembers,omitempty\" graphql:\"deleteWorkspaceMembers\""
	DeleteWorkspaces                 *WorkspacesMutationResponse               "json:\"deleteWorkspaces,omitempty\" graphql:\"deleteWorkspaces\""
	InsertApp                        *Apps                                     "json:\"insertApp,omitempty\" graphql:\"insertApp\""
	InsertApps                       *AppsMutationResponse                     "json:\"insertApps,omitempty\" graphql:\"insertApps\""
	InsertDeployment                 *Deployments                              "json:\"insertDeployment,omitempty\" graphql:\"insertDeployment\""
	InsertDeployments                *DeploymentsMutationResponse              "json:\"insertDeployments,omitempty\" graphql:\"insertDeployments\""
	InsertFeatureFlag                *FeatureFlags                             "json:\"insertFeatureFlag,omitempty\" graphql:\"insertFeatureFlag\""
	InsertFeatureFlags               *FeatureFlagsMutationResponse             "json:\"insertFeatureFlags,omitempty\" graphql:\"insertFeatureFlags\""
	InsertGithubAppInstallation      *GithubAppInstallations                   "json:\"insertGithubAppInstallation,omitempty\" graphql:\"insertGithubAppInstallation\""
	InsertGithubAppInstallations     *GithubAppInstallationsMutationResponse   "json:\"insertGithubAppInstallations,omitempty\" graphql:\"insertGithubAppInstallations\""
	InsertPaymentMethod              *PaymentMethods                           "json:\"insertPaymentMethod,omitempty\" graphql:\"insertPaymentMethod\""
	InsertPaymentMethods             *PaymentMethodsMutationResponse           "json:\"insertPaymentMethods,omitempty\" graphql:\"insertPaymentMethods\""
	InsertRunService                 *RunService                               "json:\"insertRunService,omitempty\" graphql:\"insertRunService\""
	InsertRunServiceConfig           ConfigRunServiceConfig                    "json:\"insertRunServiceConfig\" graphql:\"insertRunServiceConfig\""
	InsertRunServices                *RunServiceMutationResponse               "json:\"insertRunServices,omitempty\" graphql:\"insertRunServices\""
	InsertSecret                     ConfigEnvironmentVariable                 "json:\"insertSecret\" graphql:\"insertSecret\""
	InsertWorkspace                  *Workspaces                               "json:\"insertWorkspace,omitempty\" graphql:\"insertWorkspace\""
	InsertWorkspaceMember            *WorkspaceMembers                         "json:\"insertWorkspaceMember,omitempty\" graphql:\"insertWorkspaceMember\""
	InsertWorkspaceMemberInvite      *WorkspaceMemberInvites                   "json:\"insertWorkspaceMemberInvite,omitempty\" graphql:\"insertWorkspaceMemberInvite\""
	InsertWorkspaceMemberInvites     *WorkspaceMemberInvitesMutationResponse   "json:\"insertWorkspaceMemberInvites,omitempty\" graphql:\"insertWorkspaceMemberInvites\""
	InsertWorkspaceMembers           *WorkspaceMembersMutationResponse         "json:\"insertWorkspaceMembers,omitempty\" graphql:\"insertWorkspaceMembers\""
	InsertWorkspaces                 *WorkspacesMutationResponse               "json:\"insertWorkspaces,omitempty\" graphql:\"insertWorkspaces\""
	ReplaceRunServiceConfig          ConfigRunServiceConfig                    "json:\"replaceRunServiceConfig\" graphql:\"replaceRunServiceConfig\""
	ResetPostgresPassword            bool                                      "json:\"resetPostgresPassword\" graphql:\"resetPostgresPassword\""
	RestoreApplicationDatabase       bool                                      "json:\"restoreApplicationDatabase\" graphql:\"restoreApplicationDatabase\""
	UpdateApp                        *Apps                                     "json:\"updateApp,omitempty\" graphql:\"updateApp\""
	UpdateApps                       *AppsMutationResponse                     "json:\"updateApps,omitempty\" graphql:\"updateApps\""
	UpdateConfig                     ConfigConfig                              "json:\"updateConfig\" graphql:\"updateConfig\""
	UpdateGithubAppInstallation      *GithubAppInstallations                   "json:\"updateGithubAppInstallation,omitempty\" graphql:\"updateGithubAppInstallation\""
	UpdateGithubAppInstallations     *GithubAppInstallationsMutationResponse   "json:\"updateGithubAppInstallations,omitempty\" graphql:\"updateGithubAppInstallations\""
	UpdatePaymentMethod              *PaymentMethods                           "json:\"updatePaymentMethod,omitempty\" graphql:\"updatePaymentMethod\""
	UpdatePaymentMethods             *PaymentMethodsMutationResponse           "json:\"updatePaymentMethods,omitempty\" graphql:\"updatePaymentMethods\""
	UpdateRunServiceConfig           ConfigRunServiceConfig                    "json:\"updateRunServiceConfig\" graphql:\"updateRunServiceConfig\""
	UpdateSecret                     ConfigEnvironmentVariable                 "json:\"updateSecret\" graphql:\"updateSecret\""
	UpdateUser                       *Users                                    "json:\"updateUser,omitempty\" graphql:\"updateUser\""
	UpdateUsers                      *UsersMutationResponse                    "json:\"updateUsers,omitempty\" graphql:\"updateUsers\""
	UpdateWorkspace                  *Workspaces                               "json:\"updateWorkspace,omitempty\" graphql:\"updateWorkspace\""
	UpdateWorkspaceMember            *WorkspaceMembers                         "json:\"updateWorkspaceMember,omitempty\" graphql:\"updateWorkspaceMember\""
	UpdateWorkspaceMemberInvite      *WorkspaceMemberInvites                   "json:\"updateWorkspaceMemberInvite,omitempty\" graphql:\"updateWorkspaceMemberInvite\""
	UpdateWorkspaceMemberInvites     *WorkspaceMemberInvitesMutationResponse   "json:\"updateWorkspaceMemberInvites,omitempty\" graphql:\"updateWorkspaceMemberInvites\""
	UpdateWorkspaceMembers           *WorkspaceMembersMutationResponse         "json:\"updateWorkspaceMembers,omitempty\" graphql:\"updateWorkspaceMembers\""
	UpdateWorkspaces                 *WorkspacesMutationResponse               "json:\"updateWorkspaces,omitempty\" graphql:\"updateWorkspaces\""
	UpdateAppsMany                   []*AppsMutationResponse                   "json:\"update_apps_many,omitempty\" graphql:\"update_apps_many\""
	UpdateGithubAppInstallationsMany []*GithubAppInstallationsMutationResponse "json:\"update_githubAppInstallations_many,omitempty\" graphql:\"update_githubAppInstallations_many\""
	UpdatePaymentMethodsMany         []*PaymentMethodsMutationResponse         "json:\"update_paymentMethods_many,omitempty\" graphql:\"update_paymentMethods_many\""
	UpdateUsersMany                  []*UsersMutationResponse                  "json:\"update_users_many,omitempty\" graphql:\"update_users_many\""
	UpdateWorkspaceMemberInvitesMany []*WorkspaceMemberInvitesMutationResponse "json:\"update_workspaceMemberInvites_many,omitempty\" graphql:\"update_workspaceMemberInvites_many\""
	UpdateWorkspaceMembersMany       []*WorkspaceMembersMutationResponse       "json:\"update_workspaceMembers_many,omitempty\" graphql:\"update_workspaceMembers_many\""
	UpdateWorkspacesMany             []*WorkspacesMutationResponse             "json:\"update_workspaces_many,omitempty\" graphql:\"update_workspaces_many\""
}
type GetWorkspacesApps_Workspaces_Apps_Region struct {
	AwsName string "json:\"awsName\" graphql:\"awsName\""
}

func (t *GetWorkspacesApps_Workspaces_Apps_Region) GetAwsName() string {
	if t == nil {
		t = &GetWorkspacesApps_Workspaces_Apps_Region{}
	}
	return t.AwsName
}

type GetWorkspacesApps_Workspaces_Apps struct {
	ID        string                                   "json:\"id\" graphql:\"id\""
	Name      string                                   "json:\"name\" graphql:\"name\""
	Subdomain string                                   "json:\"subdomain\" graphql:\"subdomain\""
	Region    GetWorkspacesApps_Workspaces_Apps_Region "json:\"region\" graphql:\"region\""
}

func (t *GetWorkspacesApps_Workspaces_Apps) GetID() string {
	if t == nil {
		t = &GetWorkspacesApps_Workspaces_Apps{}
	}
	return t.ID
}
func (t *GetWorkspacesApps_Workspaces_Apps) GetName() string {
	if t == nil {
		t = &GetWorkspacesApps_Workspaces_Apps{}
	}
	return t.Name
}
func (t *GetWorkspacesApps_Workspaces_Apps) GetSubdomain() string {
	if t == nil {
		t = &GetWorkspacesApps_Workspaces_Apps{}
	}
	return t.Subdomain
}
func (t *GetWorkspacesApps_Workspaces_Apps) GetRegion() *GetWorkspacesApps_Workspaces_Apps_Region {
	if t == nil {
		t = &GetWorkspacesApps_Workspaces_Apps{}
	}
	return &t.Region
}

type GetWorkspacesApps_Workspaces struct {
	Name string                               "json:\"name\" graphql:\"name\""
	Apps []*GetWorkspacesApps_Workspaces_Apps "json:\"apps\" graphql:\"apps\""
}

func (t *GetWorkspacesApps_Workspaces) GetName() string {
	if t == nil {
		t = &GetWorkspacesApps_Workspaces{}
	}
	return t.Name
}
func (t *GetWorkspacesApps_Workspaces) GetApps() []*GetWorkspacesApps_Workspaces_Apps {
	if t == nil {
		t = &GetWorkspacesApps_Workspaces{}
	}
	return t.Apps
}

type GetHasuraAdminSecret_App_Config_Hasura struct {
	Version     *string "json:\"version,omitempty\" graphql:\"version\""
	AdminSecret string  "json:\"adminSecret\" graphql:\"adminSecret\""
}

func (t *GetHasuraAdminSecret_App_Config_Hasura) GetVersion() *string {
	if t == nil {
		t = &GetHasuraAdminSecret_App_Config_Hasura{}
	}
	return t.Version
}
func (t *GetHasuraAdminSecret_App_Config_Hasura) GetAdminSecret() string {
	if t == nil {
		t = &GetHasuraAdminSecret_App_Config_Hasura{}
	}
	return t.AdminSecret
}

type GetHasuraAdminSecret_App_Config struct {
	Hasura GetHasuraAdminSecret_App_Config_Hasura "json:\"hasura\" graphql:\"hasura\""
}

func (t *GetHasuraAdminSecret_App_Config) GetHasura() *GetHasuraAdminSecret_App_Config_Hasura {
	if t == nil {
		t = &GetHasuraAdminSecret_App_Config{}
	}
	return &t.Hasura
}

type GetHasuraAdminSecret_App struct {
	Config *GetHasuraAdminSecret_App_Config "json:\"config,omitempty\" graphql:\"config\""
}

func (t *GetHasuraAdminSecret_App) GetConfig() *GetHasuraAdminSecret_App_Config {
	if t == nil {
		t = &GetHasuraAdminSecret_App{}
	}
	return t.Config
}

type DeleteRefreshToken_DeleteAuthRefreshTokens_Returning struct {
	Typename *string "json:\"__typename,omitempty\" graphql:\"__typename\""
}

func (t *DeleteRefreshToken_DeleteAuthRefreshTokens_Returning) GetTypename() *string {
	if t == nil {
		t = &DeleteRefreshToken_DeleteAuthRefreshTokens_Returning{}
	}
	return t.Typename
}

type DeleteRefreshToken_DeleteAuthRefreshTokens struct {
	AffectedRows int64                                                   "json:\"affected_rows\" graphql:\"affected_rows\""
	Returning    []*DeleteRefreshToken_DeleteAuthRefreshTokens_Returning "json:\"returning\" graphql:\"returning\""
}

func (t *DeleteRefreshToken_DeleteAuthRefreshTokens) GetAffectedRows() int64 {
	if t == nil {
		t = &DeleteRefreshToken_DeleteAuthRefreshTokens{}
	}
	return t.AffectedRows
}
func (t *DeleteRefreshToken_DeleteAuthRefreshTokens) GetReturning() []*DeleteRefreshToken_DeleteAuthRefreshTokens_Returning {
	if t == nil {
		t = &DeleteRefreshToken_DeleteAuthRefreshTokens{}
	}
	return t.Returning
}

type GetSecrets_AppSecrets struct {
	Name  string "json:\"name\" graphql:\"name\""
	Value string "json:\"value\" graphql:\"value\""
}

func (t *GetSecrets_AppSecrets) GetName() string {
	if t == nil {
		t = &GetSecrets_AppSecrets{}
	}
	return t.Name
}
func (t *GetSecrets_AppSecrets) GetValue() string {
	if t == nil {
		t = &GetSecrets_AppSecrets{}
	}
	return t.Value
}

type CreateSecret_InsertSecret struct {
	Name  string "json:\"name\" graphql:\"name\""
	Value string "json:\"value\" graphql:\"value\""
}

func (t *CreateSecret_InsertSecret) GetName() string {
	if t == nil {
		t = &CreateSecret_InsertSecret{}
	}
	return t.Name
}
func (t *CreateSecret_InsertSecret) GetValue() string {
	if t == nil {
		t = &CreateSecret_InsertSecret{}
	}
	return t.Value
}

type DeleteSecret_DeleteSecret struct {
	Name string "json:\"name\" graphql:\"name\""
}

func (t *DeleteSecret_DeleteSecret) GetName() string {
	if t == nil {
		t = &DeleteSecret_DeleteSecret{}
	}
	return t.Name
}

type UpdateSecret_UpdateSecret struct {
	Name  string "json:\"name\" graphql:\"name\""
	Value string "json:\"value\" graphql:\"value\""
}

func (t *UpdateSecret_UpdateSecret) GetName() string {
	if t == nil {
		t = &UpdateSecret_UpdateSecret{}
	}
	return t.Name
}
func (t *UpdateSecret_UpdateSecret) GetValue() string {
	if t == nil {
		t = &UpdateSecret_UpdateSecret{}
	}
	return t.Value
}

type UpdateRunServiceConfig_UpdateRunServiceConfig struct {
	Typename *string "json:\"__typename,omitempty\" graphql:\"__typename\""
}

func (t *UpdateRunServiceConfig_UpdateRunServiceConfig) GetTypename() *string {
	if t == nil {
		t = &UpdateRunServiceConfig_UpdateRunServiceConfig{}
	}
	return t.Typename
}

type ReplaceRunServiceConfig_ReplaceRunServiceConfig struct {
	Typename *string "json:\"__typename,omitempty\" graphql:\"__typename\""
}

func (t *ReplaceRunServiceConfig_ReplaceRunServiceConfig) GetTypename() *string {
	if t == nil {
		t = &ReplaceRunServiceConfig_ReplaceRunServiceConfig{}
	}
	return t.Typename
}

type GetRunServiceInfo_RunService struct {
	AppID string "json:\"appID\" graphql:\"appID\""
}

func (t *GetRunServiceInfo_RunService) GetAppID() string {
	if t == nil {
		t = &GetRunServiceInfo_RunService{}
	}
	return t.AppID
}

type GetWorkspacesApps struct {
	Workspaces []*GetWorkspacesApps_Workspaces "json:\"workspaces\" graphql:\"workspaces\""
}

func (t *GetWorkspacesApps) GetWorkspaces() []*GetWorkspacesApps_Workspaces {
	if t == nil {
		t = &GetWorkspacesApps{}
	}
	return t.Workspaces
}

type GetHasuraAdminSecret struct {
	App *GetHasuraAdminSecret_App "json:\"app,omitempty\" graphql:\"app\""
}

func (t *GetHasuraAdminSecret) GetApp() *GetHasuraAdminSecret_App {
	if t == nil {
		t = &GetHasuraAdminSecret{}
	}
	return t.App
}

type GetConfigRawJSON struct {
	ConfigRawJSON string "json:\"configRawJSON\" graphql:\"configRawJSON\""
}

func (t *GetConfigRawJSON) GetConfigRawJSON() string {
	if t == nil {
		t = &GetConfigRawJSON{}
	}
	return t.ConfigRawJSON
}

type DeleteRefreshToken struct {
	DeleteAuthRefreshTokens *DeleteRefreshToken_DeleteAuthRefreshTokens "json:\"deleteAuthRefreshTokens,omitempty\" graphql:\"deleteAuthRefreshTokens\""
}

func (t *DeleteRefreshToken) GetDeleteAuthRefreshTokens() *DeleteRefreshToken_DeleteAuthRefreshTokens {
	if t == nil {
		t = &DeleteRefreshToken{}
	}
	return t.DeleteAuthRefreshTokens
}

type GetSecrets struct {
	AppSecrets []*GetSecrets_AppSecrets "json:\"appSecrets\" graphql:\"appSecrets\""
}

func (t *GetSecrets) GetAppSecrets() []*GetSecrets_AppSecrets {
	if t == nil {
		t = &GetSecrets{}
	}
	return t.AppSecrets
}

type CreateSecret struct {
	InsertSecret CreateSecret_InsertSecret "json:\"insertSecret\" graphql:\"insertSecret\""
}

func (t *CreateSecret) GetInsertSecret() *CreateSecret_InsertSecret {
	if t == nil {
		t = &CreateSecret{}
	}
	return &t.InsertSecret
}

type DeleteSecret struct {
	DeleteSecret *DeleteSecret_DeleteSecret "json:\"deleteSecret,omitempty\" graphql:\"deleteSecret\""
}

func (t *DeleteSecret) GetDeleteSecret() *DeleteSecret_DeleteSecret {
	if t == nil {
		t = &DeleteSecret{}
	}
	return t.DeleteSecret
}

type UpdateSecret struct {
	UpdateSecret UpdateSecret_UpdateSecret "json:\"updateSecret\" graphql:\"updateSecret\""
}

func (t *UpdateSecret) GetUpdateSecret() *UpdateSecret_UpdateSecret {
	if t == nil {
		t = &UpdateSecret{}
	}
	return &t.UpdateSecret
}

type UpdateRunServiceConfig struct {
	UpdateRunServiceConfig UpdateRunServiceConfig_UpdateRunServiceConfig "json:\"updateRunServiceConfig\" graphql:\"updateRunServiceConfig\""
}

func (t *UpdateRunServiceConfig) GetUpdateRunServiceConfig() *UpdateRunServiceConfig_UpdateRunServiceConfig {
	if t == nil {
		t = &UpdateRunServiceConfig{}
	}
	return &t.UpdateRunServiceConfig
}

type ReplaceRunServiceConfig struct {
	ReplaceRunServiceConfig ReplaceRunServiceConfig_ReplaceRunServiceConfig "json:\"replaceRunServiceConfig\" graphql:\"replaceRunServiceConfig\""
}

func (t *ReplaceRunServiceConfig) GetReplaceRunServiceConfig() *ReplaceRunServiceConfig_ReplaceRunServiceConfig {
	if t == nil {
		t = &ReplaceRunServiceConfig{}
	}
	return &t.ReplaceRunServiceConfig
}

type GetRunServiceInfo struct {
	RunService *GetRunServiceInfo_RunService "json:\"runService,omitempty\" graphql:\"runService\""
}

func (t *GetRunServiceInfo) GetRunService() *GetRunServiceInfo_RunService {
	if t == nil {
		t = &GetRunServiceInfo{}
	}
	return t.RunService
}

type GetRunServiceConfigRawJSON struct {
	RunServiceConfigRawJSON string "json:\"runServiceConfigRawJSON\" graphql:\"runServiceConfigRawJSON\""
}

func (t *GetRunServiceConfigRawJSON) GetRunServiceConfigRawJSON() string {
	if t == nil {
		t = &GetRunServiceConfigRawJSON{}
	}
	return t.RunServiceConfigRawJSON
}

const GetWorkspacesAppsDocument = `query GetWorkspacesApps {
	workspaces {
		name
		apps {
			id
			name
			subdomain
			region {
				awsName
			}
		}
	}
}
`

func (c *Client) GetWorkspacesApps(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*GetWorkspacesApps, error) {
	vars := map[string]interface{}{}

	var res GetWorkspacesApps
	if err := c.Client.Post(ctx, "GetWorkspacesApps", GetWorkspacesAppsDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const GetHasuraAdminSecretDocument = `query GetHasuraAdminSecret ($appID: uuid!) {
	app(id: $appID) {
		config(resolve: true) {
			hasura {
				version
				adminSecret
			}
		}
	}
}
`

func (c *Client) GetHasuraAdminSecret(ctx context.Context, appID string, interceptors ...clientv2.RequestInterceptor) (*GetHasuraAdminSecret, error) {
	vars := map[string]interface{}{
		"appID": appID,
	}

	var res GetHasuraAdminSecret
	if err := c.Client.Post(ctx, "GetHasuraAdminSecret", GetHasuraAdminSecretDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const GetConfigRawJSONDocument = `query GetConfigRawJSON ($appID: uuid!) {
	configRawJSON(appID: $appID, resolve: false)
}
`

func (c *Client) GetConfigRawJSON(ctx context.Context, appID string, interceptors ...clientv2.RequestInterceptor) (*GetConfigRawJSON, error) {
	vars := map[string]interface{}{
		"appID": appID,
	}

	var res GetConfigRawJSON
	if err := c.Client.Post(ctx, "GetConfigRawJSON", GetConfigRawJSONDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const DeleteRefreshTokenDocument = `mutation DeleteRefreshToken ($where: authRefreshTokens_bool_exp!) {
	deleteAuthRefreshTokens(where: $where) {
		affected_rows
		returning {
			__typename
		}
	}
}
`

func (c *Client) DeleteRefreshToken(ctx context.Context, where AuthRefreshTokensBoolExp, interceptors ...clientv2.RequestInterceptor) (*DeleteRefreshToken, error) {
	vars := map[string]interface{}{
		"where": where,
	}

	var res DeleteRefreshToken
	if err := c.Client.Post(ctx, "DeleteRefreshToken", DeleteRefreshTokenDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const GetSecretsDocument = `query GetSecrets ($appID: uuid!) {
	appSecrets(appID: $appID) {
		name
		value
	}
}
`

func (c *Client) GetSecrets(ctx context.Context, appID string, interceptors ...clientv2.RequestInterceptor) (*GetSecrets, error) {
	vars := map[string]interface{}{
		"appID": appID,
	}

	var res GetSecrets
	if err := c.Client.Post(ctx, "GetSecrets", GetSecretsDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const CreateSecretDocument = `mutation CreateSecret ($appID: uuid!, $name: String!, $value: String!) {
	insertSecret(appID: $appID, secret: {name:$name,value:$value}) {
		name
		value
	}
}
`

func (c *Client) CreateSecret(ctx context.Context, appID string, name string, value string, interceptors ...clientv2.RequestInterceptor) (*CreateSecret, error) {
	vars := map[string]interface{}{
		"appID": appID,
		"name":  name,
		"value": value,
	}

	var res CreateSecret
	if err := c.Client.Post(ctx, "CreateSecret", CreateSecretDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const DeleteSecretDocument = `mutation DeleteSecret ($appID: uuid!, $name: String!) {
	deleteSecret(appID: $appID, key: $name) {
		name
	}
}
`

func (c *Client) DeleteSecret(ctx context.Context, appID string, name string, interceptors ...clientv2.RequestInterceptor) (*DeleteSecret, error) {
	vars := map[string]interface{}{
		"appID": appID,
		"name":  name,
	}

	var res DeleteSecret
	if err := c.Client.Post(ctx, "DeleteSecret", DeleteSecretDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const UpdateSecretDocument = `mutation UpdateSecret ($appID: uuid!, $name: String!, $value: String!) {
	updateSecret(appID: $appID, secret: {name:$name,value:$value}) {
		name
		value
	}
}
`

func (c *Client) UpdateSecret(ctx context.Context, appID string, name string, value string, interceptors ...clientv2.RequestInterceptor) (*UpdateSecret, error) {
	vars := map[string]interface{}{
		"appID": appID,
		"name":  name,
		"value": value,
	}

	var res UpdateSecret
	if err := c.Client.Post(ctx, "UpdateSecret", UpdateSecretDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const UpdateRunServiceConfigDocument = `mutation UpdateRunServiceConfig ($appID: uuid!, $serviceID: uuid!, $config: ConfigRunServiceConfigUpdateInput!) {
	updateRunServiceConfig(appID: $appID, serviceID: $serviceID, config: $config) {
		__typename
	}
}
`

func (c *Client) UpdateRunServiceConfig(ctx context.Context, appID string, serviceID string, config ConfigRunServiceConfigUpdateInput, interceptors ...clientv2.RequestInterceptor) (*UpdateRunServiceConfig, error) {
	vars := map[string]interface{}{
		"appID":     appID,
		"serviceID": serviceID,
		"config":    config,
	}

	var res UpdateRunServiceConfig
	if err := c.Client.Post(ctx, "UpdateRunServiceConfig", UpdateRunServiceConfigDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const ReplaceRunServiceConfigDocument = `mutation ReplaceRunServiceConfig ($appID: uuid!, $serviceID: uuid!, $config: ConfigRunServiceConfigInsertInput!) {
	replaceRunServiceConfig(appID: $appID, serviceID: $serviceID, config: $config) {
		__typename
	}
}
`

func (c *Client) ReplaceRunServiceConfig(ctx context.Context, appID string, serviceID string, config ConfigRunServiceConfigInsertInput, interceptors ...clientv2.RequestInterceptor) (*ReplaceRunServiceConfig, error) {
	vars := map[string]interface{}{
		"appID":     appID,
		"serviceID": serviceID,
		"config":    config,
	}

	var res ReplaceRunServiceConfig
	if err := c.Client.Post(ctx, "ReplaceRunServiceConfig", ReplaceRunServiceConfigDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const GetRunServiceInfoDocument = `query GetRunServiceInfo ($serviceID: uuid!) {
	runService(id: $serviceID) {
		appID
	}
}
`

func (c *Client) GetRunServiceInfo(ctx context.Context, serviceID string, interceptors ...clientv2.RequestInterceptor) (*GetRunServiceInfo, error) {
	vars := map[string]interface{}{
		"serviceID": serviceID,
	}

	var res GetRunServiceInfo
	if err := c.Client.Post(ctx, "GetRunServiceInfo", GetRunServiceInfoDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

const GetRunServiceConfigRawJSONDocument = `query GetRunServiceConfigRawJSON ($appID: uuid!, $serviceID: uuid!, $resolve: Boolean!) {
	runServiceConfigRawJSON(appID: $appID, serviceID: $serviceID, resolve: $resolve)
}
`

func (c *Client) GetRunServiceConfigRawJSON(ctx context.Context, appID string, serviceID string, resolve bool, interceptors ...clientv2.RequestInterceptor) (*GetRunServiceConfigRawJSON, error) {
	vars := map[string]interface{}{
		"appID":     appID,
		"serviceID": serviceID,
		"resolve":   resolve,
	}

	var res GetRunServiceConfigRawJSON
	if err := c.Client.Post(ctx, "GetRunServiceConfigRawJSON", GetRunServiceConfigRawJSONDocument, &res, vars, interceptors...); err != nil {
		if c.Client.ParseDataWhenErrors {
			return &res, err
		}

		return nil, err
	}

	return &res, nil
}

var DocumentOperationNames = map[string]string{
	GetWorkspacesAppsDocument:          "GetWorkspacesApps",
	GetHasuraAdminSecretDocument:       "GetHasuraAdminSecret",
	GetConfigRawJSONDocument:           "GetConfigRawJSON",
	DeleteRefreshTokenDocument:         "DeleteRefreshToken",
	GetSecretsDocument:                 "GetSecrets",
	CreateSecretDocument:               "CreateSecret",
	DeleteSecretDocument:               "DeleteSecret",
	UpdateSecretDocument:               "UpdateSecret",
	UpdateRunServiceConfigDocument:     "UpdateRunServiceConfig",
	ReplaceRunServiceConfigDocument:    "ReplaceRunServiceConfig",
	GetRunServiceInfoDocument:          "GetRunServiceInfo",
	GetRunServiceConfigRawJSONDocument: "GetRunServiceConfigRawJSON",
}
