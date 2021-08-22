package misc

type TriggerType string

const (
	TRIGGER_MANUAL  TriggerType = "MANUAL"
	TRIGGER_CRON    TriggerType = "CRON"
	TRIGGER_RETRY   TriggerType = "RETRY"
	TRIGGER_PARENT  TriggerType = "PARENT"
	TRIGGER_API     TriggerType = "API"
	TRIGGER_MISFIRE TriggerType = "MISFIRE"
)

type AddressType int32

const (
	ADDRESS_AUTO_REGISTER   AddressType = 0
	ADDRESS_MANUAL_REGISTER AddressType = 1
)

type BlockStrategyType string

const (
	BLOCK_STRATEGY_SERIAL_EXECUTION BlockStrategyType = "BlockStrategyType"
	BLOCK_STRATEGY_DISCARD_LATER    BlockStrategyType = "DISCARD_LATER"
	BLOCK_STRATEGY_COVER_EARLY      BlockStrategyType = "COVER_EARLY"
)

const (
	NONE     string = "NONE"
	CRON     string = "CRON"
	FIX_RATE string = "FIX_RATE"
)
