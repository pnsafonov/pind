package ali

// application layer interface

type ApiLogic interface {
	AddPinMapping(vmName string, numa int) error
	GetPinMapping() map[string]int
	Remove(vmName string)
	Clean()
}
