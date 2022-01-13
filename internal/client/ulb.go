package client

import (
	"fmt"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/errors"
	"github.com/ucloud/migrate/internal/log"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"sync"
)

func (client UCloudClient) UnBindBackendToUlB(ulbId, backendId string) error {
	req := client.ULBConn.NewReleaseBackendRequest()
	req.ULBId = ucloud.String(ulbId)
	req.BackendId = ucloud.String(backendId)

	if _, err := client.ULBConn.ReleaseBackend(req); err != nil {
		return err
	}
	return nil
}
func (client UCloudClient) BindUHostToUlBVServer(ulbId, vServerId, uHostId string, port int) (string, error) {
	req := client.ULBConn.NewAllocateBackendRequest()
	req.ULBId = ucloud.String(ulbId)
	req.VServerId = ucloud.String(vServerId)
	req.ResourceType = ucloud.String("UHost")
	req.ResourceId = ucloud.String(uHostId)
	req.Port = ucloud.Int(port)

	resp, err := client.ULBConn.AllocateBackend(req)
	if err != nil {
		return "", fmt.Errorf("error in create lb attachment, %s", err)
	}

	return resp.BackendId, nil
}

//type ULBVServerInfo struct {
//	ULBId       string
//	VServerId   string
//	CubeBackendMap map[string]string
//	Policies    []ulb.ULBPolicySet
//}

type ULBCubeInfo struct {
	ULBId            string
	VServerCubeInfos []VServerCubeInfo
}
type VServerCubeInfo struct {
	VServerId      string
	CubeBackendMap map[string][]ulb.ULBBackendSet

	Policies []ulb.ULBPolicySet
}

//func (VServerCubeInfo VServerCubeInfo) Get

func (client UCloudClient) GetULBCubeInfoList(config *conf.ULBConfig) ([]ULBCubeInfo, error) {
	var ulbCubeInfos []ULBCubeInfo
	if len(config.UlBIdList) > 0 {
		var wg sync.WaitGroup
		var lock sync.Mutex
		for i := 0; i < len(config.UlBIdList); i++ {
			wg.Add(1)
			go func(ulbId string) {
				defer wg.Done()
				req := client.ULBConn.NewDescribeULBRequest()
				req.ULBId = ucloud.String(ulbId)
				resp, err := client.ULBConn.DescribeULB(req)
				if err != nil {
					log.Logger.Sugar().Warnf("[DescribeULB] got error, %s", err)
					return
				}

				for _, ulbInfo := range resp.DataSet {
					var ulbCubeInfo ULBCubeInfo
					for _, vServerInfo := range ulbInfo.VServerSet {
						var info VServerCubeInfo
						for _, backendInfo := range vServerInfo.BackendSet {
							if backendInfo.ResourceType == "CUBE" {
								if info.CubeBackendMap == nil {
									info.CubeBackendMap = make(map[string][]ulb.ULBBackendSet)
								}
								info.CubeBackendMap[backendInfo.ResourceId] = append(
									info.CubeBackendMap[backendInfo.ResourceId], backendInfo)
							}
						}

						if len(info.CubeBackendMap) > 0 {
							info.VServerId = vServerInfo.VServerId
							info.Policies = vServerInfo.PolicySet
							ulbCubeInfo.VServerCubeInfos = append(ulbCubeInfo.VServerCubeInfos, info)
						}
					}
					if len(ulbCubeInfo.VServerCubeInfos) > 0 {
						lock.Lock()
						ulbCubeInfos = append(ulbCubeInfos, ulbCubeInfo)
						lock.Unlock()
					}
				}
			}(config.UlBIdList[i])
		}
		wg.Wait()
	} else if config.ULBIdFilter != nil {
		ulbInfos, err := client.filterULBList(config.ULBIdFilter)
		if err != nil {
			return nil, err
		}
		for _, ulbInfo := range ulbInfos {
			var ulbCubeInfo ULBCubeInfo
			for _, vServerInfo := range ulbInfo.VServerSet {
				var info VServerCubeInfo
				for _, backendInfo := range vServerInfo.BackendSet {
					if backendInfo.ResourceType == "CUBE" {
						if info.CubeBackendMap == nil {
							info.CubeBackendMap = make(map[string][]ulb.ULBBackendSet)
						}
						info.CubeBackendMap[backendInfo.ResourceId] = append(
							info.CubeBackendMap[backendInfo.ResourceId], backendInfo)
					}
				}

				if len(info.CubeBackendMap) > 0 {
					info.VServerId = vServerInfo.VServerId
					info.Policies = vServerInfo.PolicySet
					ulbCubeInfo.VServerCubeInfos = append(ulbCubeInfo.VServerCubeInfos, info)
				}
			}
			if len(ulbCubeInfo.VServerCubeInfos) > 0 {
				ulbCubeInfos = append(ulbCubeInfos, ulbCubeInfo)
			}
		}
	} else {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("must set one of `ulb_id_list` and `ulb_id_filter`"))
	}

	return ulbCubeInfos, nil

}

func (client *UCloudClient) filterULBList(filter *conf.ULBIdFilter) ([]ulb.ULBSet, error) {
	if filter == nil {
		return nil, fmt.Errorf("empty ulb id filter")
	}
	req := client.ULBConn.NewDescribeULBRequest()

	if len(filter.BusinessId) > 0 {
		req.BusinessId = ucloud.String(filter.BusinessId)
	}
	if len(filter.VPCId) > 0 {
		req.VPCId = ucloud.String(filter.VPCId)
	}
	if len(filter.SubnetId) > 0 {
		req.SubnetId = ucloud.String(filter.SubnetId)
	}

	var ulbInfos []ulb.ULBSet
	var offset int
	limit := 100
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := client.ULBConn.DescribeULB(req)
		if err != nil {
			return nil, fmt.Errorf("error on reading ulb list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		ulbInfos = append(ulbInfos, resp.DataSet...)
		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	if len(ulbInfos) == 0 {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("can not found any ulb about ulb_id_filter, %v", filter))
	}
	return ulbInfos, nil
}

func (client *UCloudClient) describeVServerById(lbId, vServerId string) (*ulb.ULBVServerSet, error) {
	conn := client.ULBConn
	req := conn.NewDescribeVServerRequest()
	req.ULBId = ucloud.String(lbId)
	req.VServerId = ucloud.String(vServerId)

	resp, err := conn.DescribeVServer(req)
	if err != nil {
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, errors.NewNotFoundError(fmt.Errorf("not found VServer %s", vServerId))
	}

	return &resp.DataSet[0], nil
}

func (client *UCloudClient) DescribeBackendById(lbId, vServerId, backendId string) (*ulb.ULBBackendSet, error) {
	vServerSet, err := client.describeVServerById(lbId, vServerId)

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(vServerSet.BackendSet); i++ {
		backend := vServerSet.BackendSet[i]
		if backend.BackendId == backendId {
			return &backend, nil
		}
	}

	return nil, errors.NewNotFoundError(fmt.Errorf("not found %s", backendId))
}
