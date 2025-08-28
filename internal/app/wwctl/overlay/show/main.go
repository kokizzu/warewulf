package show

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	overlayName := args[0]
	fileName := args[1]

	overlay_, err := overlay.Get(overlayName)
	if err != nil {
		return err
	}

	overlayFile := overlay_.File(fileName)

	if NodeName == "" {
		if !util.IsFile(overlayFile) {
			return fmt.Errorf("%s: %s not found", overlayName, overlayFile)
		}
		f, err := os.ReadFile(overlayFile)
		if err != nil {
			return err
		}

		wwlog.Output("%s", string(f))
	} else {
		if !util.IsFile(overlayFile) {
			possibleFile := fmt.Sprintf("%s.ww", overlayFile)
			if filepath.Ext(overlayFile) != ".ww" && util.IsFile(possibleFile) {
				wwlog.Debug("found overlay template: %s", possibleFile)
				overlayFile = possibleFile
			} else {
				return fmt.Errorf("%s: %s not found", overlayName, overlayFile)
			}
		}
		if filepath.Ext(overlayFile) != ".ww" {
			wwlog.Warn("%s lacks the '.ww' suffix, will not be rendered in an overlay", fileName)
		}
		nodeDB, err := node.New()
		if err != nil {
			return err
		}
		nodeConf, err := nodeDB.GetNode(NodeName)
		if err == node.ErrNotFound {
			hostName, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("could not get host name: %s", err)
			}
			nodeConf = node.NewNode(hostName)
			nodeConf.ClusterName = hostName
		}
		var allNodes []node.Node
		allNodes, err = nodeDB.FindAllNodes()
		if err != nil {
			return err
		}
		tstruct, err := overlay.InitStruct(overlayName, nodeConf, allNodes)
		if err != nil {
			return err
		}
		tstruct.BuildSource = overlayFile
		buffer, backupFile, writeFile, err := overlay.RenderTemplateFile(overlayFile, tstruct)
		if err != nil {
			return err
		}
		var outBuffer bytes.Buffer
		// search for magic file name comment
		bufferScanner := bufio.NewScanner(bytes.NewReader(buffer.Bytes()))
		bufferScanner.Split(overlay.ScanLines)
		reg := regexp.MustCompile(`.*{{\s*/\*\s*file\s*["'](.*)["']\s*\*/\s*}}.*`)
		foundFileComment := false
		destFileName := strings.TrimSuffix(fileName, ".ww")
		for bufferScanner.Scan() {
			line := bufferScanner.Text()
			filenameFromTemplate := reg.FindAllStringSubmatch(line, -1)
			if len(filenameFromTemplate) != 0 {
				wwlog.Debug("Found multifile comment, new filename %s", filenameFromTemplate[0][1])
				if foundFileComment {
					if !Quiet {
						wwlog.Info("backupFile: %v", backupFile)
						wwlog.Info("writeFile: %v", writeFile)
						wwlog.Info("Filename: %s", destFileName)
					}
					wwlog.Output("%s", outBuffer.String())
					outBuffer.Reset()
				}
				destFileName = filenameFromTemplate[0][1]
				foundFileComment = true
			} else {
				_, _ = outBuffer.WriteString(line)
			}
		}
		if !Quiet {
			wwlog.Info("backupFile: %v", backupFile)
			wwlog.Info("writeFile: %v", writeFile)
			wwlog.Info("Filename: %s", destFileName)
		}
		wwlog.Output("%s", outBuffer.String())
	}
	return nil
}
