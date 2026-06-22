// Package live contains end-to-end tests that exercise the Kurrent Cloud Pulumi
// provider against a real Kurrent Cloud organization using the Pulumi Automation
// API.
//
// These tests provision (and tear down) real, billable cloud resources. They are
// skipped automatically unless Kurrent Cloud credentials are present in the
// environment. See TESTING.md at the repository root for the full list of
// environment variables, where their values come from, and how to run the suite
// locally and in CI.
package live
