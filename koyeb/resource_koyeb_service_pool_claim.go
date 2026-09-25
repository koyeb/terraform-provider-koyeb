package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// Claim wait budget and poll interval, mirroring the Python SDK's
// DEFAULT_CLAIM_WAIT_TIMEOUT and DEFAULT_CLAIM_POLL_INTERVAL; variables so
// tests can shorten them.
var (
	claimWaitTimeout  = 300 * time.Second
	claimWaitInterval = 2 * time.Second
)

// waitForClaimFulfilled polls GetClaim until the claim is FULFILLED,
// surfacing FAILED and RELEASED immediately: neither can still become
// FULFILLED. Cold claims provision a service on demand, so a pending
// claim is not yet usable.
func waitForClaimFulfilled(ctx context.Context, client *koyeb.APIClient, claimID string, timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		res, resp, err := client.PoolClaimsApi.GetClaim(ctx, claimID).Execute()
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("claim %s disappeared while waiting for fulfillment", claimID)
			}
			return err
		}
		claim := res.GetClaim()
		status := claim.GetStatus()
		switch status {
		case koyeb.POOLCLAIMSTATUS_FULFILLED:
			return nil
		case koyeb.POOLCLAIMSTATUS_FAILED:
			return fmt.Errorf("claim %s reached status FAILED", claimID)
		case koyeb.POOLCLAIMSTATUS_RELEASED:
			return fmt.Errorf("claim %s reached status RELEASED", claimID)
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("claim %s did not reach FULFILLED within %s (last status %s)", claimID, timeout, status)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for claim %s cancelled: %w", claimID, ctx.Err())
		case <-time.After(interval):
		}
	}
}

// NOTE: A claim hands out a service detached from the pool, so destroying
// the claim deletes that service via the public service API; the claim
// record itself is permanent server-side bookkeeping.
func resourceKoyebServicePoolClaim() *schema.Resource {
	return &schema.Resource{
		Description: "Claim a prewarmed service from a service pool. " +
			"The claim is idempotent: the same request ID re-claims the same service instead of claiming a second one. " +
			"Create waits until the claim is FULFILLED and the claimed service is ready, failing fast on FAILED or RELEASED. " +
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

	if err := waitForClaimFulfilled(ctx, client, d.Id(), claimWaitTimeout, claimWaitInterval); err != nil {
		return diag.Errorf("Error waiting for service pool claim to be fulfilled: %s", err)
	}

	// The server stamps FULFILLED when the service is created, not when it
	// is ready; mirror the Python SDK's wait_claim_ready so the exported
	// service_id is usable when apply reports success.
	if err := waitForResourceStatus(ctx, client.ServicesApi.GetService(ctx, res.GetServiceId()).Execute, "Service", serviceReadyStatuses, claimWaitTimeout, true, serviceTerminalStatuses...); err != nil {
		return diag.Errorf("Error waiting for claimed service to be ready: %s", err)
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
