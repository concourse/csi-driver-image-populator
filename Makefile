# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: deploy

image:
	docker build -t taylorsilva/baggageclaim . && docker push taylorsilva/baggageclaim

deploy:
	kubectl delete daemonsets.apps csi-imageplugin && \
		kubectl delete csidrivers.storage.k8s.io baggageclaim.concourse-ci.org && \
		kubectl apply -f deploy/kubernetes-latest/
