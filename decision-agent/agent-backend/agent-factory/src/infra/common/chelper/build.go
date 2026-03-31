package chelper

import (
	"fmt"
	"log"
	"time"
)

type BuildInfo struct {
	BranchName      string `json:"branch_name"`
	BuildTime       string `json:"build_time"`
	CommitID        string `json:"commit_id"`
	GoCommonVersion string `json:"go_common_version"`
	Other           map[string]string
}

func PrintBuildInfo(delay time.Duration, buildInfo *BuildInfo) {
	time.Sleep(delay)

	log.Println("build info: ")
	log.Println("\tbranchName: ", buildInfo.BranchName)
	log.Println("\tbuildTime: ", buildInfo.BuildTime)
	log.Println("\tcommitID: ", buildInfo.CommitID)
	log.Println("\tgoCommonVersion: ", buildInfo.GoCommonVersion)

	for k, v := range buildInfo.Other {
		log.Println("\t", k, ": ", v)
	}
}

func BuildVersion(buildInfo *BuildInfo) (version string) {
	version = fmt.Sprintf("branchName: %s, buildTime: %s, commitID: %s", buildInfo.BranchName, buildInfo.BuildTime, buildInfo.CommitID)
	return
}
