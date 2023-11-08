package provider

import (
	"context"
	"os"
	"strings"
	"terraform-provider-airbrake/internal/airbrake"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &AirbrakeProvider{}

const (
	DEFAULT_AIRBRAKE_URL = "https://api.airbrake.io/api/v4/"
)

var (
	envAirbrakeBaseUrl  = "AIRBRAKE_BASE_URL"
	envAirbrakeEmail    = "AIRBRAKE_EMAIL"
	envAirbrakePassword = "AIRBRAKE_PASSWORD"
	envAirbrakeApiKey   = "AIRBRAKE_API_KEY"
)

type AirbrakeProvider struct {
	version string
}

type AirbrakeProviderModel struct {
	Email    types.String `tfsdk:"email"`
	Password types.String `tfsdk:"password"`
	ApiKey   types.String `tfsdk:"api_key"`
	BaseUrl  types.String `tfsdk:"base_url"`
}

func (p *AirbrakeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "airbrake"
	resp.Version = p.version
}

func (p *AirbrakeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "Email used to connect to Airbrake.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password used to connect to Airbrake.",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Api key used to connect to Airbrake.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "The airbrake base API url with api version",
			},
		},
	}
}

func (p *AirbrakeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config AirbrakeProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Airbake API key",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	if config.Email.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"),
			"Unknown Airbrake user",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Airbake password",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_PASSWORD environment variable.",
		)
	}

	if config.BaseUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown Airbake base url",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	apikey := os.Getenv(envAirbrakeBaseUrl)
	email := os.Getenv(envAirbrakeEmail)
	password := os.Getenv(envAirbrakePassword)
	baseurl := os.Getenv(envAirbrakeApiKey)
	if len(baseurl) == 0 {
		baseurl = DEFAULT_AIRBRAKE_URL
	}

	if !config.ApiKey.IsNull() {
		apikey = config.ApiKey.ValueString()
	}

	if !config.Email.IsNull() {
		email = config.Email.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.BaseUrl.IsNull() {
		baseurl = config.BaseUrl.ValueString()
	}

	if apikey == "" && (email == "" || password == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Airbrake credentials",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API host. "+
				"Set the host value in the configuration or use the HASHICUPS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if !strings.HasSuffix(baseurl, "/") {
		baseurl = baseurl + "/"
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := airbrake.NewClient(baseurl, email, password, apikey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Connect to airbake API Client",
			"An unexpected error occurred when creating the airbrake API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Airbrake Client Error: "+err.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AirbrakeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
	}
}

func (p *AirbrakeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AirbrakeProvider{
			version: version,
		}
	}
}
