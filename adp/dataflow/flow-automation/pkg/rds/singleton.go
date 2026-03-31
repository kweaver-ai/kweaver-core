package rds

import (
	"sync"
)

var (
	confDao     ConfDao
	confDaoOnce sync.Once

	aiModelDao     AiModelDao
	aiModelDaoOnce sync.Once

	alarmRuleDao     AlarmRuleDao
	alarmRuleDaoOnce sync.Once

	contentAdminDao     ContentAmdinDao
	contentAdminDaoOnce sync.Once

	agentDao     AgentDao
	agentDaoOnce sync.Once

	dagInstanceEventRepository     DagInstanceEventRepository
	dagInstanceEventRepositoryOnce sync.Once

	dagInstanceExtDataDao     DagInstanceExtDataDao
	dagInstanceExtDataDaoOnce sync.Once

	executorDao     ExecutorDao
	executorDaoOnce sync.Once

	taskCache     TaskCache
	taskCacheOnce sync.Once

	flowStorageDao     FlowStorageDao
	flowStorageDaoOnce sync.Once

	flowFileDao     FlowFileDao
	flowFileDaoOnce sync.Once

	flowFileDownloadJobDao     FlowFileDownloadJobDao
	flowFileDownloadJobDaoOnce sync.Once

	flowTaskResumeDao     FlowTaskResumeDao
	flowTaskResumeDaoOnce sync.Once
)

func SetConfDao(dao ConfDao) {
	confDao = dao
}

func SetAiModelDao(dao AiModelDao) {
	aiModelDao = dao
}

func SetAlarmRuleDao(dao AlarmRuleDao) {
	alarmRuleDao = dao
}

func SetContentAdminDao(dao ContentAmdinDao) {
	contentAdminDao = dao
}

func SetAgentDao(dao AgentDao) {
	agentDao = dao
}

func SetDagInstanceEventRepository(repo DagInstanceEventRepository) {
	dagInstanceEventRepository = repo
}

func SetDagInstanceExtDataDao(dao DagInstanceExtDataDao) {
	dagInstanceExtDataDao = dao
}

func SetExecutorDao(dao ExecutorDao) {
	executorDao = dao
}

func SetTaskCache(cache TaskCache) {
	taskCache = cache
}

func GetConfDao() ConfDao {
	return confDao
}

func GetAiModelDao() AiModelDao {
	return aiModelDao
}

func GetAlarmRuleDao() AlarmRuleDao {
	return alarmRuleDao
}

func GetContentAdminDao() ContentAmdinDao {
	return contentAdminDao
}

func GetAgentDao() AgentDao {
	return agentDao
}

func GetDagInstanceEventRepository() DagInstanceEventRepository {
	return dagInstanceEventRepository
}

func GetDagInstanceExtDataDao() DagInstanceExtDataDao {
	return dagInstanceExtDataDao
}

func GetExecutorDao() ExecutorDao {
	return executorDao
}

func GetTaskCache() TaskCache {
	return taskCache
}

func SetFlowStorageDao(dao FlowStorageDao) {
	flowStorageDao = dao
}

func SetFlowFileDao(dao FlowFileDao) {
	flowFileDao = dao
}

func SetFlowFileDownloadJobDao(dao FlowFileDownloadJobDao) {
	flowFileDownloadJobDao = dao
}

func SetFlowTaskResumeDao(dao FlowTaskResumeDao) {
	flowTaskResumeDao = dao
}

func GetFlowStorageDao() FlowStorageDao {
	return flowStorageDao
}

func GetFlowFileDao() FlowFileDao {
	return flowFileDao
}

func GetFlowFileDownloadJobDao() FlowFileDownloadJobDao {
	return flowFileDownloadJobDao
}

func GetFlowTaskResumeDao() FlowTaskResumeDao {
	return flowTaskResumeDao
}
