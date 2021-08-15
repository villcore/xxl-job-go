package api

type AdminBizApi interface {
	Callback(callbackParamSlice []HandleCallbackParam) ReturnT
	Registry(registryParam RegistryParam) ReturnT
	RegistryRemove(registryParam RegistryParam) ReturnT
}
