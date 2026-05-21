// Package assets builds public Steam Store and Library static asset URLs from AppIDs,
// requests Storefront screenshot/movie/background URLs, and provides small helpers
// for verifying, reading, and downloading those resources.
//
// URL construction helpers are local-only. Verification and download helpers perform
// explicit HTTP requests only when called.
package assets
