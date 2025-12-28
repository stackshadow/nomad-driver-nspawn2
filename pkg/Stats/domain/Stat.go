package domain

type Stat struct {
	CPUTicks        float64
	CPUUsagePercent float64

	RamUsaged       uint64
	RamUsagePercent float64
}
