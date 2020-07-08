// Copyright (c) 2020 Zededa, Inc.
// SPDX-License-Identifier: Apache-2.0

package upgradeconverter

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/lf-edge/eve/pkg/pillar/types"
	log "github.com/sirupsen/logrus"
)

const (
	srcDirRoot = types.PersistDir + "/downloads"
	dstDir     = types.SealedDirName + "/verifier/verified"
)

// Move files from /persist/downloads/<objType>/verified/<UPPER CASE SHA>/<file>
// to /persist/vault/verifier/verified/<lower case sha>

func renameVerifiedFiles(ctxPtr *ucContext) error {
	log.Infof("renameVerifiedFiles()")
	if _, err := os.Stat(dstDir); err != nil {
		if err := os.MkdirAll(dstDir, 0700); err != nil {
			log.Error(err)
			return err
		}
	}
	renameFiles(srcDirRoot+"/"+types.AppImgObj+"/verified", dstDir,
		ctxPtr.noFlag)
	renameFiles(srcDirRoot+"/"+types.BaseOsObj+"/verified", dstDir,
		ctxPtr.noFlag)
	log.Infof("renameVerifiedFiles() DONE")
	return nil
}

// If noFlag is set we just log and no file system modifications.
func renameFiles(srcDir string, dstDir string, noFlag bool) {

	log.Infof("renameFiles(%s, %s, %t)", srcDir, dstDir, noFlag)
	if _, err := os.Stat(dstDir); err != nil {
		if err := os.MkdirAll(dstDir, 0700); err != nil {
			log.Error(err)
			return
		}
	}
	locations, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Errorf("renameFiles read directory '%s' failed: %v",
			srcDir, err)
		return
	}
	for _, location := range locations {
		sha := strings.ToLower(location.Name())
		dstFile := dstDir + "/" + sha
		// Find single file in srcDir
		innerDir := srcDir + "/" + location.Name()
		files, err := ioutil.ReadDir(innerDir)
		if err != nil {
			log.Errorf("renameFiles read directory '%s' failed: %v",
				innerDir, err)
			continue
		}
		if len(files) == 0 {
			log.Errorf("renameFiles read directory '%s' no file",
				innerDir)
			continue
		}
		if len(files) > 1 {
			log.Errorf("renameFiles read directory '%s' more than one file: %d",
				innerDir, len(files))
			continue
		}
		srcFile := innerDir + "/" + files[0].Name()
		if _, err := os.Stat(srcFile); err != nil {
			log.Errorf("renameFiles srcFile %s disappeared?: %s",
				srcFile, err)
			continue
		}
		if noFlag {
			log.Infof("renameFiles dryrun from %s to %s",
				srcFile, dstFile)
		} else {
			// Must copy due to fscrypt
			// XXX copy to tmpfile in new dir then rename
			if err := CopyFile(srcFile, dstFile); err != nil {
				log.Errorf("cp old to new failed: %s", err)
			} else {
				err := os.Remove(srcFile)
				if err != nil {
					log.Errorf("Remove old failed: %s", err)
				}
			}
		}
	}
}