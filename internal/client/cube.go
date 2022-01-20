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

func (client *UCloudClient) DeleteCube(cubeId string) error {
	req := client.CubeConn.NewDeleteCubePodRequest()
	req.CubeId = ucloud.String(cubeId)
	if _, err := client.CubeConn.DeleteCubePod(req); err != nil {
		return errors.NewAPIRequestFailedError(err)
	}
	return nil
}

func (client *UCloudClient) GetCubePodExtendInfoList(config *conf.CubeConfig) ([]cube.CubeExtendInfo, error) {
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
		cubePodInfoList, err := client.filterCubeInfoList(config.CubeIdFilter)
		if err != nil {
			return nil, errors.NewConfigValidateFailedError(err)
		}
		var idList []string
		for _, info := range cubePodInfoList {
			idList = append(idList, info.Metadata.CubeId)
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

func (client *UCloudClient) GetCubePodInfoList(config *conf.CubeConfig) ([]CubePodInfo, error) {
	var cubePodInfoList []CubePodInfo
	if len(config.CubeIdList) > 0 {
		for _, cubeId := range config.CubeIdList {
			reqPod := client.CubeConn.NewGetCubePodRequest()
			reqPod.CubeId = ucloud.String(cubeId)

			respPod, err := client.CubeConn.GetCubePod(reqPod)
			if err != nil {
				return nil, err
			}
			podByte, err := base64.StdEncoding.DecodeString(respPod.Pod)
			if err != nil {
				return nil, err
			}

			var podInfo CubePodInfo
			if err = yaml.Unmarshal(podByte, &podInfo); err != nil {
				return nil, err
			}
			cubePodInfoList = append(cubePodInfoList, podInfo)
		}
		if len(cubePodInfoList) == 0 {
			return nil, errors.NewConfigValidateFailedError(fmt.Errorf("can not found any cube about cube_id_list, %v", config.CubeIdList))
		}
		return cubePodInfoList, nil
	} else if config.CubeIdFilter != nil {
		infoList, err := client.filterCubeInfoList(config.CubeIdFilter)
		if err != nil {
			return nil, errors.NewConfigValidateFailedError(err)
		}

		if len(config.CubeIdFilter.NameRegex) > 0 {
			infoIdMap := make(map[string]CubePodInfo)
			for _, info := range infoList {
				infoIdMap[info.Metadata.CubeId] = info
			}

			var idList []string
			for _, info := range cubePodInfoList {
				idList = append(idList, info.Metadata.CubeId)
			}

			req := client.CubeConn.NewGetCubeExtendInfoRequest()
			req.CubeIds = ucloud.String(strings.Join(idList, ","))
			resp, err := client.CubeConn.GetCubeExtendInfo(req)
			if err != nil {
				return nil, errors.NewAPIRequestFailedError(err)
			}
			if len(resp.ExtendInfo) == 0 {
				return nil, errors.NewAPIRequestFailedError(fmt.Errorf("got empty cube extendInfo for filter name"))
			}

			r := regexp.MustCompile(config.CubeIdFilter.NameRegex)
			for _, v := range resp.ExtendInfo {
				if r != nil && !r.MatchString(v.Name) {
					continue
				}
				if info, ok := infoIdMap[v.CubeId]; ok {
					cubePodInfoList = append(cubePodInfoList, info)
				}
			}
			if len(cubePodInfoList) == 0 {
				return nil, errors.NewAPIRequestFailedError(fmt.Errorf("got empty cube extendInfo for filter name"))
			}
		} else {
			cubePodInfoList = infoList
		}

		return cubePodInfoList, nil
	}

	return nil, errors.NewConfigValidateFailedError(fmt.Errorf("must set one of `cube_id_list` and `cube_id_filter`"))
}

func (client *UCloudClient) filterCubeInfoList(filter *conf.CubeIdFilter) ([]CubePodInfo, error) {
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

	var infoList []CubePodInfo
	var offset int
	limit := 100
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := client.CubeConn.ListCubePod(req)
		if err != nil {
			return nil, fmt.Errorf("error on reading cube list, %s", err)
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
			if err = yaml.Unmarshal(podByte, &podInfo); err != nil {
				return nil, err
			}

			infoList = append(infoList, podInfo)
		}

		if len(resp.Pods) < limit {
			break
		}

		offset = offset + limit
	}

	if len(infoList) == 0 {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("can not found any cube about cube_id_filter, %v", filter))
	}
	return infoList, nil
}
