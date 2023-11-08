package provider

import (
	"context"
	"fmt"
	"strconv"
	"terraform-provider-airbrake/internal/airbrake"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ProjectResource{}
	_ resource.ResourceWithConfigure   = &ProjectResource{}
	_ resource.ResourceWithImportState = &ProjectResource{}
)

type ProjectResource struct {
	client *airbrake.Client
}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Id       types.String `tfsdk:"id"`
	ApiKey   types.String `tfsdk:"api_key"`
	Language types.String `tfsdk:"language"`
}

func (p ProjectResourceModel) convertToProject() airbrake.Project {
	id, _ := strconv.Atoi(p.Id.ValueString())
	return airbrake.Project{
		Id:       id,
		Name:     p.Name.ValueString(),
		APIKey:   p.ApiKey.ValueString(),
		Language: p.Language.ValueString(),
	}
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"language": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(LISTLANGUAGE...),
				},
			},
		},
	}
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var projectInput ProjectResourceModel
	diags := req.Plan.Get(ctx, &projectInput)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "create project resource", map[string]interface{}{
		"projectModel": projectInput,
	})

	// Create new project
	project, err := r.client.CreateProject(ctx, projectInput.convertToProject())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"Could not create project, unexpected error: "+err.Error(),
		)
		return
	}

	data := ProjectDataSourceModel{
		Name:     types.StringValue(project.Name),
		Id:       types.StringValue(strconv.Itoa(project.Id)),
		ApiKey:   types.StringValue(project.APIKey),
		Language: types.StringValue(project.Language),
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, _ := strconv.Atoi(state.Id.ValueString())
	project, err := r.client.GetProjectById(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Airbrake Project",
			"Could not read Airbrake Project "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	data := ProjectDataSourceModel{
		Name:     types.StringValue(project.Name),
		Id:       types.StringValue(strconv.Itoa(project.Id)),
		ApiKey:   types.StringValue(project.APIKey),
		Language: types.StringValue(project.Language),
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from project
	var projectInput ProjectResourceModel
	diags := req.Plan.Get(ctx, &projectInput)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing project
	err := r.client.UpdateProject(ctx, projectInput.convertToProject())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Airbrake project",
			"Could not update project, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &projectInput)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProject(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Airbrake project",
			"Could not delete project, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
