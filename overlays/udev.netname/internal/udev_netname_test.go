package udev_netname

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_udev_netnameOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.ImportFile("etc/warewulf/nodes.conf", "nodes.conf")
	env.ImportFile("var/lib/warewulf/overlays/udev.netname/rootfs/etc/udev/rules.d/70-persistent-net.rules.ww", "../rootfs/etc/udev/rules.d/70-persistent-net.rules.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "udev.netname",
			args: []string{"--render", "node1", "udev.netname", "etc/udev/rules.d/70-persistent-net.rules.ww"},
			log:  udev_netname,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := show.GetCommand()
			cmd.SetArgs(tt.args)
			stdout := bytes.NewBufferString("")
			stderr := bytes.NewBufferString("")
			logbuf := bytes.NewBufferString("")
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			wwlog.SetLogWriter(logbuf)
			err := cmd.Execute()
			assert.NoError(t, err)
			assert.Empty(t, stdout.String())
			assert.Empty(t, stderr.String())
			assert.Equal(t, tt.log, logbuf.String())
		})
	}
}

const udev_netname string = `backupFile: true
writeFile: true
Filename: etc/udev/rules.d/70-persistent-net.rules
# This file is autogenerated by warewulf

SUBSYSTEM=="net", ACTION=="add", ATTR{address}=="e6:92:39:49:7b:03", NAME="wwnet0"

SUBSYSTEM=="net", ACTION=="add", ATTR{address}=="9a:77:29:73:14:f1", NAME="wwnet1"
`
