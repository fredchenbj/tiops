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

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pingcap-incubator/tiops/pkg/meta"
	operator "github.com/pingcap-incubator/tiops/pkg/operation"
	"github.com/pingcap-incubator/tiops/pkg/task"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

type upgradeOptions struct {
	version string
	options operator.Options
}

func newUpgradeCmd() *cobra.Command {
	opt := upgradeOptions{}
	cmd := &cobra.Command{
		Use:   "upgrade <cluster-name>",
		Short: "Upgrade a specified TiDB cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Help()
			}
			return upgrade(args[0], opt)
		},
	}

	cmd.Flags().StringVarP(&opt.version, "target-version", "t", "", "Specify the target version")
	cmd.Flags().BoolVar(&opt.options.Force, "force", false, "Force upgrade won't transfer leader")

	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("target-version")

	return cmd
}

func versionCompare(curVersion, newVersion string) error {

	switch semver.Compare(curVersion, newVersion) {
	case -1:
		return nil
	case 1:
		if repository.Version(newVersion).IsNightly() {
			return nil
		}
		return errors.New(fmt.Sprintf("unsupport upgrade from %s to %s", curVersion, newVersion))
	default:
		return errors.New("unkown error")
	}
}

func upgrade(name string, opt upgradeOptions) error {
	metadata, err := meta.ClusterMetadata(name)
	if err != nil {
		return err
	}

	var (
		downloadCompTasks []task.Task // tasks which are used to download components
		copyCompTasks     []task.Task // tasks which are used to copy components to remote host

		uniqueComps = map[componentInfo]struct{}{}
	)

	if err := versionCompare(metadata.Version, opt.version); err != nil {
		return err
	}

	for _, comp := range metadata.Topology.ComponentsByStartOrder() {
		for _, inst := range comp.Instances() {
			version := getComponentVersion(inst.ComponentName(), opt.version)
			if version == "" {
				return errors.Errorf("unsupported component: %v", inst.ComponentName())
			}
			compInfo := componentInfo{
				component: inst.ComponentName(),
				version:   version,
			}

			// Download component from repository
			if _, found := uniqueComps[compInfo]; !found {
				uniqueComps[compInfo] = struct{}{}
				t := task.NewBuilder().
					Download(inst.ComponentName(), version).
					Build()
				downloadCompTasks = append(downloadCompTasks, t)
			}

			deployDir := inst.DeployDir()
			if !strings.HasPrefix(deployDir, "/") {
				deployDir = filepath.Join("/home/"+metadata.User+"/deploy", deployDir)
			}
			// Deploy component
			t := task.NewBuilder().
				BackupComponent(inst.ComponentName(), metadata.Version, inst.GetHost(), deployDir).
				CopyComponent(inst.ComponentName(), version, inst.GetHost(), deployDir).
				Build()
			copyCompTasks = append(copyCompTasks, t)
		}
	}

	t := task.NewBuilder().
		SSHKeySet(
			meta.ClusterPath(name, "ssh", "id_rsa"),
			meta.ClusterPath(name, "ssh", "id_rsa.pub")).
		ClusterSSH(metadata.Topology, metadata.User).
		Parallel(downloadCompTasks...).
		Parallel(copyCompTasks...).
		ClusterOperate(metadata.Topology, operator.UpgradeOperation, opt.options).
		Build()

	return t.Execute(task.NewContext())
}
