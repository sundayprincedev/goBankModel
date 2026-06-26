package systems

import (
	"fmt"
	interfaces "gobanksystem/interfaces"
)

func AuditSystem(accounts ...interfaces.Auditable) {
	for _, acc := range accounts {
		fmt.Println(acc.Audit())
	}
}
