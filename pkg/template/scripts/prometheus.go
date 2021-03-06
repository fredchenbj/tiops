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

// PrometheusScript represent the data to generate Prometheus config
type PrometheusScript struct {
	IP        string
	Port      uint64
	DeployDir string
	DataDir   string
	NumaNode  string
}

// NewPrometheusScript returns a PrometheusScript with given arguments
func NewPrometheusScript(ip, deployDir, dataDir string) *PrometheusScript {
	return &PrometheusScript{
		IP:        ip,
		Port:      9090,
		DeployDir: deployDir,
		DataDir:   dataDir,
	}
}

// WithPort set Port field of PrometheusScript
func (c *PrometheusScript) WithPort(port uint64) *PrometheusScript {
	c.Port = port
	return c
}

// WithNumaNode set NumaNode field of PrometheusScript
func (c *PrometheusScript) WithNumaNode(numa string) *PrometheusScript {
	c.NumaNode = numa
	return c
}

// Config read ${localdata.EnvNameComponentInstallDir}/templates/scripts/run_Prometheus.sh.tpl as template
// and generate the config by ConfigWithTemplate
func (c *PrometheusScript) Config() (string, error) {
	fp := path.Join(os.Getenv(localdata.EnvNameComponentInstallDir), "templates", "scripts", "run_prometheus.sh.tpl")
	tpl, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Prometheus config content by tpl
func (c *PrometheusScript) ConfigWithTemplate(tpl string) (string, error) {
	tmpl, err := template.New("Prometheus").Parse(tpl)
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return "", err
	}

	return content.String(), nil
}
