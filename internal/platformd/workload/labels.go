package workload

// SystemWorkloadLabels returns the labels used by system workloads
func SystemWorkloadLabels(name string) map[string]string {
	return map[string]string{
		"platform.chunks.cloud/workload-name": name,
		"platform.chunks.cloud/workload-type": "system",
	}
}
