package cron

import (
	"fmt"
	"gitcafe.com/ops/common/model"
	"gitcafe.com/ops/common/utils"
	"github.com/toolkits/file"
	"log"
	"os/exec"
	"path"
	"strings"
	"time"
)

func StartDesiredAgent(da *model.DesiredAgent) {
	if err := InsureDesiredAgentDirExists(da); err != nil {
		return
	}

	if err := InsureNewVersionFiles(da); err != nil {
		return
	}

	if err := StopAgentOf(da.Name); err != nil {
		return
	}

	if err := ControlStartIn(da.AgentVersionDir); err != nil {
		return
	}

	file.WriteString(path.Join(da.AgentDir, ".version"), da.Version)
}

func ControlStartIn(workdir string) error {
	out, err := ControlStatus(workdir)
	if err == nil && strings.Contains(out, "started") {
		return nil
	}

	_, err = ControlStart(workdir)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	out, err = ControlStatus(workdir)
	if err == nil && strings.Contains(out, "started") {
		return nil
	}

	return err
}

func InsureNewVersionFiles(da *model.DesiredAgent) error {
	if FilesReady(da) {
		return nil
	}

	downloadTarballCmd := exec.Command("wget", "-q", da.TarballUrl, "-O", da.TarballFilename)
	downloadTarballCmd.Dir = da.AgentVersionDir
	err := downloadTarballCmd.Run()
	if err != nil {
		log.Println("wget -q", da.TarballUrl, "-O", da.TarballFilename, "fail", err)
		return err
	}

	downloadMd5Cmd := exec.Command("wget", "-q", da.Md5Url, "-O", da.Md5Filename)
	downloadMd5Cmd.Dir = da.AgentVersionDir
	err = downloadMd5Cmd.Run()
	if err != nil {
		log.Println("wget -q", da.Md5Url, "-O", da.Md5Filename, "fail", err)
		return err
	}

	if utils.Md5sumCheck(da.AgentVersionDir, da.Md5Filename) {
		return nil
	} else {
		return fmt.Errorf("md5sum -c fail")
	}
}

func FilesReady(da *model.DesiredAgent) bool {
	if !file.IsExist(da.Md5Filepath) {
		return false
	}

	if !file.IsExist(da.TarballFilepath) {
		return false
	}

	if !file.IsExist(da.ControlFilepath) {
		return false
	}

	return utils.Md5sumCheck(da.AgentVersionDir, da.Md5Filename)
}

func InsureDesiredAgentDirExists(da *model.DesiredAgent) error {
	err := file.InsureDir(da.AgentDir)
	if err != nil {
		log.Println("insure dir", da.AgentDir, "fail", err)
		return err
	}

	err = file.InsureDir(da.AgentVersionDir)
	if err != nil {
		log.Println("insure dir", da.AgentVersionDir, "fail", err)
	}
	return err
}