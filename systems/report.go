package systems

import (
	"fmt"
	interfaces "gobanksystem/interfaces"
)

func ReportSystem(accounts ...interfaces.Reportable) {
	for _, acc := range accounts {
		fmt.Println(acc.GenerateSummary())
	}
}
