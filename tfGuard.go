// A package that will parse a Terraform plan file
// and execute a set of functions against the underlying
// resources that can be used to gate or control how
// Terraform resources are being provisioned into your
// environment.
//
// This package can be utilized to build your own
// infrastructure scanning rules with the ability
// to leverage the capabilities of a full language
// instead of being limited by Domain Specific
// Languages (DSLs).
package tfGuard
