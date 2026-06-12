package ai

// SupportedVendors lists the AI vendors that are supported by the application.
var SupportedVendors = []string{"deepseek"}

// SupportedModels returns the list of supported models for a given vendor.
func SupportedModels(vendor string) []string {
	return []string{"deepseek-chat"}
}
