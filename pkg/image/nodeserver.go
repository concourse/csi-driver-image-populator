/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package image

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/golang/glog"
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"

	"github.com/concourse/baggageclaim"
	bclient "github.com/concourse/baggageclaim/client"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	deviceID = "deviceID"
)

var (
	TimeoutError = fmt.Errorf("Timeout")
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
	Timeout  time.Duration
	execPath string
	args     []string
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	glog.V(4).Info("creating baggageclaim client")
	bagClient := bclient.NewWithHTTPClient("http://127.0.0.1:7788",
		&http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 1 * time.Minute,
			},
			Timeout: 5 * time.Minute,
		})

	glog.V(4).Info("requesting baggageclaim to create volume")
	volume, err := bagClient.CreateVolume(lager.NewLogger("client"), req.VolumeId, baggageclaim.VolumeSpec{
		Strategy:   baggageclaim.EmptyStrategy{},
		Properties: map[string]string{},
	})
	if err != nil {
		return nil, err
	}

	targetPath := req.GetTargetPath()
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	readOnly := req.GetReadonly()

	options := []string{"bind"}
	if readOnly {
		options = append(options, "ro")
	}

	mounter := mount.New("")
	path := volume.Path()
	glog.V(4).Infof("mounting baggageclaim volume at %s", path)
	if err := mounter.Mount(path, targetPath, "", options); err != nil {
		return nil, err
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	targetPath := req.GetTargetPath()
	volumeId := req.GetVolumeId()

	// Unmounting the image
	err := mount.New("").Unmount(req.GetTargetPath())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.V(4).Infof("image: volume %s/%s has been unmounted.", targetPath, volumeId)

	// baggageclaim client
	bagClient := bclient.NewWithHTTPClient("http://127.0.0.1:7788",
		&http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 1 * time.Minute,
			},
			Timeout: 5 * time.Minute,
		},
	)

	err = bagClient.DestroyVolume(lager.NewLogger("client"), volumeId)
	if err != nil {
		return nil, err
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}
