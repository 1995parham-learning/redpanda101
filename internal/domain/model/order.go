package model

type Channel int

const (
	Unknown            Channel = 0
	Web                Channel = 1
	WebFast            Channel = 2
	WebConvert         Channel = 3
	WebSimple          Channel = 4
	Android            Channel = 11
	AndroidFast        Channel = 12
	AndroidConvert     Channel = 13
	AndroidSimple      Channel = 14
	iOS                Channel = 21
	iOSConvert         Channel = 23
	API                Channel = 31
	APIInternal        Channel = 32
	APIConvert         Channel = 33
	APIInternalOld     Channel = 34
	WebV1              Channel = 41
	WebV2              Channel = 42
	SystemMargin       Channel = 51
	SystemBlock        Channel = 52
	SystemABCLiquidate Channel = 53
	SystemLiquidator   Channel = 54
	Locket             Channel = 61
)

type Order struct {
	ID          uint64
	SrcCurrency uint64
	DstCurrency uint64
	Description string
	Channel     Channel
}
