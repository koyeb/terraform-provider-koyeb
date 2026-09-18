package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// NOTE: A claim hands out a service detached from the pool, so destroying
// the claim deletes that service via the public service API; the claim
// record itself is permanent server-side bookkeeping.
func resourceKoyebServicePoolClaim() *schema.Resource {
	return &schema.Resource{
		Description: "Claim a prewarmed service from a service pool. " +
			"The claim is idempotent: the same request ID re-claims the same service instead of claiming a second one. " +
			"Destroying the claim deletes the claimed service; " +
			"the claim record itself remains on the Koyeb side.",

		CreateContext: resourceKoyebServicePoolClaimCreate,
		ReadContext:   resourceKoyebServicePoolClaimRead,
		DeleteContext: resourceKoyebServicePoolClaimDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"pool": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The service pool name",
			},
			"request_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "The claim request ID, making the claim idempotent: " +
					"re-claiming with the same ID returns the same claim instead of claiming a second service",
			},
			"service_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the service the claim was fulfilled with",
			},
			"prewarmed": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "Whether the claimed service came prewarmed from the pool. " +
					"Only known when Terraform creates the claim; claims imported " +
					"or read later report false",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The claim status (PENDING, FULFILLED, FAILED or RELEASED)",
			},
			"pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the service pool the claim was made against",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the claim was created",
			},
			"fulfilled_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the claim was fulfilled",
			},
			"released_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the claim was released",
			},
			"pool_generation": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The pool generation the claim was made against",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization ID owning the claim",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workspace ID owning the claim",
			},
			"customer_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The customer ID owning the claim",
			},
		},
	}
}

func resourceKoyebServicePoolClaimCreate(
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	name := d.Get("pool").(string)

	pools, resp, err := client.ServicePoolsApi.ListServicePools(context.Background()).Name(name).Execute()
	if err != nil {
		return diag.Errorf("Error retrieving service pool: %s (%v %v)", err, resp, pools)
	}

	var poolID string
	for _, pool := range pools.GetServicePools() {
		if pool.GetName() == name {
			poolID = pool.GetId()
			break
		}
	}
	if poolID == "" {
		return diag.Errorf("No service pool found with name %s", name)
	}

	res, resp, err := client.PoolClaimsApi.Claim(context.Background()).Body(koyeb.PoolClaimRequest{
		PoolId:    toOpt(poolID),
		RequestId: toOpt(d.Get("request_id").(string)),
	}).Execute()
	if err != nil {
		return diag.Errorf("Error claiming service pool: %s (%v %v)", err, resp, res)
	}

	// The reply carries what only the claim call sees: the service handed
	// out and whether it was prewarmed.
	d.SetId(res.GetClaimId())
	if err := d.Set("service_id", res.GetServiceId()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("prewarmed", res.GetPrewarmed()); err != nil {
		return diag.FromErr(err)
	}

	return resourceKoyebServicePoolClaimRead(ctx, d, meta)
}

func resourceKoyebServicePoolClaimRead(
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.PoolClaimsApi.GetClaim(context.Background(), d.Id()).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error retrieving service pool claim: %s (%v %v)", err, resp, res)
	}

	claim := res.GetClaim()
	var createdAt, fulfilledAt, releasedAt string
	if claim.HasCreatedAt() {
		createdAt = claim.GetCreatedAt().UTC().String()
	}
	if claim.HasFulfilledAt() {
		fulfilledAt = claim.GetFulfilledAt().UTC().String()
	}
	if claim.HasReleasedAt() {
		releasedAt = claim.GetReleasedAt().UTC().String()
	}

	for key, value := range map[string]interface{}{
		"service_id":      claim.GetServiceId(),
		"status":          string(claim.GetStatus()),
		"pool_id":         claim.GetPoolId(),
		"created_at":      createdAt,
		"fulfilled_at":    fulfilledAt,
		"released_at":     releasedAt,
		"pool_generation": claim.GetPoolGeneration(),
		"organization_id": claim.GetOrganizationId(),
		"workspace_id":    claim.GetWorkspaceId(),
		"customer_id":     claim.GetCustomerId(),
	} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceKoyebServicePoolClaimDelete(
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	serviceID := d.Get("service_id").(string)

	// A claim that never fulfilled handed out no service: nothing to
	// delete, the record alone is inert.
	if serviceID == "" {
		d.SetId("")
		return nil
	}

	res, resp, err := client.ServicesApi.DeleteService(context.Background(), serviceID).Execute()
	if err != nil {
		return diag.Errorf("Error deleting claimed service: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
