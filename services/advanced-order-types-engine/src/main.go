package main

import (
	"fmt"
	"log"
)

func main() {
	log.Println("[INFO] Initializing Advanced Order Types Engine...")

	icebergSlicer := NewIcebergSlicer()
	trailingStopEngine := NewTrailingStopEngine()
	twapSlicer := NewTWAPSlicer()
	vwapSlicer := NewVWAPSlicer()
	ocoOrchestrator := NewOCOOrchestrator()
	triggerEngine := NewTriggerEngine()

	fmt.Printf("[INFO] Advanced Order Types Engine running. Registered subsystems: Iceberg (%p), TrailingStop (%p), TWAP (%p), VWAP (%p), OCO (%p), Trigger (%p)\n",
		icebergSlicer, trailingStopEngine, twapSlicer, vwapSlicer, ocoOrchestrator, triggerEngine)
}
