// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package digitalocean

import (
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func LookupDomain(ctx *pulumi.Context, args *LookupDomainArgs, opts ...pulumi.InvokeOption) (*LookupDomainResult, error) {
	var rv LookupDomainResult
	err := ctx.Invoke("digitalocean:index/getDomain:getDomain", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getDomain.
type LookupDomainArgs struct {
	// The name of the domain.
	Name string `pulumi:"name"`
}

// A collection of values returned by getDomain.
type LookupDomainResult struct {
	// The provider-assigned unique ID for this managed resource.
	Id   string `pulumi:"id"`
	Name string `pulumi:"name"`
	Ttl  int    `pulumi:"ttl"`
	// The uniform resource name of the domain
	// * `zoneFile`: The zone file of the domain.
	Urn      string `pulumi:"urn"`
	ZoneFile string `pulumi:"zoneFile"`
}