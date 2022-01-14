package client

import (
	"fmt"
	"github.com/ucloud/migrate/internal/errors"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
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

type VServerCubeInfo struct {
	VServerId      string
	CubeBackendMap map[string][]ulb.ULBBackendSet

	Policies []ulb.ULBPolicySet
}

func (client UCloudClient) GetULBVServerInfoListAboutCube(ulbId string) ([]VServerCubeInfo, error) {
	req := client.ULBConn.NewDescribeULBRequest()
	req.ULBId = ucloud.String(ulbId)
	resp, err := client.ULBConn.DescribeULB(req)
	if err != nil {
		return nil, fmt.Errorf("[DescribeULB] got error by ULBId(%s), %s", ulbId, err)
	}

	if len(resp.DataSet) == 0 {
		return nil, fmt.Errorf("[DescribeULB] got empty ulb data by ULBId(%s)", ulbId)
	}

	ulbInfo := resp.DataSet[0]
	vServerCubeInfos := make([]VServerCubeInfo, 0)
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
			vServerCubeInfos = append(vServerCubeInfos, info)
		}
	}

	return vServerCubeInfos, nil
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
