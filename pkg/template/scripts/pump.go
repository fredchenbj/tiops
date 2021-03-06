// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package scripts

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/pingcap-incubator/tiup/pkg/localdata"
)

// PumpScript represent the data to generate Pump config
type PumpScript struct {
	NodeID    string
	IP        string
	Port      uint64
	DeployDir string
	DataDir   string
	NumaNode  string
	CommitTs  int64
	Endpoints []*PDScript
}

// NewPumpScript returns a PumpScript with given arguments
func NewPumpScript(nodeID, ip, deployDir, dataDir string) *PumpScript {
	return &PumpScript{
		NodeID:    nodeID,
		IP:        ip,
		Port:      8250,
		DeployDir: deployDir,
		DataDir:   dataDir,
	}
}

// WithPort set Port field of PumpScript
func (c *PumpScript) WithPort(port uint64) *PumpScript {
	c.Port = port
	return c
}

// WithNumaNode set NumaNode field of PumpScript
func (c *PumpScript) WithNumaNode(numa string) *PumpScript {
	c.NumaNode = numa
	return c
}

// AppendEndpoints add new PumpScript to Endpoints field
func (c *PumpScript) AppendEndpoints(ends ...*PDScript) *PumpScript {
	c.Endpoints = append(c.Endpoints, ends...)
	return c
}

// Config read ${localdata.EnvNameComponentInstallDir}/templates/scripts/run_pump.sh.tpl as template
// and generate the config by ConfigWithTemplate
func (c *PumpScript) Config() (string, error) {
	fp := path.Join(os.Getenv(localdata.EnvNameComponentInstallDir), "templates", "scripts", "run_pump.sh.tpl")
	tpl, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Pump config content by tpl
func (c *PumpScript) ConfigWithTemplate(tpl string) (string, error) {
	tmpl, err := template.New("Pump").Parse(tpl)
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return "", err
	}

	return content.String(), nil
}
