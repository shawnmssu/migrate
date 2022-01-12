package client

import (
	"encoding/base64"
	"fmt"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/errors"
	"github.com/ucloud/ucloud-sdk-go/services/cube"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

type Provider struct {
	ChargeType     string `yaml:"chargeType"`
	ContainerCount string `yaml:"containerCount"`
	CpuPlatform    string `yaml:"cpuPlatform"`
	Region         string `yaml:"region"`
	SubnetId       string `yaml:"subnetId"`
	VpcId          string `yaml:"vpcId"`
	Zone           string `yaml:"zone"`
}
type Metadata struct {
	CreateTimestamp string   `yaml:"creationTimestamp"`
	Group           string   `yaml:"group"`
	Provider        Provider `yaml:"provider"`
	CubeId          string   `yaml:"cubeId"`
}
type status struct {
	Phase string `yaml:"phase"`
	PodIp string `yaml:"podIp"`
}
type CubePodInfo struct {
	Metadata Metadata `yaml:"metadata"`
	Status   status   `yaml:"status"`
}

type CubePodExtendInfo struct {
	CubePodInfo    CubePodInfo
	CubeExtendInfo cube.CubeExtendInfo
}

func (client *UCloudClient) UnBindEIPToCube(cubeId, eipId string) error {
	req := client.UNetConn.NewUnBindEIPRequest()
	req.EIPId = ucloud.String(eipId)
	req.ResourceId = ucloud.String(cubeId)
	req.ResourceType = ucloud.String("cube")
	if _, err := client.UNetConn.UnBindEIP(req); err != nil {
		return errors.NewAPIRequestFailedError(err)
	}
	return nil
}

func (client *UCloudClient) BindEIPToCube(cubeId, eipId string) error {
	req := client.UNetConn.NewBindEIPRequest()
	req.EIPId = ucloud.String(eipId)
	req.ResourceId = ucloud.String(cubeId)
	req.ResourceType = ucloud.String("cube")
	if _, err := client.UNetConn.BindEIP(req); err != nil {
		return errors.NewAPIRequestFailedError(err)
	}
	return nil
}

func (client *UCloudClient) GetCubeCubePodExtendInfoList(config *conf.CubeConfig) ([]cube.CubeExtendInfo, error) {
	if len(config.CubeIdList) > 0 {
		req := client.CubeConn.NewGetCubeExtendInfoRequest()
		req.CubeIds = ucloud.String(strings.Join(config.CubeIdList, ","))
		resp, err := client.CubeConn.GetCubeExtendInfo(req)
		if err != nil {
			return nil, errors.NewAPIRequestFailedError(err)
		}
		if len(resp.ExtendInfo) == 0 {
			return nil, errors.NewAPIRequestFailedError(fmt.Errorf("got empty cube extendInfo"))
		}
		return resp.ExtendInfo, nil
	} else if config.CubeIdFilter != nil {
		idList, err := client.filterCubeIdList(config.CubeIdFilter)
		if err != nil {
			return nil, errors.NewConfigValidateFailedError(err)
		}
		req := client.CubeConn.NewGetCubeExtendInfoRequest()
		req.CubeIds = ucloud.String(strings.Join(idList, ","))
		resp, err := client.CubeConn.GetCubeExtendInfo(req)
		if err != nil {
			return nil, errors.NewAPIRequestFailedError(err)
		}
		if len(resp.ExtendInfo) == 0 {
			return nil, errors.NewAPIRequestFailedError(fmt.Errorf("got empty cube extendInfo"))
		}

		var filteredCubeExtendInfo []cube.CubeExtendInfo
		if len(config.CubeIdFilter.NameRegex) > 0 {
			r := regexp.MustCompile(config.CubeIdFilter.NameRegex)
			for _, v := range resp.ExtendInfo {
				if r != nil && !r.MatchString(v.Name) {
					continue
				}
				filteredCubeExtendInfo = append(filteredCubeExtendInfo, v)
			}
		} else {
			filteredCubeExtendInfo = resp.ExtendInfo
		}
		if len(filteredCubeExtendInfo) == 0 {
			return nil, errors.NewAPIRequestFailedError(fmt.Errorf("got empty cube extendInfo"))
		}

		return filteredCubeExtendInfo, nil
	}

	return nil, errors.NewConfigValidateFailedError(fmt.Errorf("must set one of `cube_id_list` and `cube_id_filter`"))
}

func (client *UCloudClient) filterCubeIdList(filter *conf.CubeIdFilter) ([]string, error) {
	if filter == nil {
		return nil, fmt.Errorf("empty cube id filter")
	}
	req := client.CubeConn.NewListCubePodRequest()
	if len(filter.Zone) > 0 {
		req.Zone = ucloud.String(filter.Zone)
	}
	if len(filter.Group) > 0 {
		req.Group = ucloud.String(filter.Group)
	}
	if len(filter.DeploymentId) > 0 {
		req.DeploymentId = ucloud.String(filter.DeploymentId)
	}
	if len(filter.VPCId) > 0 {
		req.VPCId = ucloud.String(filter.VPCId)
	}
	if len(filter.SubnetId) > 0 {
		req.SubnetId = ucloud.String(filter.SubnetId)
	}

	var idList []string
	var offset int
	limit := 100
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := client.CubeConn.ListCubePod(req)
		if err != nil {
			return nil, fmt.Errorf("error on reading subnet list, %s", err)
		}

		if resp == nil || len(resp.Pods) < 1 {
			break
		}

		for _, pod := range resp.Pods {
			podByte, err := base64.StdEncoding.DecodeString(pod)
			if err != nil {
				return nil, err
			}
			var podInfo CubePodInfo
			if err := yaml.Unmarshal(podByte, &podInfo); err != nil {
				return nil, err
			}

			idList = append(idList, podInfo.Metadata.CubeId)
		}

		if len(resp.Pods) < limit {
			break
		}

		offset = offset + limit
	}

	if len(idList) == 0 {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("can not found any cube about cube_id_filter, %v", filter))
	}
	return idList, nil
}
