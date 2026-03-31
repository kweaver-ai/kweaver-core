package service

// var uniqueIDRepo portDriven.UlidRepo

// genRedisDlmUniqueValue generates a unique value for RedisDlm
//// 【注意】使用数据库来保障唯一性，此处使用场景不用过于考虑性能问题
// func genRedisDlmUniqueValue() (value string, err error) {
//	if uniqueIDRepo == nil {
//		uniqueIDRepo = dbaulid.dbaulid.NewUlidRepo()
//	}
//
//	ctx := context.Background()
//	value, err = uniqueIDRepo.GenUniqID(ctx, consts.UniqueIDFlagRedisDlm)
//
//	return
//}
//
//// delRedisDlmUniqueValue delete the value from db after RedisDlm unlock
//func delRedisDlmUniqueValue(value string) (err error) {
//	if uniqueIDRepo == nil {
//		uniqueIDRepo = dbaulid.NewUlidRepo()
//	}
//
//	ctx := context.Background()
//	err = uniqueIDRepo.DelUniqID(ctx, consts.UniqueIDFlagRedisDlm, value)
//
//	return
//}
//
//func getDlmConf() (dlmConf *adapterdrivencmpt.adapterdrivencmpt) {
//	expiry := 20 * time.Second
//	delay := 20 * time.Millisecond
//	maxTries := 500
//
//	opts := []redsync.Option{
//		redsync.WithExpiry(expiry),
//		redsync.WithRetryDelay(delay),
//		redsync.WithTries(maxTries), //【注意】：重试次数用完后，会返回错误（加锁失败）
//		redsync.WithGenValueFunc(genRedisDlmUniqueValue),
//	}
//
//	dlmConf = &adapterdrivencmpt.RedisDlmCmpConf{
//		Options:          opts,
//		WatchDogInterval: expiry / 2,
//		DeleteValueFunc:  delRedisDlmUniqueValue,
//	}
//
//	return
//}
