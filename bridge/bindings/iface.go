package bindings

type IAliceNetFactory interface {
	IAliceNetFactoryCaller
	IAliceNetFactoryTransactor
	IAliceNetFactoryFilterer
}

type IATokenBurner interface {
	IATokenBurnerCaller
	IATokenBurnerTransactor
	IATokenBurnerFilterer
}

type IAToken interface {
	IATokenCaller
	IATokenTransactor
	IATokenFilterer
}

type IATokenMinter interface {
	IATokenMinterCaller
	IATokenMinterTransactor
	IATokenMinterFilterer
}

type IBTokenErrorCodes interface {
	IBTokenErrorCodesCaller
	IBTokenErrorCodesTransactor
	IBTokenErrorCodesFilterer
}

type IBToken interface {
	IBTokenCaller
	IBTokenTransactor
	IBTokenFilterer
}

type IETHDKGErrorCodes interface {
	IETHDKGErrorCodesCaller
	IETHDKGErrorCodesTransactor
	IETHDKGErrorCodesFilterer
}

type IETHDKG interface {
	IETHDKGCaller
	IETHDKGTransactor
	IETHDKGFilterer
}

type IGovernanceErrorCodes interface {
	IGovernanceErrorCodesCaller
	IGovernanceErrorCodesTransactor
	IGovernanceErrorCodesFilterer
}

type IGovernance interface {
	IGovernanceCaller
	IGovernanceTransactor
	IGovernanceFilterer
}

type IPublicStaking interface {
	IPublicStakingCaller
	IPublicStakingTransactor
	IPublicStakingFilterer
}

type ISnapshotsErrorCodes interface {
	ISnapshotsErrorCodesCaller
	ISnapshotsErrorCodesTransactor
	ISnapshotsErrorCodesFilterer
}

type ISnapshots interface {
	ISnapshotsCaller
	ISnapshotsTransactor
	ISnapshotsFilterer
}

type IValidatorPoolErrorCodes interface {
	IValidatorPoolErrorCodesCaller
	IValidatorPoolErrorCodesTransactor
	IValidatorPoolErrorCodesFilterer
}

type IValidatorPool interface {
	IValidatorPoolCaller
	IValidatorPoolTransactor
	IValidatorPoolFilterer
}

type IValidatorStaking interface {
	IValidatorStakingCaller
	IValidatorStakingTransactor
	IValidatorStakingFilterer
}
