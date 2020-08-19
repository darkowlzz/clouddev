// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package digitalocean

import (
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

// Volume snapshots are saved instances of a block storage volume. Use this data
// source to retrieve the ID of a DigitalOcean volume snapshot for use in other
// resources.
func LookupVolumeSnapshot(ctx *pulumi.Context, args *LookupVolumeSnapshotArgs, opts ...pulumi.InvokeOption) (*LookupVolumeSnapshotResult, error) {
	var rv LookupVolumeSnapshotResult
	err := ctx.Invoke("digitalocean:index/getVolumeSnapshot:getVolumeSnapshot", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getVolumeSnapshot.
type LookupVolumeSnapshotArgs struct {
	// If more than one result is returned, use the most recent volume snapshot.
	MostRecent *bool `pulumi:"mostRecent"`
	// The name of the volume snapshot.
	Name *string `pulumi:"name"`
	// A regex string to apply to the volume snapshot list returned by DigitalOcean. This allows more advanced filtering not supported from the DigitalOcean API. This filtering is done locally on what DigitalOcean returns.
	NameRegex *string `pulumi:"nameRegex"`
	// A "slug" representing a DigitalOcean region (e.g. `nyc1`). If set, only volume snapshots available in the region will be returned.
	Region *string `pulumi:"region"`
}

// A collection of values returned by getVolumeSnapshot.
type LookupVolumeSnapshotResult struct {
	// The date and time the volume snapshot was created.
	CreatedAt string `pulumi:"createdAt"`
	// The provider-assigned unique ID for this managed resource.
	Id string `pulumi:"id"`
	// The minimum size in gigabytes required for a volume to be created based on this volume snapshot.
	MinDiskSize int     `pulumi:"minDiskSize"`
	MostRecent  *bool   `pulumi:"mostRecent"`
	Name        *string `pulumi:"name"`
	NameRegex   *string `pulumi:"nameRegex"`
	Region      *string `pulumi:"region"`
	// A list of DigitalOcean region "slugs" indicating where the volume snapshot is available.
	Regions []string `pulumi:"regions"`
	// The billable size of the volume snapshot in gigabytes.
	Size float64 `pulumi:"size"`
	// A list of the tags associated to the volume snapshot.
	Tags []string `pulumi:"tags"`
	// The ID of the volume from which the volume snapshot originated.
	VolumeId string `pulumi:"volumeId"`
}