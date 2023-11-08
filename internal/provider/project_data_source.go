package provider

import (
	"context"
	"fmt"
	"strconv"
	"terraform-provider-airbrake/internal/airbrake"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &ProjectDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectDataSource{}
)

type ProjectDataSource struct {
	client *airbrake.Client
}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

type ProjectDataSourceModel struct {
	Name     types.String `tfsdk:"name"`
	Id       types.String `tfsdk:"id"`
	ApiKey   types.String `tfsdk:"api_key"`
	Language types.String `tfsdk:"language"`
}

func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"api_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"language": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Airbrake connexion",
			"Expected configured Airbrake connexion. Please report this issue to the provider developers.",
		)
		return
	}

	var data ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := d.client.GetProject(ctx, data.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch project", err.Error())
	}

	data = ProjectDataSourceModel{
		Name:     types.StringValue(project.Name),
		Id:       types.StringValue(strconv.Itoa(project.Id)),
		ApiKey:   types.StringValue(project.APIKey),
		Language: types.StringValue(project.Language),
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*airbrake.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}
