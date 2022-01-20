package client

import (
	"context"
	"fmt"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/errors"
	"github.com/ucloud/migrate/internal/log"
	"github.com/ucloud/migrate/internal/utils"
	"github.com/ucloud/migrate/internal/utils/retry"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

func (client *UCloudClient) BindEIPToUHost(uhostId, eipId string) error {
	req := client.UNetConn.NewBindEIPRequest()
	req.EIPId = ucloud.String(eipId)
	req.ResourceId = ucloud.String(uhostId)
	req.ResourceType = ucloud.String("uhost")

	_, err := client.UNetConn.BindEIP(req)
	if err != nil {
		return errors.NewAPIRequestFailedError(err)
	}
	return nil
}

func (client *UCloudClient) UnBindEIPToUHost(uhostId, eipId string) error {
	req := client.UNetConn.NewUnBindEIPRequest()
	req.EIPId = ucloud.String(eipId)
	req.ResourceId = ucloud.String(uhostId)
	req.ResourceType = ucloud.String("uhost")

	_, err := client.UNetConn.UnBindEIP(req)
	if err != nil {
		return errors.NewAPIRequestFailedError(err)
	}
	return nil
}

func (client *UCloudClient) PowerOffUHostInstance(uhostId string) error {
	req := client.UHostConn.NewPoweroffUHostInstanceRequest()
	req.UHostId = ucloud.String(uhostId)

	_, err := client.UHostConn.PoweroffUHostInstance(req)
	if err != nil {
		return errors.NewAPIRequestFailedError(err)
	}
	return nil
}

func (client *UCloudClient) DeleteUHostInstance(uhostId string) error {
	req := client.UHostConn.NewTerminateUHostInstanceRequest()
	req.UHostId = ucloud.String(uhostId)
	req.ReleaseEIP = ucloud.Bool(true)
	req.ReleaseUDisk = ucloud.Bool(true)

	_, err := client.UHostConn.TerminateUHostInstance(req)
	if err != nil {
		return errors.NewAPIRequestFailedError(err)
	}
	return nil
}

func (client *UCloudClient) CreateUHost(config *conf.UHostConfig, count int) ([]string, error) {
	var allUHostIds []string
	for i := 0; i <= count/10; i++ {
		groupCount := 10
		if i == count/10 {
			if count%10 > 0 {
				groupCount = count % 10
			} else {
				break
			}
		}
		uHostIds := client.createUHostInstanceGroup(config, groupCount)
		allUHostIds = append(allUHostIds, uHostIds...)
	}

	return allUHostIds, nil
}

func (client *UCloudClient) createUHostInstanceGroup(config *conf.UHostConfig, groupCount int) []string {
	var wg sync.WaitGroup
	var lock sync.Mutex
	var err error
	var uHostIds []string
	for i := 0; i < groupCount; i++ {
		wg.Add(1)
		go func(config *conf.UHostConfig) {
			defer wg.Done()
			var ids []string
			ids, err = client.createUHostInstance(config, 1)
			if err != nil {
				log.Logger.Sugar().Warnf("[CreateUHost] got error, %s", err)
				return
			}

			lock.Lock()
			uHostIds = append(uHostIds, ids...)
			lock.Unlock()
		}(config)
	}
	wg.Wait()
	return uHostIds
}

func (client *UCloudClient) createUHostInstance(config *conf.UHostConfig, maxCount int) ([]string, error) {
	if maxCount == 0 {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("got zero max count about `CreateUHostInstance`"))
	}
	req := client.UHostConn.NewCreateUHostInstanceRequest()
	req.MaxCount = ucloud.Int(maxCount)
	if len(config.Name) > 0 {
		req.Name = ucloud.String(config.Name)
	} else if len(config.NamePrefix) > 0 {
		req.Name = ucloud.String(utils.PrefixedUniqueId(config.NamePrefix))
	} else {
		req.Name = ucloud.String(utils.PrefixedUniqueId("uhost-instance-"))
	}

	if len(config.Zone) > 0 {
		req.Zone = ucloud.String(config.Zone)
	} else {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("must set zone of uhost_config"))
	}

	if len(config.ImageId) > 0 {
		req.ImageId = ucloud.String(config.ImageId)
	} else if config.ImageIdFilter != nil {
		imageId, err := client.filterImageId(config.ImageIdFilter, config.Zone)
		if err != nil {
			return nil, err
		}
		req.ImageId = ucloud.String(imageId)
	} else {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("must set one of `image_id` and `image_id_filter` about `uhost_config`"))
	}

	req.LoginMode = ucloud.String("Password")

	if len(config.Password) > 0 {
		req.Password = ucloud.String(config.Password)
	} else {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("must set password of uhost_config"))
	}

	if len(config.ChargeType) > 0 {
		req.ChargeType = ucloud.String(config.ChargeType)
	} else {
		req.ChargeType = ucloud.String("Month")
	}

	if config.Duration > 0 {
		req.Quantity = ucloud.Int(config.Duration)
	} else {
		if *req.ChargeType == "Month" {
			req.Quantity = ucloud.Int(0)
		} else {
			req.Quantity = ucloud.Int(1)
		}
	}

	if config.CPU != 0 {
		req.CPU = ucloud.Int(config.CPU)
	} else {
		req.CPU = ucloud.Int(1)
	}

	if config.Memory != 0 {
		req.Memory = ucloud.Int(config.Memory)
	} else {
		req.Memory = ucloud.Int(1024)
	}

	if len(config.Tag) > 0 {
		req.Tag = ucloud.String(config.Tag)
	}

	if len(config.MinimalCpuPlatform) > 0 {
		req.MinimalCpuPlatform = ucloud.String(config.MinimalCpuPlatform)
	} else {
		req.MinimalCpuPlatform = ucloud.String("Amd/Auto")
	}

	if len(config.MachineType) > 0 {
		req.MachineType = ucloud.String(config.MachineType)
	} else {
		req.MachineType = ucloud.String("OM")
	}

	if len(config.Disks) > 0 {
		disks := make([]uhost.UHostDisk, 0)
		for _, v := range config.Disks {
			disks = append(disks, uhost.UHostDisk{
				IsBoot: ucloud.String(v.IsBoot),
				Size:   ucloud.Int(v.Size),
				Type:   ucloud.String(v.Type),
			})
		}
		req.Disks = disks
	} else {
		return nil, errors.NewConfigValidateFailedError(fmt.Errorf("must set one of disks about system disk"))
	}

	if len(config.VPCId) > 0 {
		req.VPCId = ucloud.String(config.VPCId)
		if len(config.SubnetId) > 0 {
			req.SubnetId = ucloud.String(config.SubnetId)
		}
	}

	if len(config.SecurityGroupId) > 0 {
		req.SecurityGroupId = ucloud.String(config.SecurityGroupId)
	}

	if config.PrivateIp != "" {
		req.PrivateIp = []string{config.PrivateIp}
	}

	resp, err := client.UHostConn.CreateUHostInstance(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 8433 {
			ctx := context.TODO()
			err = retry.Config{
				Tries: 10,
				ShouldRetry: func(err error) bool {
					return errors.IsNotCompleteError(err)
				},
				RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 8 * time.Second, Multiplier: 3}).Linear,
			}.Run(ctx, func(ctx context.Context) error {
				resp, err = client.UHostConn.CreateUHostInstance(req)
				if err != nil {
					if uErr, ok = err.(uerr.Error); ok && uErr.Code() == 8433 {
						return errors.NewNotCompletedError(err)
					}
					return err
				}
				return nil
			})
			if err != nil {
				return nil, errors.NewAPIRequestFailedError(
					fmt.Errorf("[Waiting for Create UHost By PrivateIp %s] got error, %s", config.PrivateIp, err))
			}
			return resp.UHostIds, nil
		}
		return nil, errors.NewAPIRequestFailedError(err)
	}

	return resp.UHostIds, nil
}

func (client *UCloudClient) filterImageId(filter *conf.ImageIdFilter, zone string) (string, error) {
	if filter == nil {
		return "", fmt.Errorf("empty filter")
	}
	req := client.UHostConn.NewDescribeImageRequest()

	req.Zone = ucloud.String(zone)

	if len(filter.ImageType) > 0 {
		req.ImageType = ucloud.String(filter.ImageType)
	}

	if len(filter.OSType) > 0 {
		req.OsType = ucloud.String(filter.OSType)
	}

	var allImages []uhost.UHostImageSet
	var offset int
	limit := 100
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := client.UHostConn.DescribeImage(req)
		if err != nil {
			return "", errors.NewAPIRequestFailedError(fmt.Errorf("error on reading image list, %s", err))
		}

		if resp == nil || len(resp.ImageSet) < 1 {
			break
		}

		for _, v := range resp.ImageSet {
			if v.State == "Available" {
				allImages = append(allImages, v)
			}
		}

		if len(resp.ImageSet) < limit {
			break
		}

		offset = offset + limit
	}

	var filteredImages []uhost.UHostImageSet
	if len(filter.NameRegex) > 0 {
		r := regexp.MustCompile(filter.NameRegex)
		for _, image := range allImages {
			if r != nil && !r.MatchString(image.ImageName) {
				continue
			}
			filteredImages = append(filteredImages, image)
		}
	} else {
		filteredImages = allImages
	}

	var finalImages []uhost.UHostImageSet
	if len(filteredImages) > 1 && filter.MostRecent {
		sort.Slice(filteredImages, func(i, j int) bool {
			return int64(filteredImages[i].CreateTime) > int64(filteredImages[j].CreateTime)
		})

		finalImages = []uhost.UHostImageSet{filteredImages[0]}
	} else {
		finalImages = filteredImages
	}

	if len(finalImages) == 0 {
		return "", errors.NewConfigValidateFailedError(fmt.Errorf("can not found any image about image_filter, %v", filter))
	}
	return finalImages[0].ImageId, nil
}

func (client *UCloudClient) DescribeUHostById(uhostId string) (*uhost.UHostInstanceSet, error) {
	req := client.UHostConn.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{uhostId}

	resp, err := client.UHostConn.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, nil
	}

	return &resp.UHostSet[0], nil
}

func (client *UCloudClient) WaitingForUHostStatus(uHostIds []string, status string) ([]string, error) {
	expectedUHostIds := map[string]struct{}{}
	for _, id := range uHostIds {
		expectedUHostIds[id] = struct{}{}
	}
	statusUHostIds := make([]string, 0)
	ctx := context.TODO()
	err := retry.Config{
		Tries: 10,
		ShouldRetry: func(err error) bool {
			return errors.IsNotCompleteError(err)
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 10}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		for uhostId := range expectedUHostIds {
			inst, err := client.DescribeUHostById(uhostId)
			if err != nil {
				return err
			}

			if inst.State == status {
				delete(expectedUHostIds, uhostId)
				statusUHostIds = append(statusUHostIds, uhostId)
				continue
			}

			if inst.State == "Install Fail" {
				delete(expectedUHostIds, uhostId)
				log.Logger.Sugar().Warnf("Install UHost %s failed", uhostId)
				continue
			}
		}

		if len(expectedUHostIds) != 0 {
			keys := make([]string, len(expectedUHostIds))
			for k := range expectedUHostIds {
				keys = append(keys, k)
			}
			return errors.NewNotCompletedError(fmt.Errorf("waiting uhost %v to %s", keys, status))
		}

		return nil
	})

	if err != nil {
		var s []string
		for v := range expectedUHostIds {
			s = append(s, v)
		}
		return nil, fmt.Errorf("[Waiting for UHost %s] about uhostIds: %q got error, %s", strings.Join(s, ","), status, err)
	}
	return statusUHostIds, nil
}
